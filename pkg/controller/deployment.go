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
	"encoding/xml"
	"errors"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/camunda"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/controller/etree"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/metadata"
	model "github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/camundamodel"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/deploymentmodel"
	"log"
)

func (this *Controller) CreateDeployment(deployment deploymentmodel.Deployment) (err error) {
	err = this.cleanupExistingDeployment(deployment.Id)
	if err != nil {
		return err
	}
	xml := deployment.Diagram.XmlDeployed
	svg := deployment.Diagram.Svg
	if !validateXml(xml) {
		log.Println("ERROR: got invalid xml, replace with default")
		xml = camunda.CreateBlankProcess()
		svg = camunda.CreateBlankSvg()
	}
	if this.config.Debug {
		log.Println("deploy process", deployment.Id, deployment.Name, xml)
	}
	var id string
	id, err = this.camunda.DeployProcess(deployment.Name, xml, svg, UserId, "senergy")
	if err != nil {
		log.Println("WARNING: unable to deploy process to camunda ", err)
		return err
	}

	//metadata
	metadata := metadata.Metadata{
		DeploymentModel:     deployment,
		ProcessParameter:    nil,
		CamundaDeploymentId: id,
	}

	metadata.ProcessParameter, err = this.getProcessParameter(id)
	if err != nil {
		log.Println("WARNING: unable to get process parameter", err)
	}

	err = this.metadata.Store(metadata)
	if err != nil {
		log.Println("WARNING: unable to store deployment metadata ", err)
	}
	return this.backend.SendDeploymentMetadata(metadata)
}

func (this *Controller) getProcessParameter(deploymentId string) (result map[string]model.Variable, err error) {
	definition, err := this.camunda.GetDefinitionByDeploymentVid(deploymentId, UserId)
	if err != nil {
		return nil, err
	}
	return this.camunda.GetProcessParameters(definition[0].Id, UserId)
}

func (this *Controller) DeleteDeployment(id string) error {
	return this.camunda.RemoveProcess(id, UserId)
}

func (this *Controller) StartDeployment(id string) error {
	definitions, err := this.camunda.GetDefinitionByDeploymentVid(id, UserId)
	if err != nil {
		return err
	}
	if len(definitions) == 0 {
		return errors.New("no definition for deployment found: " + id)
	}
	return this.camunda.StartProcess(definitions[0].Id, "", map[string]interface{}{})
}

func (this *Controller) SendCurrentDeployments() error {
	deployments, err := this.camunda.GetDeploymentList(UserId, map[string][]string{})
	if err != nil {
		return err
	}
	ids := []string{}
	for _, depl := range deployments {
		ids = append(ids, depl.Id)
		err = this.backend.SendDeploymentUpdate(depl)
		if err != nil {
			return err
		}
	}
	err = this.backend.SendDeploymentKnownIds(ids)
	if err != nil {
		return err
	}
	knownmetadata, err := this.metadata.EnsureKnownDeployments(ids)
	if err != nil {
		return err
	}
	return this.sendKnownDeploymentMetadata(knownmetadata)
}

func (this *Controller) sendKnownDeploymentMetadata(knownmetadata []metadata.Metadata) error {
	for _, metadata := range knownmetadata {
		err := this.backend.SendDeploymentMetadata(metadata)
		if err != nil {
			return err
		}
	}
	return nil
}

// {"id_":"1b3e90fe-750a-11eb-8c7e-0242ac110006","name_":"test","deploy_time_":"2021-02-22T12:33:03.214","source_":"test","tenant_id_":"user"}}
type DeploymentInPg struct {
	Id             string `json:"id_"`
	Name           string `json:"name_"`
	DeploymentTime string `json:"deploy_time_"`
	Source         string `json:"source_"`
	TenantId       string `json:"tenant_id_"`
}

func (this *Controller) NotifyDeploymentUpdate(extra string) {
	deployment := DeploymentInPg{}
	err := json.Unmarshal([]byte(extra), &deployment)
	if err != nil {
		log.Println("ERROR: unable to unmarshal deployment in NotifyDeploymentUpdate(): ", err)
		return
	}
	err = this.backend.SendDeploymentUpdate(model.Deployment{
		Id:             deployment.Id,
		Name:           deployment.Name,
		Source:         deployment.Source,
		DeploymentTime: deployment.DeploymentTime,
		TenantId:       deployment.TenantId,
	})
	if err != nil {
		log.Println("ERROR: unable to send deployment update in NotifyDeploymentUpdate(): ", err)
		return
	}
}

func (this *Controller) NotifyDeploymentDelete(extra string) {
	deployment := DeploymentInPg{}
	err := json.Unmarshal([]byte(extra), &deployment)
	if err != nil {
		log.Println("ERROR: unable to unmarshal deployment in NotifyDeploymentDelete(): ", err)
		return
	}
	err = this.backend.SendDeploymentDelete(deployment.Id)
	if err != nil {
		log.Println("ERROR: unable to send deployment delete in NotifyDeploymentDelete(): ", err)
	}
	err = this.metadata.Remove(deployment.Id)
	if err != nil {
		log.Println("WARNING: unable to remove deployment metadata", err)
	}
}

func (this *Controller) cleanupExistingDeployment(id string) error {
	return this.DeleteDeployment(id)
}

func validateXml(xmlStr string) bool {
	if xmlStr == "" {
		return false
	}
	err := etree.NewDocument().ReadFromString(xmlStr)
	if err != nil {
		log.Println("ERROR: unable to parse xml", err)
		return false
	}
	err = xml.Unmarshal([]byte(xmlStr), new(interface{}))
	if err != nil {
		log.Println("ERROR: unable to parse xml", err)
		return false
	}
	return true
}
