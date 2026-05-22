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
	"fmt"
	"log/slog"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/camunda"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/controller/etree"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/metadata"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/camundamodel"
)

func (this *Controller) CreateDeployment(deployment model.FogDeploymentMessage) (id string, err error) {
	err = this.cleanupExistingDeployment(deployment.Id)
	if err != nil {
		return "", err
	}
	xml := deployment.Diagram.XmlDeployed

	xml, err = SetProcessId(xml, deployment.Id)
	if err != nil {
		return "", err
	}

	xml = this.replaceNotificationUrl(xml)

	xml, err = ReplaceTaskTopics(xml, this.config.TaskTopicReplace)
	if err != nil {
		return "", err
	}

	svg := deployment.Diagram.Svg
	if !validateXml(xml) {
		this.config.GetLogger().Info("got invalid xml, replace with default")
		xml = camunda.CreateBlankProcess()
		svg = camunda.CreateBlankSvg()
	}
	this.config.GetLogger().Debug("deploy process", "deploymentId", deployment.Id, "deploymentName", deployment.Name)
	id, err = this.camunda.DeployProcess(deployment.Name, xml, svg, UserId, "senergy")
	if err != nil {
		this.config.GetLogger().Warn("unable to deploy process to camunda ", "error", err)
		return "", err
	}

	incidentHandling := deployment.IncidentHandling
	if incidentHandling != nil {
		err = this.DeployIncidentsHandlerForDeploymentId(id, *incidentHandling)
		if err != nil {
			removeErr := this.camunda.RemoveProcess(id, UserId)
			if removeErr != nil {
				this.config.GetLogger().Error("unable to remove deployed process", "id", id, "removeErr", removeErr, "error", err)
			}
			return id, err
		}
	}

	//metadata
	metadata := metadata.Metadata{
		DeploymentModel:     deployment,
		ProcessParameter:    nil,
		CamundaDeploymentId: id,
	}

	metadata.ProcessParameter, err = this.getProcessParameter(id)
	if err != nil {
		this.config.GetLogger().Warn("unable to get process parameter", "error", err)
	}

	err = this.metadata.Store(metadata)
	if err != nil {
		this.config.GetLogger().Warn("unable to store deployment metadata", "error", err)
	}

	err = this.DeployConditionalEventOperators(metadata)
	if err != nil {
		this.config.GetLogger().Error("unable to deploy conditional event operators", "error", err)
		return id, err
	}

	return id, this.backend.SendDeploymentMetadata(metadata)
}

func (this *Controller) replaceNotificationUrl(xml string) string {
	return strings.ReplaceAll(xml, this.config.NotificationUrlPlaceholder, this.config.NotificationUrl)
}

func (this *Controller) getProcessParameter(deploymentId string) (result map[string]camundamodel.Variable, err error) {
	definition, err := this.camunda.GetDefinitionByDeploymentVid(deploymentId, UserId)
	if err != nil {
		return nil, err
	}
	if len(definition) == 0 {
		return nil, fmt.Errorf("no definition for deployment '%v' found", deploymentId)
	}
	return this.camunda.GetProcessParameters(definition[0].Id, UserId)
}

func (this *Controller) DeleteDeployment(id string) error {
	return this.camunda.RemoveProcess(id, UserId)
}

func (this *Controller) StartDeployment(id string, businessKey string, parameter map[string]interface{}) error {
	definitions, err := this.camunda.GetDefinitionByDeploymentVid(id, UserId)
	if err != nil {
		return err
	}
	if len(definitions) == 0 {
		return fmt.Errorf("no definition for deployment '%s' found", id)
	}
	if businessKey != "" && len(parameter) > 0 {
		err = this.metadata.StoreInstanceParameter(businessKey, parameter)
		if err != nil {
			return fmt.Errorf("unable to store instance parameter: %w", err)
		}
	}
	return this.camunda.StartProcess(definitions[0].Id, businessKey, UserId, parameter)
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
		this.config.GetLogger().Error("unable to unmarshal deployment in NotifyDeploymentUpdate()", "error", err)
		return
	}
	err = this.backend.SendDeploymentUpdate(camundamodel.Deployment{
		Id:             deployment.Id,
		Name:           deployment.Name,
		Source:         deployment.Source,
		DeploymentTime: deployment.DeploymentTime,
		TenantId:       deployment.TenantId,
	})
	if err != nil {
		this.config.GetLogger().Error("unable to send deployment update in NotifyDeploymentUpdate()", "error", err)
		return
	}
}

func (this *Controller) NotifyDeploymentDelete(extra string) {
	deployment := DeploymentInPg{}
	err := json.Unmarshal([]byte(extra), &deployment)
	if err != nil {
		this.config.GetLogger().Error("unable to unmarshal deployment in NotifyDeploymentDelete()", "error", err)
		return
	}
	err = this.backend.SendDeploymentDelete(deployment.Id)
	if err != nil {
		this.config.GetLogger().Error("unable to send deployment delete in NotifyDeploymentDelete()", "error", err)
	}
	err = this.metadata.Remove(deployment.Id)
	if err != nil {
		this.config.GetLogger().Warn("unable to remove deployment metadata", "error", err)
	}
	err = this.RemoveConditionalEventOperators(deployment.Id)
	if err != nil {
		this.config.GetLogger().Warn("unable to remove event operator", "error", err)
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
		slog.Error("unable to parse xml", "error", err)
		return false
	}
	err = xml.Unmarshal([]byte(xmlStr), new(interface{}))
	if err != nil {
		slog.Error("unable to parse xml", "error", err)
		return false
	}
	return true
}

func ReplaceTaskTopics(xml string, fromToMap map[string]string) (result string, err error) {
	defer func() {
		if r := recover(); r != nil && err == nil {
			slog.Error("Recovered Error", "error", r, "stack", debug.Stack())
			err = errors.New(fmt.Sprint("Recovered Error: ", r))
		}
	}()
	doc := etree.NewDocument()
	err = doc.ReadFromString(xml)
	if err != nil {
		return result, err
	}
	for from, to := range fromToMap {
		for _, element := range doc.FindElements("//bpmn:serviceTask[@camunda:topic='" + from + "']") {
			attr := element.SelectAttr("camunda:topic")
			if attr != nil {
				attr.Value = to
			}
		}
	}
	return doc.WriteToString()
}

func SetProcessId(xml string, id string) (result string, err error) {
	defer func() {
		if r := recover(); r != nil && err == nil {
			slog.Error("Recovered Error", "error", r, "stack", debug.Stack())
			err = errors.New(fmt.Sprint("Recovered Error: ", r))
		}
	}()
	doc := etree.NewDocument()
	err = doc.ReadFromString(xml)
	if err != nil {
		return result, err
	}
	normalizedId := "deplid_" + strings.NewReplacer("-", "_", ":", "_", "#", "_").Replace(id)
	for i, element := range doc.FindElements("//bpmn:process") {
		attr := element.SelectAttr("id")
		if attr != nil {
			if i > 0 {
				attr.Value = normalizedId + "_" + strconv.Itoa(i)
			} else {
				attr.Value = normalizedId
			}
		}
	}
	return doc.WriteToString()
}
