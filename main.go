/*
 * Copyright (c) 2020 Devtron Labs
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
 *
 */

package main

import (
	"fmt"
	_ "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	_ "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	app, err := InitializeApp()
	if err != nil {
		log.Panic(err)
	}
	//     gracefulStop start
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v", sig)
		app.Stop()
		os.Exit(0)
	}()
	//      gracefulStop end

	app.Start()

}
