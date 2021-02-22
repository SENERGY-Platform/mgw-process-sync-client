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
	"bytes"
	"context"
	"database/sql"
	"github.com/lib/pq"
	"log"
	"text/template"
	"time"
)

const MIN_RECONN = 10 * time.Second
const MAX_RECONN = time.Minute

func Listen(ctx context.Context, pgUrl string, channel string, reportErr pq.EventCallbackType) (notifications chan *pq.Notification, err error) {
	listener := pq.NewListener(pgUrl, MIN_RECONN, MAX_RECONN, reportErr)
	err = listener.Listen(channel)
	if err != nil {
		return nil, err
	}

	go func() {
		defer listener.Close()
		<-ctx.Done()
	}()
	return listener.Notify, nil
}

func RegisterNotifier(pgUrl string, setChannel string, deleteChannel string, table string) (err error) {
	db, err := sql.Open("postgres", pgUrl)
	if err != nil {
		log.Printf("Failed to connect to '%s': %s", pgUrl, err)
		return err
	}
	defer db.Close()
	return registerNotifier(db, setChannel, deleteChannel, table)
}

func registerNotifier(db *sql.DB, setChannel string, deleteChannel string, table string) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(notifyNewFunctionSql)
	if err != nil {
		log.Println("ERROR: unable to create senergy_notify_new", err)
		return err
	}

	_, err = tx.Exec(notifyOldFunctionSql)
	if err != nil {
		log.Println("ERROR: unable to create senergy_notify_old", err)
		return err
	}

	triggerSql, err := templateToString(triggerTemplate, map[string]string{
		"ChannelSet":    setChannel,
		"ChannelDelete": deleteChannel,
		"Table":         table,
	})
	if err != nil {
		return err
	}

	_, err = tx.Exec(triggerSql)
	if err != nil {
		log.Println("ERROR: unable to create trigger", err, "\n", triggerSql)
		return err
	}
	return tx.Commit()
}

func templateToString(tmpl string, values map[string]string) (result string, err error) {
	var temp bytes.Buffer
	err = template.Must(template.New("").Parse(tmpl)).Execute(&temp, values)
	return temp.String(), err
}

const notifyNewFunctionSql = `
create or replace function senergy_notify_new()
 returns trigger
 language plpgsql
as $$
declare
  channel text := TG_ARGV[0];
  payload text := row_to_json(NEW)::text;
begin
  PERFORM (
     select pg_notify(channel, payload)
  );
  RETURN NULL;
end;
$$;
`

const notifyOldFunctionSql = `
create or replace function senergy_notify_old()
 returns trigger
 language plpgsql
as $$
declare
  channel text := TG_ARGV[0];
  payload text := row_to_json(OLD)::text;
begin
  PERFORM (
     select pg_notify(channel, payload)
  );
  RETURN NULL;
end;
$$;
`

const triggerTemplate = `
DROP TRIGGER IF EXISTS notify_{{.Table}}_set
  ON {{.Table}};

DROP TRIGGER IF EXISTS notify_{{.Table}}_delete
  ON {{.Table}};

CREATE TRIGGER notify_{{.Table}}_set
AFTER INSERT OR UPDATE
ON {{.Table}}
FOR EACH ROW
EXECUTE PROCEDURE senergy_notify_new('{{.ChannelSet}}');

CREATE TRIGGER notify_{{.Table}}_delete
AFTER DELETE
ON {{.Table}}
FOR EACH ROW
EXECUTE PROCEDURE senergy_notify_old('{{.ChannelDelete}}');
`

/*
const triggerSqlTemplate = `
begin;

create or replace function senergy.notify()
 returns trigger
 language plpgsql
as $$
declare
  channel text := TG_ARGV[0];
  payload text := TG_ARGV[1];
begin
  PERFORM (
     select pg_notify(channel, payload)
  );
  RETURN NULL;
end;
$$;

DROP TRIGGER IF EXISTS notify_{{.Table}}_set
  ON {{.Table}};

DROP TRIGGER IF EXISTS notify_{{.Table}}_delete
  ON {{.Table}};

CREATE TRIGGER notify_{{.Table}}_set
AFTER INSERT OR UPDATE
ON {{.Table}}
FOR EACH ROW
EXECUTE PROCEDURE senergy.notify('{{.ChannelSet}}', row_to_json(NEW)::text);

CREATE TRIGGER notify_{{.Table}}_delete
AFTER DELETE
ON {{.Table}}
FOR EACH ROW
EXECUTE PROCEDURE senergy.notify('{{.ChannelDelete}}', row_to_json(OLD)::text);

commit;
`
*/
