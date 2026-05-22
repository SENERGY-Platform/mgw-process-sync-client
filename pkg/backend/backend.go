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

package backend

import (
	"context"
	"encoding/json"

	eventmodel "github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/model/camundamodel"
	paho "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	mqtt    paho.Client
	debug   bool
	config  configuration.Config
	handler Handler
}

type Handler interface {
	DeleteProcessInstanceHistory(id string) error
	DeleteProcessInstance(id string) error
	DeleteDeployment(id string) error
	StartDeployment(id string, businessKey string, parameter map[string]interface{}) error
	CreateDeployment(payload model.FogDeploymentMessage) (id string, err error)
	UpdateDeploymentEvents(camundaDeploymentId string, descriptions []eventmodel.EventDesc, id map[string]string, localId map[string]string) error
	HandleIncident(incident camundamodel.Incident) error
}

func New(config configuration.Config, ctx context.Context, handler Handler) (*Client, error) {
	client := &Client{
		config:  config,
		debug:   config.Debug,
		handler: handler,
	}
	options := paho.NewClientOptions().
		SetPassword(config.MqttPw).
		SetUsername(config.MqttUser).
		SetAutoReconnect(true).
		SetCleanSession(false).
		SetClientID(config.MqttClientId).
		AddBroker(config.MqttBroker).
		SetResumeSubs(true).
		SetConnectionLostHandler(func(_ paho.Client, err error) {
			config.GetLogger().Error("connection to mqtt broker lost", "error", err)
		}).
		SetOnConnectHandler(func(m paho.Client) {
			config.GetLogger().Info("(re)connected to mqtt broker")
			client.subscribe()
		})

	if config.MqttFileStoreLocation != "" {
		options = options.SetStore(paho.NewFileStore(config.MqttFileStoreLocation))
	}

	client.mqtt = paho.NewClient(options)
	if token := client.mqtt.Connect(); token.Wait() && token.Error() != nil {
		config.GetLogger().Error("error on mqtt connect", "error", token.Error())
		return nil, token.Error()
	}

	go func() {
		<-ctx.Done()
		client.mqtt.Disconnect(0)
	}()

	return client, nil
}

func (this *Client) subscribe() {
	this.mqtt.Subscribe(this.getDeploymentTopic(), 2, func(client paho.Client, message paho.Message) {
		this.config.GetLogger().Debug("recieve", "topic", message.Topic(), "payload", string(message.Payload()))
		go this.handleDeploymentCommand(message)
	})
	this.mqtt.Subscribe(this.getDeploymentDeleteTopic(), 2, func(client paho.Client, message paho.Message) {
		this.config.GetLogger().Debug("recieved", "topic", message.Topic(), "payload", string(message.Payload()))
		go this.handleDeploymentDeleteCommand(message)
	})
	this.mqtt.Subscribe(this.getProcessDeploymentStartTopic(), 2, func(client paho.Client, message paho.Message) {
		this.config.GetLogger().Debug("recieved", "topic", message.Topic(), "payload", string(message.Payload()))
		go this.handleDeploymentStartCommand(message)
	})
	this.mqtt.Subscribe(this.getProcessEventUpdateTopic(), 2, func(client paho.Client, message paho.Message) {
		this.config.GetLogger().Debug("recieved", "topic", message.Topic(), "payload", string(message.Payload()))
		go this.handleEventUpdateCommand(message)
	})
	this.mqtt.Subscribe(this.getProcessStopTopic(), 2, func(client paho.Client, message paho.Message) {
		this.config.GetLogger().Debug("recieved", "topic", message.Topic(), "payload", string(message.Payload()))
		go this.handleProcessStopCommand(message)
	})
	this.mqtt.Subscribe(this.getProcessHistoryDeleteTopic(), 2, func(client paho.Client, message paho.Message) {
		this.config.GetLogger().Debug("recieved", "topic", message.Topic(), "payload", string(message.Payload()))
		go this.handleProcessHistoryDeleteCommand(message)
	})
	this.mqtt.Subscribe(this.getProcessIncidentTopic(), 2, func(client paho.Client, message paho.Message) {
		this.config.GetLogger().Debug("recieved", "topic", message.Topic(), "payload", string(message.Payload()))
		go this.handleProcessIncident(message)
	})
}

func (this *Client) getBaseTopic() string {
	if this.config.NetworkId != "" {
		return "processes/" + this.config.NetworkId
	} else {
		return "processes"
	}
}

func (this *Client) getCommandTopic(entity string, subcommand ...string) (topic string) {
	topic = this.getBaseTopic() + "/cmd/" + entity
	for _, sub := range subcommand {
		topic = topic + "/" + sub
	}
	return
}

func (this *Client) getStateTopic(entity string, substate ...string) (topic string) {
	topic = this.getBaseTopic() + "/state/" + entity
	for _, sub := range substate {
		topic = topic + "/" + sub
	}
	return
}

func (this *Client) sendObj(topic string, message interface{}) (err error) {
	var msg []byte
	switch v := message.(type) {
	case error:
		msg = []byte(v.Error())
	default:
		msg, err = json.Marshal(message)
		if err != nil {
			return err
		}
	}
	this.config.GetLogger().Debug("sendObj", "topic", topic, "payload", string(msg))
	token := this.mqtt.Publish(topic, 2, false, msg)
	token.Wait()
	return token.Error()
}

func (this *Client) sendStr(topic string, message string) error {
	this.config.GetLogger().Debug("sendObj", "topic", topic, "payload", message)
	token := this.mqtt.Publish(topic, 2, false, message)
	token.Wait()
	return token.Error()
}

func (this *Client) GetMqttClient() paho.Client {
	return this.mqtt
}
