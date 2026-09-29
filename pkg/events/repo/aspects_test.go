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

package repo

import (
	"context"
	"reflect"
	"testing"

	eventmodel "github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/metadata"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model"
)

// the repo hands the descriptions to the local event-worker, which evaluates the aspect list
// and folds the deprecated aspect id into it. both have to arrive unchanged.
func TestRepoAspects(t *testing.T) {
	repo, err := New(context.Background(), configuration.Config{})
	if err != nil {
		t.Error(err)
		return
	}
	err = repo.AddDeployment(metadata.Metadata{
		CamundaDeploymentId: "deplid_1",
		DeploymentModel: model.FogDeploymentMessage{
			DeviceIdToLocalId:  map[string]string{"did1": "ldid1"},
			ServiceIdToLocalId: map[string]string{"sid1": "lsid1"},
			EventDescriptions: []eventmodel.EventDesc{
				{EventId: "deprecated", DeviceId: "did1", ServiceId: "sid1", AspectId: "a1"},
				{EventId: "list", DeviceId: "did1", ServiceId: "sid1", AspectIds: []string{"a1", "a2"}},
				{EventId: "both", DeviceId: "did1", ServiceId: "sid1", AspectId: "a1", AspectIds: []string{"a2"}},
			},
		},
	})
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
		{UserId: UserId, DeploymentId: "deplid_1", EventId: "deprecated", DeviceId: "did1", ServiceId: "sid1", AspectId: "a1"},
		{UserId: UserId, DeploymentId: "deplid_1", EventId: "list", DeviceId: "did1", ServiceId: "sid1", AspectIds: []string{"a1", "a2"}},
		{UserId: UserId, DeploymentId: "deplid_1", EventId: "both", DeviceId: "did1", ServiceId: "sid1", AspectId: "a1", AspectIds: []string{"a2"}},
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%#v\n%#v\n", expected, actual)
		return
	}

	wantAspects := [][]string{{"a1"}, {"a1", "a2"}, {"a2", "a1"}}
	for i, desc := range actual {
		if !reflect.DeepEqual(desc.GetAspectIds(), wantAspects[i]) {
			t.Error(desc.EventId, desc.GetAspectIds())
		}
	}
}
