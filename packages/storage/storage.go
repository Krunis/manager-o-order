package storage

import (
	"context"
	"log"
	"net"

	"github.com/Krunis/manager-o-order/packages/common"
	pb "github.com/Krunis/manager-o-order/packages/grpcapi"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
)

type StorageService struct{
	pb.UnimplementedStorageServiceServer

	port string

	lis net.Listener
	grpcServer *grpc.Server

	poolDB *pgxpool.Pool

	lifecycle common.Lifecycle
}

func (c *StorageService) Start() error{
	lis, err := net.Listen("tcp", c.port)
	if err != nil{
		return err
	}

	c.grpcServer = grpc.NewServer()

	pb.RegisterStorageServiceServer(c.grpcServer, c)

	if err := c.grpcServer.Serve(lis); err != nil{
		return err
	}

	return nil
}

func (s *StorageService) ReserveItem(ctx context.Context, req *pb.ItemRequest) (*pb.ItemResponse, error){
	var id string

	row := s.poolDB.QueryRow(ctx, `SELECT item_id
							FROM storage
							WHERE item_id = $1, name = $2, count <= $3
							`, req.Id, req.Name, req.Count)

	err := row.Scan(&id)
	if err != nil{
		return nil, err
	}

	reserveId := uuid.New()
	
	_, err = s.poolDB.Exec(ctx, `INSERT INTO reservations(id, item_id, count)
							   VALUES($1, $2, $3)
							   `, reserveId.String(), req.Id, req.Count)
	if err != nil{
		return nil, err
	}

	return &pb.ItemResponse{Accepted: true, ReserveId: reserveId.String()}, nil

}

func (s *StorageService) CancelReserve(ctx context.Context, req *pb.CancelReserveRequest) (*pb.CancelReserveResponse, error){
	tag, err := s.poolDB.Exec(ctx, `DELETE FROM reservations
							   		WHERE id = $1 
							   		`, req.ReserveId)
	if err != nil{
		return nil, err
	}
	if tag.RowsAffected() == 0{
		return &pb.CancelReserveResponse{Success: false}, nil
	}

	return &pb.CancelReserveResponse{Success: true}, nil
}

func (s *StorageService) Stop(){
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	if s.poolDB != nil{
		s.poolDB.Close()
	}

	log.Println("StorageService stopped")
}
