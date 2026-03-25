package sagacoordinator

import (
	"context"
	"time"

	"github.com/Krunis/manager-o-order/packages/common"
	"github.com/jackc/pgx/v4/pgxpool"

	pb "github.com/Krunis/manager-o-order/packages/grpcapi"
)

type SagaState struct{
	ID string
	OrderID string
	Status string
	CurrentStep int
	Payload *common.Order
	Error string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type DBSagaRepository interface {
	Save(ctx context.Context, state *SagaState) (string, error)

	Update(ctx context.Context, state *SagaState) error

	FindByID(ctx context.Context, id string) (*SagaState, error)
}

type PostgresSagaRepository struct{
	pool *pgxpool.Pool
}

type StorageClient interface{
	ReserveItem(id, name string, count uint32) (string, error)
	CancelReserve(reserveId string) error
}

type StorageGRPC struct{
	storage pb.StorageServiceClient
}

type DeliveryClient interface{
	SendToQueue(table string) error
	CancelDelivery(orderId string) error
}

type DeliveryGRPC struct{
	delivery pb.DeliveryServiceClient
}

type ConfirmationClient interface{
	SendConfirmation(confirmationEmployeeID string, confirmationType []string) error
	CancelConfirmation(confirmationId string) error
}

type ConfirmationGRPC struct{
	confirmation pb.ConfirmationServiceClient
}
