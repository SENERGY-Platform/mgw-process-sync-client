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
	"errors"
	"log/slog"
)

type VoidStorage struct {
	Debug bool
}

func (this VoidStorage) Read(deploymentId string) (Metadata, error) {
	return Metadata{}, errors.New("metadata storage disabled")
}

func (this VoidStorage) IsPlaceholder() bool {
	return true
}

func (this VoidStorage) Store(metadata Metadata) error {
	slog.Debug("try to store metadata, no storage is used")
	return nil
}

func (this VoidStorage) Remove(camundaDeploymentId string) (err error) {
	slog.Debug("try to remove metadata, no storage is used")
	return nil
}

func (this VoidStorage) EnsureKnownDeployments(knownCamundaDeploymentIds []string) (known []Metadata, err error) {
	slog.Debug("try to retrieve known metadata, no storage is used")
	return []Metadata{}, nil
}

func (this VoidStorage) List() (known []Metadata, err error) {
	slog.Debug("try to list metadata, no storage is used")
	return nil, nil
}

func (this VoidStorage) GetInstanceParameter(businessKey string) (map[string]interface{}, error) {
	slog.Debug("try to get instance parameter, no storage is used")
	return nil, errors.New("metadata storage disabled")
}

func (this VoidStorage) StoreInstanceParameter(businessKey string, params map[string]interface{}) error {
	slog.Debug("try to store instance parameter, no storage is used")
	return nil
}

func (this VoidStorage) RemoveInstanceParameter(businessKey string) error {
	slog.Debug("try to remove instance parameter, no storage is used")
	return nil
}

func (this VoidStorage) ListInstanceParameterBusinessKeys() (businessKeys []string, err error) {
	slog.Debug("try to list instance parameter business keys, no storage is used")
	return []string{}, nil
}
