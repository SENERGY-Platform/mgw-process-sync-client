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

package helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func DeployProcess(camundUrl string, name string, xml string, svg string, owner string, source string) (deploymentId string, err error) {
	result := map[string]interface{}{}
	boundary := "---------------------------" + time.Now().String()
	b := strings.NewReader(buildPayLoad(name, xml, svg, boundary, owner, source))
	resp, err := http.Post(camundUrl+"/engine-rest/deployment/create", "multipart/form-data; boundary="+boundary, b)
	if err != nil {
		log.Println("ERROR: request to processengine ", err)
		return "", err
	}
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		return "", errors.New(fmt.Sprint("unexpected status-code", resp.StatusCode, string(temp)))
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	deploymentId, ok := result["id"].(string)
	if !ok {
		temp, _ := json.Marshal(result)
		return "", errors.New("missing deployment id in response\n" + string(temp))
	}
	return deploymentId, nil
}

func buildPayLoad(name string, xml string, svg string, boundary string, owner string, deploymentSource string) string {
	segments := []string{}
	if deploymentSource == "" {
		deploymentSource = "sepl"
	}

	segments = append(segments, "Content-Disposition: form-data; name=\"data\"; "+"filename=\""+name+".bpmn\"\r\nContent-Type: text/xml\r\n\r\n"+xml+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"diagram\"; "+"filename=\""+name+".svg\"\r\nContent-Type: image/svg+xml\r\n\r\n"+svg+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"deployment-name\"\r\n\r\n"+name+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"deployment-source\"\r\n\r\n"+deploymentSource+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"tenant-id\"\r\n\r\n"+owner+"\r\n")

	return "--" + boundary + "\r\n" + strings.Join(segments, "--"+boundary+"\r\n") + "--" + boundary + "--\r\n"
}

func StartProcess(camundUrl string, processDefinitionId string) (err error) {
	message := map[string]interface{}{}

	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(message)
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", camundUrl+"/engine-rest/process-definition/"+url.QueryEscape(processDefinitionId)+"/submit-form", b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		temp, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(resp.Status + " " + string(temp))
		return
	}
	return nil
}

func RemoveHistory(camundUrl string, instanceId string) (err error) {
	req, err := http.NewRequest("DELETE", camundUrl+"/engine-rest/history/process-instance/"+url.QueryEscape(instanceId), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		temp, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(resp.Status + " " + string(temp))
		return
	}
	return nil
}

func RemoveDeployment(camundUrl string, instanceId string) (err error) {
	req, err := http.NewRequest("DELETE", camundUrl+"/engine-rest/deployment/"+url.QueryEscape(instanceId)+"?cascade=true&skipIoMappings=true", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		temp, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(resp.Status + " " + string(temp))
		return
	}
	return nil
}

const BpmnExample = `<bpmn:definitions xmlns:xsi='http://www.w3.org/2001/XMLSchema-instance' xmlns:bpmn='http://www.omg.org/spec/BPMN/20100524/MODEL' xmlns:bpmndi='http://www.omg.org/spec/BPMN/20100524/DI' xmlns:dc='http://www.omg.org/spec/DD/20100524/DC' xmlns:di='http://www.omg.org/spec/DD/20100524/DI' id='Definitions_1' targetNamespace='http://bpmn.io/schema/bpmn'><bpmn:process id='Process_1' isExecutable='true'><bpmn:startEvent id='StartEvent_1'><bpmn:outgoing>SequenceFlow_02ibfc0</bpmn:outgoing></bpmn:startEvent><bpmn:endEvent id='EndEvent_1728wjv'><bpmn:incoming>SequenceFlow_02ibfc0</bpmn:incoming></bpmn:endEvent><bpmn:sequenceFlow id='SequenceFlow_02ibfc0' sourceRef='StartEvent_1' targetRef='EndEvent_1728wjv'/></bpmn:process><bpmndi:BPMNDiagram id='BPMNDiagram_1'><bpmndi:BPMNPlane id='BPMNPlane_1' bpmnElement='Process_1'><bpmndi:BPMNShape id='_BPMNShape_StartEvent_2' bpmnElement='StartEvent_1'><dc:Bounds x='173' y='102' width='36' height='36'/></bpmndi:BPMNShape><bpmndi:BPMNShape id='EndEvent_1728wjv_di' bpmnElement='EndEvent_1728wjv'><dc:Bounds x='259' y='102' width='36' height='36'/></bpmndi:BPMNShape><bpmndi:BPMNEdge id='SequenceFlow_02ibfc0_di' bpmnElement='SequenceFlow_02ibfc0'><di:waypoint x='209' y='120'></di:waypoint><di:waypoint x='259' y='120'></di:waypoint></bpmndi:BPMNEdge></bpmndi:BPMNPlane></bpmndi:BPMNDiagram></bpmn:definitions>`
const SvgExample = `<svg height='48' version='1.1' viewBox='167 96 134 48' width='134' xmlns='http://www.w3.org/2000/svg' xmlns:xlink='http://www.w3.org/1999/xlink'><defs><marker id='sequenceflow-end-white-black-3gh21e50i1p8scvqmvrotmp9p' markerHeight='10' markerWidth='10' orient='auto' refX='11' refY='10' viewBox='0 0 20 20'><path d='M 1 5 L 11 10 L 1 15 Z' style='fill: black; stroke-width: 1px; stroke-linecap: round; stroke-dasharray: 10000, 1; stroke: black;'/></marker></defs><g class='djs-group'><g class='djs-element djs-connection' data-element-id='SequenceFlow_04zz9eb' style='display: block;'><g class='djs-visual'><path d='m  209,120L259,120 ' style='fill: none; stroke-width: 2px; stroke: black; stroke-linejoin: round; marker-end: url(&apos;#sequenceflow-end-white-black-3gh21e50i1p8scvqmvrotmp9p&apos;);'/></g><polyline class='djs-hit' points='209,120 259,120 ' style='fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;'/><rect class='djs-outline' height='12' style='fill: none;' width='62' x='203' y='114'/></g></g><g class='djs-group'><g class='djs-element djs-shape' data-element-id='StartEvent_1' style='display: block;' transform='translate(173 102)'><g class='djs-visual'><circle cx='18' cy='18' r='18' style='stroke: black; stroke-width: 2px; fill: white; fill-opacity: 0.95;'/><path d='m 8.459999999999999,11.34 l 0,12.6 l 18.900000000000002,0 l 0,-12.6 z l 9.450000000000001,5.4 l 9.450000000000001,-5.4' style='fill: white; stroke-width: 1px; stroke: black;'/></g><rect class='djs-hit' height='36' style='fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;' width='36' x='0' y='0'></rect><rect class='djs-outline' height='48' style='fill: none;' width='48' x='-6' y='-6'></rect></g></g><g class='djs-group'><g class='djs-element djs-shape' data-element-id='EndEvent_056p30q' style='display: block;' transform='translate(259 102)'><g class='djs-visual'><circle cx='18' cy='18' r='18' style='stroke: black; stroke-width: 4px; fill: white; fill-opacity: 0.95;'/></g><rect class='djs-hit' height='36' style='fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;' width='36' x='0' y='0'></rect><rect class='djs-outline' height='48' style='fill: none;' width='48' x='-6' y='-6'></rect></g></g></svg>`
const BpmnWithScriptExample = `<bpmn:definitions xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI" xmlns:dc="http://www.omg.org/spec/DD/20100524/DC" xmlns:di="http://www.omg.org/spec/DD/20100524/DI" id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn"><bpmn:process id="scipttest" isExecutable="true"><bpmn:startEvent id="StartEvent_1"><bpmn:outgoing>SequenceFlow_08boegg</bpmn:outgoing></bpmn:startEvent><bpmn:sequenceFlow id="SequenceFlow_08boegg" sourceRef="StartEvent_1" targetRef="Task_0d9kudy" /><bpmn:endEvent id="EndEvent_1lkf4gg"><bpmn:incoming>SequenceFlow_1x0w4tg</bpmn:incoming></bpmn:endEvent><bpmn:sequenceFlow id="SequenceFlow_1x0w4tg" sourceRef="Task_0d9kudy" targetRef="EndEvent_1lkf4gg" /><bpmn:scriptTask id="Task_0d9kudy" scriptFormat="JavaScript"><bpmn:incoming>SequenceFlow_08boegg</bpmn:incoming><bpmn:outgoing>SequenceFlow_1x0w4tg</bpmn:outgoing><bpmn:script>var foo = "bar"</bpmn:script></bpmn:scriptTask></bpmn:process><bpmndi:BPMNDiagram id="BPMNDiagram_1"><bpmndi:BPMNPlane id="BPMNPlane_1" bpmnElement="scipttest"><bpmndi:BPMNShape id="_BPMNShape_StartEvent_2" bpmnElement="StartEvent_1"><dc:Bounds x="173" y="102" width="36" height="36" /></bpmndi:BPMNShape><bpmndi:BPMNEdge id="SequenceFlow_08boegg_di" bpmnElement="SequenceFlow_08boegg"><di:waypoint x="209" y="120" /><di:waypoint x="260" y="120" /></bpmndi:BPMNEdge><bpmndi:BPMNShape id="EndEvent_1lkf4gg_di" bpmnElement="EndEvent_1lkf4gg"><dc:Bounds x="412" y="102" width="36" height="36" /></bpmndi:BPMNShape><bpmndi:BPMNEdge id="SequenceFlow_1x0w4tg_di" bpmnElement="SequenceFlow_1x0w4tg"><di:waypoint x="360" y="120" /><di:waypoint x="412" y="120" /></bpmndi:BPMNEdge><bpmndi:BPMNShape id="ScriptTask_1hy7qhu_di" bpmnElement="Task_0d9kudy"><dc:Bounds x="260" y="80" width="100" height="80" /></bpmndi:BPMNShape></bpmndi:BPMNPlane></bpmndi:BPMNDiagram></bpmn:definitions>`
const BPMNWithTasksExample = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI" xmlns:dc="http://www.omg.org/spec/DD/20100524/DC" xmlns:camunda="http://camunda.org/schema/1.0/bpmn" xmlns:senergy="https://senergy.infai.org" xmlns:di="http://www.omg.org/spec/DD/20100524/DI" id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
    <bpmn:process id="ExampleId" name="ExampleName" isExecutable="true" senergy:description="ExampleDesc">
        <bpmn:startEvent id="StartEvent_1">
            <bpmn:outgoing>SequenceFlow_0qjn3dq</bpmn:outgoing>
        </bpmn:startEvent>
        <bpmn:sequenceFlow id="SequenceFlow_0qjn3dq" sourceRef="StartEvent_1" targetRef="Task_1nbnl8y"/>
        <bpmn:sequenceFlow id="SequenceFlow_15v8030" sourceRef="Task_1nbnl8y" targetRef="Task_1lhzy95"/>
        <bpmn:endEvent id="EndEvent_1vhaxdr">
            <bpmn:incoming>SequenceFlow_17lypcn</bpmn:incoming>
        </bpmn:endEvent>
        <bpmn:sequenceFlow id="SequenceFlow_17lypcn" sourceRef="Task_1lhzy95" targetRef="EndEvent_1vhaxdr"/>
        <bpmn:serviceTask id="Task_1nbnl8y" name="Lighting getColorFunction" camunda:type="external" camunda:topic="pessimistic">
            <bpmn:extensionElements>
                <camunda:inputOutput>
                    <camunda:inputParameter name="payload"><![CDATA[{
	"function": {
		"id": "urn:infai:ses:measuring-function:bdb6a7c8-4a3d-4fe0-bab3-ce02e09b5869",
		"name": "",
		"concept_id": "",
		"rdf_type": ""
	},
	"characteristic_id": "urn:infai:ses:characteristic:5b4eea52-e8e5-4e80-9455-0382f81a1b43",
	"aspect": {
		"id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6",
		"name": "",
		"rdf_type": ""
	},
	"device_id": "hue",
	"service_id": "urn:infai:ses:service:99614933-4734-41b6-a131-3f96f134ee69",
	"input": {},
	"retries": 2
}]]></camunda:inputParameter>
                    <camunda:outputParameter name="outputs.b">${result.b}</camunda:outputParameter>
                    <camunda:outputParameter name="outputs.g">${result.g}</camunda:outputParameter>
                    <camunda:outputParameter name="outputs.r">${result.r}</camunda:outputParameter>
                </camunda:inputOutput>
            </bpmn:extensionElements>
            <bpmn:incoming>SequenceFlow_0qjn3dq</bpmn:incoming>
            <bpmn:outgoing>SequenceFlow_15v8030</bpmn:outgoing>
        </bpmn:serviceTask>
        <bpmn:serviceTask id="Task_1lhzy95" name="Lamp setColorFunction" camunda:type="external" camunda:topic="optimistic">
            <bpmn:extensionElements>
                <camunda:inputOutput>
                    <camunda:inputParameter name="payload"><![CDATA[{
	"function": {
		"id": "urn:infai:ses:controlling-function:c54e2a89-1fb8-4ecb-8993-a7b40b355599",
		"name": "",
		"concept_id": "",
		"rdf_type": ""
	},
	"characteristic_id": "urn:infai:ses:characteristic:5b4eea52-e8e5-4e80-9455-0382f81a1b43",
	"device_class": {
		"id": "urn:infai:ses:device-class:14e56881-16f9-4120-bb41-270a43070c86",
		"name": "",
		"image": "",
		"rdf_type": ""
	},
	"device_id": "hue",
	"service_id": "urn:infai:ses:service:67789396-d1ca-4ea9-9147-0614c6d68a2f",
	"input": {
		"b": 0,
		"g": 0,
		"r": 0
	}
}]]></camunda:inputParameter>
                    <camunda:inputParameter name="inputs.b">0</camunda:inputParameter>
                    <camunda:inputParameter name="inputs.g">255</camunda:inputParameter>
                    <camunda:inputParameter name="inputs.r">100</camunda:inputParameter>
                </camunda:inputOutput>
            </bpmn:extensionElements>
            <bpmn:incoming>SequenceFlow_15v8030</bpmn:incoming>
            <bpmn:outgoing>SequenceFlow_17lypcn</bpmn:outgoing>
        </bpmn:serviceTask>
    </bpmn:process>
    <bpmndi:BPMNDiagram id="BPMNDiagram_1">
        <bpmndi:BPMNPlane id="BPMNPlane_1" bpmnElement="ExampleId">
            <bpmndi:BPMNShape id="_BPMNShape_StartEvent_2" bpmnElement="StartEvent_1">
                <dc:Bounds x="173" y="102" width="36" height="36"/>
            </bpmndi:BPMNShape>
            <bpmndi:BPMNEdge id="SequenceFlow_0qjn3dq_di" bpmnElement="SequenceFlow_0qjn3dq">
                <di:waypoint x="209" y="120"/>
                <di:waypoint x="260" y="120"/>
            </bpmndi:BPMNEdge>
            <bpmndi:BPMNEdge id="SequenceFlow_15v8030_di" bpmnElement="SequenceFlow_15v8030">
                <di:waypoint x="360" y="120"/>
                <di:waypoint x="420" y="120"/>
            </bpmndi:BPMNEdge>
            <bpmndi:BPMNShape id="EndEvent_1vhaxdr_di" bpmnElement="EndEvent_1vhaxdr">
                <dc:Bounds x="582" y="102" width="36" height="36"/>
            </bpmndi:BPMNShape>
            <bpmndi:BPMNEdge id="SequenceFlow_17lypcn_di" bpmnElement="SequenceFlow_17lypcn">
                <di:waypoint x="520" y="120"/>
                <di:waypoint x="582" y="120"/>
            </bpmndi:BPMNEdge>
            <bpmndi:BPMNShape id="ServiceTask_04qer1y_di" bpmnElement="Task_1nbnl8y">
                <dc:Bounds x="260" y="80" width="100" height="80"/>
            </bpmndi:BPMNShape>
            <bpmndi:BPMNShape id="ServiceTask_0o8bz5b_di" bpmnElement="Task_1lhzy95">
                <dc:Bounds x="420" y="80" width="100" height="80"/>
            </bpmndi:BPMNShape>
        </bpmndi:BPMNPlane>
    </bpmndi:BPMNDiagram>
</bpmn:definitions>`
