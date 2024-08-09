package docker

import (
	"context"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/resources"
	"github.com/testcontainers/testcontainers-go"
	"log"
	"strings"
	"sync"
)

func TaskWorker(ctx context.Context, wg *sync.WaitGroup, mqttUrl string, camundaUrl string) (err error) {
	log.Println("start mgw-external-task-worker")
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "ghcr.io/senergy-platform/mgw-external-task-worker:dev",
			Env: map[string]string{
				"MQTT_BROKER":         mqttUrl,
				"CAMUNDA_URL":         camundaUrl,
				"COMPLETION_STRATEGY": "pessimistic",
				"CAMUNDA_TOPIC":       "pessimistic",
			},
			Files: []testcontainers.ContainerFile{
				{
					Reader:            strings.NewReader(resources.RepoFallbackFile),
					ContainerFilePath: "/root/devicerepo_fallback.json",
					FileMode:          777,
				},
			},
		},
		Started: true,
	})
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		log.Println("DEBUG: remove container camunda", c.Terminate(context.Background()))
	}()

	return err
}
