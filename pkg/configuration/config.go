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

package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
	EventApiPort              string `json:"event_api_port"`
	DisableEventApiHttpLogger bool   `json:"disable_event_api_http_logger"`
	EnableSwaggerUi           bool   `json:"enable_swagger_ui"`

	CamundaDb  string `json:"camunda_db"`
	CamundaUrl string `json:"camunda_url"`

	DeploymentMetadataStorage string `json:"deployment_metadata_storage"`

	InitialWaitDuration string `json:"initial_wait_duration"`

	Debug bool `json:"debug"`

	MqttBroker            string `json:"mqtt_broker"`
	MqttClientId          string `json:"mqtt_client_id"`
	MqttUser              string `json:"mqtt_user" config:"secret"`
	MqttPw                string `json:"mqtt_pw" config:"secret"`
	MqttFileStoreLocation string `json:"mqtt_file_store_location"`
	NetworkId             string `json:"network_id"`
	FullUpdateInterval    string `json:"full_update_interval"`

	HistoryCleanupInterval      string `json:"history_cleanup_interval"`
	HistoryCleanupMaxAge        string `json:"history_cleanup_max_age"`
	HistoryCleanupBatchSize     int    `json:"history_cleanup_batch_size"`
	HistoryCleanupFilterLocally bool   `json:"history_cleanup_filter_locally"`
	HistoryCleanupLocation      string `json:"history_cleanup_location"`
	NotificationUrlPlaceholder  string `json:"notification_url_placeholder"`
	NotificationUrl             string `json:"notification_url"`

	TaskTopicReplace map[string]string `json:"task_topic_replace"`
}

// loads config from json in location and used environment variables (e.g ZookeeperUrl --> ZOOKEEPER_URL)
func Load(location string) (config Config, err error) {
	file, err := os.Open(location)
	if err != nil {
		return config, err
	}
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return config, err
	}
	handleEnvironmentVars(&config)
	return config, nil
}

var camel = regexp.MustCompile("(^[^A-Z]*|[A-Z]*)([A-Z][^A-Z]+|$)")

func fieldNameToEnvName(s string) string {
	var a []string
	for _, sub := range camel.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}
	return strings.ToUpper(strings.Join(a, "_"))
}

// preparations for docker
func handleEnvironmentVars(config *Config) {
	configValue := reflect.Indirect(reflect.ValueOf(config))
	configType := configValue.Type()
	for index := 0; index < configType.NumField(); index++ {
		fieldName := configType.Field(index).Name
		fieldConfig := configType.Field(index).Tag.Get("config")
		envName := fieldNameToEnvName(fieldName)
		envValue := os.Getenv(envName)
		if envValue != "" {
			loggedEnvValue := envValue
			if strings.Contains(fieldConfig, "secret") {
				loggedEnvValue = "***"
			}
			fmt.Println("use environment variable: ", envName, " = ", loggedEnvValue)
			if configValue.FieldByName(fieldName).Kind() == reflect.Int64 {
				i, _ := strconv.ParseInt(envValue, 10, 64)
				configValue.FieldByName(fieldName).SetInt(i)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.String {
				configValue.FieldByName(fieldName).SetString(envValue)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Bool {
				b, _ := strconv.ParseBool(envValue)
				configValue.FieldByName(fieldName).SetBool(b)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Float64 {
				f, _ := strconv.ParseFloat(envValue, 64)
				configValue.FieldByName(fieldName).SetFloat(f)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Slice {
				val := []string{}
				for _, element := range strings.Split(envValue, ",") {
					val = append(val, strings.TrimSpace(element))
				}
				configValue.FieldByName(fieldName).Set(reflect.ValueOf(val))
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Map {
				value := map[string]string{}
				for _, element := range strings.Split(envValue, ",") {
					keyVal := strings.Split(element, ":")
					key := strings.TrimSpace(keyVal[0])
					val := strings.TrimSpace(keyVal[1])
					value[key] = val
				}
				configValue.FieldByName(fieldName).Set(reflect.ValueOf(value))
			}
		}
	}
}
