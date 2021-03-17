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

import "log"

type VoidStorage struct {
	Debug bool
}

func (this VoidStorage) Store(metadata Metadata) error {
	if this.Debug {
		log.Println("DEBUG: try to store metadata, no storage is used")
	}
	return nil
}

func (this VoidStorage) Remove(camundaDeploymentId string) (err error) {
	if this.Debug {
		log.Println("DEBUG: try to remove metadata from storage, no storage is used")
	}
	return nil
}

func (this VoidStorage) EnsureKnownDeployments(knownCamundaDeploymentIds []string) (known []Metadata, err error) {
	if this.Debug {
		log.Println("DEBUG: try to retrieve known metadata, no storage is used")
	}
	return []Metadata{}, nil
}
