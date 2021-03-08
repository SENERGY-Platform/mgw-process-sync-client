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
	model "github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/camundamodel"
	paho "github.com/eclipse/paho.mqtt.golang"
)

const processInstanceTopic = "process-instance"

func (this *Client) SendProcessInstanceUpdate(instance model.ProcessInstance) error {
	return this.sendObj(this.getStateTopic(processInstanceTopic), instance)
}

func (this *Client) SendProcessInstanceDelete(id string) error {
	return this.sendStr(this.getStateTopic(processInstanceTopic, "delete"), id)
}

func (this *Client) getProcessStopTopic() string {
	return this.getCommandTopic(processInstanceTopic, "delete")
}

func (this *Client) handleProcessStopCommand(message paho.Message) {
	err := this.handler.DeleteProcessInstance(string(message.Payload()))
	if err != nil {
		this.error(err)
	}
}

func (this *Client) SendProcessInstanceKnownIds(ids []string) error {
	return this.sendObj(this.getStateTopic(processInstanceTopic, "known"), ids)
}
