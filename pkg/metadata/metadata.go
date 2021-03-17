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
	"errors"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"strings"
	"time"
)

func NewStorage(config configuration.Config) (storage Storage, err error) {
	if config.DeploymentMetadataStorage == "" {
		return VoidStorage{Debug: config.Debug}, nil
	}
	if strings.HasPrefix(config.DeploymentMetadataStorage, "mongodb://") {
		return NewMongoStorage(config)
	}
	return nil, errors.New("unknown storage connection string type")
}

func NewMongoStorage(config configuration.Config) (storage Storage, err error) {
	connStr := config.DeploymentMetadataStorage
	connStrObj, err := connstring.ParseAndValidate(connStr)
	if err != nil {
		return storage, err
	}
	ctx, _ := getTimeoutContext()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
	if err != nil {
		return nil, err
	}
	m := &MongoStorage{
		config:   config,
		client:   client,
		database: connStrObj.Database,
	}
	return m, m.Init()
}

func getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
