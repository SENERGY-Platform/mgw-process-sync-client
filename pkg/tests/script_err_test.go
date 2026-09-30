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

package tests

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/controller"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/docker"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/resources"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	paho "github.com/eclipse/paho.mqtt.golang"
)

// INSERT INTO public.act_ru_incident (id_, rev_, incident_timestamp_, incident_msg_, incident_type_, execution_id_, activity_id_, proc_inst_id_, proc_def_id_, cause_incident_id_, root_cause_incident_id_, configuration_, tenant_id_, job_def_id_) VALUES ('6da046d1-5580-11ef-9030-0242ac11000a', 1, '2024-08-08 12:19:15.517000', 'Unable to evaluate script while executing activity ”Task_0fi26gl” in the process definition with id ”script_err:1:631e00d6-5580-11ef-9030-0242ac11000a”:TypeError: Cannot read property "batz" from undefined in <eval> at line number 2', 'failedJob', '6614ab9a-5580-11ef-9030-0242ac11000a', 'IntermediateThrowEvent_1jxyivh', '6613e848-5580-11ef-9030-0242ac11000a', 'script_err:1:631e00d6-5580-11ef-9030-0242ac11000a', '6da046d1-5580-11ef-9030-0242ac11000a', '6da046d1-5580-11ef-9030-0242ac11000a', '66156eec-5580-11ef-9030-0242ac11000a', 'senergy', '631e27e7-5580-11ef-9030-0242ac11000a');
func TestScriptError(t *testing.T) {
	result := runScriptErrorProcess(t, scriptErrorScenario{restartLimit: 0, localRestart: true})
	if len(result.notifications) < 2 {
		t.Error("notification count should be greater than 2")
	}
	if len(result.incidents) < 2 {
		t.Error("incident count should be greater than 2")
	}
	if result.deleteCount < 2 {
		t.Error("deleteCount count should be greater than 2")
	}
}

func TestScriptErrorStopsRestartingAtRestartLimit(t *testing.T) {
	result := runScriptErrorProcess(t, scriptErrorScenario{restartLimit: 1, localRestart: true})
	notifications := result.notifications
	if len(notifications) != 2 {
		t.Fatalf("expected one notification for the restarted and one for the finally stopped process, got %v: %#v", len(notifications), notifications)
	}
	if !strings.Contains(notifications[0], "process will be restarted") {
		t.Errorf("expected first notification to announce the restart, got %v", notifications[0])
	}
	if !strings.Contains(notifications[1], "process will not be restarted") {
		t.Errorf("expected last notification to announce the final stop, got %v", notifications[1])
	}
}

// the process-sync warden restarts a process after an incident by sending a start command with the same business key
func TestScriptErrorRejectsWardenStartAtRestartLimit(t *testing.T) {
	result := runScriptErrorProcess(t, scriptErrorScenario{
		restartLimit: 1,
		localRestart: false,
		businessKey:  "warden-bk",
		afterwards: func(start func(businessKey string)) {
			start("warden-bk") // second incident: limit reached
			time.Sleep(20 * time.Second)
			start("warden-bk") // rejected
			time.Sleep(10 * time.Second)
			start("fresh-bk") // unrelated to the blocked process
			time.Sleep(20 * time.Second)
		},
	})
	businessKeys := []string{}
	for _, incident := range result.incidents {
		wrapper := struct {
			BusinessKey string `json:"business_key"`
		}{}
		err := json.Unmarshal([]byte(incident), &wrapper)
		if err != nil {
			t.Fatal(err)
		}
		businessKeys = append(businessKeys, wrapper.BusinessKey)
	}
	if strings.Join(businessKeys, ",") != "warden-bk,warden-bk,fresh-bk" {
		t.Errorf("expected two incidents of the blocked and one of the fresh process, got %v", businessKeys)
	}
	if len(result.errors) != 1 || !strings.Contains(result.errors[0], "restart limit reached") || !strings.Contains(result.errors[0], "warden-bk") {
		t.Errorf("expected one rejected start of warden-bk, got %#v", result.errors)
	}
}

type scriptErrorScenario struct {
	restartLimit int64
	localRestart bool
	businessKey  string
	afterwards   func(start func(businessKey string)) // runs after the first minute, before the results are collected
}

type scriptErrorResult struct {
	notifications []string
	incidents     []string
	errors        []string
	deleteCount   int
}

// runScriptErrorProcess starts a process that fails on every start and collects what the sync client emits
func runScriptErrorProcess(t *testing.T, scenario scriptErrorScenario) (result scriptErrorResult) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	defer wg.Wait()
	defer cancel()

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
	config.CamundaUrl = camundaUrl
	config.CamundaDb = connstr

	_, mqttIp, err := docker.Mqtt(ctx, &wg)
	if err != nil {
		t.Error(err)
		return
	}
	config.MqttBroker = "tcp://" + mqttIp + ":1883"

	config.DeploymentMetadataStorage = t.TempDir() + "/bolt.db"
	config.IncidentRestartLimit = scenario.restartLimit

	mux := sync.Mutex{}

	notificationTestServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		msg, _ := io.ReadAll(request.Body)
		t.Log("notification:", request.URL.String(), string(msg))
		mux.Lock()
		defer mux.Unlock()
		result.notifications = append(result.notifications, string(msg))
	}))
	config.NotificationUrl = notificationTestServer.URL

	_, err = controller.New(config, ctx)
	if err != nil {
		t.Error(err)
		return
	}

	err = docker.TaskWorker(ctx, &wg, config.MqttBroker, config.CamundaUrl, "")
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(5 * time.Second)

	mqttClient := paho.NewClient(paho.NewClientOptions().
		SetAutoReconnect(true).
		SetCleanSession(true).
		AddBroker(config.MqttBroker))
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		t.Error(token.Error())
		return
	}

	deploymentId := ""

	token := mqttClient.Subscribe("#", 2, func(client paho.Client, msg paho.Message) {
		t.Log("mqtt:", msg.Topic(), string(msg.Payload()))
		mux.Lock()
		defer mux.Unlock()
		switch msg.Topic() {
		case "processes/state/deployment":
			wrapper := struct {
				Id string `json:"id"`
			}{}
			err = json.Unmarshal(msg.Payload(), &wrapper)
			if err != nil {
				t.Error(err)
			}
			deploymentId = wrapper.Id
			t.Log("use deploymentId=", deploymentId)
		case "processes/state/process-instance/delete":
			result.deleteCount = result.deleteCount + 1
		case "processes/state/incident":
			result.incidents = append(result.incidents, string(msg.Payload()))
		case "processes/state/error":
			result.errors = append(result.errors, string(msg.Payload()))
		}
	})
	if token.Wait() && token.Error() != nil {
		t.Error(token.Error())
		return
	}

	t.Run("deploy process", func(t *testing.T) {
		pl, err := json.Marshal(model.FogDeploymentMessage{
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
					Restart: scenario.localRestart,
					Notify:  true,
				},
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		token = mqttClient.Publish("processes/cmd/deployment", 2, false, pl)
		if token.Wait() && token.Error() != nil {
			t.Error(token.Error())
			return
		}
	})

	time.Sleep(5 * time.Second)

	start := func(businessKey string) {
		pl, err := json.Marshal(model.StartMessage{
			DeploymentId: deploymentId,
			BusinessKey:  businessKey,
			Parameter:    map[string]interface{}{},
		})
		if err != nil {
			t.Error(err)
			return
		}
		token := mqttClient.Publish("processes/cmd/deployment/start", 2, false, pl)
		if token.Wait() && token.Error() != nil {
			t.Error(token.Error())
			return
		}
	}

	t.Run("start process", func(t *testing.T) {
		start(scenario.businessKey)
	})

	time.Sleep(time.Minute)

	if scenario.afterwards != nil {
		scenario.afterwards(start)
	}

	mux.Lock()
	defer mux.Unlock()
	return scriptErrorResult{
		notifications: append([]string{}, result.notifications...),
		incidents:     append([]string{}, result.incidents...),
		errors:        append([]string{}, result.errors...),
		deleteCount:   result.deleteCount,
	}
}
