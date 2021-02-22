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

package devicemodel

type Device struct {
	Id           string `json:"id"`
	LocalId      string `json:"local_id,omitempty"`
	Name         string `json:"name,omitempty"`
	DeviceTypeId string `json:"device_type_id,omitempty"`
}

type DeviceType struct {
	Id            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Services      []Service `json:"services"`
	DeviceClassId string    `json:"device_class_id"`
	RdfType       string    `json:"rdf_type"`
}

type Service struct {
	Id          string      `json:"id"`
	LocalId     string      `json:"local_id,omitempty"`
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Interaction Interaction `json:"interaction,omitempty"`
	AspectIds   []string    `json:"aspect_ids,omitempty"`
	ProtocolId  string      `json:"protocol_id,omitempty"`
	Inputs      []Content   `json:"inputs,omitempty"`
	Outputs     []Content   `json:"outputs,omitempty"`
	FunctionIds []string    `json:"function_ids,omitempty"`
	RdfType     string      `json:"rdf_type,omitempty"`
}

type Interaction string

const (
	EVENT             Interaction = "event"
	REQUEST           Interaction = "request"
	EVENT_AND_REQUEST Interaction = "event+request"
)

type Protocol struct {
	Id               string            `json:"id"`
	Name             string            `json:"name"`
	Handler          string            `json:"handler"`
	ProtocolSegments []ProtocolSegment `json:"protocol_segments"`
}

type ProtocolSegment struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type DeviceGroup struct {
	Id            string                      `json:"id"`
	Name          string                      `json:"name"`
	Image         string                      `json:"image"`
	Criteria      []DeviceGroupFilterCriteria `json:"criteria"`
	DeviceIds     []string                    `json:"device_ids"`
	CriteriaShort []string                    `json:"criteria_short,omitempty"`
}

func (this *DeviceGroup) SetShortCriteria() {
	this.CriteriaShort = []string{}
	for _, criteria := range this.Criteria {
		this.CriteriaShort = append(this.CriteriaShort, criteria.Short())
	}
}

type FilterCriteria struct {
	FunctionId    string `json:"function_id"`
	AspectId      string `json:"aspect_id"`
	DeviceClassId string `json:"device_class_id"`
}

func (this FilterCriteria) Short() string {
	return this.FunctionId + "_" + this.AspectId + "_" + this.DeviceClassId
}

type DeviceGroupFilterCriteria struct {
	Interaction   Interaction `json:"interaction"`
	FunctionId    string      `json:"function_id"`
	AspectId      string      `json:"aspect_id"`
	DeviceClassId string      `json:"device_class_id"`
}

func (this DeviceGroupFilterCriteria) Short() string {
	return this.FunctionId + "_" + this.AspectId + "_" + this.DeviceClassId + "_" + string(this.Interaction)
}
