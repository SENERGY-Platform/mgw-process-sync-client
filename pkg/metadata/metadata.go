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
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"log"
	"strings"
)

func NewStorage(ctx context.Context, config configuration.Config) (storage Storage, err error) {
	if config.DeploymentMetadataStorage == "" {
		log.Println("WARNING: metadata storage not used -> disable deployment of message-events")
		return VoidStorage{Debug: config.Debug}, nil
	}
	if strings.HasPrefix(config.DeploymentMetadataStorage, "mongodb://") {
		log.Println("use mongodb for metadata storage")
		return NewMongoStorage(ctx, config)
	}
	if strings.HasSuffix(config.DeploymentMetadataStorage, ".db") {
		log.Println("use bolt for metadata storage")
		return NewBoltStorage(ctx, config)
	}
	log.Println("use badger for metadata storage")
	return NewBadgerStorage(ctx, config)
}
