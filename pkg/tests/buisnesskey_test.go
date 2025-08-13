/*
 * Copyright 2025 InfAI (CC SES)
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

package tests

import (
	"context"
	"encoding/json"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/controller"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/camundamodel"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/docker"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/helper"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/resources"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/server"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	paho "github.com/eclipse/paho.mqtt.golang"
)

func TestBusinessKey(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := server.CreateSyncEnv(ctx, wg, configuration.Config{InitialWaitDuration: "1s"})
	if err != nil {
		t.Error(err)
		return
	}

	mqtt := paho.NewClient(paho.NewClientOptions().
		SetPassword(conf.MqttPw).
		SetUsername(conf.MqttUser).
		SetAutoReconnect(true).
		SetCleanSession(false).
		SetClientID("test-client").
		AddBroker(conf.MqttBroker))
	if token := mqtt.Connect(); token.Wait() && token.Error() != nil {
		t.Error(token.Error())
		return
	}

	mqttMessages := map[string][]string{}
	mqttmux := sync.Mutex{}
	mqtt.Subscribe("processes/#", 2, func(client paho.Client, message paho.Message) {
		mqttmux.Lock()
		defer mqttmux.Unlock()
		mqttMessages[message.Topic()] = append(mqttMessages[message.Topic()], string(message.Payload()))
	})

	ctrl, err := controller.New(conf, ctx)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("create deployment 1", createTestDeployment(conf, mqtt, "d1", resources.LongProcess, helper.SvgExample))

	t.Run("wait", func(t *testing.T) { time.Sleep(2 * time.Second) })

	deploymentId := ""
	t.Run("get deploymentId", func(t *testing.T) {
		mqttmux.Lock()
		defer mqttmux.Unlock()
		deploymentstopic := "processes/" + conf.NetworkId + "/state/deployment"
		deployments := mqttMessages[deploymentstopic]
		if len(deployments) == 0 {
			t.Error("expect deployments")
			return
		}
		depl := deploymentmodel.Deployment{}
		err := json.Unmarshal([]byte(deployments[0]), &depl)
		if err != nil {
			t.Error(err)
			return
		}
		deploymentId = depl.Id
	})

	t.Run("start deployment 1", startTestDeploymentWithBusinessKey(conf, mqtt, deploymentId, "testbid"))

	t.Run("wait", func(t *testing.T) { time.Sleep(2 * time.Second) })

	t.Run("send all known", testSendAllKnown(ctrl))

	t.Run("wait", func(t *testing.T) { time.Sleep(2 * time.Second) })

	t.Run("delete deployment 1", deleteTestDeployment(conf, mqtt, &deploymentId))

	t.Run("wait", func(t *testing.T) { time.Sleep(2 * time.Second) })

	t.Run("send all known", testSendAllKnown(ctrl))

	t.Run("wait", func(t *testing.T) { time.Sleep(2 * time.Second) })

	t.Run("check mqtt messages", func(t *testing.T) {
		temp, err := json.Marshal(mqttMessages)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(string(temp))
		for _, instanceMsg := range mqttMessages["processes/test-network-id/state/process-instance"] {
			instance := camundamodel.ProcessInstance{}
			err = json.Unmarshal([]byte(instanceMsg), &instance)
			if err != nil {
				t.Error(err)
				return
			}
			if instance.BusinessKey != "testbid" {
				t.Errorf("expect business key = '%s', got '%s'", "testbid", instance.BusinessKey)
				return
			}
		}

		for _, instanceMsg := range mqttMessages["processes/test-network-id/state/process-instance-history"] {
			instance := camundamodel.HistoricProcessInstance{}
			err = json.Unmarshal([]byte(instanceMsg), &instance)
			if err != nil {
				t.Error(err)
				return
			}
			if instance.BusinessKey != "testbid" {
				t.Errorf("expect business key = '%s', got '%s'", "testbid", instance.BusinessKey)
				return
			}
		}
	})

}

func TestIncidentBusinessKey(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := server.CreateSyncEnv(ctx, wg, configuration.Config{
		InitialWaitDuration:       "1s",
		Debug:                     true,
		TaskTopicReplace:          map[string]string{"optimistic": "pessimistic"},
		DeploymentMetadataStorage: t.TempDir() + "/bolt.db",
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = docker.TaskWorker(ctx, wg, conf.MqttBroker, conf.CamundaUrl, conf.NetworkId)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(5 * time.Second)

	mqtt := paho.NewClient(paho.NewClientOptions().
		SetPassword(conf.MqttPw).
		SetUsername(conf.MqttUser).
		SetAutoReconnect(true).
		SetCleanSession(false).
		SetClientID("test-client").
		AddBroker(conf.MqttBroker))
	if token := mqtt.Connect(); token.Wait() && token.Error() != nil {
		t.Error(token.Error())
		return
	}

	mqttMessages := map[string][]string{}
	mqttmux := sync.Mutex{}
	mqtt.Subscribe("processes/#", 2, func(client paho.Client, message paho.Message) {
		mqttmux.Lock()
		defer mqttmux.Unlock()
		mqttMessages[message.Topic()] = append(mqttMessages[message.Topic()], string(message.Payload()))
	})

	ctrl, err := controller.New(conf, ctx)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("create deployment 1", createTestDeploymentWithRestart(conf, mqtt, "d1", resources.IncidentBpmn, helper.SvgExample))

	t.Run("wait", func(t *testing.T) { time.Sleep(2 * time.Second) })

	deploymentId := ""
	t.Run("get deploymentId", func(t *testing.T) {
		mqttmux.Lock()
		defer mqttmux.Unlock()
		deploymentstopic := "processes/" + conf.NetworkId + "/state/deployment"
		deployments := mqttMessages[deploymentstopic]
		if len(deployments) == 0 {
			t.Error("expect deployments")
			return
		}
		depl := deploymentmodel.Deployment{}
		err := json.Unmarshal([]byte(deployments[0]), &depl)
		if err != nil {
			t.Error(err)
			return
		}
		deploymentId = depl.Id
	})

	t.Run("start deployment 1", startTestDeploymentWithBusinessKey(conf, mqtt, deploymentId, "testbid"))

	t.Run("wait", func(t *testing.T) { time.Sleep(5 * time.Second) })

	t.Run("send all known", testSendAllKnown(ctrl))

	t.Run("wait", func(t *testing.T) { time.Sleep(5 * time.Second) })

	t.Run("delete deployment 1", deleteTestDeployment(conf, mqtt, &deploymentId))

	t.Run("wait", func(t *testing.T) { time.Sleep(2 * time.Second) })

	t.Run("send all known", testSendAllKnown(ctrl))

	t.Run("wait", func(t *testing.T) { time.Sleep(2 * time.Second) })

	t.Run("check mqtt messages", func(t *testing.T) {
		temp, err := json.Marshal(mqttMessages)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(string(temp))

		for _, instanceMsg := range mqttMessages["processes/test-network-id/state/incident"] {
			instance := camundamodel.Incident{}
			err = json.Unmarshal([]byte(instanceMsg), &instance)
			if err != nil {
				t.Error(err)
				return
			}
			if instance.BusinessKey != "testbid" {
				t.Errorf("expect business key = '%s', got '%s'", "testbid", instance.BusinessKey)
				break
			}
		}

		for _, instanceMsg := range mqttMessages["processes/test-network-id/state/process-instance"] {
			instance := camundamodel.ProcessInstance{}
			err = json.Unmarshal([]byte(instanceMsg), &instance)
			if err != nil {
				t.Error(err)
				return
			}
			if instance.BusinessKey != "testbid" {
				t.Errorf("expect business key = '%s', got '%s'", "testbid", instance.BusinessKey)
				break
			}
		}

		historicInstances := mqttMessages["processes/test-network-id/state/process-instance-history"]
		distinctInstances := []camundamodel.HistoricProcessInstance{}
		for _, instanceMsg := range historicInstances {
			instance := camundamodel.HistoricProcessInstance{}
			err = json.Unmarshal([]byte(instanceMsg), &instance)
			if err != nil {
				t.Error(err)
				return
			}
			if instance.BusinessKey != "testbid" {
				t.Errorf("expect business key = '%s', got '%s'", "testbid", instance.BusinessKey)
				break
			}
			if !slices.ContainsFunc(distinctInstances, func(element camundamodel.HistoricProcessInstance) bool {
				return element.Id == instance.Id
			}) {
				distinctInstances = append(distinctInstances, instance)
			}
		}

		if len(distinctInstances) < 2 {
			t.Error("expect at least 2 instances, got", len(distinctInstances))
		}

	})
}
