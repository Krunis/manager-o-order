package storage

import (
	"context"
	"net"

	"github.com/Krunis/manager-o-order/packages/common"
	pb "github.com/Krunis/manager-o-order/packages/grpcapi"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/google/uuid"
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

func (s *StorageService) CancelReserve(context.Context, *pb.CancelReserveRequest) (*pb.CancelReserveResponse, error){
	return nil, nil
}