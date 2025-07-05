package docker

import (
	"context"
	"docker-certs/core/eventbus"
	"docker-certs/core/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"log"
)

func ScanRunningContainers(ctx context.Context) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	f := filters.NewArgs()
	f.Add("status", "running")

	containers, err := cli.ContainerList(ctx, container.ListOptions{
		Filters: f,
	})
	if err != nil {
		return err
	}

	for _, c := range containers {
		eventbus.Publish(types.Event[Event]{
			Type: types.ContainerStarted,
			Payload: Event{
				ID:         c.ID,
				Action:     "start",
				Attributes: c.Labels,
			},
		})
	}
	log.Printf("[docker] Initial scan: published %d running containers", len(containers))
	return nil
}
