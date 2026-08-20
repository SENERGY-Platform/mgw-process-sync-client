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

package camunda

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"

	"github.com/lib/pq"
)

// UpdateDatabaseSchema brings the local camunda database up to the schema version the
// running camunda image expects. camunda does not apply minor-version schema changes on
// its own, so an updated camunda image on the mgw would otherwise find an outdated
// database. every step is guarded by its entry in ACT_GE_SCHEMA_LOG and therefore a no-op
// once it has been applied.
//
// if camunda has not created its schema yet, this returns an error and the service is
// restarted by its container. that is deliberate: apart from these update scripts camunda
// stays a black box, so we do not reason about how it reacts to a database that is only
// partially there. the tables spyOnCamundaDb needs would be missing at that point anyway.
//
// same approach as github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda/dbupdate.go;
// the ddl comes from camundas own migration scripts postgres_engine_7.22_to_7.23.sql and
// postgres_engine_7.23_to_7.24.sql (Camunda Services GmbH, Apache-2.0).
func UpdateDatabaseSchema(config configuration.Config) (err error) {
	if config.CamundaDb == "" || config.CamundaDb == "-" {
		return nil
	}
	logger := config.GetLogger()
	db, err := sql.Open("postgres", config.CamundaDb)
	if err != nil {
		logger.Error("could not open camunda db", "error", err)
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("could not begin camunda db transaction", "error", err)
		return err
	}

	err = updateDatabaseSchema_7_23(logger, tx)
	if err != nil {
		logger.Error("could not update camunda db schema to 7.23", "error", err)
		tx.Rollback()
		return err
	}

	err = updateDatabaseSchema_7_24(logger, tx)
	if err != nil {
		logger.Error("could not update camunda db schema to 7.24", "error", err)
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		logger.Error("could not commit camunda db transaction", "error", err)
		return err
	}
	return nil
}

// checkDatabaseSchemaVersion reports whether any of the given versions is recorded. a step
// passes its own version plus every later one, because a fresh camunda records only the
// version it was installed with: a new 7.24 database has no 7.23.0 entry even though it
// contains the 7.23 changes. a new step must therefore be added to the checks of all
// earlier steps as well.
func checkDatabaseSchemaVersion(db *sql.Tx, versions ...string) (hasVersion bool, err error) {
	err = db.QueryRow(`select COUNT(1) > 0 from ACT_GE_SCHEMA_LOG where version_ = any($1);`, pq.Array(versions)).Scan(&hasVersion)
	return hasVersion, err
}

func updateDatabaseSchema_7_23(logger *slog.Logger, db *sql.Tx) error {
	hasVersion, err := checkDatabaseSchemaVersion(db, "7.23.0", "7.24.0")
	if err != nil {
		return err
	}
	if hasVersion {
		return nil
	}

	logger.Info("updating camunda db schema to 7.23")

	_, err = db.Exec(`insert into ACT_GE_SCHEMA_LOG
values ('1200', CURRENT_TIMESTAMP, '7.23.0');`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`alter table ACT_HI_COMMENT
    add column if not exists REV_ integer not null
    default 1;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`alter table ACT_RU_EXECUTION add column if not exists PROC_DEF_KEY_ varchar(255);`)
	if err != nil {
		return err
	}

	return nil
}

func updateDatabaseSchema_7_24(logger *slog.Logger, db *sql.Tx) error {
	hasVersion, err := checkDatabaseSchemaVersion(db, "7.24.0")
	if err != nil {
		return err
	}
	if hasVersion {
		return nil
	}

	logger.Info("updating camunda db schema to 7.24")

	_, err = db.Exec(`insert into ACT_GE_SCHEMA_LOG
values ('1300', CURRENT_TIMESTAMP, '7.24.0');`)
	if err != nil {
		return err
	}

	return nil
}
