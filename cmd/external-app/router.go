package main

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/api/sso"
	"github.com/devtron-labs/devtron/api/team"
	"github.com/devtron-labs/devtron/api/user"
	"github.com/devtron-labs/devtron/client/dashboard"
	"github.com/devtron-labs/devtron/util"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type MuxRouter struct {
	Router          *mux.Router
	logger          *zap.SugaredLogger
	ssoLoginRouter  sso.SsoLoginRouter
	teamRouter      team.TeamRouter
	UserAuthRouter  user.UserAuthRouter
	userRouter      user.UserRouter
	clusterRouter   cluster.ClusterRouter
	dashboardRouter dashboard.DashboardRouter
}

func NewMuxRouter(
	logger *zap.SugaredLogger,
	ssoLoginRouter sso.SsoLoginRouter,
	teamRouter team.TeamRouter,
	UserAuthRouter user.UserAuthRouter,
	userRouter user.UserRouter,
	clusterRouter cluster.ClusterRouter,
	dashboardRouter dashboard.DashboardRouter,

) *MuxRouter {
	r := &MuxRouter{
		Router:          mux.NewRouter(),
		logger:          logger,
		ssoLoginRouter:  ssoLoginRouter,
		teamRouter:      teamRouter,
		UserAuthRouter:  UserAuthRouter,
		userRouter:      userRouter,
		clusterRouter:   clusterRouter,
		dashboardRouter: dashboardRouter,
	}
	return r
}
func (r *MuxRouter) Init() {
	r.Router.StrictSlash(true)
	r.Router.Path("/health").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(200)
		response := common.Response{}
		response.Code = 200
		response.Result = "OK"
		b, err := json.Marshal(response)
		if err != nil {
			b = []byte("OK")
			r.logger.Errorw("Unexpected error in apiError", "err", err)
		}
		_, _ = writer.Write(b)
	})
	baseRouter := r.Router.PathPrefix("/orchestrator/").Subrouter()
	baseRouter.Path("/version").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(200)
		response := common.Response{}
		response.Code = 200
		response.Result = util.GetDevtronVersion()
		b, err := json.Marshal(response)
		if err != nil {
			b = []byte("OK")
			r.logger.Errorw("Unexpected error in apiError", "err", err)
		}
		_, _ = writer.Write(b)
	})



	ssoLoginRouter := baseRouter.PathPrefix("/sso").Subrouter()
	r.ssoLoginRouter.InitSsoLoginRouter(ssoLoginRouter)
	teamRouter := baseRouter.PathPrefix("/team").Subrouter()
	r.teamRouter.InitTeamRouter(teamRouter)
	rootRouter := baseRouter.PathPrefix("/").Subrouter()
	r.UserAuthRouter.InitUserAuthRouter(rootRouter)
	userRouter := baseRouter.PathPrefix("/user").Subrouter()
	r.userRouter.InitUserRouter(userRouter)

	clusterRouter := baseRouter.PathPrefix("/cluster").Subrouter()
	r.clusterRouter.InitClusterRouter(clusterRouter)

	r.Router.Path("/").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/dashboard", 301)
	})
	dashboardRouter := r.Router.PathPrefix("/dashboard").Subrouter()
	r.dashboardRouter.InitDashboardRouter(dashboardRouter)

}
