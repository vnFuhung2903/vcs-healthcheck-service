package messages

import (
	"github.com/segmentio/kafka-go"
	"github.com/vnFuhung2903/vcs-healthcheck-service/pkg/env"
)

type IKafkaFactory interface {
	ConnectKafkaWriter(topic string) (*kafka.Writer, error)
}

type kafkaFactory struct {
	address string
}

func NewKafkaFactory(env env.KafkaEnv) IKafkaFactory {
	return &kafkaFactory{
		address: env.KafkaAddress,
	}
}

func (f *kafkaFactory) ConnectKafkaWriter(topic string) (*kafka.Writer, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(f.address),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
	return writer, nil
}
