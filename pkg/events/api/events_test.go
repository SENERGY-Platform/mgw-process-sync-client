/*
 * Copyright 2026 InfAI (CC SES)
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
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
)

type repoMock []model.EventDesc

func (this repoMock) Find(string, string) ([]model.EventDesc, error) {
	return this, nil
}

// the local event-worker reads the descriptions as json. a worker from before the aspect
// lists only knows aspect_id, a current one reads aspect_ids and folds aspect_id into it.
func TestEventsAspectsJson(t *testing.T) {
	server := httptest.NewServer(GetRouter(configuration.Config{DisableEventApiHttpLogger: true}, repoMock{
		{EventId: "deprecated", AspectId: "a1"},
		{EventId: "list", AspectIds: []string{"a1", "a2"}},
	}))
	defer server.Close()

	resp, err := http.Get(server.URL + "/event-descriptions?" + url.Values{"local_device_id": {"ldid1"}, "local_service_id": {"lsid1"}}.Encode())
	if err != nil {
		t.Error(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Error(resp.StatusCode)
		return
	}
	raw := []map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&raw)
	if err != nil {
		t.Error(err)
		return
	}
	if len(raw) != 2 {
		t.Error(raw)
		return
	}
	if raw[0]["aspect_id"] != "a1" {
		t.Error(raw[0])
	}
	if _, ok := raw[0]["aspect_ids"]; ok {
		t.Error("empty aspect list should be omitted", raw[0])
	}
	if !reflect.DeepEqual(raw[1]["aspect_ids"], []interface{}{"a1", "a2"}) {
		t.Error(raw[1])
	}
}
