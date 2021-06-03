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
	"testing"
)

func TestBadgerStorage(t *testing.T) {
	config := configuration.Config{
		DeploymentMetadataStorage: t.TempDir(),
		Debug:                     true,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage, err := NewStorage(ctx, config)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("test", MetadataTest(storage))
}
