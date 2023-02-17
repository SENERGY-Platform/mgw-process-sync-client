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

package repo

import (
	"context"
	eventmodel "github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/metadata"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model"
	"log"
	"sync"
)

type EventRepo struct {
	mux    sync.Mutex
	events EventIndex
}

const UserId = model.UserId

type EventIndex = map[LocalDeviceId]map[LocalServiceId][]eventmodel.EventDesc

type LocalDeviceId = string
type LocalServiceId = string

func New(context.Context, configuration.Config) (*EventRepo, error) {
	return &EventRepo{
		events: EventIndex{},
	}, nil
}

func (this *EventRepo) AddDeployment(deployment metadata.Metadata) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.events = addDeployment(this.events, deployment)
	return nil
}

func addDeployment(events EventIndex, deployment metadata.Metadata) EventIndex {
	for _, event := range deployment.DeploymentModel.EventDescriptions {
		localDeviceId, ok := deployment.DeploymentModel.DeviceIdToLocalId[event.DeviceId]
		if !ok {
			log.Printf("warning: unable to get local device id for \"%v\" --> no event handler deployed\n", event.DeviceId)
			continue
		}
		localServiceId, ok := deployment.DeploymentModel.ServiceIdToLocalId[event.ServiceId]
		if !ok {
			log.Printf("warning: unable to get local service id for \"%v\" --> no event handler deployed\n", event.ServiceId)
			continue
		}
		event.DeploymentId = deployment.CamundaDeploymentId
		events = addEvent(events, localDeviceId, localServiceId, event)
	}
	return events
}

func addEvent(events EventIndex, localDeviceId string, localServiceId string, event eventmodel.EventDesc) EventIndex {
	if _, ok := events[localDeviceId]; !ok {
		events[localDeviceId] = map[LocalServiceId][]eventmodel.EventDesc{}
	}
	event.UserId = UserId
	events[localDeviceId][localServiceId] = append(events[localDeviceId][localServiceId], event)
	return events
}

func (this *EventRepo) RemoveDeployment(deploymentId string) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	newEvents := map[LocalDeviceId]map[LocalServiceId][]eventmodel.EventDesc{}
	for localDeviceId, serviceIndex := range this.events {
		for localServiceId, events := range serviceIndex {
			for _, event := range events {
				if event.DeploymentId != deploymentId {
					newEvents = addEvent(newEvents, localDeviceId, localServiceId, event)
				}
			}
		}
	}
	this.events = newEvents
	return nil
}

func (this *EventRepo) Find(localDeviceId string, localServiceId string) (result []eventmodel.EventDesc, err error) {
	serviceIndex, ok := this.events[localDeviceId]
	if !ok {
		return []eventmodel.EventDesc{}, nil
	}
	result, ok = serviceIndex[localServiceId]
	if !ok {
		return []eventmodel.EventDesc{}, nil
	}
	return result, nil
}
