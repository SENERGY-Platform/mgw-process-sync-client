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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/controller/notification"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/metadata"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/camundamodel"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"github.com/SENERGY-Platform/service-commons/pkg/cache"
	"github.com/google/uuid"
)

type OnIncident struct {
	ProcessDefinitionId string `json:"process_definition_id" bson:"process_definition_id"`
	Restart             bool   `json:"restart" bson:"restart"`
	Notify              bool   `json:"notify" bson:"notify"`
}

// public.act_ru_incident (id_, rev_, incident_timestamp_, incident_msg_, incident_type_, execution_id_, activity_id_, proc_inst_id_, proc_def_id_, cause_incident_id_, root_cause_incident_id_, configuration_, tenant_id_, job_def_id_) VALUES ('6da046d1-5580-11ef-9030-0242ac11000a', 1, '2024-08-08 12:19:15.517000', 'Unable to evaluate script while executing activity ”Task_0fi26gl” in the process definition with id ”script_err:1:631e00d6-5580-11ef-9030-0242ac11000a”:TypeError: Cannot read property "batz" from undefined in <eval> at line number 2', 'failedJob', '6614ab9a-5580-11ef-9030-0242ac11000a', 'IntermediateThrowEvent_1jxyivh', '6613e848-5580-11ef-9030-0242ac11000a', 'script_err:1:631e00d6-5580-11ef-9030-0242ac11000a', '6da046d1-5580-11ef-9030-0242ac11000a', '6da046d1-5580-11ef-9030-0242ac11000a', '66156eec-5580-11ef-9030-0242ac11000a', 'senergy', '631e27e7-5580-11ef-9030-0242ac11000a');
type ProcessIncidentInPg struct {
	Id                  string `json:"id_"`
	Message             string `json:"incident_msg_"`
	ActivityId          string `json:"activity_id_"`
	ProcessInstanceId   string `json:"proc_inst_id_"`
	ProcessDefinitionId string `json:"proc_def_id_"`
}

func (this *Controller) NotifyIncident(extra string) {
	element := ProcessIncidentInPg{}
	err := json.Unmarshal([]byte(extra), &element)
	if err != nil {
		this.config.GetLogger().Error("unable to unmarshal process incident in NotifyIncident()", "error", err)
		return
	}

	def, err := this.camunda.GetProcessDefinition(element.ProcessDefinitionId, UserId)
	if err != nil {
		this.config.GetLogger().Warn("unable to get process def in NotifyIncident()", "error", err)
		def = camundamodel.ProcessDefinition{Name: "unknown"}
	}

	instance, err := this.camunda.GetHistoricProcessInstance(element.ProcessInstanceId, UserId)
	if err != nil {
		this.config.GetLogger().Warn("unable to get process instance in NotifyIncident()", "error", err)
		instance = camundamodel.HistoricProcessInstance{}
	}

	err = this.backend.SendIncident(camundamodel.Incident{
		Id:                  uuid.NewString(),
		ExternalTaskId:      element.ActivityId,
		ProcessInstanceId:   element.ProcessInstanceId,
		ProcessDefinitionId: element.ProcessDefinitionId,
		WorkerId:            "mgw-process-sync-client",
		ErrorMessage:        element.Message,
		Time:                time.Now(),
		TenantId:            UserId,
		DeploymentName:      def.Name,
		BusinessKey:         instance.BusinessKey,
	})
	if err != nil {
		this.config.GetLogger().Warn("unable to send incident", "error", err)
	}
}

func (this *Controller) sendPgIncident(incident ProcessIncidentInPg) {
	def, err := this.camunda.GetProcessDefinition(incident.ProcessDefinitionId, UserId)
	if err != nil {
		this.config.GetLogger().Warn("unable to get process def in NotifyIncident()", "error", err)
		def = camundamodel.ProcessDefinition{Name: "unknown"}
	}

	instance, err := this.camunda.GetHistoricProcessInstance(incident.ProcessInstanceId, UserId)
	if err != nil {
		this.config.GetLogger().Warn("unable to get process instance in NotifyIncident()", "error", err)
		instance = camundamodel.HistoricProcessInstance{}
	}

	err = this.backend.SendIncident(camundamodel.Incident{
		Id:                  uuid.NewString(),
		ExternalTaskId:      incident.ActivityId,
		ProcessInstanceId:   incident.ProcessInstanceId,
		ProcessDefinitionId: incident.ProcessDefinitionId,
		WorkerId:            "mgw-process-sync-client",
		ErrorMessage:        incident.Message,
		Time:                time.Now(),
		TenantId:            UserId,
		DeploymentName:      def.Name,
		BusinessKey:         instance.BusinessKey,
	})
	if err != nil {
		this.config.GetLogger().Warn("unable to send incident", "error", err)
	}
}

func (this *Controller) SendCurrentIncidents() (count int, err error) {
	incidents, err := this.camunda.GetIncidents(UserId)
	if err != nil {
		this.config.GetLogger().Error("unable to load current incidents", "error", err)
		return count, err
	}
	for _, incident := range incidents {
		this.sendPgIncident(ProcessIncidentInPg{
			Id:                  incident.Id,
			Message:             incident.IncidentMessage,
			ActivityId:          incident.ActivityId,
			ProcessInstanceId:   incident.ProcessInstanceId,
			ProcessDefinitionId: incident.ProcessDefinitionId,
		})
	}
	return len(incidents), nil
}

func (this *Controller) DeployIncidentsHandlerForDeploymentId(camundaDeplId string, handling deploymentmodel.IncidentHandling) error {
	definitions, err := this.camunda.GetRawDefinitionsByDeployment(camundaDeplId, UserId)
	if err != nil {
		return err
	}
	if len(definitions) == 0 {
		this.config.GetLogger().Warn("no definitions for deployment found --> no incident handling deployed", "deploymentId", camundaDeplId)
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

func (this *Controller) handleIncident(incident camundamodel.Incident) error {
	handler, ok := this.incidentsHandler[incident.ProcessDefinitionId]
	if !ok {
		this.config.GetLogger().Warn("unhandled incident for deployment", "deploymentId", incident.DeploymentName, "incident", incident)
		return nil
	}
	this.config.GetLogger().Info("incident handled", "name", incident.DeploymentName, "instance", incident.ProcessInstanceId, "businessKey", incident.BusinessKey, "restart", handler.Restart, "notify", handler.Notify, "error", incident.ErrorMessage)
	if handler.Notify {
		msg := notification.Message{
			Title:   "Fog Process-Incident in " + incident.DeploymentName,
			Message: incident.ErrorMessage,
			Topic:   notification.Topic,
		}
		if handler.Restart {
			msg.Message = msg.Message + "\n\nprocess will be restarted"
		}
		_ = notification.Send(this.config.NotificationUrl, msg)
	}
	err := this.camunda.StopProcessInstance(incident.ProcessInstanceId)
	if err != nil {
		return err
	}
	if handler.Restart {
		this.config.GetLogger().Debug("restarting process", "definitionsId", incident.ProcessDefinitionId, "name", incident.DeploymentName, "instance", incident.ProcessInstanceId, "businessKey", incident.BusinessKey)
		param, err := this.GetProcessStartParameters(incident)
		if err != nil {
			this.config.GetLogger().Error("unable to get process parameters for incident", "error", err, "incident", fmt.Sprintf("%#v", incident))
			return err
		}
		err = this.camunda.StartProcess(incident.ProcessDefinitionId, incident.BusinessKey, UserId, param)
		if err != nil {
			this.config.GetLogger().Error("unable to restart process", "error", err, "incident", fmt.Sprintf("%#v", incident))
			if incident.TenantId != "" {
				_ = notification.Send(this.config.NotificationUrl, notification.Message{
					Title:   "Fog ERROR: unable to restart process after incident in: " + incident.DeploymentName,
					Message: fmt.Sprintf("Restart-Error: %v \n\n Incident: %v \n", err, incident.ErrorMessage),
					Topic:   notification.Topic,
				})
			}
		}
	}
	return nil
}

func (this *Controller) HandleIncident(incident camundamodel.Incident) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	//for every process instance an incident may only be handled once every 5 min
	//use the cache.Use method to do incident handling, only if the process instance is not found in cache
	_, err := cache.Use[string](this.handledIncidentsCache, incident.ProcessInstanceId, func() (string, error) {
		return "", this.handleIncident(incident)
	}, cache.NoValidation, 5*time.Minute)
	return err
}

func (this *Controller) GetProcessStartParameters(incident camundamodel.Incident) (params map[string]interface{}, err error) {
	params, err = this.GetStoredStartParameters(incident.BusinessKey)
	if err == nil {
		return params, nil
	}
	if !errors.Is(err, metadata.ErrNotFound) {
		this.config.GetLogger().Warn("unable to get stored start parameters for incident", "businessKey", incident.BusinessKey, "error", err)
	}
	camundaParams, err := this.camunda.GetProcessParameters(incident.ProcessDefinitionId, UserId)
	if err != nil {
		this.config.GetLogger().Error("unable to get camunda process parameters", "error", err)
		return nil, err
	}
	if len(camundaParams) > 0 {
		return nil, fmt.Errorf("needed process start parameters not found")
	}
	return nil, nil //parameters are not needed --> error in GetStoredStartParameters is unimportant
}

func (this *Controller) GetStoredStartParameters(businessKey string) (params map[string]interface{}, err error) {
	return this.metadata.GetInstanceParameter(businessKey)
}
