package dto

import "github.com/vnFuhung2903/vcs-healthcheck-service/entities"

type KafkaStatusUpdate struct {
	ContainerId string                   `json:"container_id"`
	Status      entities.ContainerStatus `json:"status"`
	Ipv4        string                   `json:"ipv4"`
}
