package confirmation

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Krunis/manager-o-order/packages/common"
	pb "github.com/Krunis/manager-o-order/packages/grpcapi"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
)

type ConfirmationService struct {
	pb.UnimplementedConfirmationServiceServer

	port string

	lis        net.Listener
	grpcServer *grpc.Server

	poolDB *pgxpool.Pool
	dbConnectionString string

	stopOnce sync.Once

	lifecycle common.Lifecycle
}

func NewConfirmationService(dbConnectionString string) *ConfirmationService{
	ctx, cancel := context.WithCancel(context.Background())

	return &ConfirmationService{
		dbConnectionString: dbConnectionString,
		lifecycle: common.Lifecycle{Ctx: ctx, Cancel: cancel},
	}
}

func (c *ConfirmationService) Start(port string) error {
	var err error

	c.poolDB, err = common.ConnectToDB(c.lifecycle.Ctx, c.dbConnectionString)
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	c.grpcServer = grpc.NewServer()

	pb.RegisterConfirmationServiceServer(c.grpcServer, c)

	if err := c.grpcServer.Serve(lis); err != nil {
		return err
	}

	return nil
}

func (c *ConfirmationService) SendConfirmation(ctx context.Context, req *pb.ConfirmatorInfoRequest) (*pb.ConfirmatorInfoResponse, error) {
	id := uuid.New()

	for _, confType := range req.ConfirmationTypes {
		go func() {
			time.Sleep(time.Millisecond * 50)
			log.Println(confType)
		}()
	}

	_, err := c.poolDB.Exec(ctx, `INSERT INTO confirmations(id, employee_id)
						VALUES($1, $2)
						`, id, req.ConfirmationEmployeeId)
	if err != nil {
		if err.(*pgconn.PgError).Code == pgerrcode.ForeignKeyViolation{
			return nil, errors.New("unknown confirmation_employee_id")
		}
		return nil, err
	}

	return &pb.ConfirmatorInfoResponse{Sent: true, ConfirmationId: id.String()}, nil
}

func (c *ConfirmationService) CancelConfirmation(ctx context.Context, req *pb.CancelConfirmationRequest) (*pb.CancelConfirmationResponse, error) {
	go func() {
		time.Sleep(time.Millisecond * 50)
	}()

	tag, err := c.poolDB.Exec(ctx, `DELETE FROM confirmations
								  WHERE id = $1
								  `, req.ConfirmationId)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, errors.New("unknown confirmation_id")
	}

	return &pb.CancelConfirmationResponse{Success: true}, nil

}

func (c *ConfirmationService) Stop() {
	c.stopOnce.Do(func() {
		if c.grpcServer != nil {
			c.grpcServer.GracefulStop()
		}

		if c.poolDB != nil {
			c.poolDB.Close()
		}

		log.Println("ConfirmationService stopped")
	})

}
