package docker

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/vnFuhung2903/vcs-healthcheck-service/entities"
)

type IDockerClient interface {
	GetStatus(ctx context.Context, containerID string) entities.ContainerStatus
	GetIpv4(ctx context.Context, containerID string) string
}

type dockerClient struct {
	client *client.Client
}

func NewDockerClient() (IDockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &dockerClient{
		client: cli,
	}, nil
}

func (c *dockerClient) GetStatus(ctx context.Context, containerId string) entities.ContainerStatus {
	inspect, err := c.client.ContainerInspect(ctx, containerId)
	if err != nil {
		return entities.ContainerOff
	}

	var status entities.ContainerStatus
	if inspect.State.Running {
		status = entities.ContainerOn
	} else {
		status = entities.ContainerOff
	}
	return status
}

func (c *dockerClient) GetIpv4(ctx context.Context, containerId string) string {
	inspect, err := c.client.ContainerInspect(ctx, containerId)
	if err != nil {
		return ""
	}

	for _, network := range inspect.NetworkSettings.Networks {
		if network.IPAddress != "" {
			return network.IPAddress
		}
	}
	return ""
}
