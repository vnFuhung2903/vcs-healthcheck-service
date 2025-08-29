package dto

import (
	"time"

	"github.com/vnFuhung2903/vcs-healthcheck-service/entities"
)

type BulkMeta struct {
	Index  *BaseMeta `json:"index,omitempty"`
	Update *BaseMeta `json:"update,omitempty"`
}

type BaseMeta struct {
	Index string `json:"_index"`
	ID    string `json:"_id"`
}

type EsStatus struct {
	ContainerId string                   `json:"container_id"`
	Status      entities.ContainerStatus `json:"status"`
	Uptime      int64                    `json:"uptime"`
	LastUpdated time.Time                `json:"last_updated"`
	Counter     int64                    `json:"counter"`
}

type SortOrder string

const (
	Asc SortOrder = "asc"
	Dsc SortOrder = "desc"
)
