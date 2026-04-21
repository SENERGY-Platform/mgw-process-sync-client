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
	"log"
)

type ErrorMessage struct {
	NetworkId           string `json:"network_id"`
	DeploymentId        string `json:"deployment_id"`
	CamundaDeploymentId string `json:"camunda_deployment_id"`
	BusinessKey         string `json:"business_key"`
	Error               string `json:"error"`
}

func (this *Client) error(err ErrorMessage) {
	log.Println("ERROR:", err, "\n", this.sendObj(this.getStateTopic("error"), err))
}
