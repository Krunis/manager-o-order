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
	ReserveItem(ctx context.Context, item *common.Item) (string, error)
	CancelReserve(ctx context.Context, reserveId string) error
}

type StorageGRPC struct{
	storage pb.StorageServiceClient
}

type DeliveryClient interface{
	SendToQueue(ctx context.Context, table int32) error
	CancelDelivery(ctx context.Context, orderId string) error
}

type DeliveryGRPC struct{
	delivery pb.DeliveryServiceClient
}

type ConfirmationClient interface{
	SendConfirmation(ctx context.Context, confirmationEmployeeID string, confirmationType []string) (string, error)
	CancelConfirmation(ctx context.Context, confirmationId string) error
}

type ConfirmationGRPC struct{
	confirmation pb.ConfirmationServiceClient
}
