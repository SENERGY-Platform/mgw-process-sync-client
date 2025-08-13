package docker

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/resources"
	"github.com/testcontainers/testcontainers-go"
)

func TaskWorker(ctx context.Context, wg *sync.WaitGroup, mqttUrl string, camundaUrl string, networkId string) (err error) {
	log.Println("start mgw-external-task-worker")
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			AlwaysPullImage: true,
			Image:           "ghcr.io/senergy-platform/mgw-external-task-worker:dev",
			Env: map[string]string{
				"MQTT_BROKER":         mqttUrl,
				"CAMUNDA_URL":         camundaUrl,
				"COMPLETION_STRATEGY": "pessimistic",
				"CAMUNDA_TOPIC":       "pessimistic",
				"DEBUG":               "true",
				"SYNC_NETWORK_ID":     networkId,
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
		defer func() {
			log.Println("DEBUG: remove container task-worker", c.Terminate(context.Background()))
		}()
		<-ctx.Done()
		reader, err := c.Logs(context.Background())
		if err != nil {
			log.Println("ERROR: unable to get container log")
			return
		}
		buf := new(strings.Builder)
		io.Copy(buf, reader)
		fmt.Println("TASK-WORKER LOGS: ------------------------------------------")
		fmt.Println(buf.String())
		fmt.Println("\n---------------------------------------------------------------")
	}()

	return err
}
