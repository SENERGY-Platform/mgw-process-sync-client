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

package events

import (
	"context"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/events/api"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/events/repo"
)

func StartApi(ctx context.Context, config configuration.Config) (r *repo.EventRepo, err error) {
	r, err = repo.New(ctx, config)
	if err != nil {
		return r, err
	}
	err = api.Start(ctx, config, r)
	return r, err
}
