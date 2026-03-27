package sagacoordinator

import (
	"context"
	"time"

	"github.com/Krunis/manager-o-order/packages/common"
	"google.golang.org/grpc"

	pb "github.com/Krunis/manager-o-order/packages/grpcapi"
)


type SagaCoordinator struct {
	dbRepo       DBSagaRepository

	confirmation ConfirmationClient
	delivery     DeliveryClient
	storage      StorageClient
}

func (s *SagaCoordinator) StartSaga(order *common.Order) error{
	saga := &SagaState{
		OrderID: order.ID,
		Status: "ACTIVE",
		CurrentStep: 0,
		Payload: order,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	id, err := s.dbRepo.Save(ctx, saga)
	if err != nil{
		return err
	}

	saga.ID = id

	ctx, cancel = context.WithTimeout(context.Background(), time.Second * 10)
	defer cancel()

	return s.processSaga(ctx, saga)
}

func (s *SagaCoordinator) processSaga(ctx context.Context, saga *SagaState) error{

	if saga.CurrentStep == 0{
		confirmationTypes := make([]string, 0, len(saga.Payload.Items))

		for i := range len(saga.Payload.Items){
			confirmationTypes = append(confirmationTypes, saga.Payload.Items[i].ConfirmationType)
		}

		ctxClient, cancel := context.WithTimeout(ctx, time.Second * 3)

		err := s.confirmation.SendConfirmation(ctxClient, saga.Payload.EmployeeID, confirmationTypes)
		cancel()
		if err != nil{
			saga.Error = err.Error()
			// compensate
		}

	}

	if saga.CurrentStep == 1{
		ctxClient, cancel := context.WithTimeout(ctx, time.Second * 3)

		for _, item := range saga.Payload.Items{
			id, err := s.storage.ReserveItem(ctxClient, item)
			if err != nil{
				saga.Error = err.Error()
				// compensate
			}
		}
		cancel()
		
	}

	if saga.CurrentStep == 2{
		ctxClient, cancel := context.WithTimeout(ctx, time.Second * 3)

		err := s.delivery.SendToQueue(ctxClient, saga.Payload.Delivery.Table)
		cancel()
		if err != nil{
			saga.Error = err.Error()
			// compensate
		}
	}
}