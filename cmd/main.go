package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vnFuhung2903/vcs-healthcheck-service/infrastructures/databases"
	"github.com/vnFuhung2903/vcs-healthcheck-service/interfaces"
	"github.com/vnFuhung2903/vcs-healthcheck-service/pkg/docker"
	"github.com/vnFuhung2903/vcs-healthcheck-service/pkg/env"
	"github.com/vnFuhung2903/vcs-healthcheck-service/pkg/logger"
	"github.com/vnFuhung2903/vcs-healthcheck-service/usecases/services"
	"github.com/vnFuhung2903/vcs-healthcheck-service/usecases/workers"
)

func main() {
	env, err := env.LoadEnv()
	if err != nil {
		log.Fatalf("Failed to retrieve env: %v", err)
	}

	logger, err := logger.LoadLogger(env.LoggerEnv)
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}

	esRawClient, err := databases.NewElasticsearchFactory(env.ElasticsearchEnv).ConnectElasticsearch()
	if err != nil {
		log.Fatalf("Failed to create docker client: %v", err)
	}
	esClient := interfaces.NewElasticsearchClient(esRawClient)

	redisRawClient := databases.NewRedisFactory(env.RedisEnv).ConnectRedis()
	redisClient := interfaces.NewRedisClient(redisRawClient)

	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		log.Fatalf("Failed to create docker client: %v", err)
	}

	healthcheckService := services.NewHealthcheckService(esClient, logger)

	healthcheckWorker := workers.NewHealthcheckWorker(
		dockerClient,
		healthcheckService,
		redisClient,
		logger,
		10*time.Second,
	)
	healthcheckWorker.Start(1)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		logger.Info("Shutting down...")
		healthcheckWorker.Stop()
		os.Exit(0)
	}()
}
