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
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
)

type ResourceService interface {
	FetchResourceTree(appName string) (*Items, error)
	FetchPodContainerLogs(appName string, podName string, req PodContainerLogReq) (*Items, error)
}
type ResourceServiceImpl struct {
	*ArgoSession
	id       int
	location string
}

func NewResourceServiceImpl(session *ArgoSession) *ResourceServiceImpl {
	return &ResourceServiceImpl{
		ArgoSession: session,
		location:    "/api/v1/applications",
	}
}

func (impl *ResourceServiceImpl) FetchResourceTree(appName string) (*Items, error) {

	path := impl.location + "/" + appName + "/resource-tree"
	res := &Items{}
	_, _, err := impl.DoRequest(&ClientRequest{ResponseBody: res, Path: path, Method: "GET"})
	if err != nil {
		return nil, err
	}
	//writeJsonResp(w, err, resJson, http.StatusOK)
	return res, nil
}

func (impl *ResourceServiceImpl) FetchPodContainerLogs(appName string, podName string, req PodContainerLogReq) (*Items, error) {

	path := impl.location + "/" + appName + "/pods/" + podName + "/logs"
	req2 := &PodContainerLogReq{
		Namespace: "dont-use",
		Container: "application",
		Follow:    true,
	}
	_, _, err := impl.DoRequest(&ClientRequest{ResponseBody: nil, Path: path, RequestBody: req2, Method: "GET"})
	if err != nil {
		return nil, err
	}
	//writeJsonResp(w, err, resJson, http.StatusOK)
	return nil, nil
}

type Items struct {
	Items []Resource `json:"items"`
}

type Resource struct {
	v1alpha1.ResourceNode `json:",inline" protobuf:"bytes,1,opt,name=resourceNode"`
	//Children              []v1alpha1.ResourceRef `json:"children,omitempty" protobuf:"bytes,3,opt,name=children"`
}

type PodContainerLogReq struct {
	Namespace    string `json:"namespace,omitempty"`
	Container    string `json:"container,omitempty"`
	SinceSeconds string `json:"sinceSeconds,omitempty"`
	TailLines    string `json:"tailLines,omitempty"`
	Follow       bool   `json:"follow,omitempty"`
}
