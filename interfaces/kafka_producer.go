package interfaces

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type IKafkaProducer interface {
	Produce(ctx context.Context, message []byte) error
}

type kafkaProducer struct {
	writer *kafka.Writer
	topic  string
}

func NewKafkaProducer(writer *kafka.Writer, topic string) IKafkaProducer {
	return &kafkaProducer{writer: writer, topic: topic}
}

func (p *kafkaProducer) Produce(ctx context.Context, message []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(p.topic),
		Value: message,
	})
}
