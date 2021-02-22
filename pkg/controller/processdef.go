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
)

// {"id_":"ExampleId:1:686e7a53-750c-11eb-b54c-0242ac110006","rev_":1,"category_":"http://bpmn.io/schema/bpmn","name_":"ExampleName","key_":"ExampleId","version_":1,"deployment_id_":"685ee9f0-750c-11eb-b54c-0242ac110006","resource_name_":"test.bpmn","dgrm_resource_name_":"test.svg","has_start_form_key_":false,"suspension_state_":1,"tenant_id_":"user","version_tag_":null,"history_ttl_":null,"startable_":true}
type ProcessDefInPg struct {
	Id                string `json:"id_"`
	Key               string `json:"key_"`
	Category          string `json:"category_"`
	Name              string `json:"name_"`
	Version           int    `json:"version_"`
	Resource          string `json:"resource_name_"`
	DeploymentId      string `json:"deployment_id_"`
	Diagram           string `json:"dgrm_resource_name_"`
	Suspended         int    `json:"suspension_state_"`
	TenantId          string `json:"tenant_id_"`
	VersionTag        string `json:"version_tag_"`
	HistoryTimeToLive int    `json:"history_ttl_"`
}

func (this *Controller) NotifyProcessDefUpdate(extra string) {
	element := ProcessDefInPg{}
	err := json.Unmarshal([]byte(extra), &element)
	if err != nil {
		log.Println("ERROR: unable to unmarshal process def in NotifyProcessDefUpdate(): ", err)
		return
	}

	def, err := this.camunda.GetProcessDefinition(element.Id, UserId)
	if err != nil {
		log.Println("ERROR: unable to get process def in NotifyProcessDefUpdate(): ", err)
		return
	}
	/*
		// alternative to this.camunda.GetProcessDefinition(element.Id, UserId)
		def := camundamodel.ProcessDefinition{
			Id:                element.Id,
			Key:               element.Key,
			Category:          element.Category,
			//Description:       element.Description, //no description in database
			Name:              element.Name,
			Version:           element.Version,
			Resource:          element.Resource,
			DeploymentId:      element.DeploymentId,
			Diagram:           element.Diagram,
			Suspended:         element.Suspended > 0,
			TenantId:          element.TenantId,
			VersionTag:        element.VersionTag,
			HistoryTimeToLive: element.HistoryTimeToLive,
		}
		def.DeploymentId, _, err = this.vid.GetVirtualId(def.DeploymentId)
		if err != nil {
			log.Println("ERROR: unable to get process vid in NotifyProcessDefUpdate(): ", err)
			return
		}
	*/

	err = this.backend.SendProcessDefinitionUpdate(def)
	if err != nil {
		log.Println("ERROR: unable to send process def update in NotifyProcessDefUpdate(): ", err)
		return
	}
}

func (this *Controller) NotifyProcessDefDelete(extra string) {
	element := ProcessDefInPg{}
	err := json.Unmarshal([]byte(extra), &element)
	if err != nil {
		log.Println("ERROR: unable to unmarshal process def in NotifyProcessDefDelete(): ", err)
		return
	}
	err = this.backend.SendProcessDefinitionDelete(element.Id)
	if err != nil {
		log.Println("ERROR: unable to send process def delete in NotifyProcessDefDelete(): ", err)
		return
	}
}

func (this *Controller) SendCurrentProcessDefs() error {
	instances, err := this.camunda.GetProcessDefinitionList(UserId)
	if err != nil {
		return err
	}
	ids := []string{}
	for _, instance := range instances {
		ids = append(ids, instance.Id)
		err = this.backend.SendProcessDefinitionUpdate(instance)
		if err != nil {
			return err
		}
	}
	return this.backend.SendProcessDefinitionKnownIds(ids)
}
