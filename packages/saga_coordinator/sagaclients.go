package sagacoordinator

import (
	"context"
	"errors"

	"github.com/IBM/sarama"
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
		return errors.New("unknown reserve_id")
	}

	return nil
}

func (d *DeliveryGRPC) SendToQueue(ctx context.Context, table int32) error{
	resp, err := d.delivery.SendToQueue(ctx, &pb.AddressRequest{Table: table})
	if err != nil{
		return err
	}
	if !resp.Added {
		return errors.New("unknown error in Delivery")
	}

	return nil
}

func (d *DeliveryGRPC) CancelDelivery(ctx context.Context, orderId string) error{
	resp, err := d.delivery.CancelDelivery(ctx, &pb.CancelDeliveryRequest{OrderId: orderId})
	if err != nil{
		return err
	}
	if !resp.Success {
		return errors.New("unknown error in Delivery")
	}

	return nil
}

func (c *ConfirmationGRPC) SendConfirmation(ctx context.Context, confirmationEmployeeID string, confirmationType []string) (string, error){
	resp, err := c.confirmation.SendConfirmation(ctx, &pb.ConfirmatorInfoRequest{
		ConfirmationEmployeeId: confirmationEmployeeID,
		ConfirmationTypes: confirmationType,
	})
	if err != nil{
		return "", err
	}
	if !resp.Sent{
		return "", errors.New("unknown error in Confirmation")
	}

	return resp.ConfirmationId, nil
}

func (c *ConfirmationGRPC) CancelConfirmation(ctx context.Context, confirmationId string) error{
	resp, err := c.confirmation.CancelConfirmation(ctx, &pb.CancelConfirmationRequest{ConfirmationId: confirmationId})
	if err != nil{
		return err
	}
	if !resp.Success{
		return errors.New("unknown error in Confirmation")
	}

	return nil
}

func (cons *SaramaConsumer) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error{
	return cons.ConsumerGroup.Consume(ctx, topics, handler)
}

func (cons *SaramaConsumer) Close() error{
	return cons.ConsumerGroup.Close()
}