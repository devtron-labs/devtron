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

package imageScanning

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/adapter"
	bean3 "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/helper/parser"
)

func (impl ImageScanServiceImpl) GetScanResults(resourceScanQueryParams *bean3.ResourceScanQueryParams) (resp parser.ResourceScanResponseDto, err error) {
	request := &bean3.ImageScanRequest{
		ArtifactId: resourceScanQueryParams.ArtifactId,
		AppId:      resourceScanQueryParams.AppId,
		EnvId:      resourceScanQueryParams.EnvId,
	}
	respFromExecutionDetail, err := impl.FetchExecutionDetailResult(request)
	if err != nil {
		impl.Logger.Errorw("error encountered in GetScanResults", "req", request, "err", err)
		return resp, err
	}
	// build an adapter to convert the respFromExecutionDetail to the required ResourceScanResponseDto format
	return adapter.ExecutionDetailsToResourceScanResponseDto(respFromExecutionDetail), nil

}
