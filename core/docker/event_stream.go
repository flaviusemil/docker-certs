package docker

import (
	"context"
	"docker-certs/core/eventbus"
	"docker-certs/core/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"log"
)

func ListenToDockerEvents() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error createing Docker client: %v", err)
	}

	messages, errs := cli.Events(context.Background(), events.ListOptions{})
	log.Println("[docker] Listening for container events...")

	for {
		select {
		case msg := <-messages:
			if msg.Type == "container" && (msg.Action == "start" || msg.Action == "stop") {
				log.Println("Docker event received", msg)

				msgType := types.ContainerStarted

				if msg.Action == "stop" {
					msgType = types.ContainerStopped
				}

				eventbus.Publish(types.Event[Event]{
					Type: msgType,
					Payload: Event{
						ID:         msg.Actor.ID,
						Action:     string(msg.Action),
						Attributes: msg.Actor.Attributes,
					},
				})
			}
		case err := <-errs:
			log.Fatalf("Error listening for events: %v", err)
		}
	}
}
