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

package analytics

import "time"

type ControlCommand struct {
	Command string      `json:"command,omitempty"`
	Data    OperatorJob `json:"data,omitempty"`
}

type OperatorJob struct {
	ImageId         string            `json:"imageId,omitempty"`
	Agent           Agent             `json:"agent,omitempty"`
	OperatorConfig  map[string]string `json:"operatorConfig,omitempty"`
	ContainerId     string            `json:"containerId,omitempty"`
	InputTopics     []InputTopic      `json:"inputTopics,omitempty"`
	Config          FogConfig         `json:"config,omitempty"`
	Response        string            `json:"response,omitempty"`
	ResponseMessage string            `json:"responseMessage,omitempty"`
}

type FogConfig struct {
	PipelineId     string `json:"pipelineId,omitempty"`
	OutputTopic    string `json:"outputTopic,omitempty"`
	OperatorId     string `json:"operatorId,omitempty"`
	BaseOperatorId string `json:"baseOperatorId,omitempty"`
}

type InputTopic struct {
	Name        string    `json:"name,omitempty"`
	FilterType  string    `json:"filterType,omitempty"`
	FilterValue string    `json:"filterValue,omitempty"`
	Mappings    []Mapping `json:"mappings,omitempty"`
}

type Mapping struct {
	Dest   string `json:"dest,omitempty"`
	Source string `json:"source,omitempty"`
}

type AgentMessage struct {
	Type string `json:"type,omitempty"`
	Conf Agent  `json:"agent,omitempty"`
}

type Agent struct {
	Id      string    `json:"id,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
	Active  bool      `json:"active,omitempty"`
}
