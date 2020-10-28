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

import "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"

type RepositoryService interface {
	Create(repositoryRequest *v1alpha1.Repository) (repository *v1alpha1.Repository, err error)
}

type RepositoryServiceImpl struct {
	*ArgoSession
	location string
}

func NewRepositoryService(session *ArgoSession) *RepositoryServiceImpl {
	return &RepositoryServiceImpl{
		ArgoSession: session,
		location:    "/api/v1/repositories",
	}
}

func (impl RepositoryServiceImpl) Create(repositoryRequest *v1alpha1.Repository) (repository *v1alpha1.Repository, err error) {
	res := &v1alpha1.Repository{}
	_, _, err = impl.DoRequest(&ClientRequest{ResponseBody: res, Path: impl.location, Method: "POST", RequestBody: repositoryRequest})
	if err != nil {
		return nil, err
	}
	return res, nil
}
