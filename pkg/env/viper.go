package env

import (
	"errors"

	"github.com/spf13/viper"
)

type ElasticsearchEnv struct {
	ElasticsearchAddress string
}

type KafkaEnv struct {
	KafkaAddress string
}

type RedisEnv struct {
	RedisAddress  string
	RedisPassword string
	RedisDb       int
}

type WorkerEnv struct {
	Count int
}

type LoggerEnv struct {
	Level      string
	FilePath   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
}

type Env struct {
	ElasticsearchEnv ElasticsearchEnv
	KafkaEnv         KafkaEnv
	RedisEnv         RedisEnv
	LoggerEnv        LoggerEnv
	WorkerEnv        WorkerEnv
}

func LoadEnv() (*Env, error) {
	v := viper.New()
	v.AutomaticEnv()

	v.SetDefault("ELASTICSEARCH_ADDRESS", "http://localhost:9200")
	v.SetDefault("KAFKA_ADDRESS", "localhost:9092")
	v.SetDefault("REDIS_ADDRESS", "localhost:6379")
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("ZAP_LEVEL", "info")
	v.SetDefault("ZAP_FILEPATH", "./logs/app.log")
	v.SetDefault("ZAP_MAXSIZE", 100)
	v.SetDefault("ZAP_MAXAGE", 10)
	v.SetDefault("ZAP_MAXBACKUPS", 30)

	elasticsearchEnv := ElasticsearchEnv{
		ElasticsearchAddress: v.GetString("ELASTICSEARCH_ADDRESS"),
	}
	if elasticsearchEnv.ElasticsearchAddress == "" {
		return nil, errors.New("elasticsearch environment variables are empty")
	}

	kafkaEnv := KafkaEnv{
		KafkaAddress: v.GetString("KAFKA_ADDRESS"),
	}
	if kafkaEnv.KafkaAddress == "" {
		return nil, errors.New("kafka environment variables are empty")
	}

	redisEnv := RedisEnv{
		RedisAddress:  v.GetString("REDIS_ADDRESS"),
		RedisPassword: v.GetString("REDIS_PASSWORD"),
		RedisDb:       v.GetInt("REDIS_DB"),
	}
	if redisEnv.RedisAddress == "" || redisEnv.RedisDb < 0 {
		return nil, errors.New("redis environment variables are empty")
	}

	workerEnv := WorkerEnv{
		Count: v.GetInt("HEALTHCHECK_WORKER_COUNT"),
	}
	if workerEnv.Count <= 0 {
		return nil, errors.New("worker environment variables are empty")
	}

	loggerEnv := LoggerEnv{
		Level:      v.GetString("ZAP_LEVEL"),
		FilePath:   v.GetString("ZAP_FILEPATH"),
		MaxSize:    v.GetInt("ZAP_MAXSIZE"),
		MaxAge:     v.GetInt("ZAP_MAXAGE"),
		MaxBackups: v.GetInt("ZAP_MAXBACKUPS"),
	}
	if loggerEnv.Level == "" || loggerEnv.FilePath == "" || loggerEnv.MaxSize <= 0 || loggerEnv.MaxAge <= 0 || loggerEnv.MaxBackups <= 0 {
		return nil, errors.New("logger environment variables are empty or invalid")
	}

	return &Env{
		ElasticsearchEnv: elasticsearchEnv,
		KafkaEnv:         kafkaEnv,
		RedisEnv:         redisEnv,
		WorkerEnv:        workerEnv,
		LoggerEnv:        loggerEnv,
	}, nil
}
