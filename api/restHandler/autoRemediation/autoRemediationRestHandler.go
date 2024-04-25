package autoRemediation

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/autoRemediation"
	util "github.com/devtron-labs/devtron/util/event"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"time"
)

type WatcherRestHandler interface {
	SaveWatcher(w http.ResponseWriter, r *http.Request)
	GetWatcherById(w http.ResponseWriter, r *http.Request)
	DeleteWatcherById(w http.ResponseWriter, r *http.Request)
	RetrieveInterceptedEvents(w http.ResponseWriter, r *http.Request)
	UpdateWatcherById(w http.ResponseWriter, r *http.Request)
	RetrieveWatchers(w http.ResponseWriter, r *http.Request)
}
type WatcherRestHandlerImpl struct {
	watcherService  autoRemediation.WatcherService
	userAuthService user.UserService
	validator       *validator.Validate
	logger          *zap.SugaredLogger
}

type ChannelDto struct {
	Channel util.Channel `json:"channel" validate:"required"`
}

func NewWatcherRestHandlerImpl(watcherService autoRemediation.WatcherService, userAuthService user.UserService, validator *validator.Validate,
	logger *zap.SugaredLogger) *WatcherRestHandlerImpl {
	return &WatcherRestHandlerImpl{
		watcherService:  watcherService,
		userAuthService: userAuthService,
		validator:       validator,
		logger:          logger,
	}
}

func (impl WatcherRestHandlerImpl) SaveWatcher(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var watcherRequest autoRemediation.WatcherDto
	err = json.NewDecoder(r.Body).Decode(&watcherRequest)
	if err != nil {
		impl.logger.Errorw("request err, SaveWatcher", "err", err, "payload", watcherRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, SaveWatcher", "err", err, "payload", watcherRequest)
	err = impl.validator.Struct(watcherRequest)
	if err != nil {
		impl.logger.Errorw("validation err, SaveWatcher", "err", err, "payload", watcherRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//RBAC
	//token := r.Header.Get("token")

	//RBAC
	res, err := impl.watcherService.Create(watcherRequest)
	if err != nil {
		impl.logger.Errorw("service err, SaveWatcher", "err", err, "payload", watcherRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}
func (impl WatcherRestHandlerImpl) GetWatcherById(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	watcherId, err := strconv.Atoi(vars["identifier"])
	//RBAC
	//token := r.Header.Get("token")

	//RBAC
	res, err := impl.watcherService.GetWatcherById(watcherId)
	if err != nil {
		impl.logger.Errorw("service err, GetWatcherById", "err", err, "watcher id", watcherId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}
func (impl WatcherRestHandlerImpl) DeleteWatcherById(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	watcherId, err := strconv.Atoi(vars["identifier"])
	//RBAC
	//token := r.Header.Get("token")

	//RBAC
	err = impl.watcherService.DeleteWatcherById(watcherId)
	if err != nil {
		impl.logger.Errorw("service err, DeleteWatcherById", "err", err, "watcher id", watcherId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, nil, watcherId, http.StatusOK)
}
func (impl WatcherRestHandlerImpl) RetrieveInterceptedEvents(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	//RBAC
	//token := r.Header.Get("token")

	//RBAC
	res, err := impl.watcherService.RetrieveInterceptedEvents()
	if err != nil {
		impl.logger.Errorw("service err, retrieveInterceptedevents", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}
func (impl WatcherRestHandlerImpl) UpdateWatcherById(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	watcherId, err := strconv.Atoi(vars["identifier"])
	var watcherRequest autoRemediation.WatcherDto
	err = json.NewDecoder(r.Body).Decode(&watcherRequest)
	if err != nil {
		impl.logger.Errorw("request err, SaveWatcher", "err", err, "payload", watcherRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, SaveWatcher", "err", err, "payload", watcherRequest)
	err = impl.validator.Struct(watcherRequest)
	if err != nil {
		impl.logger.Errorw("validation err, SaveWatcher", "err", err, "payload", watcherRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//RBAC
	//token := r.Header.Get("token")

	//RBAC
	err = impl.watcherService.UpdateWatcherById(watcherId, watcherRequest)
	if err != nil {
		impl.logger.Errorw("service err, updateWatcherById", "err", err, "watcher id", watcherId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}
func (impl WatcherRestHandlerImpl) RetrieveWatchers(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	watchers := vars["watchers"]
	clusters := vars["clusters"]
	namespaces := vars["namespaces"]
	executionStatuses := vars["executionStatuses"]
	from, err := time.Parse(time.RFC1123, vars["from"])
	if err != nil {
		impl.logger.Errorw("request err, RetrieveWatchers", "err", err, "payload", from)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	to, err := time.Parse(time.RFC1123, vars["to"])
	if err != nil {
		impl.logger.Errorw("request err, RetrieveWatchers", "err", err, "payload", to)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	size, err := strconv.Atoi(vars["size"])
	if err != nil {
		impl.logger.Errorw("request err, RetrieveWatchers", "err", err, "payload", size)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	offset, err := strconv.Atoi(vars["offset"])
	if err != nil {
		impl.logger.Errorw("request err, RetrieveWatchers", "err", err, "payload", offset)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	searchString := vars["searchString"]

}
