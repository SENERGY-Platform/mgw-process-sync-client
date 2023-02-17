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
	_ "embed"
	"encoding/json"
	"github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/controller"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/events/repo"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/server"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	paho "github.com/eclipse/paho.mqtt.golang"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestMsgEventDeployment(t *testing.T) {
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
	mqtt.Subscribe("fog/#", 2, func(client paho.Client, message paho.Message) {
		mqttmux.Lock()
		defer mqttmux.Unlock()
		mqttMessages[message.Topic()] = append(mqttMessages[message.Topic()], string(message.Payload()))
	})
	mqtt.Subscribe("processes/#", 2, func(client paho.Client, message paho.Message) {
		mqttmux.Lock()
		defer mqttmux.Unlock()
		mqttMessages[message.Topic()] = append(mqttMessages[message.Topic()], string(message.Payload()))
	})

	_, err = controller.New(conf, ctx)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("create deployment", func(t *testing.T) {
		token := mqtt.Publish("processes/"+conf.NetworkId+"/cmd/deployment", 2, false, deploymentWithMsgEvents)
		if token.Wait(); token.Error() != nil {
			t.Error(token.Error())
		}
	})

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

	t.Run("check events", func(t *testing.T) {
		query := url.Values{
			"local_device_id":  {"ldid1"},
			"local_service_id": {"lsid1"},
		}
		resp, err := http.Get("http://localhost:" + conf.EventApiPort + "/event-descriptions?" + query.Encode())
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != 200 {
			temp, _ := io.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(temp))
			return
		}
		actual := []model.EventDesc{}
		err = json.NewDecoder(resp.Body).Decode(&actual)
		if err != nil {
			t.Error(err)
			return
		}
		expected := []model.EventDesc{
			{
				UserId:        repo.UserId,
				DeploymentId:  deploymentId,
				EventId:       "1",
				DeviceId:      "did1",
				ServiceId:     "sid1",
				DeviceGroupId: "",
				Script:        "x == 42",
				ValueVariable: "x",
			},
			{
				UserId:        repo.UserId,
				DeploymentId:  deploymentId,
				EventId:       "1-group",
				DeviceId:      "did1",
				ServiceId:     "sid1",
				DeviceGroupId: "gid1",
				Script:        "x == 42",
				ValueVariable: "x",
			},
		}

		sort.Slice(actual, func(i, j int) bool {
			return actual[i].EventId < actual[j].EventId
		})
		sort.Slice(expected, func(i, j int) bool {
			return expected[i].EventId < expected[j].EventId
		})
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("\n%#v\n%#v", expected, actual)
			return
		}
	})

}

//go:embed deployment.json
var deploymentWithMsgEvents string
