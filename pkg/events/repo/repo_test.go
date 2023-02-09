/*
 * Copyright 2023 InfAI (CC SES)
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

package repo

import (
	"context"
	eventmodel "github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/metadata"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model"
	"reflect"
	"testing"
)

func TestRepo(t *testing.T) {
	repo, err := New(context.Background(), configuration.Config{})
	if err != nil {
		t.Error(err)
		return
	}
	err = repo.AddDeployment(metadata.Metadata{
		CamundaDeploymentId: "deplid_1",
		DeploymentModel: model.FogDeploymentMessage{
			DeviceIdToLocalId: map[string]string{
				"did1": "ldid1",
				"did2": "ldid2",
			},
			ServiceIdToLocalId: map[string]string{
				"sid1": "lsid1",
				"sid2": "lsid2",
			},
			EventDescriptions: []eventmodel.EventDesc{
				{
					DeploymentId:  "nope",
					EventId:       "1",
					DeviceId:      "did1",
					ServiceId:     "sid1",
					DeviceGroupId: "",
					Script:        "x == 42",
					ValueVariable: "x",
				},
				{
					DeploymentId:  "nope",
					EventId:       "1-group",
					DeviceId:      "did1",
					ServiceId:     "sid1",
					DeviceGroupId: "gid1",
					Script:        "x == 42",
					ValueVariable: "x",
				},
				{
					DeploymentId:  "nope",
					EventId:       "1-group",
					DeviceId:      "did2",
					ServiceId:     "sid2",
					DeviceGroupId: "gid1",
					Script:        "x == 42",
					ValueVariable: "x",
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = repo.AddDeployment(metadata.Metadata{
		CamundaDeploymentId: "deplid_2",
		DeploymentModel: model.FogDeploymentMessage{
			DeviceIdToLocalId: map[string]string{
				"did1": "ldid1",
				"did2": "ldid2",
			},
			ServiceIdToLocalId: map[string]string{
				"sid1": "lsid1",
				"sid2": "lsid2",
			},
			EventDescriptions: []eventmodel.EventDesc{
				{
					DeploymentId:  "nope",
					EventId:       "removed-1",
					DeviceId:      "did1",
					ServiceId:     "sid1",
					DeviceGroupId: "",
					Script:        "x == 42",
					ValueVariable: "x",
				},
				{
					DeploymentId:  "nope",
					EventId:       "removed-2",
					DeviceId:      "did1",
					ServiceId:     "sid1",
					DeviceGroupId: "gid1",
					Script:        "x == 42",
					ValueVariable: "x",
				},
				{
					DeploymentId:  "nope",
					EventId:       "removed-2",
					DeviceId:      "did2",
					ServiceId:     "sid2",
					DeviceGroupId: "gid1",
					Script:        "x == 42",
					ValueVariable: "x",
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = repo.RemoveDeployment("deplid_2")
	if err != nil {
		t.Error(err)
		return
	}

	actual, err := repo.Find("ldid1", "lsid1")
	if err != nil {
		t.Error(err)
		return
	}

	expected := []eventmodel.EventDesc{
		{
			DeploymentId:  "deplid_1",
			EventId:       "1",
			DeviceId:      "did1",
			ServiceId:     "sid1",
			DeviceGroupId: "",
			Script:        "x == 42",
			ValueVariable: "x",
		},
		{
			DeploymentId:  "deplid_1",
			EventId:       "1-group",
			DeviceId:      "did1",
			ServiceId:     "sid1",
			DeviceGroupId: "gid1",
			Script:        "x == 42",
			ValueVariable: "x",
		},
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%#v\n%#v\n", expected, actual)
	}

}
