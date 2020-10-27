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
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
)

type ClusterService interface {
	GetClusterByServer(server string) (*v1alpha1.Cluster, error)
	ClusterList() (*v1alpha1.ClusterList, error)
	CreateCluster(cluster v1alpha1.Cluster) (*v1alpha1.Cluster, error)
	UpdateCluster(cluster v1alpha1.Cluster) (*v1alpha1.Cluster, error)
	DeleteCluster(server string) (string, error)
}

type ClusterServiceImpl struct {
	*ArgoSession
	id       int
	location string
}

func NewClusterServiceImpl(session *ArgoSession) *ClusterServiceImpl {
	return &ClusterServiceImpl{
		ArgoSession: session,
		location:    "/api/v1/clusters",
	}
}

func (impl *ClusterServiceImpl) GetClusterByServer(server string) (*v1alpha1.Cluster, error) {

	path := impl.location + "/" + server
	res := &v1alpha1.Cluster{}
	_, _, err := impl.DoRequest(&ClientRequest{ResponseBody: res, Path: path, Method: "GET"})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (impl *ClusterServiceImpl) ClusterList() (*v1alpha1.ClusterList, error) {

	path := impl.location
	res := &v1alpha1.ClusterList{}
	_, _, err := impl.DoRequest(&ClientRequest{ResponseBody: res, Path: path, Method: "GET"})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (impl *ClusterServiceImpl) CreateCluster(cluster v1alpha1.Cluster) (*v1alpha1.Cluster, error) {

	path := impl.location
	res := &v1alpha1.Cluster{}

	_, _, err := impl.DoRequest(&ClientRequest{ResponseBody: res, Path: path, RequestBody: cluster, Method: "POST"})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (impl *ClusterServiceImpl) UpdateCluster(cluster v1alpha1.Cluster) (*v1alpha1.Cluster, error) {

	path := impl.location + "/" + cluster.Server
	res := &v1alpha1.Cluster{}
	_, _, err := impl.DoRequest(&ClientRequest{ResponseBody: res, Path: path, RequestBody: cluster, Method: "PUT"})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (impl *ClusterServiceImpl) DeleteCluster(server string) (string, error) {
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
