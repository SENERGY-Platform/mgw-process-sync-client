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
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/deploymentmodel"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/server"
	paho "github.com/eclipse/paho.mqtt.golang"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGroupConvertEventDeployment(t *testing.T) {
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
		token := mqtt.Publish("processes/"+conf.NetworkId+"/cmd/deployment", 2, false, deploymentWithGroupConvertMsgEvents)
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

	t.Run("check mqtt messages", func(t *testing.T) {
		actual := []string{}
		for _, msg := range mqttMessages["fog/control"] {
			value := strings.ReplaceAll(msg, conf.CamundaUrl, "http://camundaurl")
			value = strings.ReplaceAll(value, deploymentId, "deploymentid")
			actual = append(actual, value)
		}
		expected := []string{`{"command":"startOperator","data":{"imageId":"1","agent":{"updated":"0001-01-01T00:00:00Z"},"operatorConfig":{"convertFrom":"cid2","convertTo":"cid1","eventId":"1","groupConvertFrom":"","url":"http://camundaurl/engine-rest/message","value":"1"},"inputTopics":[{"name":"event/ldid1/lsid1","filterType":"filtertype_placeholder","filterValue":"filtervalue_placeholder","mappings":[{"dest":"value","source":"value.path.to.chid2"}]}],"config":{"pipelineId":"deploymentid","outputTopic":"event-trigger","operatorId":"deploymentid_1"}}}`,
			`{"command":"startOperator","data":{"imageId":"1-group","agent":{"updated":"0001-01-01T00:00:00Z"},"operatorConfig":{"convertFrom":"","convertTo":"to_characteristic_id_1","eventId":"1-group","groupConvertFrom":"{\"event/ldid1/lsid1::value.path.to.chid2\":\"cid2\",\"event/ldid1/lsid2::path.to.chid3\":\"cid3\",\"event/ldid2/lsid1::value.path.to.chid2\":\"cid2\",\"event/ldid2/lsid2::path.to.chid3\":\"cid3\"}","url":"http://camundaurl/engine-rest/message","value":"1-group"},"inputTopics":[{"name":"event/ldid1/lsid1","filterType":"filtertype_placeholder","filterValue":"filtervalue_placeholder","mappings":[{"dest":"value","source":"value.path.to.chid2"}]},{"name":"event/ldid2/lsid1","filterType":"filtertype_placeholder","filterValue":"filtervalue_placeholder","mappings":[{"dest":"value","source":"value.path.to.chid2"}]},{"name":"event/ldid1/lsid2","filterType":"filtertype_placeholder","filterValue":"filtervalue_placeholder","mappings":[{"dest":"value","source":"path.to.chid3"}]},{"name":"event/ldid2/lsid2","filterType":"filtertype_placeholder","filterValue":"filtervalue_placeholder","mappings":[{"dest":"value","source":"path.to.chid3"}]}],"config":{"pipelineId":"deploymentid","outputTopic":"event-trigger","operatorId":"deploymentid_1-group"}}}`}
		sort.Strings(actual)
		sort.Strings(expected)
		if !reflect.DeepEqual(actual, expected) {
			t.Error(actual)
			return
		}
	})

}

const deploymentWithGroupConvertMsgEvents = `{
   "id":"",
   "name":"test-deployment-name",
   "description":"test-description",
   "diagram":{
      "xml_raw":"\u003c?xml version=\"1.0\" encoding=\"UTF-8\"?\u003e\n\u003cbpmn:definitions xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlns:bpmn=\"http://www.omg.org/spec/BPMN/20100524/MODEL\" xmlns:bpmndi=\"http://www.omg.org/spec/BPMN/20100524/DI\" xmlns:dc=\"http://www.omg.org/spec/DD/20100524/DC\" xmlns:camunda=\"http://camunda.org/schema/1.0/bpmn\" xmlns:senergy=\"https://senergy.infai.org\" xmlns:di=\"http://www.omg.org/spec/DD/20100524/DI\" id=\"Definitions_1\" targetNamespace=\"http://bpmn.io/schema/bpmn\"\u003e\n    \u003cbpmn:process id=\"ExampleId\" name=\"ExampleName\" isExecutable=\"true\" senergy:description=\"ExampleDesc\"\u003e\n        \u003cbpmn:startEvent id=\"StartEvent_1\"\u003e\n            \u003cbpmn:outgoing\u003eSequenceFlow_0qjn3dq\u003c/bpmn:outgoing\u003e\n        \u003c/bpmn:startEvent\u003e\n        \u003cbpmn:sequenceFlow id=\"SequenceFlow_0qjn3dq\" sourceRef=\"StartEvent_1\" targetRef=\"Task_1nbnl8y\"/\u003e\n        \u003cbpmn:sequenceFlow id=\"SequenceFlow_15v8030\" sourceRef=\"Task_1nbnl8y\" targetRef=\"Task_1lhzy95\"/\u003e\n        \u003cbpmn:endEvent id=\"EndEvent_1vhaxdr\"\u003e\n            \u003cbpmn:incoming\u003eSequenceFlow_17lypcn\u003c/bpmn:incoming\u003e\n        \u003c/bpmn:endEvent\u003e\n        \u003cbpmn:sequenceFlow id=\"SequenceFlow_17lypcn\" sourceRef=\"Task_1lhzy95\" targetRef=\"EndEvent_1vhaxdr\"/\u003e\n        \u003cbpmn:serviceTask id=\"Task_1nbnl8y\" name=\"Lighting getColorFunction\" camunda:type=\"external\" camunda:topic=\"pessimistic\"\u003e\n            \u003cbpmn:extensionElements\u003e\n                \u003ccamunda:inputOutput\u003e\n                    \u003ccamunda:inputParameter name=\"payload\"\u003e\u003c![CDATA[{\n\t\"function\": {\n\t\t\"id\": \"urn:infai:ses:measuring-function:bdb6a7c8-4a3d-4fe0-bab3-ce02e09b5869\",\n\t\t\"name\": \"\",\n\t\t\"concept_id\": \"\",\n\t\t\"rdf_type\": \"\"\n\t},\n\t\"characteristic_id\": \"urn:infai:ses:characteristic:5b4eea52-e8e5-4e80-9455-0382f81a1b43\",\n\t\"aspect\": {\n\t\t\"id\": \"urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6\",\n\t\t\"name\": \"\",\n\t\t\"rdf_type\": \"\"\n\t},\n\t\"device_id\": \"hue\",\n\t\"service_id\": \"urn:infai:ses:service:99614933-4734-41b6-a131-3f96f134ee69\",\n\t\"input\": {},\n\t\"retries\": 2\n}]]\u003e\u003c/camunda:inputParameter\u003e\n                    \u003ccamunda:outputParameter name=\"outputs.b\"\u003e${result.b}\u003c/camunda:outputParameter\u003e\n                    \u003ccamunda:outputParameter name=\"outputs.g\"\u003e${result.g}\u003c/camunda:outputParameter\u003e\n                    \u003ccamunda:outputParameter name=\"outputs.r\"\u003e${result.r}\u003c/camunda:outputParameter\u003e\n                \u003c/camunda:inputOutput\u003e\n            \u003c/bpmn:extensionElements\u003e\n            \u003cbpmn:incoming\u003eSequenceFlow_0qjn3dq\u003c/bpmn:incoming\u003e\n            \u003cbpmn:outgoing\u003eSequenceFlow_15v8030\u003c/bpmn:outgoing\u003e\n        \u003c/bpmn:serviceTask\u003e\n        \u003cbpmn:serviceTask id=\"Task_1lhzy95\" name=\"Lamp setColorFunction\" camunda:type=\"external\" camunda:topic=\"optimistic\"\u003e\n            \u003cbpmn:extensionElements\u003e\n                \u003ccamunda:inputOutput\u003e\n                    \u003ccamunda:inputParameter name=\"payload\"\u003e\u003c![CDATA[{\n\t\"function\": {\n\t\t\"id\": \"urn:infai:ses:controlling-function:c54e2a89-1fb8-4ecb-8993-a7b40b355599\",\n\t\t\"name\": \"\",\n\t\t\"concept_id\": \"\",\n\t\t\"rdf_type\": \"\"\n\t},\n\t\"characteristic_id\": \"urn:infai:ses:characteristic:5b4eea52-e8e5-4e80-9455-0382f81a1b43\",\n\t\"device_class\": {\n\t\t\"id\": \"urn:infai:ses:device-class:14e56881-16f9-4120-bb41-270a43070c86\",\n\t\t\"name\": \"\",\n\t\t\"image\": \"\",\n\t\t\"rdf_type\": \"\"\n\t},\n\t\"device_id\": \"hue\",\n\t\"service_id\": \"urn:infai:ses:service:67789396-d1ca-4ea9-9147-0614c6d68a2f\",\n\t\"input\": {\n\t\t\"b\": 0,\n\t\t\"g\": 0,\n\t\t\"r\": 0\n\t}\n}]]\u003e\u003c/camunda:inputParameter\u003e\n                    \u003ccamunda:inputParameter name=\"inputs.b\"\u003e0\u003c/camunda:inputParameter\u003e\n                    \u003ccamunda:inputParameter name=\"inputs.g\"\u003e255\u003c/camunda:inputParameter\u003e\n                    \u003ccamunda:inputParameter name=\"inputs.r\"\u003e100\u003c/camunda:inputParameter\u003e\n                \u003c/camunda:inputOutput\u003e\n            \u003c/bpmn:extensionElements\u003e\n            \u003cbpmn:incoming\u003eSequenceFlow_15v8030\u003c/bpmn:incoming\u003e\n            \u003cbpmn:outgoing\u003eSequenceFlow_17lypcn\u003c/bpmn:outgoing\u003e\n        \u003c/bpmn:serviceTask\u003e\n    \u003c/bpmn:process\u003e\n\u003c/bpmn:definitions\u003e",
      "xml_deployed":"\u003c?xml version=\"1.0\" encoding=\"UTF-8\"?\u003e\n\u003cbpmn:definitions xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlns:bpmn=\"http://www.omg.org/spec/BPMN/20100524/MODEL\" xmlns:bpmndi=\"http://www.omg.org/spec/BPMN/20100524/DI\" xmlns:dc=\"http://www.omg.org/spec/DD/20100524/DC\" xmlns:camunda=\"http://camunda.org/schema/1.0/bpmn\" xmlns:senergy=\"https://senergy.infai.org\" xmlns:di=\"http://www.omg.org/spec/DD/20100524/DI\" id=\"Definitions_1\" targetNamespace=\"http://bpmn.io/schema/bpmn\"\u003e\n    \u003cbpmn:process id=\"ExampleId\" name=\"ExampleName\" isExecutable=\"true\" senergy:description=\"ExampleDesc\"\u003e\n        \u003cbpmn:startEvent id=\"StartEvent_1\"\u003e\n            \u003cbpmn:outgoing\u003eSequenceFlow_0qjn3dq\u003c/bpmn:outgoing\u003e\n        \u003c/bpmn:startEvent\u003e\n        \u003cbpmn:sequenceFlow id=\"SequenceFlow_0qjn3dq\" sourceRef=\"StartEvent_1\" targetRef=\"Task_1nbnl8y\"/\u003e\n        \u003cbpmn:sequenceFlow id=\"SequenceFlow_15v8030\" sourceRef=\"Task_1nbnl8y\" targetRef=\"Task_1lhzy95\"/\u003e\n        \u003cbpmn:endEvent id=\"EndEvent_1vhaxdr\"\u003e\n            \u003cbpmn:incoming\u003eSequenceFlow_17lypcn\u003c/bpmn:incoming\u003e\n        \u003c/bpmn:endEvent\u003e\n        \u003cbpmn:sequenceFlow id=\"SequenceFlow_17lypcn\" sourceRef=\"Task_1lhzy95\" targetRef=\"EndEvent_1vhaxdr\"/\u003e\n        \u003cbpmn:serviceTask id=\"Task_1nbnl8y\" name=\"Lighting getColorFunction\" camunda:type=\"external\" camunda:topic=\"pessimistic\"\u003e\n            \u003cbpmn:extensionElements\u003e\n                \u003ccamunda:inputOutput\u003e\n                    \u003ccamunda:inputParameter name=\"payload\"\u003e\u003c![CDATA[{\n\t\"function\": {\n\t\t\"id\": \"urn:infai:ses:measuring-function:bdb6a7c8-4a3d-4fe0-bab3-ce02e09b5869\",\n\t\t\"name\": \"\",\n\t\t\"concept_id\": \"\",\n\t\t\"rdf_type\": \"\"\n\t},\n\t\"characteristic_id\": \"urn:infai:ses:characteristic:5b4eea52-e8e5-4e80-9455-0382f81a1b43\",\n\t\"aspect\": {\n\t\t\"id\": \"urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6\",\n\t\t\"name\": \"\",\n\t\t\"rdf_type\": \"\"\n\t},\n\t\"device_id\": \"hue\",\n\t\"service_id\": \"urn:infai:ses:service:99614933-4734-41b6-a131-3f96f134ee69\",\n\t\"input\": {},\n\t\"retries\": 2\n}]]\u003e\u003c/camunda:inputParameter\u003e\n                    \u003ccamunda:outputParameter name=\"outputs.b\"\u003e${result.b}\u003c/camunda:outputParameter\u003e\n                    \u003ccamunda:outputParameter name=\"outputs.g\"\u003e${result.g}\u003c/camunda:outputParameter\u003e\n                    \u003ccamunda:outputParameter name=\"outputs.r\"\u003e${result.r}\u003c/camunda:outputParameter\u003e\n                \u003c/camunda:inputOutput\u003e\n            \u003c/bpmn:extensionElements\u003e\n            \u003cbpmn:incoming\u003eSequenceFlow_0qjn3dq\u003c/bpmn:incoming\u003e\n            \u003cbpmn:outgoing\u003eSequenceFlow_15v8030\u003c/bpmn:outgoing\u003e\n        \u003c/bpmn:serviceTask\u003e\n        \u003cbpmn:serviceTask id=\"Task_1lhzy95\" name=\"Lamp setColorFunction\" camunda:type=\"external\" camunda:topic=\"optimistic\"\u003e\n            \u003cbpmn:extensionElements\u003e\n                \u003ccamunda:inputOutput\u003e\n                    \u003ccamunda:inputParameter name=\"payload\"\u003e\u003c![CDATA[{\n\t\"function\": {\n\t\t\"id\": \"urn:infai:ses:controlling-function:c54e2a89-1fb8-4ecb-8993-a7b40b355599\",\n\t\t\"name\": \"\",\n\t\t\"concept_id\": \"\",\n\t\t\"rdf_type\": \"\"\n\t},\n\t\"characteristic_id\": \"urn:infai:ses:characteristic:5b4eea52-e8e5-4e80-9455-0382f81a1b43\",\n\t\"device_class\": {\n\t\t\"id\": \"urn:infai:ses:device-class:14e56881-16f9-4120-bb41-270a43070c86\",\n\t\t\"name\": \"\",\n\t\t\"image\": \"\",\n\t\t\"rdf_type\": \"\"\n\t},\n\t\"device_id\": \"hue\",\n\t\"service_id\": \"urn:infai:ses:service:67789396-d1ca-4ea9-9147-0614c6d68a2f\",\n\t\"input\": {\n\t\t\"b\": 0,\n\t\t\"g\": 0,\n\t\t\"r\": 0\n\t}\n}]]\u003e\u003c/camunda:inputParameter\u003e\n                    \u003ccamunda:inputParameter name=\"inputs.b\"\u003e0\u003c/camunda:inputParameter\u003e\n                    \u003ccamunda:inputParameter name=\"inputs.g\"\u003e255\u003c/camunda:inputParameter\u003e\n                    \u003ccamunda:inputParameter name=\"inputs.r\"\u003e100\u003c/camunda:inputParameter\u003e\n                \u003c/camunda:inputOutput\u003e\n            \u003c/bpmn:extensionElements\u003e\n            \u003cbpmn:incoming\u003eSequenceFlow_15v8030\u003c/bpmn:incoming\u003e\n            \u003cbpmn:outgoing\u003eSequenceFlow_17lypcn\u003c/bpmn:outgoing\u003e\n        \u003c/bpmn:serviceTask\u003e\n    \u003c/bpmn:process\u003e\n\u003c/bpmn:definitions\u003e",
      "svg":"\u003csvg\u003e\u003c/svg\u003e"
   },
   "elements":[
      {
         "bpmn_id":"bpmnid",
         "group":null,
         "name":"event-name",
         "order":0,
         "time_event":null,
         "notification":null,
         "message_event":{
            "value":"1",
            "flow_id":"1",
            "event_id":"1",
            "selection":{
               "filter_criteria":{
                  "characteristic_id":"cid1",
                  "function_id":"fid1",
                  "device_class_id":null,
                  "aspect_id":"aid1"
               },
               "selection_options":null,
               "selected_device_id":"did1",
               "selected_service_id":"sid1",
               "selected_device_group_id":null
            }
         },
         "task":null
      },
      {
         "bpmn_id":"bpmnid-group",
         "group":null,
         "name":"event-name-group",
         "order":0,
         "time_event":null,
         "notification":null,
         "message_event":{
            "value":"1-group",
            "flow_id":"1-group",
            "event_id":"1-group",
            "selection":{
               "filter_criteria":{
                  "characteristic_id":"cid1",
                  "function_id":"fid1",
                  "device_class_id":null,
                  "aspect_id":"aid1"
               },
               "selection_options":null,
               "selected_device_id":null,
               "selected_service_id":null,
               "selected_device_group_id":"gid1"
            }
         },
         "task":null
      }
   ],
   "executable":true,
   "analytics_records":[
      {
         "device_event":{
            "label":"event-name (1)",
            "deployment_id":"placeholder",
            "flow_id":"1",
            "event_id":"1",
            "device_id":"did1",
            "service_id":"sid1",
            "value":"1",
            "path":"value.path.to.chid2",
            "cast_from":"cid2",
            "cast_to":"cid1"
         },
         "group_event":null
      },
      {
         "device_event":null,
         "group_event":{
            "label":"event-name-group (1-group)",
            "desc":{
               "ImportId":"",
               "Path":"",
               "DeviceGroupId":"gid1",
               "DeviceIds":null,
               "EventId":"1-group",
               "DeploymentId":"placeholder",
               "FunctionId":"fid1",
               "AspectId":"aid1",
               "FlowId":"1-group",
               "OperatorValue":"1-group",
               "CharacteristicId":"to_characteristic_id_1"
            },
            "service_ids":[
               "sid1",
               "sid2"
            ],
            "service_to_device_ids_mapping":{
               "sid1":[
                  "did1",
                  "did2"
               ],
               "sid2":[
                  "did1",
                  "did2"
               ]
            },
            "service_to_path_mapping":{
               "sid1":"value.path.to.chid2",
               "sid2":"path.to.chid3"
            },
            "service_to_path_and_characteristic":{
               "sid1":[
                  {
                     "json_path":"value.path.to.chid2",
                     "characteristic_id":"cid2"
                  }
               ],
               "sid2":[
                  {
                     "json_path":"path.to.chid3",
                     "characteristic_id":"cid3"
                  }
               ]
            }
         }
      }
   ],
   "device_id_to_local_id":{
      "did1":"ldid1",
      "did2":"ldid2"
   },
   "service_id_to_local_id":{
      "sid1":"lsid1",
      "sid2":"lsid2"
   }
}`
