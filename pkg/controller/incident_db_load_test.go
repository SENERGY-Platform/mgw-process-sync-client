/*
 * Copyright 2024 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"context"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/backend"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/camunda"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/camunda/shards"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/events"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/metadata"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/docker"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/resources"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"log"
	"sync"
	"testing"
	"time"
)

func TestLoadIncidentFromDb(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	defer wg.Wait()
	defer cancel()

	config, err := configuration.Load("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.EventApiPort, err = docker.GetFreePortStr()
	if err != nil {
		t.Error(err)
		return
	}

	connstr, camundaPgIp, _, err := docker.PostgresWithNetwork(ctx, &wg, "camunda")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("DB-CONNECTION", connstr)

	camundaUrl, err := docker.Camunda(ctx, &wg, camundaPgIp, "5432")
	if err != nil {
		t.Error(err)
		return
	}

	config.CamundaUrl = camundaUrl
	config.CamundaDb = connstr

	_, mqttIp, err := docker.Mqtt(ctx, &wg)
	if err != nil {
		t.Error(err)
		return
	}
	config.MqttBroker = "tcp://" + mqttIp + ":1883"

	config.DeploymentMetadataStorage = t.TempDir() + "/bolt.db"

	ctrl := &Controller{config: config, incidentsHandler: map[string]OnIncident{}}

	ctrl.metadata, err = metadata.NewStorage(ctx, config)
	if err != nil {
		t.Error(err)
		return
	}

	ctrl.camunda = camunda.New(config, shards.Shards(config.CamundaUrl))
	ctrl.backend, err = backend.New(config, ctx, ctrl)
	if err != nil {
		t.Error(err)
		return
	}
	ctrl.events, err = events.StartApi(ctx, config)
	if err != nil {
		t.Error(err)
		return
	}

	id, err := ctrl.CreateDeployment(model.FogDeploymentMessage{
		Deployment: deploymentmodel.Deployment{
			Version:     3,
			Id:          "test",
			Name:        "test",
			Description: "test",
			Diagram: deploymentmodel.Diagram{
				XmlRaw:      resources.ScriptErrBpmn,
				XmlDeployed: resources.ScriptErrBpmn,
				Svg:         resources.SvgExample,
			},
			Executable: true,
			IncidentHandling: &deploymentmodel.IncidentHandling{
				Restart: true,
				Notify:  true,
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(1 * time.Second)

	err = ctrl.StartDeployment(id, map[string]interface{}{})
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(20 * time.Second)

	count, err := ctrl.SendCurrentIncidents()
	if err != nil {
		t.Error(err)
		return
	}
	if count != 1 {
		t.Error(count)
	}
}
