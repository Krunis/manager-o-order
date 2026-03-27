package sagacoordinator

import (
	"context"
	"errors"

	"github.com/Krunis/manager-o-order/packages/common"
	pb "github.com/Krunis/manager-o-order/packages/grpcapi"
)

func (s *StorageGRPC) ReserveItem(ctx context.Context, item *common.Item) (string, error) {
	resp, err := s.storage.ReserveItem(ctx, &pb.ItemRequest{Id: item.ID, Name: item.Name, Count: item.Count})
	if err != nil{
		return "", err
	}
	if !resp.Accepted{
		return "", errors.New("unknown error in Storage")
	}

	return resp.ReserveId, nil
}

func (s *StorageGRPC) CancelReserve(ctx context.Context, reserveId string) error{
	resp, err := s.storage.CancelReserve(ctx, &pb.CancelReserveRequest{ReserveId: reserveId})
	if err != nil{
		return err
	}
	if !resp.Success{
		return errors.New("unknown error in Storage")
	}

	return nil
}