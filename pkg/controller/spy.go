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

package controller

import (
	"context"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/pglistener"
	"github.com/lib/pq"
	"log"
	"runtime/debug"
)

func (this *Controller) spyOnCamundaDb(ctx context.Context) (err error) {
	err = this.spyOn(ctx, "history", "ACT_HI_PROCINST", this.NotifyHistoryUpdate, this.NotifyHistoryDelete)
	if err != nil {
		return err
	}
	err = this.spyOn(ctx, "instance", "ACT_RU_EXECUTION", this.NotifyInstanceUpdate, this.NotifyInstanceDelete)
	if err != nil {
		return err
	}
	err = this.spyOn(ctx, "deployment", "ACT_RE_DEPLOYMENT", this.NotifyDeploymentUpdate, this.NotifyDeploymentDelete)
	if err != nil {
		return err
	}
	err = this.spyOn(ctx, "definition", "ACT_RE_PROCDEF", this.NotifyProcessDefUpdate, this.NotifyProcessDefDelete)
	if err != nil {
		return err
	}
	return nil
}

func (this *Controller) spyOn(ctx context.Context, channelName string, table string, notifySet func(string), notifyDelete func(string)) error {
	setChannel := "senergy_" + channelName + "_set"
	deleteChannel := "senergy_" + channelName + "_delete"
	err := pglistener.RegisterNotifier(this.config.CamundaDb, setChannel, deleteChannel, table)
	if err != nil {
		return err
	}
	notifySetChan, err := pglistener.Listen(ctx, this.config.CamundaDb, setChannel, func(event pq.ListenerEventType, err error) {
		if err != nil {
			debug.PrintStack()
			log.Fatal("FATAL:", err)
		}
	})
	if err != nil {
		return err
	}
	go func() {
		for n := range notifySetChan {
			notifySet(n.Extra)
		}
	}()

	notifyDeleteChan, err := pglistener.Listen(ctx, this.config.CamundaDb, deleteChannel, func(event pq.ListenerEventType, err error) {
		if err != nil {
			debug.PrintStack()
			log.Fatal("FATAL:", err)
		}
	})
	if err != nil {
		return err
	}
	go func() {
		for n := range notifyDeleteChan {
			notifyDelete(n.Extra)
		}
	}()

	return nil
}
