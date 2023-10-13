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

package main

import (
	"context"
	"flag"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/controller"
	cleanup "github.com/SENERGY-Platform/process-history-cleanup/pkg"
	cleanupconfig "github.com/SENERGY-Platform/process-history-cleanup/pkg/configuration"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

func main() {
	time.Sleep(5 * time.Second) //wait for routing tables in cluster

	confLocation := flag.String("config", "config.json", "configuration file")
	flag.Parse()

	config, err := configuration.Load(*confLocation)
	if err != nil {
		log.Fatal("ERROR: unable to load config ", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	_, err = controller.New(config, ctx)
	if err != nil {
		debug.PrintStack()
		log.Fatal("FATAL:", err)
	}

	historyCleanup(ctx, config)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	sig := <-shutdown
	log.Println("received shutdown signal", sig)
	cancel()
	time.Sleep(1 * time.Second) //give connections time to close gracefully
}

func historyCleanup(ctx context.Context, config configuration.Config) {
	if config.HistoryCleanupInterval != "" {
		interval, err := time.ParseDuration(config.HistoryCleanupInterval)
		if err != nil {
			log.Println("WARNING: unable to parse history cleanup interval duration", config.HistoryCleanupInterval, err)
		} else {
			ticker := time.NewTicker(interval)
			go func() {
				done := ctx.Done()
				for {
					select {
					case <-done:
						return
					case <-ticker.C:
						log.Println("start history cleanup")
						err := cleanup.RunCleanup(&cleanupconfig.ConfigStruct{
							EngineUrl:     config.CamundaUrl,
							MaxAge:        config.HistoryCleanupMaxAge,
							BatchSize:     config.HistoryCleanupBatchSize,
							FilterLocally: config.HistoryCleanupFilterLocally,
							Location:      config.HistoryCleanupLocation,
							Debug:         config.Debug,
						})
						if err != nil {
							log.Println("ERROR: in history cleanup:", err)
						}
					}
				}
			}()
		}
	}
}
