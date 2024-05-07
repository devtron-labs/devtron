package autoRemediation

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/autoRemediation"
	"github.com/devtron-labs/devtron/pkg/autoRemediation/types"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/devtron-labs/devtron/util/response"
	types2 "github.com/devtron-labs/scoop/types"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type WatcherRestHandler interface {
	SaveWatcher(w http.ResponseWriter, r *http.Request)
	GetWatcherById(w http.ResponseWriter, r *http.Request)
	DeleteWatcherById(w http.ResponseWriter, r *http.Request)
	UpdateWatcherById(w http.ResponseWriter, r *http.Request)
	RetrieveWatchers(w http.ResponseWriter, r *http.Request)
	RetrieveInterceptedEvents(w http.ResponseWriter, r *http.Request)
	GetWatchersByClusterId(w http.ResponseWriter, r *http.Request)
}

type WatcherRestHandlerImpl struct {
	watcherService  autoRemediation.WatcherService
	userAuthService user.UserService
	validator       *validator.Validate
	enforcerUtil    rbac.EnforcerUtil
	enforcer        casbin.Enforcer
	celEvaluator    resourceFilter.CELEvaluatorService
	logger          *zap.SugaredLogger
}

func NewWatcherRestHandlerImpl(watcherService autoRemediation.WatcherService, userAuthService user.UserService, validator *validator.Validate,
	enforcerUtil rbac.EnforcerUtil, enforcer casbin.Enforcer, celEvaluator resourceFilter.CELEvaluatorService, logger *zap.SugaredLogger) *WatcherRestHandlerImpl {
	return &WatcherRestHandlerImpl{
		watcherService:  watcherService,
		userAuthService: userAuthService,
		validator:       validator,
		enforcerUtil:    enforcerUtil,
		enforcer:        enforcer,
		celEvaluator:    celEvaluator,
		logger:          logger,
	}
}

func (impl WatcherRestHandlerImpl) evaluateEventExpression(expression string) error {

	params := []resourceFilter.ExpressionParam{
		{
			ParamName: "newResource",
			Type:      resourceFilter.ParamTypeObject,
		},
		{
			ParamName: "oldResource",
			Type:      resourceFilter.ParamTypeObject,
		},
		{
			ParamName: "action",
			Type:      resourceFilter.ParamTypeString,
		},
	}

	request := resourceFilter.CELRequest{
		Expression:         expression,
		ExpressionMetadata: resourceFilter.ExpressionMetadata{Params: params},
	}

	_, _, err := impl.celEvaluator.Validate(request)
	return err
}

func (impl WatcherRestHandlerImpl) SaveWatcher(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var watcherRequest types.WatcherDto
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

	// empty watcherRequest.EventConfiguration.EventExpression should be considered as pass condition. (@abhibaw)
	if watcherRequest.EventConfiguration.EventExpression != "" {
		err = impl.evaluateEventExpression(watcherRequest.EventConfiguration.EventExpression)
		if err != nil {
			impl.logger.Errorw("validation err, event expression", "eventExpression", watcherRequest.EventConfiguration.EventExpression, "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	// RBAC
	token := r.Header.Get("token")
	isSuperAdmin := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	if !isSuperAdmin {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
		return
	}
	// RBAC
	watcherRequest.Name = strings.ToLower(watcherRequest.Name)
	res, err := impl.watcherService.CreateWatcher(&watcherRequest, userId)
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
	// RBAC enforcer applying
	token := r.Header.Get("token")
	isSuperAdmin := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	if !isSuperAdmin {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
		return
	}
	// RBAC enforcer Ends
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
	// RBAC enforcer applying
	token := r.Header.Get("token")
	isSuperAdmin := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*")
	if !isSuperAdmin {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
		return
	}
	// RBAC enforcer Ends
	err = impl.watcherService.DeleteWatcherById(watcherId, userId)
	if err != nil {
		impl.logger.Errorw("service err, DeleteWatcherById", "err", err, "watcher id", watcherId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, nil, watcherId, http.StatusOK)
}

func (impl WatcherRestHandlerImpl) UpdateWatcherById(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	watcherId, err := strconv.Atoi(vars["identifier"])
	var watcherRequest types.WatcherDto
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

	// empty watcherRequest.EventConfiguration.EventExpression should be considered as pass condition. (@abhibaw)
	if watcherRequest.EventConfiguration.EventExpression != "" {
		err = impl.evaluateEventExpression(watcherRequest.EventConfiguration.EventExpression)
		if err != nil {
			impl.logger.Errorw("validation err, event expression", "eventExpression", watcherRequest.EventConfiguration.EventExpression, "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	isSuperAdmin := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
	if !isSuperAdmin {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
		return
	}
	// RBAC enforcer Ends
	err = impl.watcherService.UpdateWatcherById(watcherId, &watcherRequest, userId)
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
	queryParams := r.URL.Query()
	// RBAC enforcer applying
	token := r.Header.Get("token")
	isSuperAdmin := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	if !isSuperAdmin {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
		return
	}
	// RBAC enforcer Ends
	watcherQueryParams, err := getWatcherQueryParams(queryParams)
	if err != nil {
		impl.logger.Errorw("error in fetching watcherQueryParams", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	watchersResponse, err := impl.watcherService.FindAllWatchers(watcherQueryParams)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("service err, find all ", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, watchersResponse, http.StatusOK)
}
func getWatcherQueryParams(queryParams url.Values) (types.WatcherQueryParams, error) {
	sortOrder := queryParams.Get("order")
	sortOrder = strings.ToLower(sortOrder)
	if sortOrder == "" {
		sortOrder = "asc"
	}
	if !(sortOrder == "asc" || sortOrder == "desc") {
		return types.WatcherQueryParams{}, errors.New("sort order can only be ASC or DESC")
	}
	sortOrderBy := queryParams.Get("orderBy")
	if sortOrderBy == "" {
		sortOrderBy = "name"
		sortOrder = "asc"
	}
	if sortOrderBy == "triggeredAt" {
		if sortOrder == "desc" {
			sortOrder = "asc"
		} else {
			sortOrder = "desc"
		}
	}
	if !(sortOrderBy == "name" || sortOrderBy == "triggeredAt") {
		return types.WatcherQueryParams{}, errors.New("sort order can only be by name or triggeredAt")
	}
	sizeStr := queryParams.Get("size")
	size := 20
	if sizeStr != "" {
		var err error
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size < 0 {
			return types.WatcherQueryParams{}, errors.New("invalid size")
		}
	}
	offsetStr := queryParams.Get("offset")
	offset := 0
	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return types.WatcherQueryParams{}, errors.New("invalid size")
		}
	}
	search := queryParams.Get("search")
	search = strings.ToLower(search)
	return types.WatcherQueryParams{
		Offset:      offset,
		Size:        size,
		Search:      search,
		SortOrder:   sortOrder,
		SortOrderBy: sortOrderBy,
	}, nil

}
func (impl WatcherRestHandlerImpl) RetrieveInterceptedEvents(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	queryParams := r.URL.Query()
	interceptedEventQuery, err := getInterceptedEventsQueryParams(queryParams)
	// RBAC enforcer applying
	token := r.Header.Get("token")
	isSuperAdmin := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	if !isSuperAdmin {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
		return
	}
	// RBAC enforcer Ends
	eventsResponse, err := impl.watcherService.RetrieveInterceptedEvents(interceptedEventQuery)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("service err, find all ", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, eventsResponse, http.StatusOK)
}
func getInterceptedEventsQueryParams(queryParams url.Values) (*types.InterceptedEventQueryParams, error) {
	sortOrder := queryParams.Get("order")
	sortOrder = strings.ToLower(sortOrder)
	if sortOrder == "" {
		sortOrder = "desc"
	} else if !(sortOrder == "asc" || sortOrder == "desc") {
		return nil, errors.New("sort order can only be ASC or DESC")
	} else if sortOrder == "asc" {
		sortOrder = "desc"
	} else {
		sortOrder = "asc"
	}

	sizeStr := queryParams.Get("size")
	size := 20
	if sizeStr != "" {
		var err error
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size < 0 {
			return nil, errors.New("invalid size")
		}
	}
	offsetStr := queryParams.Get("offset")
	offset := 0
	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return nil, errors.New("invalid offset")
		}
	}
	search := queryParams.Get("search")
	search = strings.ToLower(search)
	from := queryParams.Get("from")
	var fromTime time.Time
	if from != "" {
		var err error
		fromTime, err = time.Parse(time.RFC1123, from)
		if err != nil {
			return nil, errors.New("invalid from time")
		}
	}
	to := queryParams.Get("to")
	var toTime time.Time
	if to != "" {
		var err error
		toTime, err = time.Parse(time.RFC1123, to)
		if err != nil {
			return nil, errors.New("invalid to time")
		}
	}
	watchers := queryParams.Get("watchers")
	var watchersArray []string
	if watchers != "" {
		watchersArray = strings.Split(watchers, ",")
	}

	namespaces := queryParams.Get("namespaces")
	var namespacesArray []string
	if namespaces != "" {
		namespacesArray = strings.Split(namespaces, ",")
	}
	clusterNamespacePair, clusterIds, err := app.GetNamespaceClusterMapping(namespacesArray)
	if err != nil {
		return nil, err
	}

	selectedActionsStr := queryParams.Get("selectedActions")
	var selectedActionsArray []types2.EventType
	if selectedActionsStr != "" {
		selectedActionsArray = util.Map(strings.Split(selectedActionsStr, ","), func(action string) types2.EventType {
			return types2.EventType(action)
		})
	}

	executionStatus := queryParams.Get("executionStatuses")
	var executionStatusArray []string
	if executionStatus != "" {
		executionStatusArray = strings.Split(executionStatus, ",")
	}

	return &types.InterceptedEventQueryParams{
		Offset:                  offset,
		Size:                    size,
		SortOrder:               sortOrder,
		SearchString:            search,
		From:                    fromTime,
		To:                      toTime,
		Watchers:                watchersArray,
		ClusterIds:              clusterIds,
		ClusterIdNamespacePairs: clusterNamespacePair,
		ExecutionStatus:         executionStatusArray,
		Actions:                 selectedActionsArray,
	}, nil

}
func (handler *WatcherRestHandlerImpl) GetWatchersByClusterId(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("token")
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	if !isSuperAdmin {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	clusterId, err := strconv.Atoi(vars["clusterId"])
	if err != nil {
		handler.logger.Errorw("error in getting clusterId from query param", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	watchers, err := handler.watcherService.GetWatchersByClusterId(clusterId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, watchers, http.StatusOK)
}
