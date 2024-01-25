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
	"github.com/devtron-labs/common-lib/middlewares"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/arl/statsviz"
	"github.com/casbin/casbin"
	casbinv2 "github.com/casbin/casbin/v2"
	"github.com/devtron-labs/devtron/api/util"
	"github.com/devtron-labs/devtron/client/telemetry"
	"github.com/devtron-labs/devtron/otel"
	"github.com/devtron-labs/devtron/pkg/auth/user"

	authMiddleware "github.com/devtron-labs/authenticator/middleware"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/api/router"
	"github.com/devtron-labs/devtron/api/sse"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/go-pg/pg"
	_ "github.com/lib/pq"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.uber.org/zap"
)

type App struct {
	MuxRouter     *router.MuxRouter
	Logger        *zap.SugaredLogger
	SSE           *sse.SSE
	Enforcer      *casbin.SyncedEnforcer
	EnforcerV2    *casbinv2.SyncedEnforcer
	server        *http.Server
	db            *pg.DB
	pubsubClient  *pubsub.PubSubClientServiceImpl
	posthogClient *telemetry.PosthogClient
	// used for local dev only
	serveTls                  bool
	sessionManager2           *authMiddleware.SessionManager
	OtelTracingService        *otel.OtelTracingServiceImpl
	loggingMiddleware         util.LoggingMiddleware
	userStatusCheckMiddleware util.UserStatusCheckMiddleware
}

func NewApp(router *router.MuxRouter,
	Logger *zap.SugaredLogger,
	sse *sse.SSE,
	enforcer *casbin.SyncedEnforcer,
	enforcerV2 *casbinv2.SyncedEnforcer,
	db *pg.DB,
	pubsubClient *pubsub.PubSubClientServiceImpl,
	sessionManager2 *authMiddleware.SessionManager,
	posthogClient *telemetry.PosthogClient,
	loggingMiddleware util.LoggingMiddleware,
	userStatusCheckMiddleware util.UserStatusCheckMiddleware,
) *App {
	//check argo connection
	//todo - check argo-cd version on acd integration installation
	app := &App{
		MuxRouter:                 router,
		Logger:                    Logger,
		SSE:                       sse,
		Enforcer:                  enforcer,
		EnforcerV2:                enforcerV2,
		db:                        db,
		pubsubClient:              pubsubClient,
		serveTls:                  false,
		sessionManager2:           sessionManager2,
		posthogClient:             posthogClient,
		OtelTracingService:        otel.NewOtelTracingServiceImpl(Logger),
		loggingMiddleware:         loggingMiddleware,
		userStatusCheckMiddleware: userStatusCheckMiddleware,
	}
	return app
}

func (app *App) Start() {

	app.checkAndSetupStatsviz()

	port := 8080 //TODO: extract from environment variable
	app.Logger.Debugw("starting server")
	app.Logger.Infow("starting server on ", "port", port)

	// setup tracer
	tracerProvider := app.OtelTracingService.Init(otel.OTEL_ORCHESTRASTOR_SERVICE_NAME)

	app.MuxRouter.Init()
	//authEnforcer := casbin2.Create()

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: authMiddleware.Authorizer(app.sessionManager2, user.WhitelistChecker)(app.MuxRouter.Router)}
	app.MuxRouter.Router.Use(app.loggingMiddleware.LoggingMiddleware)
	app.MuxRouter.Router.Use(app.userStatusCheckMiddleware.UserStatusCheckMiddleware)
	app.MuxRouter.Router.Use(middleware.PrometheusMiddleware)
	app.MuxRouter.Router.Use(middlewares.Recovery)

	if tracerProvider != nil {
		app.MuxRouter.Router.Use(otelmux.Middleware(otel.OTEL_ORCHESTRASTOR_SERVICE_NAME))
	}
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

func (app *App) checkAndSetupStatsviz() {
	statsvizPort, present := os.LookupEnv("STATSVIZ_PORT")
	if present && len(statsvizPort) > 0 {
		mux := http.NewServeMux()
		statsviz.Register(mux)
		go func() {
			app.Logger.Infow("statsviz server starting", "port", statsvizPort)
			err := http.ListenAndServe(fmt.Sprintf(":%s", statsvizPort), mux)
			if err != nil {
				app.Logger.Errorw("error occurred while starting statsviz server", "err", err)
			}

		}()
	}
}

func (app *App) Stop() {
	app.Logger.Info("orchestrator shutdown initiating")
	posthogCl := app.posthogClient.Client
	if posthogCl != nil {
		app.Logger.Info("flushing messages of posthog")
		posthogCl.Close()
	}
	timeoutContext, _ := context.WithTimeout(context.Background(), 5*time.Second)
	app.Logger.Infow("closing router")
	err := app.server.Shutdown(timeoutContext)
	if err != nil {
		app.Logger.Errorw("error in mux router shutdown", "err", err)
	}

	app.OtelTracingService.Shutdown()

	app.Logger.Infow("closing db connection")
	err = app.db.Close()
	if err != nil {
		app.Logger.Errorw("error in closing db connection", "err", err)
	}
	//Close not needed if you Drain.

	if err != nil {
		app.Logger.Errorw("Error in draining nats connection", "error", err)
	}

	app.Logger.Infow("housekeeping done. exiting now")
}
