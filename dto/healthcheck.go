package dto

import (
	"github.com/vnFuhung2903/vcs-healthcheck-service/entities"
)

type StatusUpdate struct {
	ContainerId string                   `json:"container_id"`
	Status      entities.ContainerStatus `json:"status"`
}
