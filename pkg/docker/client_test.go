package docker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/vnFuhung2903/vcs-healthcheck-service/entities"
)

type DockerClientSuite struct {
	suite.Suite
	client IDockerClient
	ctx    context.Context
}

func (suite *DockerClientSuite) SetupTest() {
	suite.ctx = context.Background()

	client, err := NewDockerClient()
	suite.client = client
	suite.NoError(err)
}

func TestDockerClientSuite(t *testing.T) {
	suite.Run(t, new(DockerClientSuite))
}

func (suite *DockerClientSuite) TestContainerOffLifeCycle() {
	status := suite.client.GetStatus(suite.ctx, "test-container-id")
	suite.Equal(entities.ContainerOff, status)
	ipv4 := suite.client.GetIpv4(suite.ctx, "test-container-id")
	suite.Equal("", ipv4)
}
