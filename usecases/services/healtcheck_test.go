package services

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/vnFuhung2903/vcs-healthcheck-service/dto"
	"github.com/vnFuhung2903/vcs-healthcheck-service/mocks/docker"
	"github.com/vnFuhung2903/vcs-healthcheck-service/mocks/interfaces"
	"github.com/vnFuhung2903/vcs-healthcheck-service/mocks/logger"
)

type HealthcheckServiceSuite struct {
	suite.Suite
	ctrl               *gomock.Controller
	healthcheckService IHealthcheckService
	mockDockerClient   *docker.MockIDockerClient
	mockEsClient       *interfaces.MockIElasticsearchClient
	mockLogger         *logger.MockILogger
	ctx                context.Context
}

func (s *HealthcheckServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockDockerClient = docker.NewMockIDockerClient(s.ctrl)
	s.mockEsClient = interfaces.NewMockIElasticsearchClient(s.ctrl)
	s.mockLogger = logger.NewMockILogger(s.ctrl)
	s.healthcheckService = NewHealthcheckService(s.mockEsClient, s.mockLogger)
	s.ctx = context.Background()
}

func (s *HealthcheckServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestHealthcheckServiceSuite(t *testing.T) {
	suite.Run(t, new(HealthcheckServiceSuite))
}

func (s *HealthcheckServiceSuite) TestUpdateNewDocument() {
	statusList := []dto.StatusUpdate{
		{ContainerId: "container1", Status: "ON"},
	}

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": [
                    {
                        "hits": {
                            "hits": []
                        }
                    }
                ]
            }`)),
		}
		return response, nil
	}).Times(1)

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
	            "responses": [
	                {
	                    "hits": {
	                        "hits": [
	                            {
	                                "_id": "container1",
	                                "_source": {
	                                    "container_id": "container1",
	                                    "status": "ON",
	                                    "uptime": 3600,
	                                    "last_updated": "` + time.Now().Format(time.RFC3339) + `",
										"counter": 1
	                                }
	                            }
	                        ]
	                    }
	                }
	            ]
	        }`)),
		}
		return response, nil
	}).Times(1)

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"took":1,"errors":false}`)),
		}
		return response, nil
	}).Times(1)

	s.mockLogger.EXPECT().Info("elasticsearch status retrieved successfully", gomock.Any()).Times(2)
	s.mockLogger.EXPECT().Info("elasticsearch status indexed successfully").Times(1)

	err := s.healthcheckService.Update(s.ctx, statusList, time.Hour)
	s.NoError(err)
}

func (s *HealthcheckServiceSuite) TestUpdateUpdateDocument() {
	statusList := []dto.StatusUpdate{
		{ContainerId: "container1", Status: "OFF"},
	}

	lastUpdated := time.Now().Add(-1 * time.Hour)

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": [
                    {
                        "hits": {
                            "hits": [
                                {
                                    "_id": "container1",
                                    "_source": {
                                        "container_id": "container1",
                                        "status": "ON",
                                        "uptime": 3600,
                                        "last_updated": "` + lastUpdated.Format(time.RFC3339) + `",
										"counter": 1
                                    }
                                }
                            ]
                        }
                    }
                ]
            }`)),
		}
		return response, nil
	}).Times(2)

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"took":1,"errors":false}`)),
		}
		return response, nil
	}).Times(1)

	s.mockLogger.EXPECT().Info("elasticsearch status retrieved successfully", gomock.Any()).Times(2)
	s.mockLogger.EXPECT().Info("elasticsearch status indexed successfully").Times(1)

	err := s.healthcheckService.Update(s.ctx, statusList, time.Hour)
	s.NoError(err)
}

func (s *HealthcheckServiceSuite) TestUpdateGetEsStatusError() {
	statusList := []dto.StatusUpdate{
		{ContainerId: "container1", Status: "ON"},
	}

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).Return(nil, errors.New("elasticsearch error")).Times(1)
	s.mockLogger.EXPECT().Error("failed to msearch elasticsearch status", gomock.Any()).Times(1)

	err := s.healthcheckService.Update(s.ctx, statusList, time.Hour)
	s.Error(err)
	s.Contains(err.Error(), "elasticsearch error")
}

func (s *HealthcheckServiceSuite) TestUpdateGetPreviousStatusError() {
	statusList := []dto.StatusUpdate{
		{ContainerId: "container1", Status: "ON"},
	}

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": [
                    {
                        "hits": {
                            "hits": []
                        }
                    }
                ]
            }`)),
		}
		return response, nil
	}).Times(1)

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).Return(nil, errors.New("elasticsearch error")).Times(1)
	s.mockLogger.EXPECT().Info("elasticsearch status retrieved successfully", gomock.Any()).Times(1)
	s.mockLogger.EXPECT().Error("failed to msearch elasticsearch status", gomock.Any()).Times(1)

	err := s.healthcheckService.Update(s.ctx, statusList, time.Hour)
	s.ErrorContains(err, "elasticsearch error")
}

func (s *HealthcheckServiceSuite) TestUpdateBulkError() {
	statusList := []dto.StatusUpdate{
		{ContainerId: "container1", Status: "ON"},
	}

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": [
                    {
                        "hits": {
                            "hits": []
                        }
                    }
                ]
            }`)),
		}
		return response, nil
	}).Times(2)

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).Return(nil, errors.New("bulk error")).Times(1)
	s.mockLogger.EXPECT().Info("elasticsearch status retrieved successfully", gomock.Any()).Times(2)
	s.mockLogger.EXPECT().Error("failed to bulk elasticsearch status", gomock.Any()).Times(1)

	err := s.healthcheckService.Update(s.ctx, statusList, time.Hour)
	s.Error(err)
	s.Contains(err.Error(), "bulk error")
}

func (s *HealthcheckServiceSuite) TestUpdateSameStatusUpdate() {
	statusList := []dto.StatusUpdate{
		{ContainerId: "container1", Status: "ON"},
	}

	lastUpdated := time.Now().Add(-1 * time.Hour)

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`{
                "responses": [
                    {
                        "hits": {
                            "hits": [
                                {
                                    "_id": "container1",
                                    "_source": {
                                        "container_id": "container1",
                                        "status": "ON",
                                        "uptime": 3600,
                                        "last_updated": "` + lastUpdated.Format(time.RFC3339) + `",
										"counter": 1
                                    }
                                }
                            ]
                        }
                    }
                ]
            }`)),
		}
		return response, nil
	}).Times(2)

	s.mockEsClient.EXPECT().Do(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
		response := &esapi.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"took":1,"errors":false}`)),
		}
		return response, nil
	}).Times(1)

	s.mockLogger.EXPECT().Info("elasticsearch status retrieved successfully", gomock.Any()).Times(2)
	s.mockLogger.EXPECT().Info("elasticsearch status indexed successfully").Times(1)

	err := s.healthcheckService.Update(s.ctx, statusList, time.Hour)
	s.NoError(err)
}
