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

package api

import (
	"encoding/json"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func init() {
	endpoints = append(endpoints, &Events{})
}

type Events struct{}

// Find godoc
// @Summary      finds event descriptions for event-worker
// @Description  finds event descriptions for event-worker
// @Tags         event-description
// @Produce      json
// @Param        local_device_id query string false "search event-descriptions by local device id"
// @Param        local_service_id query string false "search event-descriptions by local service id"
// @Success      200 {array} []model.EventDesc
// @Failure      500
// @Router       /event-descriptions [get]
func (this *Events) Find(config configuration.Config, router *httprouter.Router, repo Repo) {
	router.GET("/event-descriptions", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		localDeviceId := request.URL.Query().Get("local_device_id")
		localServiceId := request.URL.Query().Get("local_service_id")
		result, err := repo.Find(localDeviceId, localServiceId)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})
}
