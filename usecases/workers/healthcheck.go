package workers

import (
	"context"
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

type HealthcheckWorker struct {
	dockerClient       docker.IDockerClient
	healthcheckService services.IHealthcheckService
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
	redisClient interfaces.IRedisClient,
	logger logger.ILogger,
	interval time.Duration,
) IHealthcheckWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &HealthcheckWorker{
		dockerClient:       dockerClient,
		healthcheckService: healthcheckService,
		redisClient:        redisClient,
		logger:             logger,
		interval:           interval,
		ctx:                ctx,
		cancel:             cancel,
		wg:                 &sync.WaitGroup{},
	}
}

func (w *HealthcheckWorker) Start(numWorkers int) {
	w.wg.Add(numWorkers)
	go w.run()
}

func (w *HealthcheckWorker) Stop() {
	w.cancel()
	w.wg.Wait()
}

func (w *HealthcheckWorker) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("elasticsearch status workers stopped")
			return
		case <-ticker.C:
			w.updateHealthcheck()
		}
	}
}

func (w *HealthcheckWorker) updateHealthcheck() {
	containers, err := w.redisClient.Get(w.ctx, "containers")
	if err != nil {
		w.logger.Error("failed to get container ids from redis", zap.Error(err))
		return
	}

	statusList := make([]dto.EsStatusUpdate, 0)

	for _, container := range containers {
		status := w.dockerClient.GetStatus(w.ctx, container.ContainerId)
		if status != container.Status {
			// if err := w.containerService.Update(w.ctx, container.ContainerId, dto.ContainerUpdate{Status: status}); err != nil {
			// 	w.logger.Error("failed to update container", zap.String("container_id", container.ContainerId))
			// }
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
	w.logger.Info("elasticsearch status updated successfully")
}
