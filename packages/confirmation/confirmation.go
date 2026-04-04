package confirmation

import (
	"net"

	"github.com/Krunis/manager-o-order/packages/common"
	pb "github.com/Krunis/manager-o-order/packages/grpcapi"
	"google.golang.org/grpc"
)

type ConfirmationService struct{
	pb.UnimplementedConfirmationServiceServer

	port string

	lis net.Listener
	grpcServer *grpc.Server

	lifecycle common.Lifecycle
}

func (c *ConfirmationService) Start() error{
	lis, err := net.Listen("tcp", c.port)
	if err != nil{
		return err
	}

	c.grpcServer = grpc.NewServer()

	pb.RegisterConfirmationServiceServer(c.grpcServer, c)

	if err := c.grpcServer.Serve(lis); err != nil{
		return err
	}

	return nil
}