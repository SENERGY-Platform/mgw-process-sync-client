/*
 * Copyright 2021 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/controller"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/deploymentmodel"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/helper"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/server"
	paho "github.com/eclipse/paho.mqtt.golang"
	"sync"
	"testing"
	"time"
)

func TestSync(t *testing.T) {
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

	t.Run("create deployment 1", createTestDeployment(conf, mqtt, "d1", helper.BpmnExample, helper.SvgExample))
	t.Run("create deployment 2", createTestDeployment(conf, mqtt, "d2", helper.BpmnExample, helper.SvgExample))
	t.Run("create deployment 3", createTestDeployment(conf, mqtt, "d3", helper.BpmnExample, helper.SvgExample))

	t.Run("wait", func(t *testing.T) { time.Sleep(2 * time.Second) })

	deploymentId := ""
	t.Run("get deploymentId", func(t *testing.T) {
		mqttmux.Lock()
		defer mqttmux.Unlock()
		deploymentstopic := "processes/" + conf.NetworkId + "/deployment"
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

	t.Run("start deployment 1", startTestDeployment(conf, mqtt, &deploymentId))

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
	})

}

func deleteTestDeployment(conf configuration.Config, mqtt paho.Client, id *string) func(t *testing.T) {
	return func(t *testing.T) {
		token := mqtt.Publish("processes/"+conf.NetworkId+"/deployment/cmd/delete", 2, false, *id)
		if token.Wait(); token.Error() != nil {
			t.Error(token.Error())
		}
	}
}

func testSendAllKnown(ctrl *controller.Controller) func(t *testing.T) {
	return func(t *testing.T) {
		err := ctrl.SendCurrentStates()
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func startTestDeployment(conf configuration.Config, mqtt paho.Client, id *string) func(t *testing.T) {
	return func(t *testing.T) {
		payload, _ := json.Marshal(model.StartMessage{
			DeploymentId: *id,
			Parameter:    nil,
		})
		token := mqtt.Publish("processes/"+conf.NetworkId+"/deployment/cmd/start", 2, false, payload)
		if token.Wait(); token.Error() != nil {
			t.Error(token.Error())
		}
	}
}

func createTestDeployment(conf configuration.Config, mqtt paho.Client, name string, bpmn string, svg string) func(t *testing.T) {
	return func(t *testing.T) {
		msg, err := json.Marshal(deploymentmodel.Deployment{
			Name: name,
			Diagram: deploymentmodel.Diagram{
				XmlDeployed: bpmn,
				Svg:         svg,
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		token := mqtt.Publish("processes/"+conf.NetworkId+"/deployment/cmd", 2, false, msg)
		if token.Wait(); token.Error() != nil {
			t.Error(token.Error())
		}
	}
}
