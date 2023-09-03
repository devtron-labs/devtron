package main

import (
	"context"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/sourceController/api"
	"net/http"
	"os"
	"time"

	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type App struct {
	Logger *zap.SugaredLogger
	Router *api.Router
	server *http.Server
	db     *pg.DB
}

func NewApp(Logger *zap.SugaredLogger,
	db *pg.DB,
	Router *api.Router) *App {
	return &App{
		Logger: Logger,
		db:     db,
		Router: Router,
	}
}

type ServerConfig struct {
	SERVER_HTTP_PORT int `env:"SERVER_HTTP_PORT" envDefault:"8080"`
}

func (app *App) Start() {
	serverConfig := ServerConfig{}
	err := env.Parse(&serverConfig)
	if err != nil {
		app.Logger.Errorw("error in parsing server config from environment", "err", err)
		os.Exit(2)
	}
	httpPort := serverConfig.SERVER_HTTP_PORT
	app.Logger.Infow("starting server on ", "httpPort", httpPort)
	app.Router.Init()
	server := &http.Server{Addr: fmt.Sprintf(":%d", httpPort), Handler: app.Router.Router}
	app.server = server
	err = server.ListenAndServe()
	if err != nil {
		app.Logger.Errorw("error in startup", "err", err)
		os.Exit(2)
	}
}

func (app *App) Stop() {
	app.Logger.Infow("source controller shutdown initiating")
	timeoutContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	app.Logger.Infow("closing router")
	err := app.server.Shutdown(timeoutContext)
	if err != nil {
		app.Logger.Errorw("error in mux router shutdown", "err", err)
	}
	app.Logger.Infow("closing db connection")
	err = app.db.Close()
	if err != nil {
		app.Logger.Errorw("Error while closing DB", "error", err)
	}

	app.Logger.Infow("housekeeping done. exiting now")
}
