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
	"go.etcd.io/bbolt"
	"log"
	"time"
)

var BBOLT_BUCKET_NAME = []byte("metadata")

func NewBoltStorage(ctx context.Context, config configuration.Config) (storage *Bolt, err error) {
	storage = &Bolt{}
	storage.db, err = bbolt.Open(config.DeploymentMetadataStorage, 0666, &bbolt.Options{Timeout: 10 * time.Second})
	if err == nil {
		go func() {
			<-ctx.Done()
			log.Println("close bbolt", storage.db.Close())
		}()
	}
	err = storage.db.Update(func(tx *bbolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(BBOLT_BUCKET_NAME)
		return err
	})
	return
}

type Bolt struct {
	db *bbolt.DB
}

func (this *Bolt) Store(metadata Metadata) error {
	return this.db.Update(func(tx *bbolt.Tx) error {
		if metadata.CamundaDeploymentId == "" {
			return errors.New("missing CamundaDeploymentId")
		}
		value, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		return tx.Bucket(BBOLT_BUCKET_NAME).Put([]byte(metadata.CamundaDeploymentId), value)
	})
}

func (this *Bolt) Remove(camundaDeploymentId string) (err error) {
	return this.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(BBOLT_BUCKET_NAME).Delete([]byte(camundaDeploymentId))
	})
}

func (this *Bolt) EnsureKnownDeployments(knownCamundaDeploymentIds []string) (known []Metadata, err error) {
	err = this.db.Update(func(tx *bbolt.Tx) error {
		prefetchValues := map[string][]byte{}

		//get known ids
		knownIds := []string{}

		bucket := tx.Bucket(BBOLT_BUCKET_NAME)

		it := bucket.Cursor()
		for k, v := it.First(); k != nil; k, v = it.Next() {
			id := string(k)
			knownIds = append(knownIds, id)
			prefetchValues[id] = v
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
			err = json.Unmarshal(prefetchValues[id], &temp)
			if err != nil {
				return err
			}
			known = append(known, temp)
		}

		//delete
		for _, id := range deleteIds {
			err = bucket.Delete([]byte(id))
			if err != nil {
				return err
			}
		}

		return nil
	})
	return
}

func (this *Bolt) Read(deploymentId string) (result Metadata, err error) {
	err = this.db.View(func(tx *bbolt.Tx) error {
		return json.Unmarshal(tx.Bucket(BBOLT_BUCKET_NAME).Get([]byte(deploymentId)), &result)
	})
	return
}

func (this *Bolt) IsPlaceholder() bool {
	return false
}
