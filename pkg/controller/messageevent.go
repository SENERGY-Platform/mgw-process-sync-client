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
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/metadata"
	"log"
)

func (this *Controller) DeployConditionalEventOperators(metadata metadata.Metadata) error {
	if this.metadata.IsPlaceholder() {
		if len(metadata.DeploymentModel.EventDescriptions) > 0 {
			log.Println("WARNING: no metadata storage configured --> no message event handling")
		}
		return nil
	}
	return this.events.AddDeployment(metadata)
}

func (this *Controller) RemoveConditionalEventOperators(deploymentId string) error {
	if this.metadata.IsPlaceholder() {
		return nil
	}
	return this.events.RemoveDeployment(deploymentId)
}
