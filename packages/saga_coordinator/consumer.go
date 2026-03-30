package sagacoordinator

import (
	"log"

	"github.com/IBM/sarama"
)

func (s *SagaCoordinator) Setup(session sarama.ConsumerGroupSession) error {
	log.Println("Setup")
	return nil
}

func (s *SagaCoordinator) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("Cleanup")
	return nil
}

func (s *SagaCoordinator) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

}

func NewSaramaConsumer(brokers []string, groupID string) (*SaramaConsumer, error) {
	config := sarama.NewConfig()

	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &SaramaConsumer{
		ConsumerGroup: consumerGroup,
	}, nil
}
