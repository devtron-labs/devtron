/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package restHandler

import (
	"context"
	"encoding/json"
	"fmt"
	application2 "github.com/argoproj/argo-cd/pkg/apiclient/application"
	"github.com/devtron-labs/devtron/api/connector"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strconv"
	"strings"
)

type ArgoApplicationRestHandler interface {
	GetPodLogs(w http.ResponseWriter, r *http.Request)
	GetResourceTree(w http.ResponseWriter, r *http.Request)
	ListResourceEvents(w http.ResponseWriter, r *http.Request)
	GetResource(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Watch(w http.ResponseWriter, r *http.Request)
	ManagedResources(w http.ResponseWriter, r *http.Request)
	Rollback(w http.ResponseWriter, r *http.Request)
	GetManifests(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)

	Sync(w http.ResponseWriter, r *http.Request)
	TerminateOperation(w http.ResponseWriter, r *http.Request)
	PatchResource(w http.ResponseWriter, r *http.Request)
	DeleteResource(w http.ResponseWriter, r *http.Request)

	GetServiceLink(w http.ResponseWriter, r *http.Request)
	GetTerminalSession(w http.ResponseWriter, r *http.Request)
}

type ArgoApplicationRestHandlerImpl struct {
	client                 application.ServiceClient
	logger                 *zap.SugaredLogger
	pump                   connector.Pump
	enforcer               rbac.Enforcer
	teamService            team.TeamService
	environmentService     cluster.EnvironmentService
	enforcerUtil           rbac.EnforcerUtil
	terminalSessionHandler terminal.TerminalSessionHandler
}

func NewArgoApplicationRestHandlerImpl(client application.ServiceClient,
	pump connector.Pump,
	enforcer rbac.Enforcer,
	teamService team.TeamService,
	environmentService cluster.EnvironmentService,
	logger *zap.SugaredLogger,
	enforcerUtil rbac.EnforcerUtil,
	terminalSessionHandler terminal.TerminalSessionHandler) *ArgoApplicationRestHandlerImpl {
	return &ArgoApplicationRestHandlerImpl{
		client:                 client,
		logger:                 logger,
		pump:                   pump,
		enforcer:               enforcer,
		teamService:            teamService,
		environmentService:     environmentService,
		enforcerUtil:           enforcerUtil,
		terminalSessionHandler: terminalSessionHandler,
	}
}

func (impl ArgoApplicationRestHandlerImpl) GetTerminalSession(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	request := &terminal.TerminalSessionRequest{}
	vars := mux.Vars(r)
	request.ContainerName = vars["container"]
	request.Namespace = vars["namespace"]
	request.PodName = vars["pod"]
	request.Shell = vars["shell"]
	appId := vars["appId"]
	envId := vars["environmentId"]
	//---------auth
	id, err := strconv.Atoi(appId)
	if err != nil {
		common.WriteJsonResp(w, fmt.Errorf("appId is not integer"), nil, http.StatusBadRequest)
		return
	}
	eId, err := strconv.Atoi(envId)
	if err != nil {
		common.WriteJsonResp(w, fmt.Errorf("envId is not integer"), nil, http.StatusBadRequest)
		return
	}
	request.AppId = id
	appRbacObject := impl.enforcerUtil.GetAppRBACNameByAppId(id)
	if appRbacObject == "" {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	envRbacObject := impl.enforcerUtil.GetEnvRBACNameByAppId(id, eId)
	if envRbacObject == "" {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusBadRequest)
		return
	}
	request.EnvironmentId = eId
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, appRbacObject); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	if ok := impl.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionCreate, envRbacObject); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//---------auth end
	//TODO apply validation
	status, message, err := impl.terminalSessionHandler.GetTerminalSession(request)
	common.WriteJsonResp(w, err, message, status)
}

func (impl ArgoApplicationRestHandlerImpl) Watch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	token := r.Header.Get("token")
	query := application2.ApplicationQuery{Name: &name}
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	defer cancel()
	app, conn, err := impl.client.Watch(ctx, &query)
	defer util.Close(conn, impl.logger)
	impl.pump.StartStream(w, func() (proto.Message, error) { return app.Recv() }, err)
}

func (impl ArgoApplicationRestHandlerImpl) GetPodLogs(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	vars := mux.Vars(r)
	name := vars["name"]
	podName := vars["podName"]
	containerName := v.Get("container")
	namespace := v.Get("namespace")
	token := r.Header.Get("token")
	sinceSeconds, err := strconv.ParseInt(v.Get("sinceSeconds"), 10, 64)
	if err != nil {
		sinceSeconds = 0
	}
	follow, err := strconv.ParseBool(v.Get("follow"))
	if err != nil {
		follow = false
	}
	tailLines, err := strconv.ParseInt(v.Get("tailLines"), 10, 64)
	if err != nil {
		tailLines = 0
	}
	query := application2.ApplicationPodLogsQuery{
		Name:         &name,
		PodName:      &podName,
		Container:    containerName,
		Namespace:    namespace,
		TailLines:    tailLines,
		Follow:       follow,
		SinceSeconds: sinceSeconds,
	}
	lastEventId := r.Header.Get("Last-Event-ID")
	isReconnect := false
	if len(lastEventId) > 0 {
		lastSeenMsgId, err := strconv.ParseInt(lastEventId, 10, 64)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		lastSeenMsgId = lastSeenMsgId + 1 //increased by one ns to avoid duplicate //FIXME still not fixed
		t := v1.Unix(0, lastSeenMsgId)
		query.SinceTime = &t
		query.SinceSeconds = 0 //set this ti zero since its reconnect request
		isReconnect = true
	}
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	defer cancel()
	logs, conn, err := impl.client.PodLogs(ctx, &query)
	defer util.Close(conn, impl.logger)
	impl.pump.StartStreamWithHeartBeat(w, isReconnect, func() (*application2.LogEntry, error) { return logs.Recv() }, err)
}

func (impl ArgoApplicationRestHandlerImpl) GetResourceTree(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	token := r.Header.Get("token")
	query := application2.ResourcesQuery{
		ApplicationName: &name,
	}
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	defer cancel()
	recv, err := impl.client.ResourceTree(ctx, &query)
	impl.pump.StartMessage(w, recv, err)
}

func (impl ArgoApplicationRestHandlerImpl) ListResourceEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	v := r.URL.Query()
	resourceNameSpace := v.Get("resourceNamespace")
	resourceUID := v.Get("resourceUID")
	resourceName := v.Get("resourceName")
	token := r.Header.Get("token")
	query := &application2.ApplicationResourceEventsQuery{
		Name:              &name,
		ResourceNamespace: resourceNameSpace,
		ResourceUID:       resourceUID,
		ResourceName:      resourceName,
	}
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	defer cancel()
	recv, err := impl.client.ListResourceEvents(ctx, query)
	impl.pump.StartMessage(w, recv, err)
}

func (impl ArgoApplicationRestHandlerImpl) GetResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	v := r.URL.Query()
	nameSpace := v.Get("namespace")
	version := v.Get("version")
	group := v.Get("group")
	kind := v.Get("kind")
	resourceName := v.Get("resourceName")
	token := r.Header.Get("token")
	query := &application2.ApplicationResourceRequest{
		Name:         &name,
		Version:      version,
		Group:        group,
		Kind:         kind,
		ResourceName: resourceName,
		Namespace:    nameSpace,
	}
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	defer cancel()
	recv, err := impl.client.GetResource(ctx, query)
	impl.pump.StartMessage(w, recv, err)
}

func (impl ArgoApplicationRestHandlerImpl) List(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	name := v.Get("name")
	refresh := v.Get("refresh")
	project := v.Get("project")
	projects := make([]string, 0)
	if len(project) > 0 {
		projects = strings.Split(project, ",")
	}
	token := r.Header.Get("token")
	query := &application2.ApplicationQuery{
		Name:     &name,
		Projects: projects,
		Refresh:  &refresh,
	}
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	defer cancel()
	recv, err := impl.client.List(ctx, query)
	impl.pump.StartMessage(w, recv, err)
}

func (impl ArgoApplicationRestHandlerImpl) ManagedResources(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationName := vars["applicationName"]
	query := &application2.ResourcesQuery{
		ApplicationName: &applicationName,
	}
	token := r.Header.Get("token")
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	defer cancel()
	recv, err := impl.client.ManagedResources(ctx, query)
	impl.pump.StartMessage(w, recv, err)
}

func (impl ArgoApplicationRestHandlerImpl) Rollback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	decoder := json.NewDecoder(r.Body)
	query := new(application2.ApplicationRollbackRequest)
	err := decoder.Decode(query)
	if err != nil {
		impl.logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	query.Name = &name
	token := r.Header.Get("token")
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	defer cancel()
	recv, err := impl.client.Rollback(ctx, query)
	impl.pump.StartMessage(w, recv, err)
}

func (impl ArgoApplicationRestHandlerImpl) GetManifests(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	v := r.URL.Query()
	revision := v.Get("revision")
	query := &application2.ApplicationManifestQuery{
		Name:     &name,
		Revision: revision,
	}
	token := r.Header.Get("token")
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	defer cancel()
	recv, err := impl.client.GetManifests(ctx, query)
	impl.pump.StartMessage(w, recv, err)
}

func (impl ArgoApplicationRestHandlerImpl) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	v := r.URL.Query()
	refresh := v.Get("refresh")
	project := v.Get("project")
	projects := make([]string, 0)
	if len(project) > 0 {
		projects = strings.Split(project, ",")
	}
	token := r.Header.Get("token")
	query := &application2.ApplicationQuery{
		Name:     &name,
		Projects: projects,
		Refresh:  &refresh,
	}
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	defer cancel()
	recv, err := impl.client.Get(ctx, query)
	impl.pump.StartMessage(w, recv, err)
}

func (impl ArgoApplicationRestHandlerImpl) Sync(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	decoder := json.NewDecoder(r.Body)
	query := new(application2.ApplicationSyncRequest)
	err := decoder.Decode(query)
	query.Name = &name
	if err != nil {
		impl.logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	defer cancel()
	recv, err := impl.client.Sync(ctx, query)
	impl.pump.StartMessage(w, recv, err)
}

func (impl ArgoApplicationRestHandlerImpl) TerminateOperation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	query := application2.OperationTerminateRequest{
		Name: &name,
	}
	token := r.Header.Get("token")
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	defer cancel()
	recv, err := impl.client.TerminateOperation(ctx, &query)
	impl.pump.StartMessage(w, recv, err)
}

func (impl ArgoApplicationRestHandlerImpl) PatchResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	token := r.Header.Get("token")
	appId := vars["appId"]
	id, err := strconv.Atoi(appId)
	if err != nil {
		common.WriteJsonResp(w, fmt.Errorf("appId is not integer"), nil, http.StatusBadRequest)
		return
	}
	app := impl.enforcerUtil.GetAppRBACNameByAppId(id)
	if app == "" {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionTrigger, app); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	decoder := json.NewDecoder(r.Body)
	query := new(application2.ApplicationResourcePatchRequest)
	err = decoder.Decode(query.Patch)
	if err != nil {
		impl.logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	query.Name = &name
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	defer cancel()
	recv, err := impl.client.PatchResource(ctx, query)
	impl.pump.StartMessage(w, recv, err)
}

func (impl ArgoApplicationRestHandlerImpl) DeleteResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appNameACD := vars["appNameACD"]
	name := vars["name"]
	namespace := vars["namespace"]
	resourceName := vars["resourceName"]
	version := vars["version"]
	kind := vars["kind"]
	group := vars["group"]
	force, err := strconv.ParseBool(vars["force"])
	if err != nil {
		force = false
	}
	if name == "" || namespace == "" || resourceName == "" || version == "" || kind == "" {
		common.WriteJsonResp(w, fmt.Errorf("missing mandatory field (name | namespace | resourceName | kind)"), nil, http.StatusBadRequest)
	}
	query := new(application2.ApplicationResourceDeleteRequest)
	query.Name = &appNameACD
	query.ResourceName = name
	query.Kind = kind
	query.Version = version
	query.Force = &force
	query.Namespace = namespace
	query.Group = group
	token := r.Header.Get("token")
	appId := vars["appId"]
	envId := vars["envId"]
	id, err := strconv.Atoi(appId)
	if err != nil {
		common.WriteJsonResp(w, fmt.Errorf("appId is not integer"), nil, http.StatusBadRequest)
		return
	}
	eId, err := strconv.Atoi(envId)
	if err != nil {
		common.WriteJsonResp(w, fmt.Errorf("envId is not integer"), nil, http.StatusBadRequest)
		return
	}
	appRbacObject := impl.enforcerUtil.GetAppRBACNameByAppId(id)
	if appRbacObject == "" {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	envRbacObject := impl.enforcerUtil.GetEnvRBACNameByAppId(id, eId)
	if envRbacObject == "" {
		common.WriteJsonResp(w, fmt.Errorf("envId is incorrect"), nil, http.StatusBadRequest)
		return
	}
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionTrigger, appRbacObject); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	if ok := impl.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionTrigger, envRbacObject); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	defer cancel()
	recv, err := impl.client.DeleteResource(ctx, query)
	impl.pump.StartMessage(w, recv, err)
}

func (impl ArgoApplicationRestHandlerImpl) GetServiceLink(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	v := r.URL.Query()
	revision := v.Get("revision")
	query := &application2.ApplicationManifestQuery{
		Name:     &name,
		Revision: revision,
	}
	token := r.Header.Get("token")
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	defer cancel()
	recv, err := impl.client.GetManifests(ctx, query)

	manifests := recv.GetManifests()
	var topMap []map[string]interface{}
	serviceCounter := 0
	//port := ""
	for _, manifest := range manifests {
		lowMap := make(map[string]interface{})
		err = json.Unmarshal([]byte(manifest), &lowMap)
		//if val, ok := lowMap["kind"]; ok {
		if lowMap["kind"] == "Service" {
			serviceCounter = serviceCounter + 1
		}
		topMap = append(topMap, lowMap)
	}

	var activeService string
	var serviceName string
	var serviceNamespace string
	var port string
	var serviceLink string
	if serviceCounter > 1 {
		for _, lowMap := range topMap {
			if lowMap["kind"] == "Rollout" {
				specObj := lowMap["spec"].(map[string]interface{})
				strategyObj := specObj["strategy"].(map[string]interface{})
				blueGreenObj := strategyObj["blueGreen"].(map[string]interface{})
				activeService = blueGreenObj["activeService"].(string)
				break
			}
		}
	}
	for _, lowMap := range topMap {
		if lowMap["kind"] == "Service" {
			metaObj := lowMap["metadata"].(map[string]interface{})
			specObj := lowMap["spec"].(map[string]interface{})
			portArr := specObj["ports"].([]interface{})

			serviceName = metaObj["name"].(string)
			serviceNamespace = metaObj["namespace"].(string)
			if serviceCounter == 1 {
				for _, item := range portArr {
					itemObj := item.(map[string]interface{})
					if itemObj["name"] == name {
						portF := itemObj["port"].(float64)
						portI := int(portF)
						port = strconv.Itoa(portI)
						break
					}
				}
				serviceLink = "http://" + serviceName + "." + serviceNamespace + ":" + port
				break
			} else if serviceCounter > 1 && serviceName == activeService {
				for _, item := range portArr {
					itemObj := item.(map[string]interface{})
					if itemObj["name"] == name {
						portF := itemObj["port"].(float64)
						portI := int(portF)
						port = strconv.Itoa(portI)
						break
					}
				}
				serviceLink = "http://" + serviceName + "." + serviceNamespace + ":" + port
				break
			} else {
				continue
			}
		}
	}
	common.WriteJsonResp(w, err, serviceLink, 200)
}
