package main

import (
	"fmt"
	authMiddleware "github.com/devtron-labs/authenticator/middleware"
	"github.com/go-pg/pg"
)

type App struct {
	db             *pg.DB
	sessionManager *authMiddleware.SessionManager
}

func NewApp(db *pg.DB,
	sessionManager *authMiddleware.SessionManager) *App {
	return &App{
		db:             db,
		sessionManager: sessionManager,
	}
}
func (app *App) Start() {
	fmt.Println("starting ea module")
}

func (app *App) Stop() {

}
