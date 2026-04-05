package sagacoordinator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/Krunis/manager-o-order/packages/common"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Krunis/manager-o-order/packages/grpcapi"
)

type SagaCoordinator struct {
	dbRepo DBSagaRepository

	dbConnnectionString string

	confirmation     ConfirmationClient
	confirmationConn *grpc.ClientConn

	storage     StorageClient
	storageConn *grpc.ClientConn

	delivery     DeliveryClient
	deliveryConn *grpc.ClientConn

	consumer Consumer

	msgCh chan *sarama.ConsumerMessage

	wg sync.WaitGroup

	stopOnce sync.Once

	lifecycle common.Lifecycle
}

func NewCoordinator(dbConnnectionString string) *SagaCoordinator {
	ctx, cancel := context.WithCancel(context.Background())

	return &SagaCoordinator{
		dbConnnectionString: dbConnnectionString,
		lifecycle:           common.Lifecycle{Ctx: ctx, Cancel: cancel},
		msgCh:               make(chan *sarama.ConsumerMessage, 100),
	}
}

func (s *SagaCoordinator) StartCoordinator(confirmationAddress,
	storageAddress,
	deliveryAddress string, topics []string) error {
	var err error

	pool, err := common.ConnectToDB(s.lifecycle.Ctx, s.dbConnnectionString)
	if err != nil {
		return err
	}

	s.dbRepo = NewPostgresSagaRepository(pool)

	s.confirmationConn, err = grpc.NewClient(confirmationAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("Failed to connect to Confirmation service: %s", err)
	}

	s.confirmation = &ConfirmationGRPC{confirmation: pb.NewConfirmationServiceClient(s.confirmationConn)}

	s.storageConn, err = grpc.NewClient(storageAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("Failed to connect to Storage service: %s", err)
	}

	s.storage = &StorageGRPC{storage: pb.NewStorageServiceClient(s.storageConn)}

	s.deliveryConn, err = grpc.NewClient(deliveryAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("Failed to connect to Delivery service: %s", err)
	}

	s.delivery = &DeliveryGRPC{delivery: pb.NewDeliveryServiceClient(s.deliveryConn)}

	s.wg.Go(s.pollMsgCh)

	if err := s.startConsuming(topics); err != nil {
		return err
	}

	return nil
}

func (s *SagaCoordinator) startConsuming(topics []string) error {
	for {
		select {
		case <-s.lifecycle.Ctx.Done():
			return nil
		default:
			ctx, cancel := context.WithCancel(s.lifecycle.Ctx)
			defer cancel()

			if err := s.consumer.Consume(ctx, topics, s); err != nil {
				return err
			}
		}
	}

}

func (s *SagaCoordinator) pollMsgCh() {
	for {
		select {
		case msg := <-s.msgCh:
			var order *common.Order

			json.Unmarshal(msg.Value, &order)

			go s.startSaga(order)

		case <-s.lifecycle.Ctx.Done():
			return
		}
	}
}

func (s *SagaCoordinator) startSaga(order *common.Order) {
	saga := &SagaState{
		OrderID:     order.ID,
		Status:      "ACTIVE",
		CurrentStep: 0,
		Payload:     order,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	id, err := s.dbRepo.Save(ctx, saga)
	if err != nil {
		log.Printf("Error while save in repo: %s", err)
		return
	}

	saga.ID = id

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = s.processSaga(ctx, saga)
	if err != nil {
		log.Printf("Errow while compensating: %s", err)
	}
}

func (s *SagaCoordinator) processSaga(ctx context.Context, saga *SagaState) error {

	if saga.CurrentStep == 0 {
		confirmationTypes := make([]string, 0, len(saga.Payload.Items))

		for i := range len(saga.Payload.Items) {
			confirmationTypes = append(confirmationTypes, saga.Payload.Items[i].ConfirmationType)
		}

		ctxClient, cancel := context.WithTimeout(ctx, time.Second*2)
		defer cancel()

		confId, err := s.confirmation.SendConfirmation(ctxClient, saga.Payload.EmployeeID, confirmationTypes)
		if err != nil {
			return s.compensate(ctx, saga, 0, err)
		}

		saga.Payload.ConfirmationId = confId
		saga.CurrentStep = 1

		ctxDB, cancelDB := context.WithTimeout(ctx, time.Second*1)
		defer cancelDB()

		if err := s.dbRepo.Update(ctxDB, saga); err != nil {
			return s.compensate(ctx, saga, 0, err)
		}
	}

	if saga.CurrentStep == 1 {
		var id string
		var err error

		ctxClient, cancel := context.WithTimeout(ctx, time.Second*2)
		defer cancel()

		for _, item := range saga.Payload.Items {
			id, err = s.storage.ReserveItem(ctxClient, item)
			if err != nil {
				return s.compensate(ctx, saga, 1, err)
			}
		}

		saga.Payload.ReserveID = id
		saga.CurrentStep = 2

		ctxDB, cancelDB := context.WithTimeout(ctx, time.Second*1)
		defer cancelDB()

		if err := s.dbRepo.Update(ctxDB, saga); err != nil {
			return s.compensate(ctx, saga, 1, err)
		}

	}

	if saga.CurrentStep == 2 {
		ctxClient, cancel := context.WithTimeout(ctx, time.Second*2)
		defer cancel()

		err := s.delivery.SendToQueue(ctxClient, saga.Payload.Delivery.Table)
		if err != nil {
			return s.compensate(ctx, saga, 2, err)
		}

		ctxDB, cancelDB := context.WithTimeout(ctx, time.Second*1)
		defer cancelDB()

		if err := s.dbRepo.Update(ctxDB, saga); err != nil {
			return s.compensate(ctx, saga, 1, err)
		}
	}

	return nil
}

func (s *SagaCoordinator) compensate(ctx context.Context, saga *SagaState, failedStep int, err error) error {
	var errs []error

	saga.Status = "COMPENSATING"
	if err := s.dbRepo.Update(ctx, saga); err != nil {
		return err
	}

	log.Printf("Compensation due to error: %s", err)

	for step := failedStep - 1; step >= 0; step-- {
		switch step {
		case 0: // Отмена подтверждения
			err := s.confirmation.CancelConfirmation(ctx, saga.Payload.ConfirmationId)
			if err != nil {
				errs = append(errs, err)
			}
		case 1: // Отмена резервации
			err := s.storage.CancelReserve(ctx, saga.Payload.ReserveID)
			if err != nil {
				errs = append(errs, err)
			}
		case 2: // Отмена доставки
			err := s.delivery.CancelDelivery(ctx, saga.OrderID)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	saga.Status = "FAILED"
	saga.Error = err.Error()
	s.dbRepo.Update(ctx, saga)

	return errors.Join(errs...)
}

func (s *SagaCoordinator) Stop() error {
	var errs []string

	var res error

	s.stopOnce.Do(func() {
		s.lifecycle.Cancel()

		s.wg.Wait()

		if s.consumer != nil {
			if err := s.consumer.Close(); err != nil {
				errs = append(errs, err.Error())
			}
		}

		if s.deliveryConn != nil {
			if err := s.deliveryConn.Close(); err != nil {
				errs = append(errs, err.Error())
			}
		}

		if s.storageConn != nil {
			if err := s.storageConn.Close(); err != nil {
				errs = append(errs, err.Error())
			}
		}

		if s.confirmationConn != nil {
			if err := s.confirmationConn.Close(); err != nil {
				errs = append(errs, err.Error())
			}
		}

		if s.dbRepo != nil {
			s.dbRepo.Close()
		}

		if len(errs) > 0 {
			res = errors.New(strings.Join(errs, ", "))
		}
	})

	return res
}
