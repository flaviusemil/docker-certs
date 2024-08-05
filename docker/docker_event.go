package docker

import "github.com/docker/docker/api/types/events"

type Event struct {
	ContainerID string            `json:"container_id"`
	Action      events.Action     `json:"action"`
	Actor       map[string]string `json:"actor"`
	Hosts       []string          `json:"host"`
}
