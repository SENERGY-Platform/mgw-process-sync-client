/*
 * Copyright 2023 InfAI (CC SES)
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

package controller

import (
	"fmt"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/controller/notification"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/camundamodel"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"log"
)

type OnIncident struct {
	ProcessDefinitionId string `json:"process_definition_id" bson:"process_definition_id"`
	Restart             bool   `json:"restart" bson:"restart"`
	Notify              bool   `json:"notify" bson:"notify"`
}

func (this *Controller) DeployIncidentsHandlerForDeploymentId(camundaDeplId string, handling deploymentmodel.IncidentHandling) error {
	definitions, err := this.camunda.GetRawDefinitionsByDeployment(camundaDeplId, UserId)
	if err != nil {
		return err
	}
	if len(definitions) == 0 {
		log.Println("WARNING: no definitions for deployment found --> no incident handling deployed")
	}
	if this.incidentsHandler == nil {
		this.incidentsHandler = map[string]OnIncident{}
	}
	for _, definition := range definitions {
		this.incidentsHandler[definition.Id] = OnIncident{
			ProcessDefinitionId: definition.Id,
			Restart:             handling.Restart,
			Notify:              handling.Notify,
		}
	}
	return nil
}

func (this *Controller) HandleIncident(incident camundamodel.Incident) error {
	handler, ok := this.incidentsHandler[incident.ProcessDefinitionId]
	if !ok {
		log.Printf("unhandled incident for %v", incident.DeploymentName)
		return nil
	}
	log.Printf("handle incident for %v: notify: %v, restart: %v", incident.DeploymentName, handler.Notify, handler.Restart)
	if handler.Notify {
		msg := notification.Message{
			Title:   "Fog Process-Incident in " + incident.DeploymentName,
			Message: incident.ErrorMessage,
		}
		if handler.Restart {
			msg.Message = msg.Message + "\n\nprocess will be restarted"
		}
		_ = notification.Send(this.config.NotificationUrl, msg)
	}
	if handler.Restart {
		err := this.camunda.StartProcess(incident.ProcessDefinitionId, UserId, nil)
		if err != nil {
			log.Printf("ERROR: unable to restart process %v \n %#v \n", err, incident)
			if incident.TenantId != "" {
				_ = notification.Send(this.config.NotificationUrl, notification.Message{
					Title:   "Fog ERROR: unable to restart process after incident in: " + incident.DeploymentName,
					Message: fmt.Sprintf("Restart-Error: %v \n\n Incident: %v \n", err, incident.ErrorMessage),
				})
			}
		}
	}
	return nil
}
