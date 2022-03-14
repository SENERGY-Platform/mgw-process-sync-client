/*
 * Copyright 2019 InfAI (CC SES)
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

type Hub struct {
	Id             string   `json:"id"`
	Name           string   `json:"name"`
	Hash           string   `json:"hash"`
	DeviceLocalIds []string `json:"device_local_ids"`
}

type Content struct {
	Id                string          `json:"id"`
	ContentVariable   ContentVariable `json:"content_variable"`
	Serialization     string          `json:"serialization"`
	ProtocolSegmentId string          `json:"protocol_segment_id"`
}

type Type string

const (
	String  Type = "https://schema.org/Text"
	Integer Type = "https://schema.org/Integer"
	Float   Type = "https://schema.org/Float"
	Boolean Type = "https://schema.org/Boolean"

	List      Type = "https://schema.org/ItemList"
	Structure Type = "https://schema.org/StructuredValue"
)

type ContentVariable struct {
	Id                   string            `json:"id"`
	Name                 string            `json:"name"`
	Type                 Type              `json:"type"`
	SubContentVariables  []ContentVariable `json:"sub_content_variables"`
	CharacteristicId     string            `json:"characteristic_id"`
	Value                interface{}       `json:"value"`
	SerializationOptions []string          `json:"serialization_options"`
	UnitReference        string            `json:"unit_reference,omitempty"`
	FunctionId           string            `json:"function_id,omitempty"`
	AspectId             string            `json:"aspect_id,omitempty"`
}
