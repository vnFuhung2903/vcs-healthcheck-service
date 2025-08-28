package workers

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/vnFuhung2903/vcs-healthcheck-service/dto"
	"github.com/vnFuhung2903/vcs-healthcheck-service/interfaces"
	"github.com/vnFuhung2903/vcs-healthcheck-service/pkg/docker"
	"github.com/vnFuhung2903/vcs-healthcheck-service/pkg/logger"
	"github.com/vnFuhung2903/vcs-healthcheck-service/usecases/services"
	"go.uber.org/zap"
)

type IHealthcheckWorker interface {
	Start(numWorkers int)
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

func (w *healthcheckWorker) Start(numWorkers int) {
	w.wg.Add(numWorkers)
	go w.run()
}

func (w *healthcheckWorker) Stop() {
	w.cancel()
	w.wg.Wait()
}

func (w *healthcheckWorker) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("healthcheck status workers stopped")
			return
		case <-ticker.C:
			w.updateHealthcheck()
		}
	}
}

func (w *healthcheckWorker) updateHealthcheck() {
	containers, err := w.redisClient.Get(w.ctx, "containers")
	if err != nil {
		w.logger.Error("failed to get container ids from redis", zap.Error(err))
		return
	}

	statusList := make([]dto.EsStatusUpdate, 0)

	for _, container := range containers {
		status := w.dockerClient.GetStatus(w.ctx, container.ContainerId)
		if status != container.Status {
			updateStatus := dto.KafkaStatusUpdate{
				ContainerId: container.ContainerId,
				Status:      status,
				Ipv4:        w.dockerClient.GetIpv4(w.ctx, container.ContainerId),
			}
			kafkaMessage, err := json.Marshal(updateStatus)
			if err != nil {
				w.logger.Error("failed to marshal kafka message", zap.String("container_id", container.ContainerId), zap.Error(err))
				continue
			}
			if err := w.kafkaProducer.Produce(w.ctx, kafkaMessage); err != nil {
				w.logger.Error("failed to send kafka message", zap.String("container_id", container.ContainerId), zap.Error(err))
				continue
			}
			w.logger.Info("kafka producer sent message successfully", zap.String("container_id", container.ContainerId), zap.String("status", string(status)))
		}
		statusList = append(statusList, dto.EsStatusUpdate{
			ContainerId: container.ContainerId,
			Status:      status,
		})
	}

	if err := w.healthcheckService.UpdateStatus(w.ctx, statusList, w.interval); err != nil {
		w.logger.Error("failed to update elasticsearch status", zap.Error(err))
		return
	}
	w.logger.Info("healthcheck status updated successfully")
}
