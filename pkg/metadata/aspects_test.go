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

package metadata

import (
	"reflect"
	"testing"

	eventmodel "github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
)

// the event descriptions are read back from the storage after a restart, so the aspect
// lists and the deprecated aspect id have to survive the round trip through each storage.
func AspectsTest(storage Storage) func(t *testing.T) {
	return func(t *testing.T) {
		md := Metadata{
			DeploymentModel: model.FogDeploymentMessage{
				Deployment: deploymentmodel.Deployment{Name: "aspects"},
				EventDescriptions: []eventmodel.EventDesc{
					{EventId: "deprecated", AspectId: "a1"},
					{EventId: "list", AspectIds: []string{"a1", "a2"}},
					{EventId: "both", AspectId: "a1", AspectIds: []string{"a2"}},
				},
			},
			CamundaDeploymentId: "cdid_aspects",
		}
		err := storage.Store(md)
		if err != nil {
			t.Error(err)
			return
		}
		defer storage.Remove(md.CamundaDeploymentId)

		actual, err := storage.Read(md.CamundaDeploymentId)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual.DeploymentModel.EventDescriptions, md.DeploymentModel.EventDescriptions) {
			t.Errorf("\n%#v\n%#v", md.DeploymentModel.EventDescriptions, actual.DeploymentModel.EventDescriptions)
		}
	}
}
