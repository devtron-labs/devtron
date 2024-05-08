package restHandler

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"net/http"
	"net/http/pprof"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/user"
)

type PProfRestHandler interface {
	Index(w http.ResponseWriter, r *http.Request)
	Cmdline(w http.ResponseWriter, r *http.Request)
	Profile(w http.ResponseWriter, r *http.Request)
	Symbol(w http.ResponseWriter, r *http.Request)
	Trace(w http.ResponseWriter, r *http.Request)
	Goroutine(w http.ResponseWriter, r *http.Request)
	Threadcreate(w http.ResponseWriter, r *http.Request)
	Heap(w http.ResponseWriter, r *http.Request)
	Block(w http.ResponseWriter, r *http.Request)
	Mutex(w http.ResponseWriter, r *http.Request)
	Allocs(w http.ResponseWriter, r *http.Request)
}

type PProfRestHandlerImpl struct {
	userService user.UserService
	enforcer    casbin.Enforcer
}

func NewPProfRestHandler(userService user.UserService,
	enforcer casbin.Enforcer) *PProfRestHandlerImpl {
	return &PProfRestHandlerImpl{userService: userService,
		enforcer: enforcer,
	}
}

func (p PProfRestHandlerImpl) Index(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := p.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	pprof.Index(w, r)
}

func (p PProfRestHandlerImpl) Cmdline(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := p.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	pprof.Cmdline(w, r)
}

func (p PProfRestHandlerImpl) Profile(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := p.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	pprof.Profile(w, r)
}

func (p PProfRestHandlerImpl) Symbol(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := p.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	pprof.Symbol(w, r)
}

func (p PProfRestHandlerImpl) Trace(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := p.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	pprof.Trace(w, r)
}

func (p PProfRestHandlerImpl) Goroutine(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := p.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	pprof.Handler("goroutine").ServeHTTP(w, r)
}

func (p PProfRestHandlerImpl) Threadcreate(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := p.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	pprof.Handler("threadcreate").ServeHTTP(w, r)
}

func (p PProfRestHandlerImpl) Heap(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := p.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	pprof.Handler("heap").ServeHTTP(w, r)
}

func (p PProfRestHandlerImpl) Block(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := p.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	pprof.Handler("block").ServeHTTP(w, r)
}

func (p PProfRestHandlerImpl) Mutex(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := p.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	pprof.Handler("mutex").ServeHTTP(w, r)
}

func (p PProfRestHandlerImpl) Allocs(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := p.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	pprof.Handler("allocs").ServeHTTP(w, r)
}
