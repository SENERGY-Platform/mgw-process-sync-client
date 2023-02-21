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
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/camundamodel"
	paho "github.com/eclipse/paho.mqtt.golang"
)

const incidentTopic = "incident"

func (this *Client) getProcessIncidentTopic() string {
	return this.getStateTopic(incidentTopic)
}

func (this *Client) SendIncident(incident camundamodel.Incident) error {
	return this.sendObj(this.getProcessIncidentTopic(), incident)
}

func (this *Client) handleProcessIncident(message paho.Message) {
	incident := camundamodel.Incident{}
	err := json.Unmarshal(message.Payload(), &incident)
	if err != nil {
		this.error(err)
		return
	}
	err = this.handler.HandleIncident(incident)
	if err != nil {
		this.error(err)
		return
	}
}
