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

package backend

import (
	"encoding/json"
	eventmodel "github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/metadata"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/camundamodel"
	paho "github.com/eclipse/paho.mqtt.golang"
)

const deploymentTopic = "deployment"

func (this *Client) getDeploymentTopic() string {
	return this.getCommandTopic(deploymentTopic)
}

func (this *Client) handleDeploymentCommand(message paho.Message) {
	deployment := model.FogDeploymentMessage{}
	err := json.Unmarshal(message.Payload(), &deployment)
	if err != nil {
		this.error(err)
	}
	_, err = this.handler.CreateDeployment(deployment)
	if err != nil {
		this.error(err)
	}
}

type EventDescriptionsUpdate struct {
	CamundaDeploymentId string                 `json:"camunda_deployment_id"`
	EventDescriptions   []eventmodel.EventDesc `json:"event_descriptions"`
	DeviceIdToLocalId   map[string]string      `json:"device_id_to_local_id"`
	ServiceIdToLocalId  map[string]string      `json:"service_id_to_local_id"`
}

func (this *Client) getProcessEventUpdateTopic() string {
	return this.getCommandTopic(deploymentTopic, "event-descriptions")
}

func (this *Client) handleEventUpdateCommand(message paho.Message) {
	msg := EventDescriptionsUpdate{}
	err := json.Unmarshal(message.Payload(), &msg)
	if err != nil {
		this.error(err)
	}
	err = this.handler.UpdateDeploymentEvents(msg.CamundaDeploymentId, msg.EventDescriptions, msg.DeviceIdToLocalId, msg.DeviceIdToLocalId)
	if err != nil {
		this.error(err)
	}
}

func (this *Client) getProcessDeploymentStartTopic() string {
	return this.getCommandTopic(deploymentTopic, "start")
}

func (this *Client) handleDeploymentStartCommand(message paho.Message) {
	msg := model.StartMessage{}
	err := json.Unmarshal(message.Payload(), &msg)
	if err != nil {
		this.error(err)
	}
	err = this.handler.StartDeployment(msg.DeploymentId, msg.BusinessKey, msg.Parameter)
	if err != nil {
		this.error(err)
	}
}

func (this *Client) getDeploymentDeleteTopic() string {
	return this.getCommandTopic(deploymentTopic, "delete")
}

func (this *Client) handleDeploymentDeleteCommand(message paho.Message) {
	err := this.handler.DeleteDeployment(string(message.Payload()))
	if err != nil {
		this.error(err)
	}
}

func (this *Client) SendDeploymentKnownIds(ids []string) error {
	return this.sendObj(this.getStateTopic(deploymentTopic, "known"), ids)
}

func (this *Client) SendDeploymentUpdate(instance camundamodel.Deployment) error {
	return this.sendObj(this.getStateTopic(deploymentTopic), instance)
}

func (this *Client) SendDeploymentDelete(id string) error {
	return this.sendStr(this.getStateTopic(deploymentTopic, "delete"), id)
}

func (this *Client) SendDeploymentMetadata(metadata metadata.Metadata) error {
	return this.sendObj(this.getStateTopic(deploymentTopic, "metadata"), metadata)
}
