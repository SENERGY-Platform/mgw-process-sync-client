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

package camundamodel

type Variable struct {
	Value     interface{} `json:"value"`
	Type      string      `json:"type"`
	ValueInfo interface{} `json:"valueInfo"`
}

// /engine-rest/process-definition/"+url.QueryEscape(processDefinitionId)+"/start
// /engine-rest/process-instance/"+url.QueryEscape(id)
// /engine-rest/process-instance/"+url.QueryEscape(id)
type ProcessInstance struct {
	Id             string `json:"id,omitempty"`
	DefinitionId   string `json:"definitionId,omitempty"`
	BusinessKey    string `json:"businessKey,omitempty"`
	CaseInstanceId string `json:"caseInstanceId,omitempty"`
	Ended          bool   `json:"ended,omitempty"`
	Suspended      bool   `json:"suspended,omitempty"`
	TenantId       string `json:"tenantId,omitempty"`
}

// /engine-rest/process-instance?tenantIdIn="+url.QueryEscape(userId)
type ProcessInstances = []ProcessInstance

// /engine-rest/process-definition/"+url.QueryEscape(id)
type ProcessDefinition struct {
	Id                string `json:"id,omitempty"`
	Key               string `json:"key,omitempty"`
	Category          string `json:"category,omitempty"`
	Description       string `json:"description,omitempty"`
	Name              string `json:"name,omitempty"`
	Version           int    `json:"Version,omitempty"`
	Resource          string `json:"resource,omitempty"`
	DeploymentId      string `json:"deploymentId,omitempty"`
	Diagram           string `json:"diagram,omitempty"`
	Suspended         bool   `json:"suspended,omitempty"`
	TenantId          string `json:"tenantId,omitempty"`
	VersionTag        string `json:"versionTag,omitempty"`
	HistoryTimeToLive int    `json:"historyTimeToLive,omitempty"`
}

func (this *ProcessDefinition) SetDeploymentId(id string) {
	this.DeploymentId = id
}

func (this *ProcessDefinition) GetDeploymentId() (id string) {
	return this.DeploymentId
}

// /engine-rest/process-definition?deploymentId="+url.QueryEscape(id)
type ProcessDefinitions = []ProcessDefinition

// /engine-rest/deployment/"+url.QueryEscape(id)
// /engine-rest/deployment/"+url.QueryEscape(deploymentId)
// /engine-rest/deployment/" + id + "?cascade=true
type Deployment struct {
	Id             string      `json:"id"`
	Name           string      `json:"name"`
	Source         string      `json:"source"`
	DeploymentTime interface{} `json:"deploymentTime"`
	TenantId       string      `json:"tenantId"`
}

func (this *Deployment) SetDeploymentId(id string) {
	this.Id = id
}

func (this *Deployment) GetDeploymentId() (id string) {
	return this.Id
}

// /engine-rest/deployment?tenantIdIn="+url.QueryEscape(userId)+"&"+params.Encode()
type Deployments = []Deployment

// /engine-rest/process-instance/count?tenantIdIn="+url.QueryEscape(userId)
// /engine-rest/incident/count?processDefinitionId="+url.QueryEscape(definitionId)
type Count struct {
	Count int64 `json:"count"`
}

// /engine-rest/history/process-instance/"+url.QueryEscape(id)
type HistoricProcessInstance struct {
	Id                       string  `json:"id"`
	SuperProcessInstanceId   string  `json:"superProcessInstanceId"`
	SuperCaseInstanceId      string  `json:"superCaseInstanceId"`
	CaseInstanceId           string  `json:"caseInstanceId"`
	ProcessDefinitionName    string  `json:"processDefinitionName"`
	ProcessDefinitionKey     string  `json:"processDefinitionKey"`
	ProcessDefinitionVersion float64 `json:"processDefinitionVersion"`
	ProcessDefinitionId      string  `json:"processDefinitionId"`
	BusinessKey              string  `json:"businessKey"`
	StartTime                string  `json:"startTime"`
	EndTime                  string  `json:"endTime"`
	DurationInMillis         float64 `json:"durationInMillis"`
	StartUserId              string  `json:"startUserId"`
	StartActivityId          string  `json:"startActivityId"`
	DeleteReason             string  `json:"deleteReason"`
	TenantId                 string  `json:"tenantId"`
	State                    string  `json:"state"`
}

// /engine-rest/history/process-instance?processDefinitionId="+url.QueryEscape(id)
// /engine-rest/history/process-instance?processDefinitionId="+url.QueryEscape(id)+"&finished=true
// /engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId)
// /engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId)+"&finished=true
type HistoricProcessInstances = []HistoricProcessInstance

type ExtendedDeployment struct {
	Deployment
	Diagram      string `json:"diagram"`
	DefinitionId string `json:"definition_id"`
	Error        string `json:"error"`
}

type HistoricProcessInstancesWithTotal = struct {
	Total int64                    `json:"total"`
	Data  HistoricProcessInstances `json:"data"`
}
