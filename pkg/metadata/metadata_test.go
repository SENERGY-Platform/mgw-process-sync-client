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

package metadata

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/camundamodel"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
)

func MetadataTest(storage Storage) func(t *testing.T) {
	return func(t *testing.T) {
		md1 := Metadata{
			DeploymentModel: model.FogDeploymentMessage{Deployment: deploymentmodel.Deployment{
				Name: "dpl1",
			}},
			ProcessParameter: map[string]camundamodel.Variable{
				"var_1": {Type: "string"},
			},
			CamundaDeploymentId: "cdid1",
		}

		md2 := Metadata{
			DeploymentModel: model.FogDeploymentMessage{Deployment: deploymentmodel.Deployment{
				Name: "dpl2",
			}},
			ProcessParameter: map[string]camundamodel.Variable{
				"var_1": {Type: "string"},
			},
			CamundaDeploymentId: "cdid2",
		}

		md3 := Metadata{
			DeploymentModel: model.FogDeploymentMessage{Deployment: deploymentmodel.Deployment{
				Name: "dpl3",
			}},
			ProcessParameter: map[string]camundamodel.Variable{
				"var_1": {Type: "string"},
			},
			CamundaDeploymentId: "cdid3",
		}

		err := storage.Store(md1)
		if err != nil {
			t.Error(err)
			return
		}

		err = storage.Store(md2)
		if err != nil {
			t.Error(err)
			return
		}

		err = storage.Store(md3)
		if err != nil {
			t.Error(err)
			return
		}

		list, err := storage.List()
		if err != nil {
			t.Error(err)
			return
		}

		if !reflect.DeepEqual(list, []Metadata{md1, md2, md3}) {
			t.Error(list)
		}

		known, err := storage.EnsureKnownDeployments([]string{"cdid1", "cdid2", "cdid3"})
		if err != nil {
			t.Error(err)
			return
		}

		if !reflect.DeepEqual(known, []Metadata{md1, md2, md3}) {
			t.Error(known)
		}

		known, err = storage.EnsureKnownDeployments([]string{"cdid1", "cdid3"})
		if err != nil {
			t.Error(err)
			return
		}

		if !reflect.DeepEqual(known, []Metadata{md1, md3}) {
			t.Error(known)
		}

		err = storage.Remove("cdid1")
		if err != nil {
			t.Error(err)
			return
		}

		known, err = storage.EnsureKnownDeployments([]string{"cdid1", "cdid3"})
		if err != nil {
			t.Error(err)
			return
		}

		if !reflect.DeepEqual(known, []Metadata{md3}) {
			t.Error(known)
		}

		actual, err := storage.Read("cdid3")
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, md3) {
			t.Error(actual, md3)
		}
	}
}

func ParameterTest(storage Storage) func(t *testing.T) {
	return func(t *testing.T) {
		err := storage.StoreInstanceParameter("bk1", map[string]interface{}{"foo": "bar", "num": 42.0})
		if err != nil {
			t.Error(err)
			return
		}
		err = storage.StoreInstanceParameter("bk2", map[string]interface{}{"foo": "batz", "num": 13.0})
		if err != nil {
			t.Error(err)
			return
		}
		list, err := storage.ListInstanceParameterBusinessKeys()
		if err != nil {
			t.Error(err)
			return
		}
		sort.Strings(list)
		if !reflect.DeepEqual(list, []string{"bk1", "bk2"}) {
			t.Error(list)
			return
		}
		bk1, err := storage.GetInstanceParameter("bk1")
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(bk1, map[string]interface{}{"foo": "bar", "num": 42.0}) {
			t.Error(bk1)
			return
		}
		bk2, err := storage.GetInstanceParameter("bk2")
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(bk2, map[string]interface{}{"foo": "batz", "num": 13.0}) {
			t.Error(bk2)
			return
		}
		err = storage.RemoveInstanceParameter("bk1")
		if err != nil {
			t.Error(err)
			return
		}
		list, err = storage.ListInstanceParameterBusinessKeys()
		if err != nil {
			t.Error(err)
			return
		}
		sort.Strings(list)
		if !reflect.DeepEqual(list, []string{"bk2"}) {
			t.Error(list)
			return
		}

		bk1, err = storage.GetInstanceParameter("bk1")
		if !errors.Is(err, ErrNotFound) {
			t.Error("expected ErrNotFound")
			return
		}
		bk2, err = storage.GetInstanceParameter("bk2")
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(bk2, map[string]interface{}{"foo": "batz", "num": 13.0}) {
			t.Error(bk2)
			return
		}

		err = storage.RemoveInstanceParameter("bk1")
		if err != nil {
			t.Error(err) //remove should not fail if key does not exist
			return
		}
	}
}
