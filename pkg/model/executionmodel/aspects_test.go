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

package executionmodel

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/devicemodel"
)

func TestTaskAspectsJson(t *testing.T) {
	t.Run("reads a task that carries only the deprecated aspect", func(t *testing.T) {
		task := Task{}
		err := json.Unmarshal([]byte(`{"version":3,"aspect":{"id":"a1"}}`), &task)
		if err != nil {
			t.Error(err)
			return
		}
		if task.Aspect == nil || task.Aspect.Id != "a1" || len(task.Aspects) != 0 {
			t.Error(task.Aspect, task.Aspects)
		}
	})
	t.Run("reads a task with an aspect list", func(t *testing.T) {
		task := Task{}
		err := json.Unmarshal([]byte(`{"version":3,"aspects":[{"id":"a1"},{"id":"a2"}]}`), &task)
		if err != nil {
			t.Error(err)
			return
		}
		if task.Aspect != nil || !reflect.DeepEqual(task.Aspects, []devicemodel.AspectNode{{Id: "a1"}, {Id: "a2"}}) {
			t.Error(task.Aspect, task.Aspects)
		}
	})
	t.Run("omits unset aspects", func(t *testing.T) {
		temp, err := json.Marshal(Task{Version: 3})
		if err != nil {
			t.Error(err)
			return
		}
		m := map[string]interface{}{}
		err = json.Unmarshal(temp, &m)
		if err != nil {
			t.Error(err)
			return
		}
		if _, ok := m["aspect"]; ok {
			t.Error(string(temp))
		}
		if _, ok := m["aspects"]; ok {
			t.Error(string(temp))
		}
	})
}
