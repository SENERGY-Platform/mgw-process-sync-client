/*
 * Copyright 2026 InfAI (CC SES)
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

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFilterCriteriaShortAspects(t *testing.T) {
	t.Run("uses the deprecated aspect id as a one element list", func(t *testing.T) {
		deprecated := FilterCriteria{FunctionId: "f", AspectId: "a1", DeviceClassId: "dc"}
		list := FilterCriteria{FunctionId: "f", AspectIds: []string{"a1"}, DeviceClassId: "dc"}
		if deprecated.Short() != "f_a1_dc" || list.Short() != deprecated.Short() {
			t.Error(deprecated.Short(), list.Short())
		}
	})
	t.Run("renders the list independent of its order", func(t *testing.T) {
		a := FilterCriteria{FunctionId: "f", AspectIds: []string{"a2", "a1"}, DeviceClassId: "dc"}
		b := FilterCriteria{FunctionId: "f", AspectIds: []string{"a1", "a2"}, DeviceClassId: "dc"}
		if a.Short() != "f_a1,a2_dc" || a.Short() != b.Short() {
			t.Error(a.Short(), b.Short())
		}
	})
	t.Run("prefers the list over the deprecated aspect id", func(t *testing.T) {
		c := FilterCriteria{FunctionId: "f", AspectId: "a3", AspectIds: []string{"a1", "a2"}, DeviceClassId: "dc"}
		if c.Short() != "f_a1,a2_dc" {
			t.Error(c.Short())
		}
	})
}

func TestDeviceGroupFilterCriteriaShortAspects(t *testing.T) {
	deprecated := DeviceGroupFilterCriteria{Interaction: EVENT, FunctionId: "f", AspectId: "a1", DeviceClassId: "dc"}
	single := DeviceGroupFilterCriteria{Interaction: EVENT, FunctionId: "f", AspectIds: []string{"a1"}, DeviceClassId: "dc"}
	if deprecated.Short() != "f_a1_dc_event" || single.Short() != deprecated.Short() {
		t.Error(deprecated.Short(), single.Short())
	}
	list := DeviceGroupFilterCriteria{Interaction: EVENT, FunctionId: "f", AspectIds: []string{"a2", "a1"}, DeviceClassId: "dc"}
	if list.Short() != "f_a1,a2_dc_event" {
		t.Error(list.Short())
	}
}

func TestDeviceGroupAspectsJson(t *testing.T) {
	t.Run("reads criteria stored before the aspect lists", func(t *testing.T) {
		group := DeviceGroup{}
		err := json.Unmarshal([]byte(`{"criteria":[{"interaction":"event","function_id":"f","aspect_id":"a1","device_class_id":"dc"}]}`), &group)
		if err != nil {
			t.Error(err)
			return
		}
		group.SetShortCriteria()
		if !reflect.DeepEqual(group.CriteriaShort, []string{"f_a1_dc_event"}) {
			t.Error(group.CriteriaShort)
		}
	})
	t.Run("reads criteria with aspect lists", func(t *testing.T) {
		group := DeviceGroup{}
		err := json.Unmarshal([]byte(`{"criteria":[{"interaction":"event","function_id":"f","aspect_ids":["a2","a1"],"device_class_id":"dc"}]}`), &group)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(group.Criteria[0].AspectIds, []string{"a2", "a1"}) {
			t.Error(group.Criteria[0].AspectIds)
		}
		group.SetShortCriteria()
		if !reflect.DeepEqual(group.CriteriaShort, []string{"f_a1,a2_dc_event"}) {
			t.Error(group.CriteriaShort)
		}
	})
	t.Run("omits an empty aspect list", func(t *testing.T) {
		temp, err := json.Marshal(FilterCriteria{FunctionId: "f", AspectId: "a1", DeviceClassId: "dc"})
		if err != nil {
			t.Error(err)
			return
		}
		if string(temp) != `{"function_id":"f","aspect_id":"a1","device_class_id":"dc"}` {
			t.Error(string(temp))
		}
	})
}
