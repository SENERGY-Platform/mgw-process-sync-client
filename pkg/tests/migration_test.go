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

package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/camunda"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/controller"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/docker"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/helper"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/server"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	paho "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/lib/pq"
)

// process-engine images before and after the camunda update; same pair the
// camunda-engine-wrapper migration test uses
const camundaTagBeforeUpdate = "v1.0.4"
const camundaTagAfterUpdate = "v1.0.5"

// TestMigration replaces the camunda container while its database stays in place and
// checks that camunda.UpdateDatabaseSchema brings the schema along: state created before
// the update is still readable afterwards and new deployments still work.
func TestMigration(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//postgres, mqtt and mongo outlive the camunda update, only camunda is replaced
	config := configuration.Config{
		InitialWaitDuration: "1s",
		Debug:               true,
		NetworkId:           "test-network-id",
		MqttClientId:        "test-sync-client",
	}

	var camundaPgIp string
	var err error
	config.CamundaDb, camundaPgIp, _, err = docker.PostgresWithNetwork(ctx, wg, "camunda")
	if err != nil {
		t.Error(err)
		return
	}

	_, mqttIp, err := docker.Mqtt(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	config.MqttBroker = fmt.Sprintf("tcp://%s:%s", mqttIp, "1883")

	mongoPort, _, err := docker.Mongo(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	config.DeploymentMetadataStorage = "mongodb://localhost:" + mongoPort + "/metadata"

	mqtt := paho.NewClient(paho.NewClientOptions().
		SetPassword(config.MqttPw).
		SetUsername(config.MqttUser).
		SetAutoReconnect(true).
		SetCleanSession(false).
		SetClientID("test-client").
		AddBroker(config.MqttBroker))
	if token := mqtt.Connect(); token.Wait() && token.Error() != nil {
		t.Error(token.Error())
		return
	}

	mqttMessages := map[string][]string{}
	mqttmux := sync.Mutex{}
	mqtt.Subscribe("processes/#", 2, func(client paho.Client, message paho.Message) {
		mqttmux.Lock()
		defer mqttmux.Unlock()
		mqttMessages[message.Topic()] = append(mqttMessages[message.Topic()], string(message.Payload()))
	})

	//----- before the camunda update -----

	camundaCtx, stopCamunda := context.WithCancel(ctx)
	defer stopCamunda()
	ctrlCtx, stopCtrl := context.WithCancel(ctx)
	defer stopCtrl()

	beforeConfig := config
	beforeConfig.CamundaUrl, err = docker.CamundaWithTag(camundaCtx, wg, camundaPgIp, "5432", camundaTagBeforeUpdate)
	if err != nil {
		t.Error(err)
		return
	}
	beforeConfig.EventApiPort, err = docker.GetFreePortStr()
	if err != nil {
		t.Error(err)
		return
	}
	//the schema camunda brought with it, before the client had a chance to migrate anything
	t.Run("schema before the client starts", func(t *testing.T) {
		versions, err := getCamundaSchemaLogVersions(config.CamundaDb)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log("schema versions of", camundaTagBeforeUpdate+":", versions)
		if len(versions) == 0 {
			t.Error("expect camunda to have created its schema")
		}
		if slices.Contains(versions, "7.24.0") {
			t.Error("expect a schema older than 7.24 before the update, otherwise this test proves nothing", versions)
		}
	})

	_, err = controller.New(beforeConfig, ctrlCtx)
	if err != nil {
		t.Error(err)
		return
	}

	//controller.New calls camunda.UpdateDatabaseSchema, so the schema is migrated by now
	t.Run("client migrated the schema on start", func(t *testing.T) {
		versions, err := getCamundaSchemaLogVersions(config.CamundaDb)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log("schema versions after the client started:", versions)
		if !slices.Contains(versions, "7.24.0") {
			t.Error("expect schema version 7.24.0 after the client started", versions)
		}
	})

	t.Run("columns added by the migration exist", func(t *testing.T) {
		for table, column := range map[string]string{
			"act_hi_comment":   "rev_",
			"act_ru_execution": "proc_def_key_",
		} {
			exists, err := camundaColumnExists(config.CamundaDb, table, column)
			if err != nil {
				t.Error(err)
				return
			}
			if !exists {
				t.Error("expect column after the migration", table, column)
			}
		}
	})

	t.Run("running the migration again changes nothing", func(t *testing.T) {
		err := camunda.UpdateDatabaseSchema(config)
		if err != nil {
			t.Error(err)
			return
		}
		versions, err := getCamundaSchemaLogVersions(config.CamundaDb)
		if err != nil {
			t.Error(err)
			return
		}
		if got := countOccurrences(versions, "7.24.0"); got != 1 {
			t.Error("expect exactly one 7.24.0 entry, a second run must not insert again", got, versions)
		}
	})

	t.Run("create deployment before update", createTestDeployment(beforeConfig, mqtt, "before-update", helper.BpmnExample, helper.SvgExample))

	t.Run("wait", func(t *testing.T) { time.Sleep(5 * time.Second) })

	deploymentId := ""
	t.Run("get deployment id before update", func(t *testing.T) {
		mqttmux.Lock()
		defer mqttmux.Unlock()
		deployments := mqttMessages["processes/"+config.NetworkId+"/state/deployment"]
		if len(deployments) == 0 {
			t.Error("expect deployment state message before the update")
			return
		}
		depl := deploymentmodel.Deployment{}
		err := json.Unmarshal([]byte(deployments[0]), &depl)
		if err != nil {
			t.Error(err)
			return
		}
		deploymentId = depl.Id
		if deploymentId == "" {
			t.Error("expect deployment id")
		}
	})

	t.Run("start deployment before update", startTestDeployment(beforeConfig, mqtt, &deploymentId))

	t.Run("wait", func(t *testing.T) { time.Sleep(5 * time.Second) })

	t.Run("check instance before update", func(t *testing.T) {
		mqttmux.Lock()
		defer mqttmux.Unlock()
		if len(mqttMessages["processes/"+config.NetworkId+"/state/process-instance"]) == 0 &&
			len(mqttMessages["processes/"+config.NetworkId+"/state/process-instance-history"]) == 0 {
			t.Error("expect instance or history state message before the update")
		}
		if len(mqttMessages["processes/"+config.NetworkId+"/state/error"]) != 0 {
			t.Error("expect no error messages before the update", mqttMessages["processes/"+config.NetworkId+"/state/error"])
		}
	})

	//----- the camunda update -----

	t.Run("stop camunda and client", func(t *testing.T) {
		stopCtrl()
		stopCamunda()
		//the replacement must not run against the old container while it is still terminating
		time.Sleep(5 * time.Second)
	})

	t.Run("forget messages from before the update", func(t *testing.T) {
		mqttmux.Lock()
		defer mqttmux.Unlock()
		mqttMessages = map[string][]string{}
	})

	//----- after the camunda update -----

	afterConfig := config
	t.Run("start updated camunda and client", func(t *testing.T) {
		var err error
		afterConfig.CamundaUrl, err = docker.CamundaWithTag(ctx, wg, camundaPgIp, "5432", camundaTagAfterUpdate)
		if err != nil {
			t.Error(err)
			return
		}
		afterConfig.EventApiPort, err = docker.GetFreePortStr()
		if err != nil {
			t.Error(err)
			return
		}
		//runs UpdateDatabaseSchema again, this time as a no-op
		_, err = controller.New(afterConfig, ctx)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("wait", func(t *testing.T) { time.Sleep(5 * time.Second) })

	t.Run("schema untouched by the second client start", func(t *testing.T) {
		versions, err := getCamundaSchemaLogVersions(config.CamundaDb)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log("schema versions after the update:", versions)
		if got := countOccurrences(versions, "7.24.0"); got != 1 {
			t.Error("expect exactly one 7.24.0 entry", got, versions)
		}
	})

	t.Run("deployment survived the update", func(t *testing.T) {
		mqttmux.Lock()
		defer mqttmux.Unlock()
		known := mqttMessages["processes/"+config.NetworkId+"/state/deployment/known"]
		if len(known) == 0 {
			t.Error("expect known deployments after the update")
			return
		}
		ids := []string{}
		err := json.Unmarshal([]byte(known[len(known)-1]), &ids)
		if err != nil {
			t.Error(err)
			return
		}
		if !slices.Contains(ids, deploymentId) {
			t.Error("expect the deployment from before the update to still be known", deploymentId, ids)
		}
	})

	t.Run("create deployment after update", createTestDeployment(afterConfig, mqtt, "after-update", helper.BpmnExample, helper.SvgExample))

	t.Run("wait", func(t *testing.T) { time.Sleep(5 * time.Second) })

	newDeploymentId := ""
	t.Run("get deployment id after update", func(t *testing.T) {
		mqttmux.Lock()
		defer mqttmux.Unlock()
		deployments := mqttMessages["processes/"+config.NetworkId+"/state/deployment"]
		for _, msg := range deployments {
			depl := deploymentmodel.Deployment{}
			err := json.Unmarshal([]byte(msg), &depl)
			if err != nil {
				t.Error(err)
				return
			}
			if depl.Id != deploymentId {
				newDeploymentId = depl.Id
			}
		}
		if newDeploymentId == "" {
			t.Error("expect a new deployment after the update", deployments)
		}
	})

	t.Run("start deployment after update", startTestDeployment(afterConfig, mqtt, &newDeploymentId))

	t.Run("wait", func(t *testing.T) { time.Sleep(5 * time.Second) })

	t.Run("check state after update", func(t *testing.T) {
		mqttmux.Lock()
		defer mqttmux.Unlock()
		if len(mqttMessages["processes/"+config.NetworkId+"/state/process-instance"]) == 0 &&
			len(mqttMessages["processes/"+config.NetworkId+"/state/process-instance-history"]) == 0 {
			t.Error("expect instance or history state message after the update")
		}
		if len(mqttMessages["processes/"+config.NetworkId+"/state/error"]) != 0 {
			t.Error("expect no error messages after the update", mqttMessages["processes/"+config.NetworkId+"/state/error"])
		}
	})

	t.Run("log mqtt messages", func(t *testing.T) {
		mqttmux.Lock()
		defer mqttmux.Unlock()
		temp, err := json.Marshal(mqttMessages)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(string(temp))
	})
}

func getCamundaSchemaLogVersions(camundaDb string) (versions []string, err error) {
	db, err := sql.Open("postgres", camundaDb)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(`select version_ from ACT_GE_SCHEMA_LOG order by id_;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		version := ""
		err = rows.Scan(&version)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	return versions, rows.Err()
}

func camundaColumnExists(camundaDb string, table string, column string) (exists bool, err error) {
	db, err := sql.Open("postgres", camundaDb)
	if err != nil {
		return false, err
	}
	defer db.Close()
	err = db.QueryRow(`select count(1) > 0 from information_schema.columns where table_name = $1 and column_name = $2;`, table, column).Scan(&exists)
	return exists, err
}

func countOccurrences(list []string, element string) (count int) {
	for _, e := range list {
		if e == element {
			count = count + 1
		}
	}
	return count
}

// TestMigrationOnFreshDatabase covers the start of a new mgw, where the service can come up
// before camunda has created its schema. UpdateDatabaseSchema fails there on purpose so the
// service is restarted: camunda stays a black box apart from the update scripts, and the
// tables the client works on would be missing at that point anyway.
func TestMigrationOnFreshDatabase(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := configuration.Config{Debug: true}
	var err error
	config.CamundaDb, _, _, err = docker.PostgresWithNetwork(ctx, wg, "camunda")
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("no schema log yet", func(t *testing.T) {
		versions, err := getCamundaSchemaLogVersions(config.CamundaDb)
		if err == nil {
			t.Error("expect the schema log to be missing on a fresh database", versions)
		}
	})

	t.Run("update schema fails", func(t *testing.T) {
		err := camunda.UpdateDatabaseSchema(config)
		if err == nil {
			t.Error("expect an error while camunda has not created its schema yet")
			return
		}
		t.Log("expected error:", err)
	})

	t.Run("nothing was created", func(t *testing.T) {
		_, err := getCamundaSchemaLogVersions(config.CamundaDb)
		if err == nil {
			t.Error("expect the failed update to have created no table")
		}
	})

	t.Run("without camunda db configured", func(t *testing.T) {
		for _, camundaDb := range []string{"", "-"} {
			err := camunda.UpdateDatabaseSchema(configuration.Config{CamundaDb: camundaDb, Debug: true})
			if err != nil {
				t.Error(camundaDb, err)
			}
		}
	})
}

// TestMigrationOnCurrentCamunda pins the version check against a fresh install of the image
// we currently run: camunda records only the version it was installed with, so a step must
// not fire just because its own entry is missing while a later one is present. against an
// up to date database the update has to be a complete no-op.
func TestMigrationOnCurrentCamunda(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, _, err := server.CreateCamundaEnv(ctx, wg, configuration.Config{Debug: true})
	if err != nil {
		t.Error(err)
		return
	}

	before, err := getCamundaSchemaLogVersions(config.CamundaDb)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("schema versions of a fresh camunda:", before)

	err = camunda.UpdateDatabaseSchema(config)
	if err != nil {
		t.Error(err)
		return
	}

	after, err := getCamundaSchemaLogVersions(config.CamundaDb)
	if err != nil {
		t.Error(err)
		return
	}
	if !slices.Equal(before, after) {
		t.Error("expect the schema log of an up to date database to be untouched", before, after)
	}
}
