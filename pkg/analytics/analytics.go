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

package analytics

import (
	"encoding/json"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
	"runtime/debug"
	"time"
)

type Analytics struct {
	client paho.Client
}

func NewWithClient(client paho.Client) *Analytics {
	return &Analytics{client: client}
}

func (this *Analytics) Send(command ControlCommand) {
	temp, err := json.Marshal(command)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
	}
	token := this.client.Publish("fog/control", 2, false, temp)
	if !token.WaitTimeout(10 * time.Second) {
		log.Println("ERROR: timeout while trying to publish fog/control message")
		return
	}
	if err := token.Error(); err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
	}
}
