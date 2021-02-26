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

package controller

import (
	"encoding/json"
	"log"
	"mgw-process-sync-client/pkg/model/camundamodel"
)

func (this *Controller) DeleteProcessInstanceHistory(id string) error {
	return this.camunda.RemoveProcessInstanceHistory(id, UserId)
}

//{"id_":"6b84bb04-750c-11eb-b54c-0242ac110006","proc_inst_id_":"6b84bb04-750c-11eb-b54c-0242ac110006","business_key_":null,"proc_def_key_":"ExampleId","proc_def_id_":"ExampleId:1:686e7a53-750c-11eb-b54c-0242ac110006","start_time_":"2021-02-22T12:49:36.886","end_time_":null,"removal_time_":null,"duration_":null,"start_user_id_":null,"start_act_id_":"StartEvent_1","end_act_id_":null,"super_process_instance_id_":null,"root_proc_inst_id_":"6b84bb04-750c-11eb-b54c-0242ac110006","super_case_instance_id_":null,"case_inst_id_":null,"delete_reason_":null,"tenant_id_":"user","state_":"ACTIVE"}
type ProcessInstanceHistoryInPg struct {
	Id                     string  `json:"id_"`
	SuperProcessInstanceId string  `json:"super_process_instance_id_"`
	SuperCaseInstanceId    string  `json:"super_case_instance_id_"`
	CaseInstanceId         string  `json:"case_inst_id_"`
	ProcessDefinitionKey   string  `json:"proc_def_key_"`
	ProcessDefinitionId    string  `json:"proc_def_id_"`
	BusinessKey            string  `json:"business_key_"`
	StartTime              string  `json:"start_time_"`
	EndTime                string  `json:"end_time_"`
	DurationInMillis       float64 `json:"duration_"`
	StartUserId            string  `json:"start_user_id_"`
	StartActivityId        string  `json:"start_act_id_"`
	DeleteReason           string  `json:"delete_reason_"`
	TenantId               string  `json:"tenant_id_"`
	State                  string  `json:"state_"`
}

func (this *Controller) NotifyHistoryUpdate(extra string) {
	element := ProcessInstanceHistoryInPg{}
	err := json.Unmarshal([]byte(extra), &element)
	if err != nil {
		log.Println("ERROR: unable to unmarshal history in NotifyHistoryUpdate(): ", err)
		return
	}
	history := camundamodel.HistoricProcessInstance{
		Id:                     element.Id,
		SuperProcessInstanceId: element.SuperProcessInstanceId,
		SuperCaseInstanceId:    element.SuperCaseInstanceId,
		CaseInstanceId:         element.CaseInstanceId,
		ProcessDefinitionKey:   element.ProcessDefinitionKey,
		ProcessDefinitionId:    element.ProcessDefinitionId,
		BusinessKey:            element.BusinessKey,
		StartTime:              element.StartTime,
		EndTime:                element.EndTime,
		DurationInMillis:       element.DurationInMillis,
		StartUserId:            element.StartUserId,
		StartActivityId:        element.StartActivityId,
		DeleteReason:           element.DeleteReason,
		TenantId:               element.TenantId,
		State:                  element.State,
	}

	definition, err := this.camunda.GetProcessDefinition(element.ProcessDefinitionId, UserId)
	if err != nil {
		log.Println("WARNING: unable to get process definition in NotifyHistoryUpdate(): ", err)
		err = nil
	} else {
		history.ProcessDefinitionName = definition.Name
		history.ProcessDefinitionVersion = float64(definition.Version)
	}

	err = this.backend.SendProcessHistoryUpdate(history)
	if err != nil {
		log.Println("ERROR: unable to send history update in SendProcessHistoryUpdate(): ", err)
		return
	}
}

func (this *Controller) NotifyHistoryDelete(extra string) {
	element := ProcessInstanceHistoryInPg{}
	err := json.Unmarshal([]byte(extra), &element)
	if err != nil {
		log.Println("ERROR: unable to unmarshal history in NotifyHistoryDelete(): ", err)
		return
	}
	err = this.backend.SendProcessHistoryDelete(element.Id)
	if err != nil {
		log.Println("ERROR: unable to send history update in NotifyHistoryDelete(): ", err)
		return
	}
}

func (this *Controller) SendCurrentHistories() error {
	instances, err := this.camunda.GetProcessInstanceHistoryList(UserId)
	if err != nil {
		return err
	}
	ids := []string{}
	for _, instance := range instances {
		ids = append(ids, instance.Id)
		err = this.backend.SendProcessHistoryUpdate(instance)
		if err != nil {
			return err
		}
	}
	return this.backend.SendProcessHistoryKnownIds(ids)
}
