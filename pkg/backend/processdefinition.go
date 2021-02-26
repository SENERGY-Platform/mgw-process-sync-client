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

package backend

import (
	model "mgw-process-sync-client/pkg/model/camundamodel"
)

const processProcessDefinitionTopic = "process-definition"

func (this *Client) SendProcessDefinitionUpdate(instance model.ProcessDefinition) error {
	return this.send(this.getStateTopic(processProcessDefinitionTopic), instance)
}

func (this *Client) SendProcessDefinitionDelete(id string) error {
	return this.send(this.getStateTopic(processProcessDefinitionTopic, "delete"), id)
}

func (this *Client) SendProcessDefinitionKnownIds(ids []string) error {
	return this.send(this.getStateTopic(processProcessDefinitionTopic, "known"), ids)
}
