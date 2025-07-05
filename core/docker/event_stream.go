package docker

import (
	"context"
	"docker-certs/core/configs"
	"docker-certs/core/eventbus"
	"docker-certs/core/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"log"
	"time"
)

func ListenToDockerEvents(ctx context.Context) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}

	appConfig := configs.GetConfig()

	f := filters.NewArgs()
	f.Add("type", "container")
	f.Add("event", "start")
	f.Add("event", "stop")

	log.Println("[docker] Listening for container events...")

	for {
		msgs, errs := cli.Events(ctx, events.ListOptions{Filters: f})
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-msgs:
				if !ok {
					log.Println("[docker] event stream closed")
					time.Sleep(time.Second)
					goto reconnect
				}
				if appConfig.Debug {
					log.Println("[docker] Docker event received", m)
				}
				msgType := types.ContainerStarted

				if m.Action == "stop" {
					msgType = types.ContainerStopped
				}

				eventbus.Publish(types.Event[Event]{
					Type: msgType,
					Payload: Event{
						ID:         m.Actor.ID,
						Action:     string(m.Action),
						Attributes: m.Actor.Attributes,
					},
				})
			case err := <-errs:
				log.Println("[docker] event stream error:", err)
				time.Sleep(time.Second)
				goto reconnect
			}
		}
	reconnect:
	}
}
