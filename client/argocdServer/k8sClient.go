/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package argocdServer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/sync/common"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"go.uber.org/zap"
	"io/ioutil"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"path/filepath"
	"text/template"
	"time"
)

type AppTemplate struct {
	ApplicationName string
	Namespace       string
	TargetNamespace string
	TargetServer    string
	Project         string
	ValuesFile      string
	RepoPath        string
	RepoUrl         string
	AutoSyncEnabled bool
}

const (
	TimeoutSlow                 = 30 * time.Second
	ARGOCD_APPLICATION_TEMPLATE = "./scripts/argo-assets/APPLICATION_TEMPLATE.tmpl"
)

type ArgoK8sClient interface {
	CreateAcdApp(ctx context.Context, appRequest *AppTemplate, applicationTemplatePath string) (string, error)
	GetArgoApplication(k8sConfig *bean.ArgoK8sConfig, appName string) (*v1alpha1.Application, error)
	DeleteArgoApplication(ctx context.Context, k8sConfig *bean.ArgoK8sConfig, appName string, cascadeDelete bool) error
	SyncArgoApplication(ctx context.Context, k8sConfig *bean.ArgoK8sConfig, appName string) error
	RefreshApp(ctx context.Context, k8sConfig *bean.ArgoK8sConfig, appName, refreshType string) error
	TerminateApp(ctx context.Context, k8sConfig *bean.ArgoK8sConfig, appName string) error
}
type ArgoK8sClientImpl struct {
	logger  *zap.SugaredLogger
	k8sUtil *k8s.K8sServiceImpl
}

func NewArgoK8sClientImpl(logger *zap.SugaredLogger,
	k8sUtil *k8s.K8sServiceImpl,
) *ArgoK8sClientImpl {
	return &ArgoK8sClientImpl{
		logger:  logger,
		k8sUtil: k8sUtil,
	}
}

const DevtronInstalationNs = "devtroncd"

// Tprintf passed template string is formatted usign its operands and returns the resulting string.
// Spaces are added between operands when neither is a string.
func (impl ArgoK8sClientImpl) tprintf(tmpl string, data interface{}) (string, error) {
	t := template.Must(template.New("tpl").Parse(tmpl))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (impl ArgoK8sClientImpl) CreateAcdApp(ctx context.Context, appRequest *AppTemplate, applicationTemplatePath string) (string, error) {
	chartYamlContent, err := ioutil.ReadFile(filepath.Clean(applicationTemplatePath))
	if err != nil {
		impl.logger.Errorw("err in reading template", "err", err)
		return "", err
	}
	applicationRequestString, err := impl.tprintf(string(chartYamlContent), appRequest)
	if err != nil {
		impl.logger.Errorw("error in rendering application template", "req", appRequest, "err", err)
		return "", err
	}

	config, err := impl.k8sUtil.GetK8sInClusterRestConfig()
	if err != nil {
		impl.logger.Errorw("error in getting in cluster rest config", "err", err)
		return "", err
	}
	config.GroupVersion = &schema.GroupVersion{Group: "argoproj.io", Version: "v1alpha1"}
	config.NegotiatedSerializer = serializer.NewCodecFactory(runtime.NewScheme())
	config.APIPath = "/apis"
	config.Timeout = TimeoutSlow
	err = impl.CreateArgoApplication(ctx, appRequest.Namespace, applicationRequestString, config)
	if err != nil {
		impl.logger.Errorw("error in creating acd application", "err", err)
		return "", err
	}
	impl.logger.Infow("argo application created successfully", "name", appRequest.ApplicationName)
	return appRequest.ApplicationName, nil
}

func (impl ArgoK8sClientImpl) CreateArgoApplication(ctx context.Context, namespace string, application string, config *rest.Config) error {
	client, err := rest.RESTClientFor(config)
	if err != nil {
		return fmt.Errorf("error creating argo cd app")
	}
	impl.logger.Debugw("creating argo application resource", "application", application)
	res, err := client.
		Post().
		Resource("applications").
		Namespace(namespace).
		Body([]byte(application)).
		Do(ctx).Raw()

	if err != nil {
		impl.logger.Errorw("error in argo application resource creation", "namespace", namespace, "res", res, "err", err)
		return impl.handleArgoAppCreationError(res, err)
	}

	impl.logger.Infow("argo app create successfully", "namespace", namespace, "res", string(res))
	return err
}

func (impl ArgoK8sClientImpl) handleArgoAppGetError(res []byte, err error) error {
	// default error set
	apiError := &util.ApiError{
		InternalMessage: "error getting argo cd application",
		UserMessage:     "error getting argo cd application",
	}
	return impl.convertArgoK8sClientError(apiError, res, err)
}

func (impl ArgoK8sClientImpl) handleArgoAppCreationError(res []byte, err error) error {
	// default error set
	apiError := &util.ApiError{
		InternalMessage: "error creating argo cd app",
		UserMessage:     "error creating argo cd app",
	}
	return impl.convertArgoK8sClientError(apiError, res, err)
}

func (impl ArgoK8sClientImpl) convertArgoK8sClientError(apiError *util.ApiError, res []byte, err error) error {
	// error override for errors.StatusError
	if statusError := (&k8sError.StatusError{}); errors.As(err, &statusError) {
		apiError.HttpStatusCode = int(statusError.Status().Code)
		apiError.InternalMessage = statusError.Error()
		apiError.UserMessage = statusError.Error()
	}
	response := make(map[string]interface{})
	jsonErr := json.Unmarshal(res, &response)
	if jsonErr != nil {
		impl.logger.Errorw("unmarshal error on app update status", "err", jsonErr)
		return apiError
	}
	// error override if API response exists, as response errors are more readable
	if response != nil {
		if statusCode, ok := response["code"]; apiError.HttpStatusCode == 0 && ok {
			if statusCodeFloat, ok := statusCode.(float64); ok {
				apiError.HttpStatusCode = int(statusCodeFloat)
			}
		}
		if response["message"] != nil {
			errMsg := response["message"].(string)
			apiError.InternalMessage = errMsg
			apiError.UserMessage = errMsg
		}
	}
	return apiError
}

func (impl ArgoK8sClientImpl) GetArgoApplication(k8sConfig *bean.ArgoK8sConfig, appName string) (*v1alpha1.Application, error) {

	config := k8sConfig.RestConfig
	config.GroupVersion = &schema.GroupVersion{Group: "argoproj.io", Version: "v1alpha1"}
	config.NegotiatedSerializer = serializer.NewCodecFactory(runtime.NewScheme())
	config.APIPath = "/apis"
	config.Timeout = TimeoutSlow
	client, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}
	impl.logger.Infow("get argo cd application", "req", appName)
	//acdApplication := &v1alpha12.Application{}
	//opts := metav1.GetOptions{}
	res, err := client.
		Get().
		Namespace(k8sConfig.AcdNamespace).
		Resource("applications").
		Name(appName).
		//VersionedParams(&opts, metav1.ParameterCodec).
		Do(context.Background()).Raw()
	response := make(map[string]interface{})
	if err != nil {
		impl.logger.Errorw("error in get argo application", "err", err)
		return nil, impl.handleArgoAppGetError(res, err)
	}
	err = json.Unmarshal(res, &response)
	if err != nil {
		impl.logger.Errorw("unmarshal error on app update status", "err", err)
		return nil, fmt.Errorf("error get argo cd app")
	}
	application, err := GetAppObject(response)
	if err != nil {
		impl.logger.Errorw("error in getting app object", "deploymentAppName", appName, "err", err)
		return nil, err
	}
	impl.logger.Infow("get argo cd application", "res", response, "err", err)
	return application, err
}

func (impl ArgoK8sClientImpl) DeleteArgoApplication(ctx context.Context, k8sConfig *bean.ArgoK8sConfig, appName string, cascadeDelete bool) error {

	metadata := make(map[string]interface{})
	if cascadeDelete {
		//patchJSON = `{"metadata": {"finalizers": ["resources-finalizer.argocd.argoproj.io"]}}`
		metadata = map[string]interface{}{
			"metadata": map[string]interface{}{
				"finalizers": []string{"resources-finalizer.argocd.argoproj.io"},
			},
		}
	} else {
		//patchJSON = `{"metadata": {"finalizers": null}}`
		metadata = map[string]interface{}{
			"metadata": map[string]interface{}{
				"finalizers": nil,
			},
		}
	}
	err := impl.patchApplicationObject(ctx, k8sConfig, appName, metadata)
	if err != nil {
		impl.logger.Errorw("error in patching argo application", "appName", appName, "err", err)
		return err
	}

	_, err = impl.k8sUtil.DeleteResource(ctx, k8sConfig.RestConfig, v1alpha1.ApplicationSchemaGroupVersionKind, k8sConfig.AcdNamespace, appName, true)
	if err != nil {
		impl.logger.Errorw("error in patching argo application", "acdAppName", appName, "err", err)
		return err
	}

	return nil
}

func (impl ArgoK8sClientImpl) SyncArgoApplication(ctx context.Context, k8sConfig *bean.ArgoK8sConfig, appName string) error {
	impl.logger.Infow("syncing argo application", "appName", appName)
	operation := map[string]interface{}{
		"operation": map[string]interface{}{
			"sync": map[string]interface{}{
				"prune": true,
			},
			"retry": map[string]interface{}{
				"limit": 1,
			},
		},
	}
	err := impl.patchApplicationObject(ctx, k8sConfig, appName, operation)
	if err != nil {
		impl.logger.Errorw("error in patching argo application", "appName", appName, "err", err)
		return err
	}
	return nil
}

func (impl ArgoK8sClientImpl) patchApplicationObject(ctx context.Context, k8sConfig *bean.ArgoK8sConfig, appName string, patch map[string]interface{}) error {
	impl.logger.Infow("patching argo application", "appName", appName, "patch", patch)
	patchJSON, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("error marshaling metadata: %w", err)
	}
	patchType := types.MergePatchType
	applicationGVK := v1alpha1.ApplicationSchemaGroupVersionKind
	for attempt := 0; attempt < 5; attempt++ {
		_, err := impl.k8sUtil.PatchResourceRequest(ctx, k8sConfig.RestConfig, patchType, string(patchJSON), appName, k8sConfig.AcdNamespace, applicationGVK)
		if err != nil {
			if !k8sError.IsConflict(err) {
				return fmt.Errorf("error patching json in application %s: %w", appName, err)
			}
		} else {
			impl.logger.Infof("patch for application", "appName", appName, "err", err)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (impl ArgoK8sClientImpl) RefreshApp(ctx context.Context, k8sConfig *bean.ArgoK8sConfig, appName, refreshType string) error {
	impl.logger.Infow("refreshing argo application", "appName", appName)
	metadata := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{
				bean.AnnotationKeyRefresh: refreshType,
			},
		},
	}
	err := impl.patchApplicationObject(ctx, k8sConfig, appName, metadata)
	if err != nil {
		impl.logger.Errorw("error in patching argo application", "appName", appName, "err", err)
		return err
	}
	return nil
}

func (impl ArgoK8sClientImpl) TerminateApp(ctx context.Context, k8sConfig *bean.ArgoK8sConfig, appName string) error {
	impl.logger.Infow("terminating argo application", "appName", appName)
	terminatePatch := map[string]interface{}{
		"status": map[string]interface{}{
			"operationState": map[string]interface{}{
				"phase": common.OperationTerminating,
			},
		},
	}
	err := impl.patchApplicationObject(ctx, k8sConfig, appName, terminatePatch)
	if err != nil {
		impl.logger.Errorw("error in patching argo application", "appName", appName, "err", err)
		return err
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		app, err := impl.GetArgoApplication(k8sConfig, appName)
		if err != nil {
			impl.logger.Errorw("error in getting argo application", "appName", appName, "err", err)
			return err
		}
		if app.Operation == nil || app.Status.OperationState.Phase == common.OperationFailed {
			break
		}
		if time.Now().After(deadline) {
			return errors.New("timed out waiting for argo application to terminate")
		}
	}
	return nil
}
