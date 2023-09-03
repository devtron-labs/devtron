package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type Router struct {
	logger *zap.SugaredLogger
	Router *mux.Router
}

func NewRouter(logger *zap.SugaredLogger) *Router {
	return &Router{logger: logger, Router: mux.NewRouter()}
}

func (r Router) Init() {
	r.Router.StrictSlash(true)
	r.Router.Path("/health").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(200)
		response := Response{}
		response.Code = 200
		response.Result = "OK"
		b, err := json.Marshal(response)
		if err != nil {
			b = []byte("OK")
			r.logger.Errorw("Unexpected error in apiError", "err", err)
		}
		_, _ = writer.Write(b)
	})

}
