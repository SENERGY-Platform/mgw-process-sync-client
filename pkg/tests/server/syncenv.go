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

func CreateSyncEnv(ctx context.Context, wg *sync.WaitGroup, initConf configuration.Config) (config configuration.Config, err error) {
	config, _, err = CreateCamundaEnv(ctx, wg, initConf)
	if err != nil {
		return config, err
	}
	mqttport, _, err := docker.Mqtt(ctx, wg)
	if err != nil {
		return config, err
	}
	config.MqttBroker = "tcp://localhost:" + mqttport
	config.MqttClientId = "test-sync-client"
	config.NetworkId = "test-network-id"
	return config, nil
}
