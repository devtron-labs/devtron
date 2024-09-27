package main

import (
	"embed"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/util/dir"
	"go.uber.org/zap"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

//go:embed ui
var staticFiles embed.FS

const DefaultPort = 8080

type App struct {
	MuxRouter *MuxRouter
	Logger    *zap.SugaredLogger
	server    *http.Server
}

func NewApp(MuxRouter *MuxRouter,
	Logger *zap.SugaredLogger) *App {
	return &App{
		MuxRouter: MuxRouter,
		Logger:    Logger,
	}
}
func (app *App) Start() {
	freePort, err := app.GetFreePort()
	if err != nil {
		app.Logger.Warn("not able to extract free port so using default port ", DefaultPort)
		freePort = DefaultPort
	}
	port := freePort
	fmt.Println("starting on port ", port)
	app.writePortToFile(port)
	app.Logger.Infow("starting server on ", "port", port)
	app.MuxRouter.Init()
	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: app.MuxRouter.Router}
	app.MuxRouter.Router.Use(middleware.PrometheusMiddleware)
	fileSystem := http.FS(staticFiles)
	const DashboardPathPrefix = "/dashboard"
	app.MuxRouter.Router.PathPrefix(DashboardPathPrefix).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestURI := r.URL.Path
		partialURL := strings.Replace(requestURI, DashboardPathPrefix, "", 1)

		baseFolder := "./ui"
		finalPath := path.Join(baseFolder, partialURL)
		file, err := fileSystem.Open(finalPath)
		if err != nil || partialURL == "" || partialURL == "/" {
			finalPath = path.Join(baseFolder, "./index.html")
		}
		file, err = fileSystem.Open(finalPath)
		if err != nil {
			common.WriteJsonResp(w, err, finalPath, http.StatusInternalServerError)
			return
		}
		stat, err := file.Stat()
		if err != nil {
			common.WriteJsonResp(w, err, finalPath, http.StatusInternalServerError)
			return
		}
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
	})
	app.server = server
	err = server.ListenAndServe()
	if err != nil {
		app.Logger.Errorw("error in startup", "err", err)
		os.Exit(2)
	}
}

func (app *App) GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func (app *App) Stop() {
	app.Logger.Info("stopping k8s client App")
}

func (app *App) writePortToFile(port int) {
	err, devtronDirPath := dir.CheckOrCreateDevtronDir()
	if err != nil {
		app.Logger.Fatal("error occurred while creating dir", "err", err)
	}
	portPath := path.Join(devtronDirPath, "./process.port")
	err = os.WriteFile(portPath, []byte(strconv.Itoa(port)), fs.ModePerm)
	if err != nil {
		app.Logger.Fatal("error occurred while saving port number to path", "port", port, "path", portPath)
	}
}
