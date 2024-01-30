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
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/client/argocdServer/connection"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
)

type ServiceClient interface {
	// ResourceTree	returns the status for all Apps deployed via ArgoCd
	ResourceTree(ctx context.Context, query *application.ResourcesQuery) (*argoApplication.ResourceTreeResponse, error)

	// Patch an ArgoCd application
	Patch(ctx context.Context, query *application.ApplicationPatchRequest) (*v1alpha1.Application, error)

	// GetResource returns single application resource
	GetResource(ctxt context.Context, query *application.ApplicationResourceRequest) (*application.ApplicationResourceResponse, error)

	// Get returns an application by name
	Get(ctx context.Context, query *application.ApplicationQuery) (*v1alpha1.Application, error)

	// Update updates an application
	Update(ctx context.Context, query *application.ApplicationUpdateRequest) (*v1alpha1.Application, error)

	// Sync syncs an application to its target state
	Sync(ctx context.Context, query *application.ApplicationSyncRequest) (*v1alpha1.Application, error)

	// Delete deletes an application
	Delete(ctx context.Context, query *application.ApplicationDeleteRequest) (*application.ApplicationResponse, error)
}

type ServiceClientImpl struct {
	logger                  *zap.SugaredLogger
	argoCDConnectionManager connection.ArgoCDConnectionManager
}

func NewApplicationClientImpl(
	logger *zap.SugaredLogger,
	argoCDConnectionManager connection.ArgoCDConnectionManager) *ServiceClientImpl {
	return &ServiceClientImpl{
		logger:                  logger,
		argoCDConnectionManager: argoCDConnectionManager,
	}
}

func (c ServiceClientImpl) Patch(ctxt context.Context, query *application.ApplicationPatchRequest) (*v1alpha1.Application, error) {
	ctx, cancel := context.WithTimeout(ctxt, argoApplication.TimeoutLazy)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := c.argoCDConnectionManager.GetConnection(token)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.Patch(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) Get(ctxt context.Context, query *application.ApplicationQuery) (*v1alpha1.Application, error) {
	ctx, cancel := context.WithTimeout(ctxt, argoApplication.TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := c.argoCDConnectionManager.GetConnection(token)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.Get(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) Update(ctxt context.Context, query *application.ApplicationUpdateRequest) (*v1alpha1.Application, error) {
	ctx, cancel := context.WithTimeout(ctxt, argoApplication.TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := c.argoCDConnectionManager.GetConnection(token)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.Update(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) Sync(ctxt context.Context, query *application.ApplicationSyncRequest) (*v1alpha1.Application, error) {
	ctx, cancel := context.WithTimeout(ctxt, argoApplication.TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, argoApplication.NewErrUnauthorized("Unauthorized")
	}
	conn := c.argoCDConnectionManager.GetConnection(token)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.Sync(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) GetResource(ctxt context.Context, query *application.ApplicationResourceRequest) (*application.ApplicationResourceResponse, error) {
	ctx, cancel := context.WithTimeout(ctxt, argoApplication.TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := c.argoCDConnectionManager.GetConnection(token)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	return asc.GetResource(ctx, query)
}

func (c ServiceClientImpl) Delete(ctxt context.Context, query *application.ApplicationDeleteRequest) (*application.ApplicationResponse, error) {
	ctx, cancel := context.WithTimeout(ctxt, argoApplication.TimeoutSlow)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := c.argoCDConnectionManager.GetConnection(token)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	return asc.Delete(ctx, query)
}

func (c ServiceClientImpl) ResourceTree(ctxt context.Context, query *application.ResourcesQuery) (*argoApplication.ResourceTreeResponse, error) {
	//all the apps deployed via argo are fetching status from here
	ctx, cancel := context.WithTimeout(ctxt, argoApplication.TimeoutSlow)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	conn := c.argoCDConnectionManager.GetConnection(token)
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
	hash := ""
	if app != nil {
		appResp, err := app.Recv()
		if err == nil {
			// https://github.com/argoproj/argo-cd/issues/11234 workaround
			updateNodeHealthStatus(resp, appResp)

			status = string(appResp.Application.Status.Health.Status)
			hash = appResp.Application.Status.Sync.Revision
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
	return &argoApplication.ResourceTreeResponse{resp, newReplicaSets, status, hash, podMetadata, conditions}, err
}

func (c ServiceClientImpl) buildPodMetadata(resp *v1alpha1.ApplicationTree, responses []*argoApplication.Result) (podMetaData []*argoApplication.PodMetadata, newReplicaSets []string) {
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
				} else {
					rolloutManifests = append(rolloutManifests, manifest)
				}
			} else if kind == "Deployment" {
				manifest := make(map[string]interface{})
				err := json.Unmarshal([]byte(manifestFromResponse), &manifest)
				if err != nil {
					c.logger.Error(err)
				} else {
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
			newReplicaSet := getRolloutNewReplicaSetName(rolloutManifest, replicaSetManifests)
			if len(newReplicaSet) > 0 {
				newReplicaSets = append(newReplicaSets, newReplicaSet)
			}
		}
	}

	for _, deploymentManifest := range deploymentManifests {
		if _, ok := deploymentManifest["kind"]; ok {
			newReplicaSet := getDeploymentNewReplicaSetName(deploymentManifest, replicaSetManifests)
			if len(newReplicaSet) > 0 {
				newReplicaSets = append(newReplicaSets, newReplicaSet)
			}
		}
	}

	if _, ok := statefulSetManifest["kind"]; ok {
		newPodNames = getStatefulSetNewPods(statefulSetManifest, podManifests)
	}

	if _, ok := daemonSetManifest["kind"]; ok {
		newPodNames = getDaemonSetNewPods(daemonSetManifest, podManifests, controllerRevisionManifests)
	}

	if _, ok := jobsManifest["kind"]; ok {
		newPodNames = getJobsNewPods(jobsManifest, podManifests)
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
		results := buildPodMetadataFromReplicaSet(resp, newReplicaSets, replicaSetManifests)
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
