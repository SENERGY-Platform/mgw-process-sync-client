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

package deploymentmodel

import (
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/deviceselectionmodel"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/importmodel"
)

type Selection struct {
	FilterCriteria        FilterCriteria                   `json:"filter_criteria"`
	SelectionOptions      []SelectionOption                `json:"selection_options"`
	SelectedDeviceId      *string                          `json:"selected_device_id"`
	SelectedServiceId     *string                          `json:"selected_service_id"`
	SelectedDeviceGroupId *string                          `json:"selected_device_group_id"`
	SelectedImportId      *string                          `json:"selected_import_id"`
	SelectedPath          *deviceselectionmodel.PathOption `json:"selected_path"`
}

type SelectionOption struct {
	Device      *Device                                      `json:"device"`
	Services    []Service                                    `json:"services"`
	DeviceGroup *DeviceGroup                                 `json:"device_group"`
	Import      *importmodel.Import                          `json:"import"`
	ImportType  *importmodel.ImportType                      `json:"importType"`
	PathOptions map[string][]deviceselectionmodel.PathOption `json:"path_options,omitempty"`
}

type Device struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type DeviceGroup struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Service struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type FilterCriteria struct {
	CharacteristicId *string `json:"characteristic_id"` //TODO: remove?
	FunctionId       *string `json:"function_id"`
	DeviceClassId    *string `json:"device_class_id"`
	AspectId         *string `json:"aspect_id"`
}

func (this FilterCriteria) ToDeviceTypeFilter() (result deviceselectionmodel.FilterCriteria) {
	if this.FunctionId != nil {
		result.FunctionId = *this.FunctionId
	}
	if this.DeviceClassId != nil {
		result.DeviceClassId = *this.DeviceClassId
	}
	if this.AspectId != nil {
		result.AspectId = *this.AspectId
	}
	return
}
