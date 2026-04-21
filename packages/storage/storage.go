package storage

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/Krunis/manager-o-order/packages/common"
	pb "github.com/Krunis/manager-o-order/packages/grpcapi"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
)

type StorageService struct {
	pb.UnimplementedStorageServiceServer

	port string

	lis        net.Listener
	grpcServer *grpc.Server

	poolDB *pgxpool.Pool
	dbConnectionString string

	stopOnce sync.Once

	lifecycle common.Lifecycle
}

func NewStorageService(dbConnectionString string) *StorageService{
	ctx, cancel := context.WithCancel(context.Background())

	return &StorageService{
		dbConnectionString: dbConnectionString,
		lifecycle: common.Lifecycle{Ctx: ctx, Cancel: cancel},
	}
}


func (c *StorageService) Start(port string) error {
	var err error

	c.poolDB, err = common.ConnectToDB(c.lifecycle.Ctx, c.dbConnectionString)
	if err != nil{
		return err
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	c.grpcServer = grpc.NewServer()

	pb.RegisterStorageServiceServer(c.grpcServer, c)

	if err := c.grpcServer.Serve(lis); err != nil {
		return err
	}

	return nil
}

func (s *StorageService) ReserveItem(ctx context.Context, req *pb.ItemRequest) (*pb.ItemResponse, error) {
	var id string

	err := s.poolDB.QueryRow(ctx, `SELECT item_id
							FROM storage
							WHERE item_id = $1 AND item_name = $2 AND count <= $3
							`, req.Id, req.Name, req.Count).Scan(&id)
	log.Println("first")
	if err != nil {
		return nil, err
	}

	reserveId := uuid.New()

	_, err = s.poolDB.Exec(ctx, `INSERT INTO reservations(id, item_id, count)
							   VALUES($1, $2, $3)
							   `, reserveId.String(), req.Id, req.Count)
	log.Println("second")
	if err != nil {
		return nil, err
	}

	return &pb.ItemResponse{Accepted: true, ReserveId: reserveId.String()}, nil

}

func (s *StorageService) CancelReserve(ctx context.Context, req *pb.CancelReserveRequest) (*pb.CancelReserveResponse, error) {
	tag, err := s.poolDB.Exec(ctx, `DELETE FROM reservations
							   		WHERE id = $1 
							   		`, req.ReserveId)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return &pb.CancelReserveResponse{Success: false}, nil
	}

	return &pb.CancelReserveResponse{Success: true}, nil
}

func (s *StorageService) Stop() {
	s.stopOnce.Do(func() {
		if s.grpcServer != nil {
			s.grpcServer.GracefulStop()
		}

		if s.poolDB != nil {
			s.poolDB.Close()
		}

		log.Println("StorageService stopped")
	})

}
