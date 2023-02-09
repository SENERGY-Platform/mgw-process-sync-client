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

package controller

import (
	eventmodel "github.com/SENERGY-Platform/event-worker/pkg/model"
	"log"
	"runtime/debug"
)

func (this *Controller) UpdateDeploymentEvents(camundaDeploymentId string, descriptions []eventmodel.EventDesc, deviceMapping map[string]string, serviceMapping map[string]string) error {
	if this.metadata.IsPlaceholder() {
		return nil
	}
	m, err := this.metadata.Read(camundaDeploymentId)
	if err != nil {
		log.Println("ERROR: unable to update events", err)
		debug.PrintStack()
		return nil
	}
	m.DeploymentModel.EventDescriptions = descriptions
	m.DeploymentModel.DeviceIdToLocalId = deviceMapping
	m.DeploymentModel.ServiceIdToLocalId = serviceMapping
	err = this.metadata.Store(m)
	if err != nil {
		log.Println("ERROR: unable to update events", err)
		debug.PrintStack()
		return err
	}
	err = this.RemoveConditionalEventOperators(camundaDeploymentId)
	if err != nil {
		return err
	}
	err = this.DeployConditionalEventOperators(m)
	if err != nil {
		return err
	}
	return this.backend.SendDeploymentMetadata(m)
}
