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

package configuration

import "testing"

func TestEnvironmentVariableSetsIntField(t *testing.T) {
	t.Setenv("HISTORY_CLEANUP_BATCH_SIZE", "42")
	config := Config{HistoryCleanupBatchSize: 100}
	handleEnvironmentVars(&config)
	if config.HistoryCleanupBatchSize != 42 {
		t.Fatalf("expected 42, got %v", config.HistoryCleanupBatchSize)
	}
}

func TestEnvironmentVariableSetsInt64Field(t *testing.T) {
	t.Setenv("INCIDENT_RESTART_LIMIT", "7")
	config := Config{IncidentRestartLimit: 3}
	handleEnvironmentVars(&config)
	if config.IncidentRestartLimit != 7 {
		t.Fatalf("expected 7, got %v", config.IncidentRestartLimit)
	}
}
