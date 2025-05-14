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

package adapter

import (
	"encoding/json"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	k8sApiV1 "k8s.io/api/core/v1"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetConfigMapJson(configMapSecretDto types.ConfigMapSecretDto) (string, error) {
	configMapBody := GetConfigMapBody(configMapSecretDto)
	configMapJson, err := json.Marshal(configMapBody)
	if err != nil {
		return "", err
	}
	return string(configMapJson), err
}

func GetSecretJson(configMapSecretDto types.ConfigMapSecretDto) (string, error) {
	secretBody, err := GetSecretBody(configMapSecretDto)
	if err != nil {
		return "", err
	}
	secretJson, err := json.Marshal(secretBody)
	if err != nil {
		return "", err
	}
	return string(secretJson), err
}

func GetConfigMapSecretDto(configSecretMap bean.ConfigSecretMap, ownerRef k8sMetaV1.OwnerReference, isSecret bool) (types.ConfigMapSecretDto, error) {
	configDataMap, err := configSecretMap.GetDataMap()
	if err != nil {
		return types.ConfigMapSecretDto{}, err
	}
	configMapSecretDto := types.ConfigMapSecretDto{
		Name:     configSecretMap.Name,
		Data:     configDataMap,
		OwnerRef: ownerRef,
	}
	return updateBinaryDataInConfigMapSecretDto(configSecretMap, configMapSecretDto, isSecret), nil
}

func GetConfigMapBody(configMapSecretDto types.ConfigMapSecretDto) k8sApiV1.ConfigMap {
	configMap := k8sApiV1.ConfigMap{
		TypeMeta: k8sMetaV1.TypeMeta{
			Kind:       commonBean.ConfigMapKind,
			APIVersion: commonBean.V1VERSION,
		},
		ObjectMeta: k8sMetaV1.ObjectMeta{
			Name:            configMapSecretDto.Name,
			OwnerReferences: []k8sMetaV1.OwnerReference{configMapSecretDto.OwnerRef},
		},
		Data: configMapSecretDto.Data,
	}
	return getConfigMapBodyEnt(configMapSecretDto, configMap)
}

func GetSecretBody(configMapSecretDto types.ConfigMapSecretDto) (k8sApiV1.Secret, error) {
	secretDataMap := make(map[string][]byte)
	// adding handling to get base64 decoded value in map value
	cmsDataMarshaled, err := json.Marshal(configMapSecretDto.Data)
	if err != nil {
		return k8sApiV1.Secret{}, err
	}
	err = json.Unmarshal(cmsDataMarshaled, &secretDataMap)
	if err != nil {
		return k8sApiV1.Secret{}, err
	}
	secret := k8sApiV1.Secret{
		TypeMeta: k8sMetaV1.TypeMeta{
			Kind:       commonBean.SecretKind,
			APIVersion: commonBean.V1VERSION,
		},
		ObjectMeta: k8sMetaV1.ObjectMeta{
			Name:            configMapSecretDto.Name,
			OwnerReferences: []k8sMetaV1.OwnerReference{configMapSecretDto.OwnerRef},
		},
		Data: secretDataMap,
		Type: k8sApiV1.SecretTypeOpaque,
	}
	return secret, nil
}
