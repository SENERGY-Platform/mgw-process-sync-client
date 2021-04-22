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
	"context"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/camundamodel"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/deploymentmodel"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/docker"
	"reflect"
	"sync"
	"testing"
)

func TestMongoStorage(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mongoPort, _, err := docker.Mongo(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	config := configuration.Config{
		DeploymentMetadataStorage: "mongodb://localhost:" + mongoPort + "/metadata",
		Debug:                     true,
	}

	storage, err := NewStorage(config)
	if err != nil {
		t.Error(err)
		return
	}

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

	err = storage.Store(md1)
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
}
