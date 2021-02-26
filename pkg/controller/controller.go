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

package controller

import (
	"context"
	"log"
	"mgw-process-sync-client/pkg/backend"
	"mgw-process-sync-client/pkg/camunda"
	"mgw-process-sync-client/pkg/camunda/shards"
	"mgw-process-sync-client/pkg/configuration"
	"time"
)

const UserId = "senergy"

func New(config configuration.Config, ctx context.Context) (ctrl *Controller, err error) {
	ctrl = &Controller{config: config}
	if err != nil {
		return ctrl, err
	}
	ctrl.camunda = camunda.New(shards.Shards(config.CamundaUrl))
	ctrl.backend, err = backend.New(config, ctx, ctrl)
	if err != nil {
		return ctrl, err
	}
	err = ctrl.spyOnCamundaDb(ctx)
	if err != nil {
		return ctrl, err
	}

	wait, err := time.ParseDuration(config.InitialWaitDuration)
	if err != nil {
		log.Println("WARNING: unable to parse initial wait duration", config.InitialWaitDuration, err)
	} else {
		time.Sleep(wait) //wait for outstanding commands
	}
	return ctrl, ctrl.SendCurrentStates()
}

type Controller struct {
	config  configuration.Config
	backend *backend.Client
	camunda *camunda.Camunda
}

func (this *Controller) SendCurrentStates() (err error) {
	err = this.SendCurrentDeployments()
	if err != nil {
		return err
	}
	err = this.SendCurrentProcessDefs()
	if err != nil {
		return err
	}
	err = this.SendCurrentInstances()
	if err != nil {
		return err
	}
	err = this.SendCurrentHistories()
	if err != nil {
		return err
	}
	return nil
}
