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

package model

import (
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/deploymentmodel"
)

type StartMessage struct {
	DeploymentId string                 `json:"deployment_id"`
	Parameter    map[string]interface{} `json:"parameter"`
}

type FogDeploymentMessage struct {
	deploymentmodel.Deployment
	AnalyticsRecords   []AnalyticsRecord `json:"analytics_records"`
	DeviceIdToLocalId  map[string]string `json:"device_id_to_local_id"`
	ServiceIdToLocalId map[string]string `json:"service_id_to_local_id"`
}

type DeviceEventAnalyticsRecord struct {
	Label        string `json:"label"`
	DeploymentId string `json:"deployment_id"`
	FlowId       string `json:"flow_id"`
	EventId      string `json:"event_id"`
	DeviceId     string `json:"device_id"`
	ServiceId    string `json:"service_id"`
	Value        string `json:"value"`
	Path         string `json:"path"`
	CastFrom     string `json:"cast_from"`
	CastTo       string `json:"cast_to"`
}

type GroupEventAnalyticsRecord struct {
	Label                          string                             `json:"label"`
	Desc                           GroupEventDescription              `json:"desc"`
	ServiceIds                     []string                           `json:"service_ids"`
	ServiceToDeviceIdsMapping      map[string][]string                `json:"service_to_device_ids_mapping"`
	ServiceToPathMapping           map[string]string                  `json:"service_to_path_mapping"`
	ServiceToPathAndCharacteristic map[string][]PathAndCharacteristic `json:"service_to_path_and_characteristic"`
}

type PathAndCharacteristic struct {
	JsonPath         string `json:"json_path"`
	CharacteristicId string `json:"characteristic_id"`
}

type GroupEventDescription struct {
	ImportId         string
	Path             string
	DeviceGroupId    string
	DeviceIds        []string //optional
	EventId          string
	DeploymentId     string
	FunctionId       string
	AspectId         string
	FlowId           string
	OperatorValue    string
	CharacteristicId string
}

type AnalyticsRecord struct {
	DeviceEvent *DeviceEventAnalyticsRecord `json:"device_event"`
	GroupEvent  *GroupEventAnalyticsRecord  `json:"group_event"`
}
