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

		ctxClient, cancel := context.WithTimeout(ctx, time.Second * 2)
		defer cancel()

		confId, err := s.confirmation.SendConfirmation(ctxClient, saga.Payload.EmployeeID, confirmationTypes)
		if err != nil{
			return s.compensate(ctx, saga, 0, err)
		}

		saga.Payload.ConfirmationId = confId
		saga.CurrentStep = 1

		ctxDB, cancelDB := context.WithTimeout(ctx, time.Second * 1)
		defer cancelDB()

		if err := s.dbRepo.Update(ctxDB, saga); err != nil{
			return s.compensate(ctx, saga, 0, err)
		}
	}

	if saga.CurrentStep == 1{
		var id string
		var err error

		ctxClient, cancel := context.WithTimeout(ctx, time.Second * 2)
		defer cancel()

		for _, item := range saga.Payload.Items{
			id, err = s.storage.ReserveItem(ctxClient, item)
			if err != nil{
				return s.compensate(ctx, saga, 1, err)
			}
		}

		saga.Payload.ReserveID = id
		saga.CurrentStep = 2
		
		ctxDB, cancelDB := context.WithTimeout(ctx, time.Second * 1)
		defer cancelDB()

		if err := s.dbRepo.Update(ctxDB, saga); err != nil{
			return s.compensate(ctx, saga, 0, err)
		}
		
	}

	if saga.CurrentStep == 2{
		ctxClient, cancel := context.WithTimeout(ctx, time.Second * 2)

		err := s.delivery.SendToQueue(ctxClient, saga.Payload.Delivery.Table)
		cancel()
		if err != nil{
			saga.Error = err.Error()
			// compensate
		}
	}

	return nil
}

func (s *SagaCoordinator) compensate(ctx context.Context, saga *SagaState, failedStep int, err error) error{
	var res error

	saga.Status = "COMPENSATING"
	s.dbRepo.Update(ctx, saga)

	 for step := failedStep - 1; step >= 0; step-- {
        switch step {
        case 0: // Отмена в ресторане
            s.confirmation.CancelConfirmation(ctx, saga.Payload.ConfirmationId)
        case 1: // Возврат денег
            s.storage.CancelReserve(ctx, saga.Payload.ReserveID)
        case 2: // Отмена доставки
            s.delivery.CancelDelivery(ctx, saga.OrderID)
        }
        
    }
    
    saga.Status = "FAILED"
    saga.Error = err.Error()
    s.dbRepo.Update(ctx, saga)
    
    // Отправляем событие о провале
    // o.producer.Publish("order.events", OrderFailedEvent{saga.OrderID, err})
    
    return err
}