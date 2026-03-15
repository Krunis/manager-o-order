package apigateway

import (
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
	config.Producer.Return.Errors = true

	config.Producer.Partitioner = sarama.NewHashPartitioner

	config.Producer.Compression = sarama.CompressionSnappy

	config.Producer.Timeout = 30 * time.Second
	config.Net.DialTimeout = 30 * time.Second
	config.Net.ReadTimeout = 30 * time.Second
	config.Net.WriteTimeout = 30 * time.Second

	return sarama.NewSyncProducer([]string{port}, config)
}