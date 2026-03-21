package sagacoordinator

import "github.com/Krunis/manager-o-order/packages/common"


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

	s.dbRepo.Save()

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