/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package adapter

import (
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/devtron-labs/common-lib/utils/k8sObjectsUtil"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func GetArgoPodMetadata(podMetadatas []*gRPC.PodMetadata) []*argoApplication.PodMetadata {
	if len(podMetadatas) == 0 {
		return []*argoApplication.PodMetadata{}
	}
	resp := make([]*argoApplication.PodMetadata, 0, len(podMetadatas))
	for _, podMetadata := range podMetadatas {
		argoPodMetadata := &argoApplication.PodMetadata{
			Name:                podMetadata.Name,
			UID:                 podMetadata.Uid,
			Containers:          GetArgoContainers(podMetadata.Containers),
			InitContainers:      GetArgoContainers(podMetadata.InitContainers),
			IsNew:               podMetadata.IsNew,
			EphemeralContainers: GetArgoEphemeralContainers(podMetadata.EphemeralContainers),
		}
		resp = append(resp, argoPodMetadata)
	}
	return resp
}

func GetArgoEphemeralContainers(ephemeralContainers []*gRPC.EphemeralContainerData) []*k8sObjectsUtil.EphemeralContainerData {
	if len(ephemeralContainers) == 0 {
		return []*k8sObjectsUtil.EphemeralContainerData{}
	}
	resp := make([]*k8sObjectsUtil.EphemeralContainerData, 0, len(ephemeralContainers))
	for _, ephemeralContainer := range ephemeralContainers {
		argoEphemeralContainerData := &k8sObjectsUtil.EphemeralContainerData{
			Name:       ephemeralContainer.GetName(),
			IsExternal: ephemeralContainer.GetIsExternal(),
		}
		resp = append(resp, argoEphemeralContainerData)
	}
	return resp
}

func GetArgoContainers(containers []string) []*string {
	if len(containers) == 0 {
		return []*string{}
	}
	resp := make([]*string, 0, len(containers))
	for _, container := range containers {
		argoContainer := &container
		resp = append(resp, argoContainer)
	}
	return resp
}

func GetArgoApplicationTreeForNodes(nodes []*gRPC.ResourceNode) (*v1alpha1.ApplicationTree, error) {
	if len(nodes) == 0 {
		return &v1alpha1.ApplicationTree{}, nil
	}
	argoNodes := make([]v1alpha1.ResourceNode, 0, len(nodes))
	for _, node := range nodes {
		argoResourceNode := v1alpha1.ResourceNode{
			ResourceRef: v1alpha1.ResourceRef{
				Group:     node.Group,
				Version:   node.Version,
				Kind:      node.Kind,
				Namespace: node.Namespace,
				Name:      node.Name,
				UID:       node.Uid,
			},
			ParentRefs:      GetArgoParentRefs(node.GetParentRefs()),
			Info:            GetArgoInfoItems(node.GetInfo()),
			NetworkingInfo:  GetArgoNetworkingInfo(node.GetNetworkingInfo()),
			ResourceVersion: node.ResourceVersion,
			Images:          nil, //TODO: do we use this?? and set this to null ??
			Health:          GetArgoHealthStatus(node.Health),
		}
		var (
			createdAtTime time.Time
			err           error
		)
		if !createdAtTime.IsZero() {
			createdAtTime, err = time.Parse(time.RFC3339, node.CreatedAt)
			if err != nil {
				return nil, err
			}
		}
		argoResourceNode.CreatedAt = &metav1.Time{createdAtTime}
		argoNodes = append(argoNodes, argoResourceNode)
	}
	return &v1alpha1.ApplicationTree{
		Nodes: argoNodes,
	}, nil
}

func GetArgoHealthStatus(status *gRPC.HealthStatus) *v1alpha1.HealthStatus {
	if status == nil {
		return nil
	}
	return &v1alpha1.HealthStatus{
		Status:  health.HealthStatusCode(status.GetStatus()),
		Message: status.GetMessage(),
	}

}

func GetArgoNetworkingInfo(info *gRPC.ResourceNetworkingInfo) *v1alpha1.ResourceNetworkingInfo {
	if info == nil {
		return &v1alpha1.ResourceNetworkingInfo{}
	}
	return &v1alpha1.ResourceNetworkingInfo{
		Labels: info.GetLabels(),
	}

}

func GetArgoInfoItems(infoItems []*gRPC.InfoItem) []v1alpha1.InfoItem {
	if len(infoItems) == 0 {
		return []v1alpha1.InfoItem{}
	}
	resp := make([]v1alpha1.InfoItem, 0, len(infoItems))
	for _, infoItem := range infoItems {
		argoInfoItem := v1alpha1.InfoItem{
			Name:  infoItem.Name,
			Value: infoItem.Value,
		}
		resp = append(resp, argoInfoItem)

	}
	return resp

}

func GetArgoParentRefs(parentRefs []*gRPC.ResourceRef) []v1alpha1.ResourceRef {
	if len(parentRefs) == 0 {
		return []v1alpha1.ResourceRef{}
	}

	resp := make([]v1alpha1.ResourceRef, 0, len(parentRefs))

	for _, parentRef := range parentRefs {
		ref := v1alpha1.ResourceRef{
			Group:     parentRef.Group,
			Version:   parentRef.Version,
			Kind:      parentRef.Kind,
			Namespace: parentRef.Namespace,
			Name:      parentRef.Name,
			UID:       parentRef.Uid,
		}
		resp = append(resp, ref)
	}
	return resp
}
