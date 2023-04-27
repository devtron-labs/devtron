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

type ClusterConfig struct {
	BearerToken     string `json:"bearerToken"`
	TLSClientConfig `json:"tlsClientConfig"`
}
type TLSClientConfig struct {
	Insecure bool `json:"insecure"`
}
type ClusterRequest struct {
	Name   string        `json:"name"`
	Server string        `json:"server"`
	Config ClusterConfig `json:"config"`
}
type ClusterResponse struct {
	Id            string        `json:"id"`
	Server        string        `json:"server"`
	Name          string        `json:"name"`
	ServerVersion string        `json:"serverVersion"`
	Config        ClusterConfig `json:"config"`
}

type ArgoClusterService interface {
	GetClusterByServer(server string) (*v1alpha1.Cluster, error)
	ClusterList() (*v1alpha1.ClusterList, error)
	CreateCluster(cluster v1alpha1.Cluster) (*v1alpha1.Cluster, error)
	CreateClusterWithGitposConfigured(request *ClusterRequest, acdToken string) (*ClusterRequest, error)
	UpdateCluster(cluster v1alpha1.Cluster) (*v1alpha1.Cluster, error)
	DeleteCluster(server string) (string, error)
}

type ArgoClusterServiceImpl struct {
	*ArgoSession
	id       int
	location string
}

func NewArgoClusterServiceImpl(session *ArgoSession) *ArgoClusterServiceImpl {
	return &ArgoClusterServiceImpl{
		ArgoSession: session,
		location:    "/api/v1/clusters?upsert=true",
	}
}

func (impl *ArgoClusterServiceImpl) GetClusterByServer(server string) (*v1alpha1.Cluster, error) {

	path := impl.location + "/" + server
	res := &v1alpha1.Cluster{}
	_, _, err := impl.DoRequest(&ClientRequest{ResponseBody: res, Path: path, Method: "GET"})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (impl *ArgoClusterServiceImpl) ClusterList() (*v1alpha1.ClusterList, error) {

	path := impl.location
	res := &v1alpha1.ClusterList{}
	_, _, err := impl.DoRequest(&ClientRequest{ResponseBody: res, Path: path, Method: "GET"})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (impl *ArgoClusterServiceImpl) CreateCluster(cluster v1alpha1.Cluster) (*v1alpha1.Cluster, error) {

	path := impl.location
	res := &v1alpha1.Cluster{}

	_, _, err := impl.DoRequest(&ClientRequest{ResponseBody: res, Path: path, RequestBody: cluster, Method: "POST"})
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (impl *ArgoClusterServiceImpl) CreateClusterWithGitposConfigured(request *ClusterRequest, acdToken string) (*ClusterRequest, error) {
	path := impl.location
	res := &ClusterRequest{}
	_, _, err := impl.DoRequestForArgo(&ClientRequest{ResponseBody: res, Path: path, RequestBody: request, Method: "POST"}, acdToken)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (impl *ArgoClusterServiceImpl) UpdateCluster(cluster v1alpha1.Cluster) (*v1alpha1.Cluster, error) {

	path := impl.location + "/" + cluster.Server
	res := &v1alpha1.Cluster{}
	_, _, err := impl.DoRequest(&ClientRequest{ResponseBody: res, Path: path, RequestBody: cluster, Method: "PUT"})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (impl *ArgoClusterServiceImpl) DeleteCluster(server string) (string, error) {
	res := ""
	path := impl.location + "/" + server
	_, _, err := impl.DoRequest(&ClientRequest{ResponseBody: &res, Path: path, Method: "DELETE"})
	if err != nil {
		return "", err
	}
	return res, nil
}

type CubeConfigCreate struct {
	Context    string `json:"context"`
	InCluster  bool   `json:"inCluster"`
	Kubeconfig string `json:"kubeconfig"`
	Upsert     bool   `json:"upsert"`
}
