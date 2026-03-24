package sagacoordinator

import (
	"context"

	"github.com/Krunis/manager-o-order/packages/common"

	pb "github.com/Krunis/manager-o-order/packages/grpcapi"
)


type SagaCoordinator struct {
	dbRepo       DBSagaRepository
	delivery     pb.DeliveryClient
	confirmation pb.ConfirmationClient
	storage      pb.StorageClient
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
		s.confirmation.SendConfirmation()
	}

	if saga.CurrentStep == 1{
		s.storage.GetItem()
	}

	if saga.CurrentStep == 2{
		s.delivery.SendToQueue()
	}
}