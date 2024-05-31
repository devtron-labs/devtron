/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package main

import (
	"fmt"
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
