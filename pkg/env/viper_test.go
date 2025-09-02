package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ViperSuite struct {
	suite.Suite
}

func TestViperSuite(t *testing.T) {
	suite.Run(t, new(ViperSuite))
}

func (suite *ViperSuite) SetupTest() {
	envVars := []string{
		"KAFKA_ADDRESS",
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_CONTAINER_DB",
		"REDIS_ADDRESS",
		"REDIS_PASSWORD",
		"REDIS_DB",
		"ZAP_LEVEL",
		"ZAP_FILEPATH",
		"ZAP_MAXSIZE",
		"ZAP_MAXAGE",
		"ZAP_MAXBACKUPS",
	}

	for _, env := range envVars {
		os.Unsetenv(env)
	}
}

func (suite *ViperSuite) createEnvVars(vars map[string]string) {
	for k, v := range vars {
		err := os.Setenv(k, v)
		suite.Require().NoError(err)
	}
}

func (suite *ViperSuite) TestLoadEnv() {
	envContent := map[string]string{
		"ZAP_LEVEL":      "info",
		"ZAP_FILEPATH":   "/tmp/app.log",
		"ZAP_MAXSIZE":    "100",
		"ZAP_MAXAGE":     "30",
		"ZAP_MAXBACKUPS": "5",
	}

	suite.createEnvVars(envContent)
	env, err := LoadEnv()
	suite.NoError(err)
	suite.NotNil(env)

	suite.Equal("info", env.LoggerEnv.Level)
	suite.Equal("/tmp/app.log", env.LoggerEnv.FilePath)
	suite.Equal(100, env.LoggerEnv.MaxSize)
	suite.Equal(30, env.LoggerEnv.MaxAge)
	suite.Equal(5, env.LoggerEnv.MaxBackups)
}

func (suite *ViperSuite) TestLoadEnvInvalidLoggerValues() {
	envContent := map[string]string{
		"ZAP_MAXSIZE":    "invalid_number",
		"ZAP_MAXAGE":     "invalid_number",
		"ZAP_MAXBACKUPS": "invalid_number",
	}
	suite.createEnvVars(envContent)
	env, err := LoadEnv()

	suite.Error(err)
	suite.Nil(env)
}

func (suite *ViperSuite) TestLoadEnvInvalidRedisValues() {
	envContent := map[string]string{
		"REDIS_DB": "-1",
	}
	suite.createEnvVars(envContent)
	env, err := LoadEnv()

	suite.Error(err)
	suite.Nil(env)
}
