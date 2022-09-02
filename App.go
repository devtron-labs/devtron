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
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/casbin/casbin"
	authMiddleware "github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/devtron/api/router"
	"github.com/devtron-labs/devtron/api/sse"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/go-pg/pg"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type App struct {
	MuxRouter    *router.MuxRouter
	Logger       *zap.SugaredLogger
	SSE          *sse.SSE
	Enforcer     *casbin.SyncedEnforcer
	server       *http.Server
	db           *pg.DB
	pubsubClient *pubsub.PubSubClient
	// used for local dev only
	serveTls        bool
	sessionManager2 *authMiddleware.SessionManager
}

func NewApp(router *router.MuxRouter,
	Logger *zap.SugaredLogger,
	sse *sse.SSE,
	versionService argocdServer.VersionService,
	enforcer *casbin.SyncedEnforcer,
	db *pg.DB,
	pubsubClient *pubsub.PubSubClient,
	sessionManager2 *authMiddleware.SessionManager,
) *App {
	//check argo connection
	//todo - check argo-cd version on acd integration installation
	app := &App{
		MuxRouter:       router,
		Logger:          Logger,
		SSE:             sse,
		Enforcer:        enforcer,
		db:              db,
		pubsubClient:    pubsubClient,
		serveTls:        false,
		sessionManager2: sessionManager2,
	}
	return app
}

func (app *App) Start() {
	port := 8080 //TODO: extract from environment variable
	app.Logger.Debugw("starting server")
	app.Logger.Infow("starting server on ", "port", port)
	app.MuxRouter.Init()
	//authEnforcer := casbin2.Create()

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: authMiddleware.Authorizer(app.sessionManager2, user.WhitelistChecker)(app.MuxRouter.Router)}
	app.MuxRouter.Router.Use(middleware.PrometheusMiddleware)
	app.server = server
	var err error
	if app.serveTls {
		cert, err := tls.LoadX509KeyPair(
			"localhost.crt",
			"localhost.key",
		)
		if err != nil {
			log.Fatal(err)
		}
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		err = server.ListenAndServeTLS("", "")
	} else {
		err = server.ListenAndServe()
	}
	//err := http.ListenAndServe(fmt.Sprintf(":%d", port), auth.Authorizer(app.Enforcer, app.sessionManager)(app.MuxRouter.Router))
	if err != nil {
		app.Logger.Errorw("error in startup", "err", err)
		os.Exit(2)
	}
}

func (app *App) Stop() {
	app.Logger.Infow("orchestrator shutdown initiating")
	timeoutContext, _ := context.WithTimeout(context.Background(), 5*time.Second)
	app.Logger.Infow("closing router")
	err := app.server.Shutdown(timeoutContext)
	if err != nil {
		app.Logger.Errorw("error in mux router shutdown", "err", err)
	}
	app.Logger.Infow("closing db connection")
	err = app.db.Close()
	if err != nil {
		app.Logger.Errorw("error in closing db connection", "err", err)
	}
	//Close not needed if you Drain.
	err = app.pubsubClient.Conn.Drain()

	if err != nil {
		app.Logger.Errorw("Error in draining nats connection", "error", err)
	}

	app.Logger.Infow("housekeeping done. exiting now")
}
