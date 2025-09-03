package interfaces

import (
	"context"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/vnFuhung2903/vcs-healthcheck-service/dto"
	"github.com/vnFuhung2903/vcs-healthcheck-service/entities"
)

type KafkaProducerTestSuite struct {
	suite.Suite
	ctx           context.Context
	kafkaProducer IKafkaProducer
	topic         string
}

func (s *KafkaProducerTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.topic = "test-topic"
	writer := &kafka.Writer{}
	s.kafkaProducer = NewKafkaProducer(writer, s.topic)
}

func TestKafkaProducerTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaProducerTestSuite))
}

func (s *KafkaProducerTestSuite) TestAddMessage() {
	msg := dto.KafkaStatusUpdate{
		ContainerId: "container-123",
		Status:      entities.ContainerOn,
		Ipv4:        "192.168.1.100",
	}

	kafkaMsg, err := s.kafkaProducer.AddMessage(msg)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), []byte(entities.ContainerOn), kafkaMsg.Key)

	err = s.kafkaProducer.Produce(s.ctx, []kafka.Message{kafkaMsg})
	s.Error(err)
}
