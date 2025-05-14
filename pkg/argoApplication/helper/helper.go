/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package helper

import (
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"strconv"
	"strings"
)

func DecodeExternalArgoAppId(appId string) (*bean.ArgoAppIdentifier, error) {
	component := strings.Split(appId, "|")
	if len(component) != 3 {
		return nil, fmt.Errorf("malformed app id %s", appId)
	}
	clusterId, err := strconv.Atoi(component[0])
	if err != nil {
		return nil, err
	}
	if clusterId <= 0 {
		return nil, fmt.Errorf("target cluster is not provided")
	}
	return &bean.ArgoAppIdentifier{
		ClusterId: clusterId,
		Namespace: component[1],
		AppName:   component[2],
	}, nil
}

func ConvertClusterBeanToGrpcConfig(cluster repository.Cluster) *gRPC.ClusterConfig {
	config := &gRPC.ClusterConfig{
		ApiServerUrl:          cluster.ServerUrl,
		Token:                 cluster.Config[commonBean.BearerToken],
		ClusterId:             int32(cluster.Id),
		ClusterName:           cluster.ClusterName,
		InsecureSkipTLSVerify: cluster.InsecureSkipTlsVerify,
	}
	if cluster.InsecureSkipTlsVerify == false {
		config.KeyData = cluster.Config[commonBean.TlsKey]
		config.CertData = cluster.Config[commonBean.CertData]
		config.CaData = cluster.Config[commonBean.CertificateAuthorityData]
	}
	return config

}

func GetHealthSyncStatusDestinationServerAndManagedResourcesForArgoK8sRawObject(obj map[string]interface{}) (string,
	string, string, []*bean.ArgoManagedResource) {
	var healthStatus, syncStatus, destinationServer string
	argoManagedResources := make([]*bean.ArgoManagedResource, 0)
	if specObjRaw, ok := obj[commonBean.Spec]; ok {
		specObj := specObjRaw.(map[string]interface{})
		if destinationObjRaw, ok2 := specObj[bean.Destination]; ok2 {
			destinationObj := destinationObjRaw.(map[string]interface{})
			if destinationServerIf, ok3 := destinationObj[bean.Server]; ok3 {
				destinationServer = destinationServerIf.(string)
			}
		}
	}
	if statusObjRaw, ok := obj[commonBean.K8sClusterResourceStatusKey]; ok {
		statusObj := statusObjRaw.(map[string]interface{})
		if healthObjRaw, ok2 := statusObj[commonBean.K8sClusterResourceHealthKey]; ok2 {
			healthObj := healthObjRaw.(map[string]interface{})
			if healthStatusIf, ok3 := healthObj[commonBean.K8sClusterResourceStatusKey]; ok3 {
				healthStatus = healthStatusIf.(string)
			}
		}
		if syncObjRaw, ok2 := statusObj[commonBean.K8sClusterResourceSyncKey]; ok2 {
			syncObj := syncObjRaw.(map[string]interface{})
			if syncStatusIf, ok3 := syncObj[commonBean.K8sClusterResourceStatusKey]; ok3 {
				syncStatus = syncStatusIf.(string)
			}
		}
		if resourceObjsRaw, ok2 := statusObj[commonBean.K8sClusterResourceResourcesKey]; ok2 {
			resourceObjs := resourceObjsRaw.([]interface{})
			argoManagedResources = make([]*bean.ArgoManagedResource, 0, len(resourceObjs))
			for _, resourceObjRaw := range resourceObjs {
				argoManagedResource := &bean.ArgoManagedResource{}
				resourceObj := resourceObjRaw.(map[string]interface{})
				if groupRaw, ok := resourceObj[commonBean.K8sClusterResourceGroupKey]; ok {
					argoManagedResource.Group = groupRaw.(string)
				}
				if kindRaw, ok := resourceObj[commonBean.K8sClusterResourceKindKey]; ok {
					argoManagedResource.Kind = kindRaw.(string)
				}
				if versionRaw, ok := resourceObj[commonBean.K8sClusterResourceVersionKey]; ok {
					argoManagedResource.Version = versionRaw.(string)
				}
				if nameRaw, ok := resourceObj[commonBean.K8sClusterResourceMetadataNameKey]; ok {
					argoManagedResource.Name = nameRaw.(string)
				}
				if namespaceRaw, ok := resourceObj[commonBean.K8sClusterResourceNamespaceKey]; ok {
					argoManagedResource.Namespace = namespaceRaw.(string)
				}
				argoManagedResources = append(argoManagedResources, argoManagedResource)
			}
		}
	}
	return healthStatus, syncStatus, destinationServer, argoManagedResources
}
