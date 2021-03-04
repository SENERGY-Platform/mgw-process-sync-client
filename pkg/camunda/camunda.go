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

package camunda

import (
	"bytes"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/camunda/request"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/camunda/shards"
	model "github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/camundamodel"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"errors"
	"net/http"

	"encoding/json"

	"log"
)

type Camunda struct {
	shards shards.Shards
}

func New(shards shards.Shards) *Camunda {
	return &Camunda{shards: shards}
}

func (this *Camunda) StartProcess(processDefinitionId string, userId string, parameter map[string]interface{}) (err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return err
	}

	message := createStartMessage(parameter)

	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(message)
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", shard+"/engine-rest/process-definition/"+url.QueryEscape(processDefinitionId)+"/submit-form", b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		temp, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(resp.Status + " " + string(temp))
		return
	}
	return nil
}

func createStartMessage(parameter map[string]interface{}) map[string]interface{} {
	if len(parameter) == 0 {
		return map[string]interface{}{}
	}
	variables := map[string]interface{}{}
	for key, val := range parameter {
		variables[key] = map[string]interface{}{
			"value": val,
		}
	}
	return map[string]interface{}{"variables": variables}
}

func (this *Camunda) GetProcessParameters(processDefinitionId string, userId string) (result map[string]model.Variable, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	req, err := http.NewRequest("GET", shard+"/engine-rest/process-definition/"+url.QueryEscape(processDefinitionId)+"/form-variables", nil)
	if err != nil {
		return result, err
	}
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	if resp.StatusCode != http.StatusOK {
		temp, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(resp.Status + " " + string(temp))
		return
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

func (this *Camunda) StartProcessGetId(processDefinitionId string, userId string, parameter map[string]interface{}) (result model.ProcessInstance, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}

	message := createStartMessage(parameter)

	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(message)
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", shard+"/engine-rest/process-definition/"+url.QueryEscape(processDefinitionId)+"/submit-form", b)
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	if resp.StatusCode != http.StatusOK {
		temp, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(resp.Status + " " + string(temp))
		return
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

func (this *Camunda) CheckProcessDefinitionAccess(id string, userId string) (err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return err
	}
	definition := model.ProcessDefinition{}
	err = request.Get(shard+"/engine-rest/process-definition/"+url.QueryEscape(id), &definition)
	if err == nil && definition.TenantId != userId {
		err = errors.New("access denied")
	}
	return
}

func (this *Camunda) CheckDeploymentAccess(id string, userId string) (err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return err
	}
	wrapper := model.Deployment{}
	err = request.Get(shard+"/engine-rest/deployment/"+url.QueryEscape(id), &wrapper)
	if err != nil {
		return err
	}
	if wrapper.Id == "" {
		return CamundaDeploymentUnknown
	}
	if wrapper.TenantId != userId {
		err = AccessDenied
	}
	return
}

func (this *Camunda) CheckProcessInstanceAccess(id string, userId string) (err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return err
	}
	wrapper := model.ProcessInstance{}
	err = request.Get(shard+"/engine-rest/process-instance/"+url.QueryEscape(id), &wrapper)
	if err == nil && wrapper.TenantId != userId {
		err = errors.New("access denied")
	}
	return
}

func (this *Camunda) CheckHistoryAccess(id string, userId string) (definitionId string, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return definitionId, err
	}
	wrapper := model.HistoricProcessInstance{}
	err = request.Get(shard+"/engine-rest/history/process-instance/"+url.QueryEscape(id), &wrapper)
	if err == nil && wrapper.TenantId != userId {
		err = errors.New("access denied")
	}
	return wrapper.ProcessDefinitionId, err
}

func (this *Camunda) RemoveProcessInstance(id string, userId string) (err error) {
	////DELETE "/engine-rest/process-instance/" + processInstanceId
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return err
	}
	client := &http.Client{}
	request, err := http.NewRequest("DELETE", shard+"/engine-rest/process-instance/"+url.QueryEscape(id)+"?skipIoMappings=true", nil)
	if err != nil {
		return
	}
	resp, err := client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if !(resp.StatusCode == 200 || resp.StatusCode == 204) {
		msg, _ := ioutil.ReadAll(resp.Body)
		err = errors.New("error on delete in engine for " + shard + "/engine-rest/process-instance/" + url.QueryEscape(id) + ": " + resp.Status + " " + string(msg))
	}
	return
}

func (this *Camunda) RemoveProcessInstanceHistory(id string, userId string) (err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return err
	}
	//DELETE "/engine-rest/history/process-instance/" + processInstanceId
	client := &http.Client{}
	request, err := http.NewRequest("DELETE", shard+"/engine-rest/history/process-instance/"+url.QueryEscape(id), nil)
	if err != nil {
		return
	}
	resp, err := client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if err == nil && !(resp.StatusCode == 200 || resp.StatusCode == 204) {
		msg, _ := ioutil.ReadAll(resp.Body)
		err = errors.New("error on delete in engine for " + shard + "/engine-rest/history/process-instance/" + url.QueryEscape(id) + ": " + resp.Status + " " + string(msg))
	}
	return
}

func (this *Camunda) GetProcessInstanceHistoryByProcessDefinition(id string, userId string) (result model.HistoricProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/history/process-instance?processDefinitionId="
	err = request.Get(shard+"/engine-rest/history/process-instance?processDefinitionId="+url.QueryEscape(id), &result)
	return
}
func (this *Camunda) GetProcessInstanceHistoryByProcessDefinitionFinished(id string, userId string) (result model.HistoricProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/history/process-instance?processDefinitionId="
	err = request.Get(shard+"/engine-rest/history/process-instance?processDefinitionId="+url.QueryEscape(id)+"&finished=true", &result)
	return
}
func (this *Camunda) GetProcessInstanceHistoryByProcessDefinitionUnfinished(id string, userId string) (result model.HistoricProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/history/process-instance?processDefinitionId="
	err = request.Get(shard+"/engine-rest/history/process-instance?processDefinitionId="+url.QueryEscape(id)+"&unfinished=true", &result)
	return
}

func (this *Camunda) GetProcessInstanceHistoryList(userId string) (result model.HistoricProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-instance"
	err = request.Get(shard+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId), &result)
	return
}

func (this *Camunda) GetFilteredProcessInstanceHistoryList(userId string, query url.Values) (result model.HistoricProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	query.Del("tenantIdIn")
	err = request.Get(shard+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId)+"&"+query.Encode(), &result)
	return
}

func (this *Camunda) GetProcessInstanceHistoryListFinished(userId string) (result model.HistoricProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-instance"
	err = request.Get(shard+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId)+"&finished=true", &result)
	return
}
func (this *Camunda) GetProcessInstanceHistoryListUnfinished(userId string) (result model.HistoricProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-instance"
	err = request.Get(shard+"/engine-rest/history/process-instance?tenantIdIn="+url.QueryEscape(userId)+"&unfinished=true", &result)
	return
}
func (this *Camunda) GetProcessInstanceCount(userId string) (result model.Count, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-instance/count"
	err = request.Get(shard+"/engine-rest/process-instance/count?tenantIdIn="+url.QueryEscape(userId), &result)
	return
}
func (this *Camunda) GetProcessInstanceList(userId string) (result model.ProcessInstances, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-instance"
	err = request.Get(shard+"/engine-rest/process-instance?tenantIdIn="+url.QueryEscape(userId), &result)
	return
}

func (this *Camunda) GetProcessDefinition(id string, userId string) (result model.ProcessDefinition, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-definition/" + processDefinitionId
	err = request.Get(shard+"/engine-rest/process-definition/"+url.QueryEscape(id), &result)
	if err != nil {
		return
	}
	return
}

func (this *Camunda) GetProcessDefinitionList(userId string) (result model.ProcessDefinitions, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/process-definition/" + processDefinitionId
	err = request.Get(shard+"/engine-rest/process-definition", &result)
	return
}

func (this *Camunda) GetProcessDefinitionDiagram(id string, userId string) (resp *http.Response, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return resp, err
	}
	// "/engine-rest/process-definition/" + processDefinitionId + "/diagram"
	resp, err = http.Get(shard + "/engine-rest/process-definition/" + url.QueryEscape(id) + "/diagram")
	return
}
func (this *Camunda) GetDeploymentList(userId string, params url.Values) (result model.Deployments, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	// "/engine-rest/deployment?tenantIdIn="+userId
	params.Del("tenantIdIn")
	path := shard + "/engine-rest/deployment?tenantIdIn=" + url.QueryEscape(userId) + "&" + params.Encode()
	err = request.Get(path, &result)
	return
}

var UnknownVid = errors.New("unknown vid")
var CamundaDeploymentUnknown = errors.New("deployment unknown in camunda")
var AccessDenied = errors.New("access denied")

func (this *Camunda) GetDefinitionByDeploymentVid(id string, userId string) (result model.ProcessDefinitions, err error) {
	//"/engine-rest/process-definition?deploymentId=
	result, err = this.GetRawDefinitionsByDeployment(id, userId)
	return
}

func (this *Camunda) GetRawDefinitionsByDeployment(deploymentId string, userId string) (result model.ProcessDefinitions, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	err = request.Get(shard+"/engine-rest/process-definition?deploymentId="+url.QueryEscape(deploymentId), &result)
	return
}

func (this *Camunda) GetDeployment(deploymentId string, userId string) (result model.Deployment, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	//"/engine-rest/deployment/" + id
	err = request.Get(shard+"/engine-rest/deployment/"+url.QueryEscape(deploymentId), &result)
	return
}

func (this *Camunda) GetDeploymentCountByShard(deploymentId string, shard string) (result model.Count, err error) {
	err = request.Get(shard+"/engine-rest/deployment/count?id="+url.QueryEscape(deploymentId), &result)
	return
}

func buildPayLoad(name string, xml string, svg string, boundary string, owner string, deploymentSource string) string {
	segments := []string{}
	if deploymentSource == "" {
		deploymentSource = "sepl"
	}

	segments = append(segments, "Content-Disposition: form-data; name=\"data\"; "+"filename=\""+name+".bpmn\"\r\nContent-Type: text/xml\r\n\r\n"+xml+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"diagram\"; "+"filename=\""+name+".svg\"\r\nContent-Type: image/svg+xml\r\n\r\n"+svg+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"deployment-name\"\r\n\r\n"+name+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"deployment-source\"\r\n\r\n"+deploymentSource+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"tenant-id\"\r\n\r\n"+owner+"\r\n")

	return "--" + boundary + "\r\n" + strings.Join(segments, "--"+boundary+"\r\n") + "--" + boundary + "--\r\n"
}

//returns original deploymentId (not vid)
func (this *Camunda) DeployProcess(name string, xml string, svg string, owner string, source string) (deploymentId string, err error) {
	responseWrapper, err := this.deployProcess(name, xml, svg, owner, source)
	if err != nil {
		log.Println("ERROR: unable to decode process engine deployment response", err)
		return deploymentId, err
	}
	ok := false
	deploymentId, ok = responseWrapper["id"].(string)
	if !ok {
		log.Println("ERROR: unable to interpret process engine deployment response", responseWrapper)
		if responseWrapper["type"] == "ProcessEngineException" {
			log.Println("DEBUG: try deploying placeholder process")
			responseWrapper, err = this.deployProcess(name, CreateBlankProcess(), CreateBlankSvg(), owner, source)
			deploymentId, ok = responseWrapper["id"].(string)
			if !ok {
				log.Println("ERROR: unable to deploy placeholder process", responseWrapper)
				err = errors.New("unable to interpret process engine deployment response")
				return
			}
		} else {
			log.Println("ERROR: unable to deploy placeholder process", responseWrapper)
			err = errors.New("unable to interpret process engine deployment response")
			return
		}
	}
	if err == nil && deploymentId == "" {
		err = errors.New("process-engine didnt deploy process: " + xml)
	}
	return
}

func (this *Camunda) deployProcess(name string, xml string, svg string, owner string, source string) (result map[string]interface{}, err error) {
	shard, err := this.shards.EnsureShardForUser(owner)
	if err != nil {
		return result, err
	}
	result = map[string]interface{}{}
	boundary := "---------------------------" + time.Now().String()
	b := strings.NewReader(buildPayLoad(name, xml, svg, boundary, owner, source))
	resp, err := http.Post(shard+"/engine-rest/deployment/create", "multipart/form-data; boundary="+boundary, b)
	if err != nil {
		log.Println("ERROR: request to processengine ", err)
		return result, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

//uses original deploymentId (not vid)
func (this *Camunda) RemoveProcess(deploymentId string, userId string) (err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return err
	}
	return this.RemoveProcessForShard(deploymentId, shard)
}

func (this *Camunda) RemoveProcessForShard(deploymentId string, shard string) (err error) {
	count, err := this.GetDeploymentCountByShard(deploymentId, shard)
	if err != nil {
		return err
	}
	if count.Count == 0 {
		return nil
	}
	client := &http.Client{}
	url := shard + "/engine-rest/deployment/" + deploymentId + "?cascade=true&skipIoMappings=true"
	request, err := http.NewRequest("DELETE", url, nil)
	_, err = client.Do(request)
	return
}

func (this *Camunda) RemoveProcessFromAllShards(deploymentId string) (err error) {
	shards, err := this.shards.GetShards()
	if err != nil {
		return err
	}
	for _, shard := range shards {
		err = this.RemoveProcessForShard(deploymentId, shard)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Camunda) GetExtendedDeploymentList(userId string, params url.Values) (result []model.ExtendedDeployment, err error) {
	deployments, err := this.GetDeploymentList(userId, params)
	if err != nil {
		return result, err
	}
	for _, deployment := range deployments {
		extended, err := this.GetExtendedDeployment(deployment, userId)
		if err != nil {
			result = append(result, model.ExtendedDeployment{Deployment: deployment, Error: err.Error()})
			err = nil
		} else {
			result = append(result, extended)
		}
	}
	return
}

func (this *Camunda) GetExtendedDeployment(deployment model.Deployment, userId string) (result model.ExtendedDeployment, err error) {
	definition, err := this.GetDefinitionByDeploymentVid(deployment.Id, userId)
	if err != nil {
		return result, err
	}
	if len(definition) < 1 {
		return result, errors.New("missing definition for given deployment")
	}
	if len(definition) > 1 {
		return result, errors.New("more than one definition for given deployment")
	}
	svgResp, err := this.GetProcessDefinitionDiagram(definition[0].Id, userId)
	if err != nil {
		return result, err
	}
	svg, err := ioutil.ReadAll(svgResp.Body)
	if err != nil {
		return result, err
	}
	return model.ExtendedDeployment{Deployment: deployment, Diagram: string(svg), DefinitionId: definition[0].Id}, nil
}

func (this *Camunda) GetProcessInstanceHistoryListWithTotal(userId string, searchtype string, searchvalue string, limit string, offset string, sortby string, sortdirection string, finished bool) (result model.HistoricProcessInstancesWithTotal, err error) {
	shard, err := this.shards.EnsureShardForUser(userId)
	if err != nil {
		return result, err
	}
	params := url.Values{
		"tenantIdIn":  []string{userId},
		"maxResults":  []string{limit},
		"firstResult": []string{offset},
		"sortBy":      []string{sortby},
		"sortOrder":   []string{sortdirection},
	}
	if searchtype != "" && searchvalue != "" {
		if searchtype == "processDefinitionId" {
			params["processDefinitionId"] = []string{searchvalue}
		}
		if searchtype == "processDefinitionNameLike" {
			params["processDefinitionNameLike"] = []string{"%" + searchvalue + "%"}
		}

	}
	if finished {
		params["finished"] = []string{"true"}
	} else {
		params["unfinished"] = []string{"true"}
	}

	temp := model.HistoricProcessInstances{}
	err = request.Get(shard+"/engine-rest/history/process-instance?"+params.Encode(), &temp)
	if err != nil {
		return
	}
	for _, process := range temp {
		result.Data = append(result.Data, process)
	}

	count := model.Count{}
	err = request.Get(shard+"/engine-rest/history/process-instance/count?"+params.Encode(), &count)
	result.Total = count.Count
	return
}

func CreateBlankSvg() string {
	return `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.2" id="Layer_1" x="0px" y="0px" viewBox="0 0 20 16" xml:space="preserve">
<path fill="#D61F33" d="M10,0L0,16h20L10,0z M11,13.908H9v-2h2V13.908z M9,10.908v-6h2v6H9z"/>
</svg>`
}

func CreateBlankProcess() string {
	templ := `<bpmn:definitions xmlns:xsi='http://www.w3.org/2001/XMLSchema-instance' xmlns:bpmn='http://www.omg.org/spec/BPMN/20100524/MODEL' xmlns:bpmndi='http://www.omg.org/spec/BPMN/20100524/DI' xmlns:dc='http://www.omg.org/spec/DD/20100524/DC' id='Definitions_1' targetNamespace='http://bpmn.io/schema/bpmn'><bpmn:process id='PROCESSID' isExecutable='true'><bpmn:startEvent id='StartEvent_1'/></bpmn:process><bpmndi:BPMNDiagram id='BPMNDiagram_1'><bpmndi:BPMNPlane id='BPMNPlane_1' bpmnElement='PROCESSID'><bpmndi:BPMNShape id='_BPMNShape_StartEvent_2' bpmnElement='StartEvent_1'><dc:Bounds x='173' y='102' width='36' height='36'/></bpmndi:BPMNShape></bpmndi:BPMNPlane></bpmndi:BPMNDiagram></bpmn:definitions>`
	return strings.Replace(templ, "PROCESSID", "id_"+strconv.FormatInt(time.Now().Unix(), 10), 1)
}
