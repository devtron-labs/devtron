package main

import (
	"fmt"
	authMiddleware "github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"os"
)

type App struct {
	db             *pg.DB
	sessionManager *authMiddleware.SessionManager
	MuxRouter      *MuxRouter
	Logger         *zap.SugaredLogger

	server *http.Server
}

func NewApp(db *pg.DB,
	sessionManager *authMiddleware.SessionManager,
	MuxRouter *MuxRouter,
	Logger *zap.SugaredLogger) *App {
	return &App{
		db:             db,
		sessionManager: sessionManager,
		MuxRouter:      MuxRouter,
		Logger:         Logger,
	}
}
func (app *App) Start() {
	fmt.Println("starting ea module")

	port := 8080 //TODO: extract from environment variable
	app.Logger.Debugw("starting server")
	app.Logger.Infow("starting server on ", "port", port)
	app.MuxRouter.Init()
	//authEnforcer := casbin2.Create()

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: authMiddleware.Authorizer(app.sessionManager, user.WhitelistChecker)(app.MuxRouter.Router)}
	app.MuxRouter.Router.Use(middleware.PrometheusMiddleware)
	app.server = server
	err := server.ListenAndServe()
	if err != nil {
		app.Logger.Errorw("error in startup", "err", err)
		os.Exit(2)
	}
}

func (app *App) Stop() {

}
