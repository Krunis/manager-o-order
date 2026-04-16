package apigateway

import (
	"errors"
	"log"
	"time"

	"github.com/IBM/sarama"
)

func NewSaramaProducer(address string) (sarama.SyncProducer, error){
	if address == ""{
		return nil, errors.New("no kafka address")
	}

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

	return sarama.NewSyncProducer([]string{address}, config)
}

func (g *GatewayServer) sendInKafka(topic string, key, value []byte) error{
	part, offset, err := g.saramaProducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key: sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	})
	if err != nil{
		log.Printf("Failed to sent: %s", err)
		return err
	}

	log.Printf("Sent in %d partition, %d offset", part, offset)

	return nil
}