package workers

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/suite"
	"github.com/vnFuhung2903/vcs-healthcheck-service/entities"
	"github.com/vnFuhung2903/vcs-healthcheck-service/mocks/docker"
	"github.com/vnFuhung2903/vcs-healthcheck-service/mocks/interfaces"
	"github.com/vnFuhung2903/vcs-healthcheck-service/mocks/logger"
	"github.com/vnFuhung2903/vcs-healthcheck-service/mocks/services"
)

type HealthcheckWorkerSuite struct {
	suite.Suite
	ctrl                   *gomock.Controller
	healthcheckWorker      IHealthcheckWorker
	mockDockerClient       *docker.MockIDockerClient
	mockHealthcheckService *services.MockIHealthcheckService
	mockKafkaProducer      *interfaces.MockIKafkaProducer
	mockRedisClient        *interfaces.MockIRedisClient
	mockLogger             *logger.MockILogger
}

func (s *HealthcheckWorkerSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockDockerClient = docker.NewMockIDockerClient(s.ctrl)
	s.mockHealthcheckService = services.NewMockIHealthcheckService(s.ctrl)
	s.mockKafkaProducer = interfaces.NewMockIKafkaProducer(s.ctrl)
	s.mockRedisClient = interfaces.NewMockIRedisClient(s.ctrl)
	s.mockLogger = logger.NewMockILogger(s.ctrl)

	s.healthcheckWorker = NewHealthcheckWorker(
		s.mockDockerClient,
		s.mockHealthcheckService,
		s.mockKafkaProducer,
		s.mockRedisClient,
		s.mockLogger,
		2*time.Second,
	)
}

func (s *HealthcheckWorkerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestHealthcheckWorkerSuite(t *testing.T) {
	suite.Run(t, new(HealthcheckWorkerSuite))
}

func (s *HealthcheckWorkerSuite) TestHealthcheckWorkerStatusChange() {
	containers := []entities.ContainerWithStatus{
		{ContainerId: "1", Status: entities.ContainerOn},
	}

	s.mockRedisClient.EXPECT().
		Get(gomock.Any(), "containers").
		Return(containers, nil)

	s.mockDockerClient.EXPECT().
		GetStatus(gomock.Any(), "1").
		Return(entities.ContainerOff)

	s.mockDockerClient.EXPECT().
		GetIpv4(gomock.Any(), "1").
		Return("192.168.1.100")

	s.mockKafkaProducer.EXPECT().
		AddMessage(gomock.Any()).
		Return(kafka.Message{}, nil)

	s.mockKafkaProducer.EXPECT().
		Produce(gomock.Any(), gomock.Any()).
		Return(nil)

	s.mockHealthcheckService.EXPECT().
		Update(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	s.mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info("healthcheck worker stopped").AnyTimes()

	s.healthcheckWorker.Start()
	time.Sleep(3 * time.Second)

	s.healthcheckWorker.Stop()
}

func (s *HealthcheckWorkerSuite) TestHealthcheckWorkerRedisError() {
	s.mockRedisClient.EXPECT().
		Get(gomock.Any(), "containers").
		Return(nil, errors.New("redis error"))

	s.mockLogger.EXPECT().Error("failed to get containers from redis", gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info("healthcheck worker stopped").AnyTimes()

	s.healthcheckWorker.Start()
	time.Sleep(3 * time.Second)

	s.healthcheckWorker.Stop()
}

func (s *HealthcheckWorkerSuite) TestHealthcheckWorkerKafkaMessageError() {
	containers := []entities.ContainerWithStatus{
		{ContainerId: "1", Status: entities.ContainerOn},
	}

	s.mockRedisClient.EXPECT().
		Get(gomock.Any(), "containers").
		Return(containers, nil)

	s.mockDockerClient.EXPECT().
		GetStatus(gomock.Any(), "1").
		Return(entities.ContainerOff)

	s.mockDockerClient.EXPECT().
		GetIpv4(gomock.Any(), "1").
		Return("192.168.1.100")

	s.mockKafkaProducer.EXPECT().
		AddMessage(gomock.Any()).
		Return(kafka.Message{}, errors.New("kafka message error"))

	s.mockHealthcheckService.EXPECT().
		Update(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	s.mockLogger.EXPECT().Error("failed to create kafka message", gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info("healthcheck worker stopped").AnyTimes()

	s.healthcheckWorker.Start()
	time.Sleep(3 * time.Second)

	s.healthcheckWorker.Stop()
}

func (s *HealthcheckWorkerSuite) TestHealthcheckWorkerNoStatusChange() {
	containers := []entities.ContainerWithStatus{
		{ContainerId: "1", Status: entities.ContainerOn},
	}

	s.mockRedisClient.EXPECT().
		Get(gomock.Any(), "containers").
		Return(containers, nil)

	s.mockDockerClient.EXPECT().
		GetStatus(gomock.Any(), "1").
		Return(entities.ContainerOn)

	s.mockHealthcheckService.EXPECT().
		Update(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	s.mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info("healthcheck worker stopped").AnyTimes()

	s.healthcheckWorker.Start()
	time.Sleep(3 * time.Second)

	s.healthcheckWorker.Stop()
}

func (s *HealthcheckWorkerSuite) TestHealthcheckWorkerUpdateStatusError() {
	containers := []entities.ContainerWithStatus{
		{ContainerId: "1", Status: entities.ContainerOn},
	}

	s.mockRedisClient.EXPECT().
		Get(gomock.Any(), "containers").
		Return(containers, nil)

	s.mockDockerClient.EXPECT().
		GetStatus(gomock.Any(), "1").
		Return(entities.ContainerOn)

	s.mockHealthcheckService.EXPECT().
		Update(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("update status error"))

	s.mockLogger.EXPECT().Error("failed to update status", gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info("healthcheck worker stopped").AnyTimes()

	s.healthcheckWorker.Start()
	time.Sleep(3 * time.Second)

	s.healthcheckWorker.Stop()
}

func (s *HealthcheckWorkerSuite) TestHealthcheckWorkerKafkaProduceError() {
	containers := []entities.ContainerWithStatus{
		{ContainerId: "1", Status: entities.ContainerOn},
	}

	s.mockRedisClient.EXPECT().
		Get(gomock.Any(), "containers").
		Return(containers, nil)

	s.mockDockerClient.EXPECT().
		GetStatus(gomock.Any(), "1").
		Return(entities.ContainerOff)

	s.mockDockerClient.EXPECT().
		GetIpv4(gomock.Any(), "1").
		Return("192.168.1.100")

	s.mockKafkaProducer.EXPECT().
		AddMessage(gomock.Any()).
		Return(kafka.Message{}, nil)

	s.mockKafkaProducer.EXPECT().
		Produce(gomock.Any(), gomock.Any()).
		Return(errors.New("kafka produce error"))

	s.mockHealthcheckService.EXPECT().
		Update(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	s.mockLogger.EXPECT().Error("failed to produce kafka messages", gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().Info("healthcheck worker stopped").AnyTimes()

	s.healthcheckWorker.Start()
	time.Sleep(3 * time.Second)

	s.healthcheckWorker.Stop()
}
