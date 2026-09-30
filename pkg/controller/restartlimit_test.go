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

package controller

import (
	"io"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/camundamodel"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func incidentOf(definition string, businessKey string) camundamodel.Incident {
	return camundamodel.Incident{ProcessDefinitionId: definition, BusinessKey: businessKey, ProcessInstanceId: "instance-changes-with-every-restart"}
}

func record(limit *RestartLimit, incident camundamodel.Incident, times int) (blocked []bool) {
	for range times {
		blocked = append(blocked, limit.RecordIncident(incident))
	}
	return blocked
}

func TestRestartLimitBlocksProcessAfterMoreIncidentsThanTheLimit(t *testing.T) {
	limit := NewRestartLimit(3, "1h", discardLogger)
	got := record(limit, incidentOf("def", "bk"), 5)
	expected := []bool{false, false, false, true, true}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
	if !limit.Blocked("def", "bk") {
		t.Fatal("expected starts of the blocked process to be rejected")
	}
}

func TestRestartLimitCountsBusinessKeysAndDefinitionsSeparately(t *testing.T) {
	limit := NewRestartLimit(1, "1h", discardLogger)
	record(limit, incidentOf("def", "bk1"), 2)
	record(limit, incidentOf("def", "bk2"), 1)
	record(limit, incidentOf("other", "bk1"), 1)
	if !limit.Blocked("def", "bk1") {
		t.Fatal("expected def/bk1 to be blocked")
	}
	if limit.Blocked("def", "bk2") || limit.Blocked("other", "bk1") {
		t.Fatal("expected processes below the limit to stay startable")
	}
}

func TestRestartLimitForgetsIncidentsOutsideTheWindow(t *testing.T) {
	limit := NewRestartLimit(1, "1h", discardLogger)
	now := time.Now()
	limit.now = func() time.Time { return now }
	limit.RecordIncident(incidentOf("def", "bk"))
	now = now.Add(time.Hour)
	if limit.RecordIncident(incidentOf("def", "bk")) {
		t.Fatal("expected an incident after the window to count as the first one")
	}
}

func TestRestartLimitKeepsBlockAfterTheWindow(t *testing.T) {
	limit := NewRestartLimit(1, "1h", discardLogger)
	now := time.Now()
	limit.now = func() time.Time { return now }
	record(limit, incidentOf("def", "bk"), 2)
	now = now.Add(24 * time.Hour)
	if !limit.Blocked("def", "bk") {
		t.Fatal("expected the block to outlast the window, the warden would restart the loop otherwise")
	}
}

func TestRestartLimitZeroNeverBlocks(t *testing.T) {
	limit := NewRestartLimit(0, "1h", discardLogger)
	for i, blocked := range record(limit, incidentOf("def", "bk"), 10) {
		if blocked {
			t.Fatalf("expected incident %v not to block", i)
		}
	}
	if limit.Blocked("def", "bk") {
		t.Fatal("expected no block without limit")
	}
}

func TestRestartLimitFallsBackToDefaultWindowOnInvalidValue(t *testing.T) {
	limit := NewRestartLimit(1, "not-a-duration", discardLogger)
	if limit.Window() != defaultRestartWindow {
		t.Fatalf("expected %v, got %v", defaultRestartWindow, limit.Window())
	}
}
