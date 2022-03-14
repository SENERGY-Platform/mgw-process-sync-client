/*
 * Copyright 2020 InfAI (CC SES)
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

package deviceselectionmodel

import (
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/importmodel"
)

type Selectable struct {
	Device             *devicemodel.Device      `json:"device"`
	Services           []devicemodel.Service    `json:"services"`
	DeviceGroup        *devicemodel.DeviceGroup `json:"device_group,omitempty"`
	Import             *importmodel.Import      `json:"import,omitempty"`
	ImportType         *importmodel.ImportType  `json:"importType,omitempty"`
	ServicePathOptions map[string][]PathOption  `json:"servicePathOptions,omitempty"`
}

type FilterCriteriaAndSet []FilterCriteria

type FilterCriteriaOrSet []FilterCriteria

type FilterCriteria struct {
	FunctionId    string `json:"function_id"`
	DeviceClassId string `json:"device_class_id"`
	AspectId      string `json:"aspect_id"`
}

type BulkRequestElement struct {
	Id                string                   `json:"id"`
	FilterInteraction *devicemodel.Interaction `json:"filter_interaction"`
	FilterProtocols   []string                 `json:"filter_protocols"`
	Criteria          FilterCriteriaAndSet     `json:"criteria"`
	IncludeGroups     bool                     `json:"include_groups"`
	IncludeImports    bool                     `json:"include_imports"`
}

type BulkRequest []BulkRequestElement

type BulkResult []BulkResultElement

type BulkResultElement struct {
	Id          string       `json:"id"`
	Selectables []Selectable `json:"selectables"`
}

type PathOption struct {
	Path             string                 `json:"path"`
	CharacteristicId string                 `json:"characteristicId"`
	AspectNode       devicemodel.AspectNode `json:"aspectNode"`
	FunctionId       string                 `json:"functionId"`
	IsVoid           bool                   `json:"isVoid"`
	Value            interface{}            `json:"value,omitempty"`
	Type             string                 `json:"type,omitempty"`
	Configurables    []Configurable         `json:"configurables,omitempty"`
}

type Configurable struct {
	Path             string                 `json:"path"`
	CharacteristicId string                 `json:"characteristic_id"`
	AspectNode       devicemodel.AspectNode `json:"aspect_node"`
	FunctionId       string                 `json:"function_id"`
	Value            interface{}            `json:"value,omitempty"`
	Type             string                 `json:"type,omitempty"`
}
