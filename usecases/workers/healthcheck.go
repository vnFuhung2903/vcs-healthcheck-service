package workers

import (
	"context"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/vnFuhung2903/vcs-healthcheck-service/dto"
	"github.com/vnFuhung2903/vcs-healthcheck-service/interfaces"
	"github.com/vnFuhung2903/vcs-healthcheck-service/pkg/docker"
	"github.com/vnFuhung2903/vcs-healthcheck-service/pkg/logger"
	"github.com/vnFuhung2903/vcs-healthcheck-service/usecases/services"
	"go.uber.org/zap"
)

type IHealthcheckWorker interface {
	Start()
	Stop()
}

type healthcheckWorker struct {
	dockerClient       docker.IDockerClient
	healthcheckService services.IHealthcheckService
	kafkaProducer      interfaces.IKafkaProducer
	redisClient        interfaces.IRedisClient
	logger             logger.ILogger
	interval           time.Duration
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 *sync.WaitGroup
}

func NewHealthcheckWorker(
	dockerClient docker.IDockerClient,
	healthcheckService services.IHealthcheckService,
	kafkaProducer interfaces.IKafkaProducer,
	redisClient interfaces.IRedisClient,
	logger logger.ILogger,
	interval time.Duration,
) IHealthcheckWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &healthcheckWorker{
		dockerClient:       dockerClient,
		healthcheckService: healthcheckService,
		kafkaProducer:      kafkaProducer,
		redisClient:        redisClient,
		logger:             logger,
		interval:           interval,
		ctx:                ctx,
		cancel:             cancel,
		wg:                 &sync.WaitGroup{},
	}
}

func (w *healthcheckWorker) Start() {
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				return
			case <-ticker.C:
				w.healthcheck()
			}
		}
	}()
}

func (w *healthcheckWorker) Stop() {
	w.cancel()
	w.wg.Wait()
	w.logger.Info("healthcheck worker stopped")
}

func (w *healthcheckWorker) healthcheck() {
	containers, err := w.redisClient.Get(w.ctx, "containers")
	if err != nil {
		w.logger.Error("failed to get containers from redis", zap.Error(err))
		return
	}

	batch := make([]dto.StatusUpdate, 0, len(containers))
	kafkaMessages := make([]kafka.Message, 0, len(containers))

	for _, container := range containers {
		status := w.dockerClient.GetStatus(w.ctx, container.ContainerId)
		batch = append(batch, dto.StatusUpdate{
			ContainerId: container.ContainerId,
			Status:      status,
		})

		if status != container.Status {
			kafkaMessage, err := w.kafkaProducer.AddMessage(dto.KafkaStatusUpdate{
				ContainerId: container.ContainerId,
				Status:      status,
				Ipv4:        w.dockerClient.GetIpv4(w.ctx, container.ContainerId),
			})
			if err != nil {
				w.logger.Error("failed to create kafka message", zap.Error(err))
				continue
			}
			kafkaMessages = append(kafkaMessages, kafkaMessage)
		}
	}

	if len(batch) > 0 {
		if err := w.healthcheckService.Update(w.ctx, batch, w.interval); err != nil {
			w.logger.Error("failed to update status", zap.Error(err))
		}
	}

	if len(kafkaMessages) > 0 {
		if err := w.kafkaProducer.Produce(w.ctx, kafkaMessages); err != nil {
			w.logger.Error("failed to produce kafka messages", zap.Error(err))
		}
	}

	w.logger.Info("healthcheck completed", zap.Int("status_changed:", len(kafkaMessages)))
}
