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

package bean

import (
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"github.com/devtron-labs/devtron/api/bean"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type WorkflowTemplate struct {
	WorkflowId       int
	WorkflowRunnerId int
	v1.PodSpec
	ConfigMaps             []bean.ConfigSecretMap
	Secrets                []bean.ConfigSecretMap
	TTLValue               *int32
	WorkflowRequestJson    string
	WorkflowNamePrefix     string
	WfControllerInstanceID string
	ClusterConfig          *rest.Config
	Namespace              string
	ArchiveLogs            bool
	BlobStorageConfigured  bool
	BlobStorageS3Config    *blob_storage.BlobStorageS3Config
	CloudProvider          blob_storage.BlobStorageType
	AzureBlobConfig        *blob_storage.AzureBlobConfig
	GcpBlobConfig          *blob_storage.GcpBlobConfig
	CloudStorageKey        string
	PrePostDeploySteps     []*StepObject
	RefPlugins             []*RefPluginObject
	TerminationGracePeriod int
	WorkflowType           string
}

const (
	CI_WORKFLOW_NAME        = "ci"
	CI_WORKFLOW_WITH_STAGES = "ci-stages-with-env"
	CiStage                 = "CI"
	JobStage                = "JOB"
	CdStage                 = "CD"
	CD_WORKFLOW_NAME        = "cd"
	CD_WORKFLOW_WITH_STAGES = "cd-stages-with-env"
)

func (workflowTemplate *WorkflowTemplate) GetEntrypoint() string {
	switch workflowTemplate.WorkflowType {
	case CI_WORKFLOW_NAME:
		return CI_WORKFLOW_WITH_STAGES
	case CD_WORKFLOW_NAME:
		return CD_WORKFLOW_WITH_STAGES
	default:
		return ""
	}
}

func (workflowTemplate *WorkflowTemplate) CreateObjectMetadata() *v12.ObjectMeta {

	switch workflowTemplate.WorkflowType {
	case CI_WORKFLOW_NAME:
		return &v12.ObjectMeta{
			GenerateName: workflowTemplate.WorkflowNamePrefix + "-",
			Labels:       map[string]string{"devtron.ai/workflow-purpose": "ci"},
		}
	case CD_WORKFLOW_NAME:
		return &v12.ObjectMeta{
			GenerateName: workflowTemplate.WorkflowNamePrefix + "-",
			Annotations:  map[string]string{"workflows.argoproj.io/controller-instanceid": workflowTemplate.WfControllerInstanceID},
			Labels:       map[string]string{"devtron.ai/workflow-purpose": "cd"},
		}
	default:
		return nil
	}
}
