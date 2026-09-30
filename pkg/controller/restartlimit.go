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
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/camundamodel"
)

const defaultRestartWindow = time.Hour

var ErrRestartLimitReached = errors.New("process start rejected: restart limit reached")

// RestartLimit bounds how often a process is restarted after incidents.
// A restart creates a new process instance, so incidents are counted per process definition and business key:
// a process that fails the same way on every start would otherwise be restarted (and notified) forever.
// Restarts come from the local incident handling and from the process-sync warden, which starts the process
// again with the same business key. Every restart is preceded by an incident, so counting incidents covers both.
// A process with more than limit incidents within the window is blocked. The block does not expire with the window,
// because the warden keeps retrying and would restart the loop; it lasts until the service restarts.
type RestartLimit struct {
	limit     int64
	window    time.Duration
	now       func() time.Time
	mux       sync.Mutex
	incidents map[string][]time.Time
	blocked   map[string]bool
}

func NewRestartLimit(limit int64, window string, logger *slog.Logger) *RestartLimit {
	duration, err := time.ParseDuration(window)
	if err != nil || duration <= 0 {
		logger.Warn("unable to parse incident restart window; use default", "incident_restart_window", window, "default", defaultRestartWindow.String(), "error", err)
		duration = defaultRestartWindow
	}
	return &RestartLimit{
		limit:     limit,
		window:    duration,
		now:       time.Now,
		incidents: map[string][]time.Time{},
		blocked:   map[string]bool{},
	}
}

// RecordIncident counts the incident and reports whether its process is blocked from being started again.
func (this *RestartLimit) RecordIncident(incident camundamodel.Incident) (blocked bool) {
	if this.limit <= 0 {
		return false
	}
	this.mux.Lock()
	defer this.mux.Unlock()
	key := restartLimitKey(incident.ProcessDefinitionId, incident.BusinessKey)
	if this.blocked[key] {
		return true
	}
	now := this.now()
	for k, times := range this.incidents {
		recent := times[:0]
		for _, t := range times {
			if now.Sub(t) < this.window {
				recent = append(recent, t)
			}
		}
		if len(recent) == 0 {
			delete(this.incidents, k)
		} else {
			this.incidents[k] = recent
		}
	}
	this.incidents[key] = append(this.incidents[key], now)
	if int64(len(this.incidents[key])) > this.limit {
		delete(this.incidents, key)
		this.blocked[key] = true
		return true
	}
	return false
}

// Blocked reports whether starts of the process with this business key are rejected.
func (this *RestartLimit) Blocked(processDefinitionId string, businessKey string) bool {
	if this.limit <= 0 {
		return false
	}
	this.mux.Lock()
	defer this.mux.Unlock()
	return this.blocked[restartLimitKey(processDefinitionId, businessKey)]
}

func (this *RestartLimit) Limit() int64 {
	return this.limit
}

func (this *RestartLimit) Window() time.Duration {
	return this.window
}

func restartLimitKey(processDefinitionId string, businessKey string) string {
	return processDefinitionId + "/" + businessKey
}
