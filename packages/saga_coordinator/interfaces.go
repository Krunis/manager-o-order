package sagacoordinator

import (
	"context"
	"time"

	"github.com/Krunis/manager-o-order/packages/common"
	"github.com/jackc/pgx/v4/pgxpool"
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
	Save(ctx context.Context, state *SagaState) error

	Update(ctx context.Context, state *SagaState) error

	FindByID(ctx context.Context, id string) (*SagaState, error)
}

type PostgresSagaRepository struct{
	pool *pgxpool.Pool
}
