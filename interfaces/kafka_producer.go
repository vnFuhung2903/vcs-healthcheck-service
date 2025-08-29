package interfaces

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"github.com/vnFuhung2903/vcs-healthcheck-service/dto"
)

type IKafkaProducer interface {
	AddMessage(msg dto.KafkaStatusUpdate) (kafka.Message, error)
	Produce(ctx context.Context, msgs []kafka.Message) error
}

type kafkaProducer struct {
	writer *kafka.Writer
	topic  string
}

func NewKafkaProducer(writer *kafka.Writer, topic string) IKafkaProducer {
	return &kafkaProducer{writer: writer, topic: topic}
}

func (p *kafkaProducer) AddMessage(msg dto.KafkaStatusUpdate) (kafka.Message, error) {
	message, err := json.Marshal(msg)
	if err != nil {
		return kafka.Message{}, err
	}
	return kafka.Message{
		Key:   []byte(msg.Status),
		Value: message,
	}, nil
}

func (p *kafkaProducer) Produce(ctx context.Context, msgs []kafka.Message) error {
	if len(msgs) == 0 {
		return nil
	}
	return p.writer.WriteMessages(ctx, msgs...)
}
