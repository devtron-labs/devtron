/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package main

import (
	"fmt"
	_ "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	_ "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	util2 "github.com/devtron-labs/devtron/util"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	globalEnvVariables, err := util2.GetEnvironmentVariables()
	if err != nil {
		log.Println("error while getting env variables reason:", err)
		return
	}
	if globalEnvVariables.GlobalEnvVariables.ExecuteWireNilChecker {
		CheckIfNilInWire()
		return
	}
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
