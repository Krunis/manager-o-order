package sagacoordinator

import (
	"context"

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

	return s.processSaga(saga)
}

func (s *SagaCoordinator) processSaga(saga *SagaState) error{

	if saga.CurrentStep == 0{
		confirmationTypes := make([]string, 0, len(saga.Payload.Items))

		for i := range len(saga.Payload.Items){
			confirmationTypes = append(confirmationTypes, saga.Payload.Items[i].ConfirmationType)
		}

		err := s.confirmation.SendConfirmation(saga.Payload.EmployeeID, confirmationTypes)
		if err != nil{
			saga.Error = err.Error()
			// compensate
		}
	}

	if saga.CurrentStep == 1{
		s.storage.ReserveItem(sa)
	}

	if saga.CurrentStep == 2{
		s.delivery.SendToQueue()
	}
}