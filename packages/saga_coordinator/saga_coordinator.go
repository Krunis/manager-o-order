package sagacoordinator

import (
	"context"

	"github.com/Krunis/manager-o-order/packages/common"
)


type SagaCoordinator struct {
	dbRepo       DBSagaRepository
	delivery     int
	confirmation int
	storage      int
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

	}

	if saga.CurrentStep == 1{
		
	}

	if saga.CurrentStep == 2{
		
	}
}