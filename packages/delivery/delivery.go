package delivery

import (
	"context"
	"net"

	"github.com/Krunis/manager-o-order/packages/common"
	pb "github.com/Krunis/manager-o-order/packages/grpcapi"
	"google.golang.org/grpc"
)

type DeliveryService struct{
	pb.UnimplementedDeliveryServiceServer

	port string

	lis net.Listener
	grpcServer *grpc.Server

	

	lifecycle common.Lifecycle
}

func (d *DeliveryService) Start() error{
	lis, err := net.Listen("tcp", d.port)
	if err != nil{
		return err
	}

	d.grpcServer = grpc.NewServer()

	pb.RegisterDeliveryServiceServer(d.grpcServer, d)

	if err := d.grpcServer.Serve(lis); err != nil{
		return err
	}

	return nil
}

func (d *DeliveryService) SendToQueue(ctx context.Context, req *pb.AddressRequest) (*pb.AddressResponse, error){
	
}

func (d *DeliveryService) CancelDelivery(ctx context.Context, req *pb.CancelDeliveryRequest) (*pb.CancelDeliveryResponse, error){

}