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
	"strings"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/reposerver/apiclient"
	"github.com/argoproj/argo-cd/v2/util/settings"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	Degraded    = "Degraded"
	Healthy     = "Healthy"
	Progressing = "Progressing"
	Suspended   = "Suspended"
	TimeoutFast = 10 * time.Second
	TimeoutSlow = 30 * time.Second
	TimeoutLazy = 60 * time.Second
	HIBERNATING = "HIBERNATING"
	SUCCEEDED   = "Succeeded"
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
	NewGenerationReplicaSets []string                        `json:"newGenerationReplicaSets"`
	Status                   string                          `json:"status"`
	PodMetadata              []*PodMetadata                  `json:"podMetadata"`
	Conditions               []v1alpha1.ApplicationCondition `json:"conditions"`
}

type PodMetadata struct {
	Name           string    `json:"name"`
	UID            string    `json:"uid"`
	Containers     []*string `json:"containers"`
	InitContainers []*string `json:"initContainers"`
	IsNew          bool      `json:"isNew"`
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
	ctx, cancel := context.WithTimeout(ctxt, TimeoutLazy)
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
	ctx, cancel := context.WithTimeout(ctxt, TimeoutSlow)
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
	ctx, cancel := context.WithTimeout(ctxt, TimeoutSlow)
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
	ctx, cancel := context.WithTimeout(ctxt, TimeoutSlow)
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
	ctx, cancel := context.WithTimeout(ctxt, TimeoutSlow)
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
	podMetadata, newReplicaSets := c.buildPodMetadata(resp, responses)

	appQuery := application.ApplicationQuery{Name: query.ApplicationName}
	app, err := asc.Watch(ctxt, &appQuery)
	var conditions = make([]v1alpha1.ApplicationCondition, 0)
	status := "Unknown"
	if app != nil {
		appResp, err := app.Recv()
		if err == nil {
			status = string(appResp.Application.Status.Health.Status)
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
	return &ResourceTreeResponse{resp, newReplicaSets, status, podMetadata, conditions}, err
}

func (c ServiceClientImpl) buildPodMetadata(resp *v1alpha1.ApplicationTree, responses []*Result) (podMetaData []*PodMetadata, newReplicaSets []string) {
	rolloutManifests := make([]map[string]interface{}, 0)
	statefulSetManifest := make(map[string]interface{})
	deploymentManifests := make([]map[string]interface{}, 0)
	daemonSetManifest := make(map[string]interface{})
	replicaSetManifests := make([]map[string]interface{}, 0)
	podManifests := make([]map[string]interface{}, 0)
	controllerRevisionManifests := make([]map[string]interface{}, 0)
	jobsManifest := make(map[string]interface{})
	var parentWorkflow []string
	for _, response := range responses {
		if response != nil && response.Response != nil {
			kind := *response.Request.Kind
			manifestFromResponse := *response.Response.Manifest
			if kind == "Rollout" {
				manifest := make(map[string]interface{})
				err := json.Unmarshal([]byte(manifestFromResponse), &manifest)
				if err != nil {
					c.logger.Error(err)
				}else{
					rolloutManifests = append(rolloutManifests, manifest)
				}
			} else if kind == "Deployment" {
				manifest := make(map[string]interface{})
				err := json.Unmarshal([]byte(manifestFromResponse), &manifest)
				if err != nil {
					c.logger.Error(err)
				}else{
					deploymentManifests = append(deploymentManifests, manifest)
				}
			} else if kind == "StatefulSet" {
				err := json.Unmarshal([]byte(manifestFromResponse), &statefulSetManifest)
				if err != nil {
					c.logger.Error(err)
				}
			} else if kind == "DaemonSet" {
				err := json.Unmarshal([]byte(manifestFromResponse), &daemonSetManifest)
				if err != nil {
					c.logger.Error(err)
				}
			} else if kind == "ReplicaSet" {
				manifest := make(map[string]interface{})
				err := json.Unmarshal([]byte(manifestFromResponse), &manifest)
				if err != nil {
					c.logger.Error(err)
				}
				replicaSetManifests = append(replicaSetManifests, manifest)
			} else if kind == "Pod" {
				manifest := make(map[string]interface{})
				err := json.Unmarshal([]byte(manifestFromResponse), &manifest)
				if err != nil {
					c.logger.Error(err)
				}
				podManifests = append(podManifests, manifest)
			} else if kind == "ControllerRevision" {
				manifest := make(map[string]interface{})
				err := json.Unmarshal([]byte(manifestFromResponse), &manifest)
				if err != nil {
					c.logger.Error(err)
				}
				controllerRevisionManifests = append(controllerRevisionManifests, manifest)
			} else if kind == "Job" {
				err := json.Unmarshal([]byte(manifestFromResponse), &jobsManifest)
				if err != nil {
					c.logger.Error(err)
				}
			}
		}
	}
	newPodNames := make(map[string]bool, 0)
	// for rollout we compare pod hash
	for _, rolloutManifest := range rolloutManifests {
		if _, ok := rolloutManifest["kind"]; ok {
			newReplicaSet := c.getRolloutNewReplicaSetName(rolloutManifest, replicaSetManifests)
			if len(newReplicaSet) > 0 {
				newReplicaSets = append(newReplicaSets, newReplicaSet)
			}
		}
	}

	for _, deploymentManifest := range deploymentManifests {
		if _, ok := deploymentManifest["kind"]; ok {
			newReplicaSet := c.getDeploymentNewReplicaSetName(deploymentManifest, replicaSetManifests)
			if len(newReplicaSet) > 0 {
				newReplicaSets = append(newReplicaSets, newReplicaSet)
			}
		}
	}

	if _, ok := statefulSetManifest["kind"]; ok {
		newPodNames = c.getStatefulSetNewPods(statefulSetManifest, podManifests)
	}

	if _, ok := daemonSetManifest["kind"]; ok {
		newPodNames = c.getDaemonSetNewPods(daemonSetManifest, podManifests, controllerRevisionManifests)
	}

	if _, ok := jobsManifest["kind"]; ok {
		newPodNames = c.getJobsNewPods(jobsManifest, podManifests)
	}

	for _, node := range resp.Nodes {
		if node.Kind == "Workflow" {
			parentWorkflow = append(parentWorkflow, node.Name)
		}
	}

	for _, node := range resp.Nodes {
		if node.Kind == "Pod" {
			if contains(parentWorkflow, node.Name) {
				newPodNames[node.Name] = true
			}
		}
	}

	//podMetaData := make([]*PodMetadata, 0)
	duplicateCheck := make(map[string]bool)
	if len(newReplicaSets) > 0 {
		results := c.buildPodMetadataFromReplicaSet(resp, newReplicaSets, replicaSetManifests)
		for _, meta := range results {
			duplicateCheck[meta.Name] = true
			podMetaData = append(podMetaData, meta)
		}
	}
	if newPodNames != nil {
		results := buildPodMetadataFromPod(resp, podManifests, newPodNames)
		for _, meta := range results {
			if _, ok := duplicateCheck[meta.Name]; !ok {
				podMetaData = append(podMetaData, meta)
			}
		}
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
		if node.Kind == "StatefulSet" || node.Kind == "DaemonSet" || node.Kind == "Workflow" {
			needPods = true
		}

		if node.Kind == "CronJob" || node.Kind == "Job" {
			queryNodes = append(queryNodes, node)
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
				c.logger.Errorw("GRPC_GET_RESOURCE", "data", request, "timeTaken", time.Since(startTime), "err", err)
			} else {
				c.logger.Debugw("GRPC_GET_RESOURCE", "data", request, "timeTaken", time.Since(startTime))
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
	initContainerMapping := make(map[string][]*string)
	for _, pod := range podManifests {
		containerMapping[getResourceName(pod)] = getPodContainers(pod)
	}

	for _, pod := range podManifests {
		initContainerMapping[getResourceName(pod)] = getPodInitContainers(pod)
	}

	for _, node := range resp.Nodes {
		if node.Kind == "Pod" {
			isNew := newPodNames[node.Name]
			metadata := PodMetadata{Name: node.Name, UID: node.UID, Containers: containerMapping[node.Name], InitContainers: initContainerMapping[node.Name], IsNew: isNew}
			podMetadata = append(podMetadata, &metadata)
		}
	}
	return
}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if strings.HasPrefix(v, s) {
			return true
		}
	}
	return false
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
func getPodInitContainers(resource map[string]interface{}) []*string {
	containers := make([]*string, 0)
	if s, ok := resource["spec"]; ok {
		if sm, ok := s.(map[string]interface{}); ok {
			if c, ok := sm["initContainers"]; ok {
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

func (c ServiceClientImpl) buildPodMetadataFromReplicaSet(resp *v1alpha1.ApplicationTree, newReplicaSets []string, replicaSetManifests []map[string]interface{}) (podMetadata []*PodMetadata) {
	replicaSets := make(map[string]map[string]interface{})
	for _, replicaSet := range replicaSetManifests {
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
			if parentName != "" {
				isNew := false
				for _, newReplicaSet := range newReplicaSets {
					if parentName == newReplicaSet {
						isNew = true
						break
					}
				}
				replicaSet := replicaSets[parentName]
				containers, intContainers := getReplicaSetContainers(replicaSet)
				metadata := PodMetadata{Name: node.Name, UID: node.UID, Containers: containers, InitContainers: intContainers, IsNew: isNew}
				podMetadata = append(podMetadata, &metadata)
			}
		}
	}
	return
}

func getReplicaSetContainers(resource map[string]interface{}) (containers []*string, intContainers []*string) {
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
							///initContainers.name
							if c, ok := tmsm["initContainers"]; ok {
								if cas, ok := c.([]interface{}); ok {
									for _, ca := range cas {
										if cam, ok := ca.(map[string]interface{}); ok {
											if n, ok := cam["name"]; ok {
												if cn, ok := n.(string); ok {
													intContainers = append(intContainers, &cn)
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
	return containers, intContainers
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
	resourceName := resource.Name
	kind := resource.Kind
	group := resource.Group
	version := resource.Version
	namespace := resource.Namespace
	request := &application.ApplicationResourceRequest{
		Name:         name,
		ResourceName: &resourceName,
		Kind:         &kind,
		Group:        &group,
		Version:      &version,
		Namespace:    &namespace,
	}
	return request
}

func (c ServiceClientImpl) getJobsNewPods(jobManifest map[string]interface{}, podManifests []map[string]interface{}) (newPodNames map[string]bool) {
	newPodNames = make(map[string]bool, 0)
	for _, pod := range podManifests {
		newPodNames[getResourceName(pod)] = true
	}

	//TODO - new or old logic
	return
}
