package workers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/vnFuhung2903/vcs-healthcheck-service/dto"
	"github.com/vnFuhung2903/vcs-healthcheck-service/entities"
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
	job                chan entities.ContainerWithStatus
	wg                 *sync.WaitGroup
	batchSize          int
}

func NewHealthcheckWorker(
	dockerClient docker.IDockerClient,
	healthcheckService services.IHealthcheckService,
	kafkaProducer interfaces.IKafkaProducer,
	redisClient interfaces.IRedisClient,
	logger logger.ILogger,
	interval time.Duration,
	batchSize int,
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
		job:                make(chan entities.ContainerWithStatus, 10000),
		wg:                 &sync.WaitGroup{},
		batchSize:          batchSize,
	}
}

func (w *healthcheckWorker) Start(numWorkers int) {
	for range numWorkers {
		w.wg.Add(1)
		go w.run()
	}

	ticker := time.NewTicker(w.interval)

	go func() {
		defer fmt.Println("ticker goroutine stopped")
		defer ticker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				return
			case <-ticker.C:
				containers, err := w.redisClient.Get(w.ctx, "containers")
				if err != nil {
					w.logger.Error("failed to get containers from redis", zap.Error(err))
					continue
				}
				for _, c := range containers {
					select {
					case w.job <- c:
					case <-w.ctx.Done():
						return
					default:
						w.logger.Warn("job queue full, dropping", zap.String("containerId", c.ContainerId))
					}
				}
			}
		}
	}()
}

func (w *healthcheckWorker) Stop() {
	w.cancel()
	close(w.job)
	w.wg.Wait()
	w.logger.Info("all healthcheck status workers stopped")
}

func (w *healthcheckWorker) run() {
	defer w.wg.Done()
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	batch := make([]dto.StatusUpdate, 0)
	kafkaMessages := make([]kafka.Message, 0)

	for {
		select {
		case <-w.ctx.Done():
			w.updateHealthcheck(&batch, &kafkaMessages)
			return
		case c, ok := <-w.job:
			if !ok {
				w.updateHealthcheck(&batch, &kafkaMessages)
				return
			}
			status := w.dockerClient.GetStatus(w.ctx, c.ContainerId)
			batch = append(batch, dto.StatusUpdate{ContainerId: c.ContainerId, Status: status})
			if status != c.Status {
				kafkaMessage, err := w.kafkaProducer.AddMessage(dto.KafkaStatusUpdate{
					ContainerId: c.ContainerId,
					Status:      status,
					Ipv4:        w.dockerClient.GetIpv4(w.ctx, c.ContainerId),
				})
				if err != nil {
					w.logger.Error("failed to create kafka message", zap.Error(err))
					continue
				}
				kafkaMessages = append(kafkaMessages, kafkaMessage)
			}
			if len(batch) >= w.batchSize {
				w.updateHealthcheck(&batch, &kafkaMessages)
			}
		case <-ticker.C:
			w.updateHealthcheck(&batch, &kafkaMessages)
		}
	}
}

func (w *healthcheckWorker) updateHealthcheck(batch *[]dto.StatusUpdate, kafkaMessages *[]kafka.Message) {
	if err := w.healthcheckService.UpdateStatus(w.ctx, *batch, w.interval); err != nil {
		w.logger.Error("failed to update status", zap.Error(err))
	}

	if err := w.kafkaProducer.Produce(w.ctx, *kafkaMessages); err != nil {
		w.logger.Error("failed to produce kafka messages", zap.Error(err))
	}

	*batch = (*batch)[:0]
	*kafkaMessages = (*kafkaMessages)[:0]
}
