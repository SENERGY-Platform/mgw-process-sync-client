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

package metadata

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/dgraph-io/badger/v3"
	"log"
)

var BADGER_PREFETCH = true

func NewBadgerStorage(ctx context.Context, config configuration.Config) (storage *Badger, err error) {
	storage = &Badger{}
	storage.db, err = badger.Open(badger.DefaultOptions(config.DeploymentMetadataStorage))
	if err == nil {
		go func() {
			<-ctx.Done()
			log.Println("close badger", storage.db.Close())
		}()
	}
	return
}

type Badger struct {
	db *badger.DB
}

func (this *Badger) Store(metadata Metadata) error {
	return this.db.Update(func(tx *badger.Txn) error {
		if metadata.CamundaDeploymentId == "" {
			return errors.New("missing CamundaDeploymentId")
		}
		value, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		return tx.Set([]byte(metadata.CamundaDeploymentId), value)
	})
}

func (this *Badger) Remove(camundaDeploymentId string) (err error) {
	return this.db.Update(func(tx *badger.Txn) error {
		return tx.Delete([]byte(camundaDeploymentId))
	})
}

func (this *Badger) EnsureKnownDeployments(knownCamundaDeploymentIds []string) (known []Metadata, err error) {
	err = this.db.Update(func(tx *badger.Txn) error {
		prefetchValues := map[string][]byte{}

		//get known ids
		knownIds := []string{}
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = BADGER_PREFETCH
		it := tx.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			id := string(item.Key())
			knownIds = append(knownIds, id)
			if BADGER_PREFETCH {
				err = item.Value(func(v []byte) error {
					prefetchValues[id] = v
					return nil
				})
				if err != nil {
					return err
				}
			}
		}

		//index of requested ids
		requestedIds := map[string]bool{}
		for _, id := range knownCamundaDeploymentIds {
			requestedIds[id] = true
		}

		//find ids to delete and ids to fetch
		deleteIds := []string{}
		fetchIds := []string{}
		for _, id := range knownIds {
			if requestedIds[id] {
				fetchIds = append(fetchIds, id)
			} else {
				deleteIds = append(deleteIds, id)
			}
		}

		//fetch
		for _, id := range fetchIds {
			temp := Metadata{}
			if BADGER_PREFETCH {
				err = json.Unmarshal(prefetchValues[id], &temp)
				if err != nil {
					return err
				}
			} else {
				item, err := tx.Get([]byte(id))
				if err != nil {
					return err
				}
				err = item.Value(func(val []byte) error {
					return json.Unmarshal(val, &temp)
				})
				if err != nil {
					return err
				}
			}
			known = append(known, temp)
		}

		//delete
		for _, id := range deleteIds {
			err = tx.Delete([]byte(id))
			if err != nil {
				return err
			}
		}

		return nil
	})
	return
}

func (this *Badger) Read(deploymentId string) (result Metadata, err error) {
	err = this.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get([]byte(deploymentId))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &result)
		})
	})
	return
}

func (this *Badger) IsPlaceholder() bool {
	return false
}
