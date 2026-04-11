package delivery

import (
	"context"
	"errors"
	"log"
	"net"

	"github.com/Krunis/manager-o-order/packages/common"
	pb "github.com/Krunis/manager-o-order/packages/grpcapi"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
)

type DeliveryService struct {
	pb.UnimplementedDeliveryServiceServer

	port string

	lis        net.Listener
	grpcServer *grpc.Server

	poolDB *pgxpool.Pool

	lifecycle common.Lifecycle
}

func (d *DeliveryService) Start() error {
	lis, err := net.Listen("tcp", d.port)
	if err != nil {
		return err
	}

	d.grpcServer = grpc.NewServer()

	pb.RegisterDeliveryServiceServer(d.grpcServer, d)

	if err := d.grpcServer.Serve(lis); err != nil {
		return err
	}

	return nil
}

func (d *DeliveryService) SendToQueue(ctx context.Context, req *pb.AddressRequest) (*pb.AddressResponse, error) {
	tx, err := d.poolDB.Begin(ctx)
	if err != nil{
		log.Println(err)
		return nil, errors.New("internal server error")
	}
	defer tx.Rollback(ctx)
	
	tag, err := tx.Exec(ctx, `INSERT INTO delivery(order_id, table)
							  VALUES($1, $2)`,
							  req.OrderId, req.Table)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0{
		return &pb.AddressResponse{Added: false}, nil
	}

	tag, err = tx.Exec(ctx, `UPDATE orders
							 SET delivery_status=$1
							 WHERE id=$2`,
							 "WAITING", req.OrderId)
	if err != nil{
		return nil, err
	}
	if tag.RowsAffected() == 0{
		return nil, errors.New("no order with requested ID")
	}

	if err := tx.Commit(ctx); err != nil{
		return nil, err
	}

	return &pb.AddressResponse{Added: true}, nil
}

func (d *DeliveryService) CancelDelivery(ctx context.Context, req *pb.CancelDeliveryRequest) (*pb.CancelDeliveryResponse, error) {
	tx, err := d.poolDB.Begin(ctx)
	if err != nil{
		return nil, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
							  DELETE FROM delivery
							  WHERE order_id=$1`,
							  req.OrderId)
	if err != nil{
		return nil, err
	}
	if tag.RowsAffected() == 0{
		return &pb.CancelDeliveryResponse{Success: false}, nil
	}

	tag, err = tx.Exec(ctx, `
							 UPDATE orders
							 SET delivery_status=$1
							 WHERE id=$2`,
							 "CANCELLED", req.OrderId)
	if err != nil{
		return nil, err
	}
	if tag.RowsAffected() == 0{
		return &pb.CancelDeliveryResponse{Success: false}, nil
	}

	if err := tx.Commit(ctx); err != nil{
		return nil, err
	}

	return &pb.CancelDeliveryResponse{Success: true}, nil
}
