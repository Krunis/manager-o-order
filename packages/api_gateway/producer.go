package apigateway

import (
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"
)

func NewSaramaProducer(port string) (sarama.SyncProducer, error){
	config := sarama.NewConfig()

	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Idempotent = true
	config.Net.MaxOpenRequests = 1

	config.Producer.Retry.Max = 10
	config.Producer.Retry.Backoff = 100 * time.Millisecond

	config.Producer.Return.Successes = true

	config.Producer.Partitioner = sarama.NewHashPartitioner

	config.Producer.Compression = sarama.CompressionSnappy

	config.Producer.Timeout = 30 * time.Second
	config.Net.DialTimeout = 30 * time.Second
	config.Net.ReadTimeout = 30 * time.Second
	config.Net.WriteTimeout = 30 * time.Second

	return sarama.NewSyncProducer([]string{port}, config)
}

func (g *GatewayServer) sendInKafka(order *Order) error{
	key := sarama.ByteEncoder([]byte(order.IdempotencyKey))

	valueBytes, err := json.Marshal(order)
	if err != nil{
		return err
	}

	value := sarama.ByteEncoder(valueBytes)	

	part, offset, err := g.saramaProducer.SendMessage(&sarama.ProducerMessage{
		Topic: "orders-new",
		Key: key,
		Value: value,
	})
	if err != nil{
		log.Printf("Failed to sent: %s", err)
		return err
	}

	log.Printf("Sent in %d partition, %d offset", part, offset)

	return nil
}