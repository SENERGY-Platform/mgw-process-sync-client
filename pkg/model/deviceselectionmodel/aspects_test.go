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

package deviceselectionmodel

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/devicemodel"
)

func TestAspectsJson(t *testing.T) {
	t.Run("filter criteria", func(t *testing.T) {
		criteria := FilterCriteria{}
		err := json.Unmarshal([]byte(`{"function_id":"f","aspect_id":"a1","aspect_ids":["a1","a2"]}`), &criteria)
		if err != nil {
			t.Error(err)
			return
		}
		if criteria.AspectId != "a1" || !reflect.DeepEqual(criteria.AspectIds, []string{"a1", "a2"}) {
			t.Error(criteria)
		}
	})
	t.Run("path option", func(t *testing.T) {
		option := PathOption{}
		err := json.Unmarshal([]byte(`{"path":"p","aspectNode":{"id":"a1"},"aspectNodes":[{"id":"a1"},{"id":"a2"}]}`), &option)
		if err != nil {
			t.Error(err)
			return
		}
		if option.AspectNode.Id != "a1" || !reflect.DeepEqual(option.AspectNodes, []devicemodel.AspectNode{{Id: "a1"}, {Id: "a2"}}) {
			t.Error(option)
		}
	})
	t.Run("configurable", func(t *testing.T) {
		configurable := Configurable{}
		err := json.Unmarshal([]byte(`{"path":"p","aspect_node":{"id":"a1"},"aspect_nodes":[{"id":"a1"},{"id":"a2"}]}`), &configurable)
		if err != nil {
			t.Error(err)
			return
		}
		if configurable.AspectNode.Id != "a1" || !reflect.DeepEqual(configurable.AspectNodes, []devicemodel.AspectNode{{Id: "a1"}, {Id: "a2"}}) {
			t.Error(configurable)
		}
	})
}
