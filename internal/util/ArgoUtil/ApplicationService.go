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

package ArgoUtil

import (
	"context"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
)

type ApplicationService interface {
	GetAll(ctx context.Context) (*v1alpha1.ApplicationList, error)
	CreateApplication(appRequest *v1alpha1.Application, ctx context.Context) (appResponse *v1alpha1.Application, err error)
	Delete(appName string, ctx context.Context) error
}
type ApplicationServiceImpl struct {
	*ArgoSession
	location string
}

func NewApplicationServiceImpl(session *ArgoSession) *ApplicationServiceImpl {
	return &ApplicationServiceImpl{
		ArgoSession: session,
		location:    "/api/v1/applications",
	}
}

func (impl *ApplicationServiceImpl) GetAll(ctx context.Context) (*v1alpha1.ApplicationList, error) {
	res := &v1alpha1.ApplicationList{}
	_, _, err := impl.DoRequest(&ClientRequest{ResponseBody: res, Path: impl.location, Method: "GET"})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (impl *ApplicationServiceImpl) CreateApplication(appRequest *v1alpha1.Application, ctx context.Context) (appResponse *v1alpha1.Application, err error) {
	res := &v1alpha1.Application{}
	_, _, err = impl.DoRequest(&ClientRequest{ResponseBody: res, Path: impl.location, Method: "POST", RequestBody: appRequest})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (impl *ApplicationServiceImpl) Delete(appName string, ctx context.Context) error {
	res := &v1alpha1.Application{}
	_, _, err := impl.DoRequest(&ClientRequest{
		ResponseBody: res,
		Method:       "DELETE",
		Path:         impl.location + "/" + appName,
	})
	return err
}
