/*
 * Copyright 2020 InfAI (CC SES)
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

package deploymentmodel

const CurrentVersion int64 = 3

type Deployment struct {
	Version     int64     `json:"version"`
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Diagram     Diagram   `json:"diagram"`
	Elements    []Element `json:"elements"`
	Executable  bool      `json:"executable"`
}

type Diagram struct {
	XmlRaw      string `json:"xml_raw"`
	XmlDeployed string `json:"xml_deployed"`
	Svg         string `json:"svg"`
}

type Element struct {
	BpmnId       string        `json:"bpmn_id"`
	Group        *string       `json:"group"`
	Name         string        `json:"name"`
	Order        int64         `json:"order"`
	TimeEvent    *TimeEvent    `json:"time_event"`
	Notification *Notification `json:"notification"`
	MessageEvent *MessageEvent `json:"message_event"`
	Task         *Task         `json:"task"`
}

type TimeEvent struct {
	Type string `json:"type"`
	Time string `json:"time"`
}

type Notification struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type MessageEvent struct {
	Value     string    `json:"value"`
	FlowId    string    `json:"flow_id"`
	EventId   string    `json:"event_id"`
	Selection Selection `json:"selection"`
}

type Task struct {
	Retries   int64             `json:"retries"`
	Parameter map[string]string `json:"parameter"`
	Selection Selection         `json:"selection"`
}
