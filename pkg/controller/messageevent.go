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
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/analytics"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/metadata"
	"log"
	"runtime/debug"
	"strings"
	"time"
)

func (this *Controller) DeployMessageEventOperators(metadata metadata.Metadata) {
	if this.metadata.IsPlaceholder() {
		if len(metadata.DeploymentModel.AnalyticsRecords) > 0 {
			log.Println("WARNING: no metadata storage configured --> no message event handling")
		}
		return
	}
	const envelopePrefix = "value."
	for _, record := range metadata.DeploymentModel.AnalyticsRecords {
		if record.DeviceEvent != nil {
			localDeviceId := metadata.DeploymentModel.DeviceIdToLocalId[record.DeviceEvent.DeviceId]
			localServiceId := metadata.DeploymentModel.ServiceIdToLocalId[record.DeviceEvent.ServiceId]
			path := record.DeviceEvent.Path
			if strings.HasPrefix(path, envelopePrefix) {
				path = path[len(envelopePrefix):]
			}
			this.sendAnalyticsCommand(analytics.ControlCommand{
				Command: "startOperator",
				Data: analytics.OperatorJob{
					ImageId: record.DeviceEvent.FlowId,
					InputTopics: []analytics.InputTopic{
						{
							Name: "event/" + localDeviceId + "/" + localServiceId,
							Mappings: []analytics.Mapping{
								{
									Dest:   "value",
									Source: path,
								},
							},
						},
					},
					OperatorConfig: this.getOperatorConfig(
						record.DeviceEvent.EventId,
						record.DeviceEvent.Value,
						record.DeviceEvent.CastFrom,
						record.DeviceEvent.CastTo),
					Config: analytics.FogConfig{
						PipelineId:  metadata.CamundaDeploymentId,
						OutputTopic: "event-trigger",
						OperatorId:  metadata.CamundaDeploymentId + "_" + record.DeviceEvent.EventId,
					},
				},
			})
		}
		if record.GroupEvent != nil {
			inputTopics := []analytics.InputTopic{}
			for _, serviceId := range record.GroupEvent.ServiceIds {
				path := record.GroupEvent.ServiceToPathMapping[serviceId]
				if path == "" {
					log.Println("WARNING: missing path for service in DeployGroup()", serviceId, " --> skip service for group event deployment")
					continue
				}
				if strings.HasPrefix(path, envelopePrefix) {
					path = path[len(envelopePrefix):]
				}
				for _, deviceId := range record.GroupEvent.ServiceToDeviceIdsMapping[serviceId] {
					localDeviceId := metadata.DeploymentModel.DeviceIdToLocalId[deviceId]
					localServiceId := metadata.DeploymentModel.ServiceIdToLocalId[serviceId]

					inputTopics = append(inputTopics, analytics.InputTopic{
						Name: "event/" + localDeviceId + "/" + localServiceId,
						Mappings: []analytics.Mapping{
							{
								Dest:   "value",
								Source: path,
							},
						},
					})
				}
			}

			command := analytics.ControlCommand{
				Command: "startOperator",
				Data: analytics.OperatorJob{
					ImageId:     record.GroupEvent.Desc.FlowId,
					InputTopics: inputTopics,
					OperatorConfig: this.getOperatorConfig(
						record.GroupEvent.Desc.EventId,
						record.GroupEvent.Desc.OperatorValue,
						"",
						""),
					Config: analytics.FogConfig{
						PipelineId:  metadata.CamundaDeploymentId,
						OutputTopic: "event-trigger",
						OperatorId:  metadata.CamundaDeploymentId + "_" + record.GroupEvent.Desc.EventId,
					},
				},
			}
			this.sendAnalyticsCommand(command)
		}
	}
}

func (this *Controller) RemoveMessageEventOperators(deploymentId string) {
	if this.metadata.IsPlaceholder() {
		return
	}
	metadata, err := this.metadata.Read(deploymentId)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return
	}
	//use elements because no case distinction between group and device events needed and EventId should match
	for _, element := range metadata.DeploymentModel.Elements {
		if element.MessageEvent != nil {
			this.sendAnalyticsCommand(analytics.ControlCommand{
				Command: "stopOperator",
				Data: analytics.OperatorJob{
					Config: analytics.FogConfig{
						PipelineId: metadata.CamundaDeploymentId,
						OperatorId: metadata.CamundaDeploymentId + "_" + element.MessageEvent.EventId,
					},
				},
			})
		}
	}
}

func (this *Controller) getOperatorConfig(eventId string, value string, from string, to string) map[string]string {
	return map[string]string{
		"value":       value,
		"url":         this.config.CamundaUrl + "/engine-rest/message",
		"eventId":     eventId,
		"convertFrom": from,
		"convertTo":   to,
	}
}

func (this *Controller) sendAnalyticsCommand(command analytics.ControlCommand) {
	go func() {
		for this.analytics == nil {
			log.Println("analytics mqtt client not connected; wait 1s")
			time.Sleep(1 * time.Second)
		}
		this.analytics.Send(command)
	}()
}
