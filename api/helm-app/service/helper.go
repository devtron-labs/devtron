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

package service

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/api/helm-app/service/bean"
	"net/http"
	"strconv"
	"strings"
)

func DecodeExternalAppAppId(appId string) (*bean.AppIdentifier, error) {
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
	return &bean.AppIdentifier{
		ClusterId:   clusterId,
		Namespace:   component[1],
		ReleaseName: component[2],
	}, nil
}

func HibernateReqAdaptor(hibernateRequest *openapi.HibernateRequest) *gRPC.HibernateRequest {
	req := &gRPC.HibernateRequest{}
	for _, reqObject := range hibernateRequest.GetResources() {
		obj := &gRPC.ObjectIdentifier{
			Group:     *reqObject.Group,
			Kind:      *reqObject.Kind,
			Version:   *reqObject.Version,
			Name:      *reqObject.Name,
			Namespace: *reqObject.Namespace,
		}
		req.ObjectIdentifier = append(req.ObjectIdentifier, obj)
	}
	return req
}

func HibernateResponseAdaptor(in []*gRPC.HibernateStatus) []*openapi.HibernateStatus {
	var resStatus []*openapi.HibernateStatus
	for _, status := range in {
		resObj := &openapi.HibernateStatus{
			Success:      &status.Success,
			ErrorMessage: &status.ErrorMsg,
			TargetObject: &openapi.HibernateTargetObject{
				Group:     &status.TargetObject.Group,
				Kind:      &status.TargetObject.Kind,
				Version:   &status.TargetObject.Version,
				Name:      &status.TargetObject.Name,
				Namespace: &status.TargetObject.Namespace,
			},
		}
		resStatus = append(resStatus, resObj)
	}
	return resStatus
}

func GetStatusCode(err error) int {
	if err.Error() == "unauthorized" {
		return http.StatusForbidden
	}
	return http.StatusInternalServerError
}
