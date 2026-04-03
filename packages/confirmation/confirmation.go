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
	
}