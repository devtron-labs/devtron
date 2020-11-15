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

package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argoproj/argo-cd/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/reposerver/apiclient"
	"github.com/argoproj/argo-cd/util/settings"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"time"
)

const (
	Degraded    = "Degraded"
	Healthy     = "Healthy"
	Progressing = "Progressing"
	Suspended   = "Suspended"
	TimeoutFast = 10 * time.Second
	TimeoutSlow = 30 * time.Second
	HIBERNATING = "HIBERNATING"
)

type ServiceClient interface {
	List(ctxt context.Context, query *application.ApplicationQuery) (*v1alpha1.ApplicationList, error)
	PodLogs(ctxt context.Context, query *application.ApplicationPodLogsQuery) (application.ApplicationService_PodLogsClient, *grpc.ClientConn, error)
	ListResourceEvents(ctx context.Context, query *application.ApplicationResourceEventsQuery) (*v1.EventList, error)
	ResourceTree(ctx context.Context, query *application.ResourcesQuery) (*ResourceTreeResponse, error)
	// Watch returns stream of application change events.
	Watch(ctx context.Context, in *application.ApplicationQuery) (application.ApplicationService_WatchClient, *grpc.ClientConn, error)
	ManagedResources(ctx context.Context, query *application.ResourcesQuery) (*application.ManagedResourcesResponse, error)
	// Rollback syncs an application to its target state
	Rollback(ctx context.Context, query *application.ApplicationRollbackRequest) (*v1alpha1.Application, error)
	// Patch patch an application
	Patch(ctx context.Context, query *application.ApplicationPatchRequest) (*v1alpha1.Application, error)
	// GetManifests returns application manifests
	GetManifests(ctx context.Context, query *application.ApplicationManifestQuery) (*apiclient.ManifestResponse, error)
	// GetResource returns single application resource
	GetResource(ctxt context.Context, query *application.ApplicationResourceRequest) (*application.ApplicationResourceResponse, error)
	// Get returns an application by name
	Get(ctx context.Context, query *application.ApplicationQuery) (*v1alpha1.Application, error)
	// Create creates an application
	Create(ctx context.Context, query *application.ApplicationCreateRequest) (*v1alpha1.Application, error)
	// Update updates an application
	Update(ctx context.Context, query *application.ApplicationUpdateRequest) (*v1alpha1.Application, error)
	// Sync syncs an application to its target state
	Sync(ctx context.Context, query *application.ApplicationSyncRequest) (*v1alpha1.Application, error)
	// TerminateOperation terminates the currently running operation
	TerminateOperation(ctx context.Context, query *application.OperationTerminateRequest) (*application.OperationTerminateResponse, error)
	// PatchResource patch single application resource
	PatchResource(ctx context.Context, query *application.ApplicationResourcePatchRequest) (*application.ApplicationResourceResponse, error)
	// DeleteResource deletes a single application resource
	DeleteResource(ctx context.Context, query *application.ApplicationResourceDeleteRequest) (*application.ApplicationResponse, error)
	// Delete deletes an application
	Delete(ctx context.Context, query *application.ApplicationDeleteRequest) (*application.ApplicationResponse, error)
}

type Result struct {
	Response *application.ApplicationResourceResponse
	Error    error
	Request  *application.ApplicationResourceRequest
}

type ResourceTreeResponse struct {
	*v1alpha1.ApplicationTree
	NewGenerationReplicaSet string                          `json:"newGenerationReplicaSet"`
	Status                  string                          `json:"status"`
	PodMetadata             []*PodMetadata                  `json:"podMetadata"`
	Conditions              []v1alpha1.ApplicationCondition `json:"conditions"`
}

type PodMetadata struct {
	Name       string    `json:"name"`
	UID        string    `json:"uid"`
	Containers []*string `json:"containers"`
	IsNew      bool      `json:"isNew"`
}

type Manifests struct {
	rolloutManifest     map[string]interface{}
	deploymentManifest  map[string]interface{}
	replicaSetManifests []map[string]interface{}
	serviceManifests    []map[string]interface{}
}

type ServiceClientImpl struct {
	settings *settings.ArgoCDSettings
	logger   *zap.SugaredLogger
}

func NewApplicationClientImpl(
	settings *settings.ArgoCDSettings,
	logger *zap.SugaredLogger,
) *ServiceClientImpl {
	return &ServiceClientImpl{
		settings: settings,
		logger:   logger,
	}
}

func (c ServiceClientImpl) ManagedResources(ctxt context.Context, query *application.ResourcesQuery) (*application.ManagedResourcesResponse, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.ManagedResources(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) Rollback(ctxt context.Context, query *application.ApplicationRollbackRequest) (*v1alpha1.Application, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.Rollback(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) Patch(ctxt context.Context, query *application.ApplicationPatchRequest) (*v1alpha1.Application, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.Patch(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) GetManifests(ctxt context.Context, query *application.ApplicationManifestQuery) (*apiclient.ManifestResponse, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.GetManifests(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) Get(ctxt context.Context, query *application.ApplicationQuery) (*v1alpha1.Application, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.Get(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) Update(ctxt context.Context, query *application.ApplicationUpdateRequest) (*v1alpha1.Application, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.Update(ctx, query)
	return resp, err
}

type ErrUnauthorized struct {
	message string
}

func NewErrUnauthorized(message string) *ErrUnauthorized {
	return &ErrUnauthorized{
		message: message,
	}
}
func (e *ErrUnauthorized) Error() string {
	return e.message
}

func (c ServiceClientImpl) Sync(ctxt context.Context, query *application.ApplicationSyncRequest) (*v1alpha1.Application, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, NewErrUnauthorized("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.Sync(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) TerminateOperation(ctxt context.Context, query *application.OperationTerminateRequest) (*application.OperationTerminateResponse, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.TerminateOperation(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) PatchResource(ctxt context.Context, query *application.ApplicationResourcePatchRequest) (*application.ApplicationResourceResponse, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.PatchResource(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) DeleteResource(ctxt context.Context, query *application.ApplicationResourceDeleteRequest) (*application.ApplicationResponse, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.DeleteResource(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) List(ctxt context.Context, query *application.ApplicationQuery) (*v1alpha1.ApplicationList, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.List(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) PodLogs(ctx context.Context, query *application.ApplicationPodLogsQuery) (application.ApplicationService_PodLogsClient, *grpc.ClientConn, error) {
	token, _ := ctx.Value("token").(string)
	conn := argocdServer.GetConnection(token, c.settings)
	//defer conn.Close()
	asc := application.NewApplicationServiceClient(conn)
	logs, err := asc.PodLogs(ctx, query)
	return logs, conn, err
}

func (c ServiceClientImpl) Watch(ctx context.Context, query *application.ApplicationQuery) (application.ApplicationService_WatchClient, *grpc.ClientConn, error) {
	token, _ := ctx.Value("token").(string)
	conn := argocdServer.GetConnection(token, c.settings)
	//defer conn.Close()
	asc := application.NewApplicationServiceClient(conn)
	logs, err := asc.Watch(ctx, query)
	return logs, conn, err
}

func (c ServiceClientImpl) GetResource(ctxt context.Context, query *application.ApplicationResourceRequest) (*application.ApplicationResourceResponse, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	return asc.GetResource(ctx, query)
}

func (c ServiceClientImpl) Delete(ctxt context.Context, query *application.ApplicationDeleteRequest) (*application.ApplicationResponse, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	return asc.Delete(ctx, query)
}

func (c ServiceClientImpl) ListResourceEvents(ctxt context.Context, query *application.ApplicationResourceEventsQuery) (*v1.EventList, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.ListResourceEvents(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) Create(ctxt context.Context, query *application.ApplicationCreateRequest) (*v1alpha1.Application, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.Create(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) ResourceTree(ctxt context.Context, query *application.ResourcesQuery) (*ResourceTreeResponse, error) {
	ctx, cancel := context.WithTimeout(ctxt, TimeoutSlow)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	conn := argocdServer.GetConnection(token, c.settings)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	c.logger.Debugw("GRPC_GET_RESOURCETREE", "req", query)
	resp, err := asc.ResourceTree(ctx, query)
	if err != nil {
		c.logger.Errorw("GRPC_GET_RESOURCETREE", "req", query, "err", err)
		return nil, err
	}
	responses := parseResult(resp, query, ctx, asc, err, c)
	podMetadata, newReplicaSet := c.buildPodMetadata(resp, responses)

	appQuery := application.ApplicationQuery{Name: query.ApplicationName}
	app, err := asc.Watch(ctxt, &appQuery)
	var conditions = make([]v1alpha1.ApplicationCondition, 0)
	status := "Unknown"
	if app != nil {
		appResp, err := app.Recv()
		if err == nil {
			status = appResp.Application.Status.Health.Status
			conditions = appResp.Application.Status.Conditions
			for _, condition := range conditions {
				if condition.Type != v1alpha1.ApplicationConditionSharedResourceWarning {
					status = "Degraded"
				}
			}
			if status == "" {
				status = "Unknown"
			}
		}
	}
	return &ResourceTreeResponse{resp, newReplicaSet, status, podMetadata, conditions}, err
}

func (c ServiceClientImpl) buildPodMetadata(resp *v1alpha1.ApplicationTree, responses []*Result) (podMetadata []*PodMetadata, newReplicaSet string) {
	rolloutManifest := make(map[string]interface{})
	statefulSetManifest := make(map[string]interface{})
	deploymentManifest := make(map[string]interface{})
	daemonSetManifest := make(map[string]interface{})
	replicaSetManifests := make([]map[string]interface{}, 0)
	podManifests := make([]map[string]interface{}, 0)
	controllerRevisionManifests := make([]map[string]interface{}, 0)
	for _, response := range responses {
		if response != nil && response.Response != nil && response.Request.Kind == "Rollout" {
			err := json.Unmarshal([]byte(response.Response.Manifest), &rolloutManifest)
			if err != nil {
				c.logger.Error(err)
			}
		} else if response != nil && response.Response != nil && response.Request.Kind == "Deployment" {
			err := json.Unmarshal([]byte(response.Response.Manifest), &deploymentManifest)
			if err != nil {
				c.logger.Error(err)
			}
		} else if response != nil && response.Response != nil && response.Request.Kind == "StatefulSet" {
			err := json.Unmarshal([]byte(response.Response.Manifest), &statefulSetManifest)
			if err != nil {
				c.logger.Error(err)
			}
		} else if response != nil && response.Response != nil && response.Request.Kind == "DaemonSet" {
			err := json.Unmarshal([]byte(response.Response.Manifest), &daemonSetManifest)
			if err != nil {
				c.logger.Error(err)
			}
		} else if response != nil && response.Response != nil && response.Request.Kind == "ReplicaSet" {
			manifest := make(map[string]interface{})
			err := json.Unmarshal([]byte(response.Response.Manifest), &manifest)
			if err != nil {
				c.logger.Error(err)
			}
			replicaSetManifests = append(replicaSetManifests, manifest)
		} else if response != nil && response.Response != nil && response.Request.Kind == "Pod" {
			manifest := make(map[string]interface{})
			err := json.Unmarshal([]byte(response.Response.Manifest), &manifest)
			if err != nil {
				c.logger.Error(err)
			}
			podManifests = append(podManifests, manifest)
		} else if response != nil && response.Response != nil && response.Request.Kind == "ControllerRevision" {
			manifest := make(map[string]interface{})
			err := json.Unmarshal([]byte(response.Response.Manifest), &manifest)
			if err != nil {
				c.logger.Error(err)
			}
			controllerRevisionManifests = append(controllerRevisionManifests, manifest)
		}
	}
	newPodNames := make(map[string]bool, 0)
	// for rollout we compare pod hash
	if _, ok := rolloutManifest["kind"]; ok {
		newReplicaSet = c.getRolloutNewReplicaSetName(rolloutManifest, replicaSetManifests)
	}

	if _, ok := deploymentManifest["kind"]; ok {
		newReplicaSet = c.getDeploymentNewReplicaSetName(deploymentManifest, replicaSetManifests)
	}

	if _, ok := statefulSetManifest["kind"]; ok {
		newPodNames = c.getStatefulSetNewPods(statefulSetManifest, podManifests)
	}

	if _, ok := daemonSetManifest["kind"]; ok {
		newPodNames = c.getDaemonSetNewPods(daemonSetManifest, podManifests, controllerRevisionManifests)
	}

	if newReplicaSet != "" {
		podMetadata = buildPodMetadataFromReplicaSet(resp, newReplicaSet, replicaSetManifests)
	} else {
		podMetadata = buildPodMetadataFromPod(resp, podManifests, newPodNames)
	}
	return
}

func parseResult(resp *v1alpha1.ApplicationTree, query *application.ResourcesQuery, ctx context.Context, asc application.ApplicationServiceClient, err error, c ServiceClientImpl) []*Result {
	var responses = make([]*Result, 0)
	qCount := 0
	response := make(chan Result)
	if err != nil || resp == nil || len(resp.Nodes) == 0 {
		return responses
	}
	needPods := false
	queryNodes := make([]v1alpha1.ResourceNode, 0)
	podParents := make([]string, 0)
	for _, node := range resp.Nodes {
		if node.Kind == "Pod" {
			for _, pr := range node.ParentRefs {
				podParents = append(podParents, pr.Name)
			}
		}
	}
	for _, node := range resp.Nodes {
		if node.Kind == "Rollout" || node.Kind == "Deployment" || node.Kind == "StatefulSet" || node.Kind == "DaemonSet" {
			queryNodes = append(queryNodes, node)
		}
		if node.Kind == "ReplicaSet" {
			for _, pr := range podParents {
				if pr == node.Name {
					queryNodes = append(queryNodes, node)
					break
				}
			}
		}
		if node.Kind == "StatefulSet" || node.Kind == "DaemonSet" {
			needPods = true
		}
	}

	c.logger.Debugw("needPods", "pods", needPods)

	if needPods {
		for _, node := range resp.Nodes {
			if node.Kind == "Pod" {
				queryNodes = append(queryNodes, node)
			}
		}
	}

	relevantCR := make(map[string]bool)
	for _, node := range resp.Nodes {
		prefix := ""
		if len(node.ParentRefs) > 0 {
			for _, p := range node.ParentRefs {
				if p.Kind == "DaemonSet" {
					prefix = p.Name
				}
			}
		}
		if node.Kind == "Pod" {
			relevantCR[prefix+"-"+node.NetworkingInfo.Labels["controller-revision-hash"]] = true
		}
	}

	for _, node := range resp.Nodes {
		if node.Kind == "ControllerRevision" {
			if ok := relevantCR[node.Name]; ok {
				queryNodes = append(queryNodes, node)
			}
		}
	}

	for _, node := range queryNodes {
		rQuery := transform(node, query.ApplicationName)
		qCount++
		go func(request application.ApplicationResourceRequest) {
			ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
			defer cancel()
			startTime := time.Now()
			res, err := asc.GetResource(ctx, &request)
			if err != nil {
				c.logger.Errorw("GRPC_GET_RESOURCE", "data", request, "timeTaken", time.Now().Sub(startTime), "err", err)
			} else {
				c.logger.Debugw("GRPC_GET_RESOURCE", "data", request, "timeTaken", time.Now().Sub(startTime))
			}
			if res != nil || err != nil {
				response <- Result{Response: res, Error: err, Request: &request}
			} else {
				response <- Result{Response: nil, Error: fmt.Errorf("connection closed by client"), Request: &request}
			}
		}(*rQuery)
	}

	if qCount == 0 {
		return responses
	}

	rCount := 0
	for {
		select {
		case msg, ok := <-response:
			if ok {
				if msg.Error == nil {
					responses = append(responses, &msg)
				}
			}
			rCount++
		}
		if qCount == rCount {
			break
		}
	}
	return responses
}

func (c ServiceClientImpl) getDeploymentNewReplicaSetName(deploymentManifest map[string]interface{}, replicaSetManifests []map[string]interface{}) (newReplicaSet string) {
	d, err := json.Marshal(deploymentManifest)
	if err != nil {
		return
	}
	deployment := &v12.Deployment{}
	err = json.Unmarshal(d, deployment)
	if err != nil {
		return
	}
	dPodHash := util.ComputeHash(&deployment.Spec.Template, deployment.Status.CollisionCount)
	for _, rs := range replicaSetManifests {
		r, err := json.Marshal(rs)
		if err != nil {
			return
		}
		replicaset := &v12.ReplicaSet{}
		err = json.Unmarshal(r, replicaset)
		if err != nil {
			continue
		}
		rsCopy := replicaset.Spec.DeepCopy()
		labels := make(map[string]string)
		for k, v := range rsCopy.Template.Labels {
			if k != "pod-template-hash" {
				labels[k] = v
			}
		}
		rsCopy.Template.Labels = labels
		podHash := util.ComputeHash(&rsCopy.Template, deployment.Status.CollisionCount)
		if podHash == dPodHash {
			newReplicaSet = getResourceName(rs)
		}
	}
	return
}

func (c ServiceClientImpl) getDaemonSetNewPods(daemonSetManifest map[string]interface{}, podManifests []map[string]interface{}, controllerRevisionManifests []map[string]interface{}) (newPodNames map[string]bool) {
	d, err := json.Marshal(daemonSetManifest)
	if err != nil {
		return
	}
	daemonSet := &v12.DaemonSet{}
	err = json.Unmarshal(d, daemonSet)
	if err != nil {
		return
	}
	latestRevision := ""
	latestGen := 0
	newPodNames = make(map[string]bool, 0)
	for _, crm := range controllerRevisionManifests {
		rev := int(crm["revision"].(float64))
		if latestGen < rev {
			latestGen = rev
			latestRevision = getDaemonSetPodControllerRevisionHash(crm)
		}
	}
	for _, pod := range podManifests {
		podRevision := getDaemonSetPodControllerRevisionHash(pod)
		if latestRevision == podRevision {
			newPodNames[getResourceName(pod)] = true
		}
	}
	return
}

func getDaemonSetPodControllerRevisionHash(pod map[string]interface{}) string {
	if md, ok := pod["metadata"]; ok {
		if mdm, ok := md.(map[string]interface{}); ok {
			if l, ok := mdm["labels"]; ok {
				if lm, ok := l.(map[string]interface{}); ok {
					if h, ok := lm["controller-revision-hash"]; ok {
						if hs, ok := h.(string); ok {
							return hs
						}
					}
				}
			}
		}
	}
	return ""
}

func (c ServiceClientImpl) getStatefulSetNewPods(statefulSetManifest map[string]interface{}, podManifests []map[string]interface{}) (newPodNames map[string]bool) {
	newPodNames = make(map[string]bool, 0)
	updateRevision := getStatefulSetUpdateRevision(statefulSetManifest)
	for _, pod := range podManifests {
		podRevision := getStatefulSetPodControllerRevisionHash(pod)
		if updateRevision == podRevision {
			newPodNames[getResourceName(pod)] = true
		}
	}
	return
}

func getStatefulSetUpdateRevision(statefulSet map[string]interface{}) string {
	if s, ok := statefulSet["status"]; ok {
		if sm, ok := s.(map[string]interface{}); ok {
			if cph, ok := sm["updateRevision"]; ok {
				if cphs, ok := cph.(string); ok {
					return cphs
				}
			}
		}
	}
	return ""
}

func getStatefulSetPodControllerRevisionHash(pod map[string]interface{}) string {
	if md, ok := pod["metadata"]; ok {
		if mdm, ok := md.(map[string]interface{}); ok {
			if l, ok := mdm["labels"]; ok {
				if lm, ok := l.(map[string]interface{}); ok {
					if h, ok := lm["controller-revision-hash"]; ok {
						if hs, ok := h.(string); ok {
							return hs
						}
					}
				}
			}
		}
	}
	return ""
}

func (c ServiceClientImpl) getRolloutNewReplicaSetName(rManifest map[string]interface{}, replicaSetManifests []map[string]interface{}) (newReplicaSet string) {
	rPodHash := getRolloutPodHash(rManifest)
	for _, rs := range replicaSetManifests {
		podHash := getRolloutPodTemplateHash(rs)
		if podHash == rPodHash {
			newReplicaSet = getResourceName(rs)
		}
	}
	return newReplicaSet
}

func getRolloutPodHash(rollout map[string]interface{}) string {
	if s, ok := rollout["status"]; ok {
		if sm, ok := s.(map[string]interface{}); ok {
			if cph, ok := sm["currentPodHash"]; ok {
				if cphs, ok := cph.(string); ok {
					return cphs
				}
			}
		}
	}
	return ""
}

func getRolloutPodTemplateHash(pod map[string]interface{}) string {
	if md, ok := pod["metadata"]; ok {
		if mdm, ok := md.(map[string]interface{}); ok {
			if l, ok := mdm["labels"]; ok {
				if lm, ok := l.(map[string]interface{}); ok {
					if h, ok := lm["rollouts-pod-template-hash"]; ok {
						if hs, ok := h.(string); ok {
							return hs
						}
					}
				}
			}
		}
	}
	return ""
}

func buildPodMetadataFromPod(resp *v1alpha1.ApplicationTree, podManifests []map[string]interface{}, newPodNames map[string]bool) (podMetadata []*PodMetadata) {
	containerMapping := make(map[string][]*string)
	for _, pod := range podManifests {
		containerMapping[getResourceName(pod)] = getPodContainers(pod)
	}
	for _, node := range resp.Nodes {
		if node.Kind == "Pod" {
			isNew := newPodNames[node.Name]
			metadata := PodMetadata{Name: node.Name, UID: node.UID, Containers: containerMapping[node.Name], IsNew: isNew}
			podMetadata = append(podMetadata, &metadata)
		}
	}
	return
}

func getPodContainers(resource map[string]interface{}) []*string {
	containers := make([]*string, 0)
	if s, ok := resource["spec"]; ok {
		if sm, ok := s.(map[string]interface{}); ok {
			if c, ok := sm["containers"]; ok {
				if cas, ok := c.([]interface{}); ok {
					for _, ca := range cas {
						if cam, ok := ca.(map[string]interface{}); ok {
							if n, ok := cam["name"]; ok {
								if cn, ok := n.(string); ok {
									containers = append(containers, &cn)
								}
							}
						}
					}
				}
			}
		}
	}
	return containers
}

func buildPodMetadataFromReplicaSet(resp *v1alpha1.ApplicationTree, newReplicaSet string, replicaSetManifests []map[string]interface{}) (podMetadata []*PodMetadata) {
	containerMapping := make(map[string][]*string)
	replicaSets := make(map[string]map[string]interface{})
	for _, replicaSet := range replicaSetManifests {
		containerMapping[getResourceName(replicaSet)] = getReplicaSetContainers(replicaSet)
		replicaSets[getResourceName(replicaSet)] = replicaSet
	}
	for _, node := range resp.Nodes {
		if node.Kind == "Pod" {
			parentName := ""
			for _, p := range node.ParentRefs {
				if p.Kind == "ReplicaSet" {
					parentName = p.Name
				}
			}
			isNew := parentName == newReplicaSet
			replicaSet := replicaSets[parentName]
			metadata := PodMetadata{Name: node.Name, UID: node.UID, Containers: getReplicaSetContainers(replicaSet), IsNew: isNew}
			podMetadata = append(podMetadata, &metadata)
		}
	}
	return
}

func getReplicaSetContainers(resource map[string]interface{}) []*string {
	containers := make([]*string, 0)
	if s, ok := resource["spec"]; ok {
		if sm, ok := s.(map[string]interface{}); ok {
			if t, ok := sm["template"]; ok {
				if tm, ok := t.(map[string]interface{}); ok {
					if tms, ok := tm["spec"]; ok {
						if tmsm, ok := tms.(map[string]interface{}); ok {
							if c, ok := tmsm["containers"]; ok {
								if cas, ok := c.([]interface{}); ok {
									for _, ca := range cas {
										if cam, ok := ca.(map[string]interface{}); ok {
											if n, ok := cam["name"]; ok {
												if cn, ok := n.(string); ok {
													containers = append(containers, &cn)
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return containers
}

func getResourceName(resource map[string]interface{}) string {
	if md, ok := resource["metadata"]; ok {
		if mdm, ok := md.(map[string]interface{}); ok {
			if h, ok := mdm["name"]; ok {
				if hs, ok := h.(string); ok {
					return hs
				}
			}
		}
	}
	return ""
}

func transform(resource v1alpha1.ResourceNode, name *string) *application.ApplicationResourceRequest {
	request := &application.ApplicationResourceRequest{
		Name:         name,
		ResourceName: resource.Name,
		Kind:         resource.Kind,
		Group:        resource.Group,
		Version:      resource.Version,
		Namespace:    resource.Namespace,
	}

	return request
}
