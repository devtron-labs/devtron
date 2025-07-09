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
	GetArgoApplication(k8sConfig *bean.ArgoK8sConfig, appName string) (map[string]interface{}, error)
	DeleteArgoApplication(ctx context.Context, k8sConfig *bean.ArgoK8sConfig, appName string, cascadeDelete bool) error
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

func (impl ArgoK8sClientImpl) GetArgoApplication(k8sConfig *bean.ArgoK8sConfig, appName string) (map[string]interface{}, error) {

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
	impl.logger.Infow("get argo cd application", "res", response, "err", err)
	return response, err
}

func (impl ArgoK8sClientImpl) DeleteArgoApplication(ctx context.Context, k8sConfig *bean.ArgoK8sConfig, appName string, cascadeDelete bool) error {

	patchType := types.MergePatchType
	patchJSON := ""

	if cascadeDelete {
		patchJSON = `{"metadata": {"finalizers": ["resources-finalizer.argocd.argoproj.io"]}}`
	} else {
		patchJSON = `{"metadata": {"finalizers": null}}`
	}

	applicationGVK := v1alpha1.ApplicationSchemaGroupVersionKind

	_, err := impl.k8sUtil.PatchResourceRequest(ctx, k8sConfig.RestConfig, patchType, patchJSON, appName, k8sConfig.AcdNamespace, applicationGVK)
	if err != nil {
		impl.logger.Errorw("error in patching argo application", "err", err)
		return err
	}

	_, err = impl.k8sUtil.DeleteResource(ctx, k8sConfig.RestConfig, applicationGVK, k8sConfig.AcdNamespace, appName, true)
	if err != nil {
		impl.logger.Errorw("error in patching argo application", "acdAppName", appName, "err", err)
		return err
	}

	return nil
}
