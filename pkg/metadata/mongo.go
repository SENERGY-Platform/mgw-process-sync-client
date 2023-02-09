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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"log"
	"reflect"
	"strings"
	"time"
)

const DeploymentMetadataMongoCollection = "deployment_metadata"

func NewMongoStorage(ctx context.Context, config configuration.Config) (storage Storage, err error) {
	connStr := config.DeploymentMetadataStorage
	connStrObj, err := connstring.ParseAndValidate(connStr)
	if err != nil {
		return storage, err
	}
	timeout, _ := context.WithTimeout(ctx, 10*time.Second)
	client, err := mongo.Connect(timeout, options.Client().ApplyURI(connStr))
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		timeout, _ := getTimeoutContext()
		log.Println("close mongo connection", client.Disconnect(timeout))
	}()
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

type MongoStorage struct {
	client      *mongo.Client
	database    string
	config      configuration.Config
	idFieldName string
}

func (this *MongoStorage) Read(deploymentId string) (result Metadata, err error) {
	ctx, _ := getTimeoutContext()
	err = this.getCollection().FindOne(ctx, bson.M{this.idFieldName: deploymentId}).Decode(&result)
	return
}

func (this *MongoStorage) IsPlaceholder() bool {
	return false
}

func (this *MongoStorage) Init() (err error) {
	this.idFieldName, err = this.initCollectionIndex(this.getCollection(), "camunda_deployment_id_index", Metadata{}, "CamundaDeploymentId")
	return err
}

func (this *MongoStorage) Store(metadata Metadata) error {
	ctx, _ := getTimeoutContext()
	_, err := this.getCollection().ReplaceOne(
		ctx,
		bson.M{
			this.idFieldName: metadata.CamundaDeploymentId,
		},
		metadata,
		options.Replace().SetUpsert(true))
	return err
}

func (this *MongoStorage) Remove(camundaDeploymentId string) (err error) {
	ctx, _ := getTimeoutContext()
	_, err = this.getCollection().DeleteOne(
		ctx,
		bson.M{
			this.idFieldName: camundaDeploymentId,
		})
	return err
}

// removes unknown deployments
// returns all all known deployment metadata
func (this *MongoStorage) EnsureKnownDeployments(knownCamundaDeploymentIds []string) (known []Metadata, err error) {
	ctx, _ := getTimeoutContext()
	collection := this.getCollection()
	_, err = collection.DeleteMany(
		ctx,
		bson.M{
			this.idFieldName: bson.M{"$nin": knownCamundaDeploymentIds},
		})
	if err != nil {
		return
	}
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		element := Metadata{}
		err = cursor.Decode(&element)
		if err != nil {
			return nil, err
		}
		known = append(known, element)
	}
	err = cursor.Err()
	return
}

func (this *MongoStorage) List() (known []Metadata, err error) {
	ctx, _ := getTimeoutContext()
	collection := this.getCollection()
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		element := Metadata{}
		err = cursor.Decode(&element)
		if err != nil {
			return nil, err
		}
		known = append(known, element)
	}
	err = cursor.Err()
	return
}

func (this *MongoStorage) getCollection() (collection *mongo.Collection) {
	return this.client.Database(this.database).Collection(DeploymentMetadataMongoCollection)
}

func (this *MongoStorage) initCollectionIndex(collection *mongo.Collection, indexname string, obj interface{}, fieldName string) (bsonPath string, err error) {
	bsonPath, err = getBsonFieldPath(obj, fieldName)
	if err != nil {
		return
	}
	err = this.ensureIndex(collection, indexname, bsonPath, true, false)
	return
}

func (this *MongoStorage) ensureIndex(collection *mongo.Collection, indexname string, indexKey string, asc bool, unique bool) error {
	ctx, _ := getTimeoutContext()
	var direction int32 = -1
	if asc {
		direction = 1
	}
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bsonx.Doc{{indexKey, bsonx.Int32(direction)}},
		Options: options.Index().SetName(indexname).SetUnique(unique),
	})
	return err
}

func getBsonFieldPath(obj interface{}, path string) (bsonPath string, err error) {
	t := reflect.TypeOf(obj)
	pathParts := strings.Split(path, ".")
	bsonPathParts := []string{}
	for _, name := range pathParts {
		field, found := t.FieldByName(name)
		if !found {
			return "", errors.New("field path '" + path + "' not found at '" + name + "'")
		}
		tags, err := bsoncodec.DefaultStructTagParser.ParseStructTags(field)
		if err != nil {
			return bsonPath, err
		}
		bsonPathParts = append(bsonPathParts, tags.Name)
		t = field.Type
	}
	bsonPath = strings.Join(bsonPathParts, ".")
	return
}
