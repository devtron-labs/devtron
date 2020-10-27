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
	"github.com/devtron-labs/devtron/api/router"
	"github.com/devtron-labs/devtron/api/sse"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/pkg/user"
	"fmt"
	"github.com/argoproj/argo-cd/util/session"
	"github.com/casbin/casbin"
	"github.com/go-pg/pg"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"time"
)

type App struct {
	MuxRouter      *router.MuxRouter
	Logger         *zap.SugaredLogger
	SSE            *sse.SSE
	Enforcer       *casbin.Enforcer
	sessionManager *session.SessionManager
	server         *http.Server
	db             *pg.DB
	pubsubClient   *pubsub.PubSubClient
}

func NewApp(router *router.MuxRouter,
	Logger *zap.SugaredLogger,
	sse *sse.SSE,
	manager *session.SessionManager,
	versionService argocdServer.VersionService,
	enforcer *casbin.Enforcer,
	db *pg.DB,
	pubsubClient *pubsub.PubSubClient) *App {
	//check argo connection
	err := versionService.CheckVersion()
	if err != nil {
		log.Panic(err)
	}
	app := &App{MuxRouter: router, Logger: Logger, SSE: sse, Enforcer: enforcer, sessionManager: manager, db: db, pubsubClient: pubsubClient}
	return app
}

func (app *App) Start() {

	/*	RequestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "Time (in seconds) spent serving HTTP requests.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route", "status_code", "ws"})
		prometheus.MustRegister(RequestDuration)
		prometheus.Ins*/
	//instrument:=middleware.Instrument{
	//	Duration: RequestDuration,
	//}
	port := 8080 //TODO: extract from environment variable
	app.Logger.Debugw("starting server")
	app.Logger.Infow("starting server on ", "port", port)
	app.MuxRouter.Init()
	//authEnforcer := casbin2.Create()

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: user.Authorizer(app.Enforcer, app.sessionManager)(app.MuxRouter.Router)}

	app.MuxRouter.Router.Use(middleware.PrometheusMiddleware)
	app.server = server
	err := server.ListenAndServe()
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
	nc := app.pubsubClient.Conn.NatsConn()
	//closing nats
	err = app.pubsubClient.Conn.Close()
	if err != nil {
		app.Logger.Errorw("error in closing stan", "err", err)
	}
	err = nc.Drain()
	if err != nil {
		app.Logger.Errorw("error in draining nats", "err", err)
	}
	nc.Close()
	app.Logger.Infow("housekeeping done. exiting now")
}
