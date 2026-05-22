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
	"errors"
	"slices"

	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/camundamodel"
)

func (this *Controller) DeleteProcessInstance(id string) error {
	instance, err := this.camunda.GetHistoricProcessInstance(id, UserId)
	if err == nil {
		this.metadata.RemoveInstanceParameter(instance.BusinessKey)
	}
	return this.camunda.RemoveProcessInstance(id, UserId)
}

// {"id_":"6b84bb04-750c-11eb-b54c-0242ac110006","rev_":1,"root_proc_inst_id_":"6b84bb04-750c-11eb-b54c-0242ac110006","proc_inst_id_":"6b84bb04-750c-11eb-b54c-0242ac110006","business_key_":null,"parent_id_":null,"proc_def_id_":"ExampleId:1:686e7a53-750c-11eb-b54c-0242ac110006","super_exec_":null,"super_case_exec_":null,"case_inst_id_":null,"act_id_":null,"act_inst_id_":"6b84bb04-750c-11eb-b54c-0242ac110006","is_active_":false,"is_concurrent_":false,"is_scope_":true,"is_event_scope_":false,"suspension_state_":1,"cached_ent_state_":0,"sequence_counter_":2,"tenant_id_":"user"}
type ProcessInstanceInPg struct {
	Id               string  `json:"id_"`
	DefinitionId     string  `json:"proc_def_id_"`
	BusinessKey      string  `json:"business_key_"`
	CaseInstanceId   string  `json:"case_inst_id_"`
	Active           bool    `json:"is_active_"`
	TenantId         string  `json:"tenant_id_"`
	EndTime          *string `json:"end_time_"`
	ParentInstanceId *string `json:"parent_id_"` //usable to check if process instance ist root element
}

func (this *Controller) NotifyInstanceUpdate(extra string) {
	element := ProcessInstanceInPg{}
	err := json.Unmarshal([]byte(extra), &element)
	if err != nil {
		this.config.GetLogger().Error("unable to unmarshal instance in NotifyInstanceUpdate()", "error", err)
		return
	}
	//forward only root instances
	if isRootInstance(element) {
		instance := camundamodel.ProcessInstance{
			Id:             element.Id,
			DefinitionId:   element.DefinitionId,
			BusinessKey:    element.BusinessKey,
			CaseInstanceId: element.CaseInstanceId,
			Ended:          element.EndTime != nil,
			Suspended:      !element.Active,
			TenantId:       element.TenantId,
		}
		err = this.backend.SendProcessInstanceUpdate(instance)
		if err != nil {
			this.config.GetLogger().Error("unable to send process instance update in NotifyInstanceUpdate()", "error", err)
			return
		}
	}
}

func (this *Controller) NotifyInstanceDelete(extra string) {
	element := ProcessInstanceInPg{}
	err := json.Unmarshal([]byte(extra), &element)
	if err != nil {
		this.config.GetLogger().Error("unable to unmarshal process instance in NotifyInstanceDelete()", "error", err)
		return
	}
	//forward only root instances
	if isRootInstance(element) {
		err = this.backend.SendProcessInstanceDelete(element.Id)
		if err != nil {
			this.config.GetLogger().Error("unable to send process instance delete in NotifyInstanceDelete()", "error", err)
			return
		}
	}
}

func isRootInstance(pgInstance ProcessInstanceInPg) bool {
	return pgInstance.ParentInstanceId == nil
}

func (this *Controller) SendCurrentInstances() error {
	instances, err := this.camunda.GetProcessInstanceList(UserId)
	if err != nil {
		return err
	}
	ids := []string{}
	for _, instance := range instances {
		ids = append(ids, instance.Id)
		err = this.backend.SendProcessInstanceUpdate(instance)
		if err != nil {
			return err
		}
	}
	return this.backend.SendProcessInstanceKnownIds(ids)
}

func (this *Controller) RemoveUnknownInstanceBusinessKeys() error {
	instances, err := this.camunda.GetProcessInstanceHistoryList(UserId)
	if err != nil {
		return err
	}
	knownKeys := []string{}
	for _, instance := range instances {
		knownKeys = append(knownKeys, instance.BusinessKey)
	}
	storedKeys, err := this.metadata.ListInstanceParameterBusinessKeys()
	if err != nil {
		return err
	}
	for _, key := range storedKeys {
		if !slices.Contains(knownKeys, key) {
			temperr := this.metadata.RemoveInstanceParameter(key)
			this.config.GetLogger().Error("unable to remove instance parameter", "error", temperr, "businessKey", key)
			err = errors.Join(err, temperr)
		}
	}
	return err
}
