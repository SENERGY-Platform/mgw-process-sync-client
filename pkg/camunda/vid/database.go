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

package vid

import (
	"database/sql"

	_ "github.com/lib/pq"
)

var CreateVidTable = `CREATE TABLE IF NOT EXISTS VidRelation (
	ID					SERIAL PRIMARY KEY,
	DeploymentId		VARCHAR(255),
	VirtualId			VARCHAR(255)
);
CREATE INDEX IF NOT EXISTS vid_index ON VidRelation (VirtualId);
CREATE INDEX IF NOT EXISTS did_index ON VidRelation (DeploymentId);
`

type DbInterface interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func InitDb(pgConn string) (db *sql.DB, err error) {
	db, err = sql.Open("postgres", pgConn)
	if err != nil {
		return
	}
	_, err = db.Exec(CreateVidTable)
	return db, err
}
