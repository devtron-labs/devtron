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

package imageDigestPolicy

import (
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

func QualifierMappingDao(qualifierId, identifierKey, IdentifierValueInt int, identifierValueName string, userId int32) *resourceQualifiers.QualifierMapping {
	return &resourceQualifiers.QualifierMapping{
		ResourceId:            resourceQualifiers.ImageDigestResourceId,
		ResourceType:          resourceQualifiers.ImageDigest,
		QualifierId:           qualifierId,
		IdentifierKey:         identifierKey,
		IdentifierValueInt:    IdentifierValueInt,
		IdentifierValueString: identifierValueName,
		Active:                true,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
}

type DigestPolicyConfigurationRequest struct {
	PipelineId    int
	ClusterId     int
	EnvironmentId int
}

func (request DigestPolicyConfigurationRequest) getQualifierMappingScope() *resourceQualifiers.Scope {
	return &resourceQualifiers.Scope{
		EnvId:      request.EnvironmentId,
		ClusterId:  request.ClusterId,
		PipelineId: request.PipelineId,
	}
}

type DigestPolicyConfigurationResponse struct {
	DigestConfiguredForPipeline     bool
	DigestConfiguredForEnvOrCluster bool
}

func (config DigestPolicyConfigurationResponse) UseDigestForTrigger() bool {
	return config.DigestConfiguredForPipeline
}
