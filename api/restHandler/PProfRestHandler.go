package restHandler

import (
"github.com/devtron-labs/devtron/api/restHandler/common"
"github.com/devtron-labs/devtron/pkg/user"
"net/http"
"net/http/pprof"
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
}

func NewPProfRestHandler(userService user.UserService) *PProfRestHandlerImpl {
	return &PProfRestHandlerImpl{userService: userService}
}

func (p PProfRestHandlerImpl) Index(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isActionUserSuperAdmin, err := p.userService.IsSuperAdmin(int(userId))
	if err != nil {
		common.WriteJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}
	if isActionUserSuperAdmin {
		pprof.Index(w, r)
	}
}

func (p PProfRestHandlerImpl) Cmdline(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isActionUserSuperAdmin, err := p.userService.IsSuperAdmin(int(userId))
	if err != nil {
		common.WriteJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}
	if isActionUserSuperAdmin {
		pprof.Cmdline(w, r)
	}
}

func (p PProfRestHandlerImpl) Profile(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isActionUserSuperAdmin, err := p.userService.IsSuperAdmin(int(userId))
	if err != nil {
		common.WriteJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}
	if isActionUserSuperAdmin {
		pprof.Profile(w, r)
	}
}

func (p PProfRestHandlerImpl) Symbol(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isActionUserSuperAdmin, err := p.userService.IsSuperAdmin(int(userId))
	if err != nil {
		common.WriteJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}
	if isActionUserSuperAdmin {
		pprof.Symbol(w, r)
	}
}

func (p PProfRestHandlerImpl) Trace(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isActionUserSuperAdmin, err := p.userService.IsSuperAdmin(int(userId))
	if err != nil {
		common.WriteJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}
	if isActionUserSuperAdmin {
		pprof.Trace(w, r)
	}
}

func (p PProfRestHandlerImpl) Goroutine(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isActionUserSuperAdmin, err := p.userService.IsSuperAdmin(int(userId))
	if err != nil {
		common.WriteJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}
	if isActionUserSuperAdmin {
		pprof.Handler("goroutine").ServeHTTP(w, r)
	}
}

func (p PProfRestHandlerImpl) Threadcreate(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isActionUserSuperAdmin, err := p.userService.IsSuperAdmin(int(userId))
	if err != nil {
		common.WriteJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}
	if isActionUserSuperAdmin {
		pprof.Handler("threadcreate").ServeHTTP(w, r)
	}
}

func (p PProfRestHandlerImpl) Heap(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isActionUserSuperAdmin, err := p.userService.IsSuperAdmin(int(userId))
	if err != nil {
		common.WriteJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}
	if isActionUserSuperAdmin {
		pprof.Handler("heap").ServeHTTP(w, r)
	}
}

func (p PProfRestHandlerImpl) Block(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isActionUserSuperAdmin, err := p.userService.IsSuperAdmin(int(userId))
	if err != nil {
		common.WriteJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}
	if isActionUserSuperAdmin {
		pprof.Handler("block").ServeHTTP(w, r)
	}
}

func (p PProfRestHandlerImpl) Mutex(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isActionUserSuperAdmin, err := p.userService.IsSuperAdmin(int(userId))
	if err != nil {
		common.WriteJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}
	if isActionUserSuperAdmin {
		pprof.Handler("mutex").ServeHTTP(w, r)
	}
}

func (p PProfRestHandlerImpl) Allocs(w http.ResponseWriter, r *http.Request) {
	userId, err := p.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isActionUserSuperAdmin, err := p.userService.IsSuperAdmin(int(userId))
	if err != nil {
		common.WriteJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}
	if isActionUserSuperAdmin {
		pprof.Handler("allocs").ServeHTTP(w, r)
	}
}


