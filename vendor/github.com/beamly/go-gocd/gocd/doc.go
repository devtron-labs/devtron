// Copyright 2017-2018 Drew J. Sonne. All rights reserved.
//
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.

/*
Package gocd provides a client for using the GoCD Server API.

Usage:

	import "github.com/beamly/go-gocd/gocd"

Construct a new GoCD client and supply the URL to your GoCD server and if required, username and password. Then use the
various services on the client to access different parts of the GoCD API.
For example:

	package main
	import (
		"github.com/beamly/go-gocd/gocd"
		"context"
		"fmt"
	)

	func main() {
		cfg := gocd.Configuration{
			Server: "https://my_gocd/go/",
			Username: "ApiUser",
			Password: "MySecretPassword",
		}

		c := cfg.Client()

		// list all agents in use by the GoCD Server
		var a []*gocd.Agent
		var err error
		var r *gocd.APIResponse
		if a, r, err = c.Agents.List(context.Background()); err != nil {
			if r.HTTP.StatusCode == 404 {
				fmt.Println("Couldn't find agent")
			} else {
				panic(err)
			}
		}

		fmt.Println(a)
	}

If you wish to use your own http client, you can use the following idiom


	package main

	import (
		"github.com/beamly/go-gocd/gocd"
		"net/http"
		"context"
	)

	func main() {
		client := gocd.NewClient(
			&gocd.Configuration{},
			&http.Client{},
		)
		client.Login(context.Background())
	}

The services of a client divide the API into logical chunks and correspond to
the structure of the GoCD API documentation at
https://api.gocd.org/current/.

*/
package gocd
