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

package server

import (
	"context"
	"mgw-process-sync/pkg/configuration"
	"mgw-process-sync/pkg/tests/docker"
	"sync"
)

func CreateCamundaEnv(ctx context.Context, wg *sync.WaitGroup, initConf configuration.Config) (config configuration.Config, camundaUrl string, err error) {
	config = initConf
	var camundaPgIp string
	config.CamundaDb, camundaPgIp, _, err = docker.PostgresWithNetwork(ctx, wg, "camunda")
	if err != nil {
		return config, camundaUrl, err
	}

	camundaUrl, err = docker.Camunda(ctx, wg, camundaPgIp, "5432")
	if err != nil {
		return config, camundaUrl, err
	}

	return config, camundaUrl, nil
}
