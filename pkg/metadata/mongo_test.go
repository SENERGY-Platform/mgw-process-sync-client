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
	"sync"
	"testing"

	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/docker"
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

	storage, err := NewStorage(ctx, config)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("metadata", MetadataTest(storage))
	t.Run("parameter", ParameterTest(storage))
}
