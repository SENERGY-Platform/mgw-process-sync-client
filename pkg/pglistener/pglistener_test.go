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

package pglistener

import (
	"context"
	"encoding/json"
	"github.com/lib/pq"
	"log"
	"mgw-process-sync/pkg/configuration"
	"mgw-process-sync/pkg/tests/helper"
	"mgw-process-sync/pkg/tests/server"
	"sync"
	"testing"
	"time"
)

func TestListen(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, camundaUrl, err := server.CreateCamundaEnv(ctx, wg, configuration.Config{})
	if err != nil {
		t.Error(err)
		return
	}

	const deploymentUpdateChannel = "senergy_depl_update"
	const deploymentDeleteChannel = "senergy_depl_delete"
	t.Run("add notifier ACT_RE_PROCDEF", func(t *testing.T) {
		err = RegisterNotifier(conf.CamundaDb, deploymentUpdateChannel, deploymentDeleteChannel, "ACT_RE_DEPLOYMENT")
		if err != nil {
			t.Error(err)
			return
		}
	})

	const processDefUpdateChannel = "senergy_procdef_update"
	const processDefDeleteChannel = "senergy_procdef_delete"
	t.Run("add notifier ACT_RE_PROCDEF", func(t *testing.T) {
		err = RegisterNotifier(conf.CamundaDb, processDefUpdateChannel, processDefDeleteChannel, "ACT_RE_PROCDEF")
		if err != nil {
			t.Error(err)
			return
		}
	})

	const instanceHistoryUpdateChannel = "senergy_instancehistory_update"
	const instanceHistoryDeleteChannel = "senergy_instancehistory_delete"
	t.Run("add notifier ACT_HI_PROCINST", func(t *testing.T) {
		err = RegisterNotifier(conf.CamundaDb, instanceHistoryUpdateChannel, instanceHistoryDeleteChannel, "ACT_HI_PROCINST")
		if err != nil {
			t.Error(err)
			return
		}
	})

	const executionUpdateChannel = "senergy_execution_update"
	const executionDeleteChannel = "senergy_execution_delete"
	t.Run("add notifier ACT_RU_EXECUTION", func(t *testing.T) {
		err = RegisterNotifier(conf.CamundaDb, executionUpdateChannel, executionDeleteChannel, "ACT_RU_EXECUTION")
		if err != nil {
			t.Error(err)
			return
		}
	})

	deplEvents := []map[string]interface{}{}
	t.Run("listen to ACT_RE_DEPLOYMENT", func(t *testing.T) {
		notifications, err := Listen(ctx, conf.CamundaDb, deploymentUpdateChannel, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Error(err, event)
			}
			return
		})
		if err != nil {
			t.Error(err)
			return
		}
		go func() {
			for n := range notifications {
				t.Log(n)
				definition := map[string]interface{}{}
				json.Unmarshal([]byte(n.Extra), &definition)
				deplEvents = append(deplEvents, definition)
			}
			log.Println("end of notifications in ACT_RE_DEPLOYMENT")
		}()
	})

	t.Run("listen to ACT_RE_DEPLOYMENT delete", func(t *testing.T) {
		notifications, err := Listen(ctx, conf.CamundaDb, deploymentDeleteChannel, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Error(err, event)
			}
			return
		})
		if err != nil {
			t.Error(err)
			return
		}
		go func() {
			for n := range notifications {
				t.Log(n)
			}
			log.Println("end of notifications in ACT_RE_DEPLOYMENT delete")
		}()
	})

	definitionEvents := []map[string]interface{}{}
	t.Run("listen to ACT_RE_PROCDEF", func(t *testing.T) {
		notifications, err := Listen(ctx, conf.CamundaDb, processDefUpdateChannel, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Error(err, event)
			}
			return
		})
		if err != nil {
			t.Error(err)
			return
		}
		go func() {
			for n := range notifications {
				t.Log(n)
				definition := map[string]interface{}{}
				json.Unmarshal([]byte(n.Extra), &definition)
				definitionEvents = append(definitionEvents, definition)
			}
			log.Println("end of notifications in ACT_RE_PROCDEF")
		}()
	})

	instanceHistoryEvents := []map[string]interface{}{}
	t.Run("listen to ACT_HI_PROCINST", func(t *testing.T) {
		notifications, err := Listen(ctx, conf.CamundaDb, instanceHistoryUpdateChannel, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Error(err, event)
			}
			return
		})
		if err != nil {
			t.Error(err)
			return
		}
		go func() {
			for n := range notifications {
				t.Log(n)
				element := map[string]interface{}{}
				json.Unmarshal([]byte(n.Extra), &element)
				instanceHistoryEvents = append(instanceHistoryEvents, element)
			}
			log.Println("end of notifications in ACT_HI_PROCINST")
		}()
	})

	executionEvents := []map[string]interface{}{}
	t.Run("listen to ACT_RU_EXECUTION", func(t *testing.T) {
		notifications, err := Listen(ctx, conf.CamundaDb, executionUpdateChannel, func(event pq.ListenerEventType, err error) {
			if err != nil {
				t.Error(err, event)
			}
			return
		})
		if err != nil {
			t.Error(err)
			return
		}
		go func() {
			for n := range notifications {
				t.Log(n)
				exec := map[string]interface{}{}
				json.Unmarshal([]byte(n.Extra), &exec)
				executionEvents = append(executionEvents, exec)
			}
			log.Println("end of notifications in ACT_RU_EXECUTION")
		}()
	})

	t.Run("create process", func(t *testing.T) {
		_, err = helper.DeployProcess(camundaUrl, "test", helper.BPMNWithTasksExample, helper.SvgExample, "user", "test")
		if err != nil {
			t.Error(err)
			return
		}
		time.Sleep(5 * time.Second)
	})

	t.Run("start process", func(t *testing.T) {
		if len(definitionEvents) < 1 {
			t.Error("expect 1 definition")
			return
		}
		id, ok := definitionEvents[0]["id_"].(string)
		if !ok {
			t.Error("expect id as string")
			return
		}
		err = helper.StartProcess(camundaUrl, id)
		if err != nil {
			t.Error(err)
			return
		}
		time.Sleep(5 * time.Second)
	})

	t.Run("remove process", func(t *testing.T) {
		if len(deplEvents) < 1 {
			t.Error("expect 1 instanceHistoryEvents")
			return
		}
		id, ok := deplEvents[0]["id_"].(string)
		if !ok {
			t.Error("expect id as string")
			return
		}
		err = helper.RemoveDeployment(camundaUrl, id)
		if err != nil {
			t.Error(err)
			return
		}
		time.Sleep(5 * time.Second)
	})

	t.Run("check events", func(t *testing.T) {
		time.Sleep(10 * time.Second)
		if len(definitionEvents) == 0 {
			t.Error(len(definitionEvents), definitionEvents)
		}

		if len(instanceHistoryEvents) == 0 {
			t.Error(len(instanceHistoryEvents), instanceHistoryEvents)
		}

		if len(executionEvents) == 0 {
			t.Error(len(executionEvents), executionEvents)
		}
	})

}
