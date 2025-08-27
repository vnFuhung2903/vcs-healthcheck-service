package dto

import (
	"time"

	"github.com/vnFuhung2903/vcs-healthcheck-service/entities"
)

type EsStatus struct {
	ContainerId string                   `json:"container_id"`
	Status      entities.ContainerStatus `json:"status"`
	Uptime      int64                    `json:"uptime"`
	LastUpdated time.Time                `json:"last_updated"`
	Counter     int64                    `json:"counter"`
}

type EsStatusUpdate struct {
	ContainerId string                   `json:"container_id"`
	Status      entities.ContainerStatus `json:"status"`
}

type KafkaStatusUpdate struct {
	ContainerId string                   `json:"container_id"`
	Status      entities.ContainerStatus `json:"status"`
	Ipv4        string                   `json:"ipv4"`
}

type SortOrder string

const (
	Asc SortOrder = "asc"
	Dsc SortOrder = "desc"
)
