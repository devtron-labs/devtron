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

package adaptor

import (
	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	bean2 "github.com/devtron-labs/devtron/pkg/workflow/dag/bean"
	"time"
)

func GetBuildArtifact(request *bean2.CiArtifactWebhookRequest, ciPipelineId int, materialJson []byte, createdOn, updatedOn time.Time) *repository.CiArtifact {
	return &repository.CiArtifact{
		Image:              request.Image,
		ImageDigest:        request.ImageDigest,
		MaterialInfo:       string(materialJson),
		DataSource:         request.DataSource,
		PipelineId:         ciPipelineId,
		WorkflowId:         request.WorkflowId,
		ScanEnabled:        request.IsScanEnabled,
		IsArtifactUploaded: request.IsArtifactUploaded, // for backward compatibility
		Scanned:            false,
		TargetPlatforms:    utils.ConvertTargetPlatformListToString(request.TargetPlatforms),
		AuditLog:           sql.AuditLog{CreatedBy: request.UserId, UpdatedBy: request.UserId, CreatedOn: createdOn, UpdatedOn: updatedOn},
	}
}
