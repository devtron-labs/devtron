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
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.uber.org/zap"
	"io/ioutil"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
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
	GetArgoApplication(namespace string, appName string, cluster *repository.Cluster) (map[string]interface{}, error)
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

func (impl ArgoK8sClientImpl) handleArgoAppCreationError(res []byte, err error) error {
	// default error set
	apiError := &util.ApiError{
		InternalMessage: "error creating argo cd app",
		UserMessage:     "error creating argo cd app",
	}
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

func (impl ArgoK8sClientImpl) GetArgoApplication(namespace string, appName string, cluster *repository.Cluster) (map[string]interface{}, error) {

	config, err := rest.InClusterConfig()
	if err != nil {
		impl.logger.Errorw("error in cluster config", "err", err)
		return nil, err
	}
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
		Namespace(namespace).
		Resource("applications").
		Name(appName).
		//VersionedParams(&opts, metav1.ParameterCodec).
		Do(context.Background()).Raw()
	response := make(map[string]interface{})
	if err != nil {
		err := json.Unmarshal(res, &response)
		if err != nil {
			impl.logger.Errorw("unmarshal error on app update status", "err", err)
			return nil, fmt.Errorf("error get argo cd app")
		}
	}
	impl.logger.Infow("get argo cd application", "res", response, "err", err)
	return response, err
}
