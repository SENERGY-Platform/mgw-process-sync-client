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

const processInstanceHistoryTopic = "process-instance-history"

func (this *Client) getProcessHistoryDeleteTopic() string {
	return this.getCommandTopic(processInstanceHistoryTopic, "delete")
}

func (this *Client) handleProcessHistoryDeleteCommand(message paho.Message) {
	err := this.handler.DeleteProcessInstanceHistory(string(message.Payload()))
	if err != nil {
		this.error(err)
	}
}

func (this *Client) SendProcessHistoryUpdate(instance model.HistoricProcessInstance) error {
	return this.send(this.getStateTopic(processInstanceHistoryTopic), instance)
}

func (this *Client) SendProcessHistoryDelete(id string) error {
	return this.send(this.getStateTopic(processInstanceHistoryTopic, "delete"), id)
}

func (this *Client) SendProcessHistoryKnownIds(ids []string) error {
	return this.send(this.getStateTopic(processInstanceHistoryTopic, "known"), ids)
}
