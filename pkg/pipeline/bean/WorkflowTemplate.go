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
	ConfigMaps               []bean.ConfigSecretMap
	Secrets                  []bean.ConfigSecretMap
	TTLValue                 *int32
	WorkflowRequestJson      string
	WorkflowNamePrefix       string
	WfControllerInstanceID   string
	ClusterConfig            *rest.Config
	Namespace                string
	ArchiveLogs              bool
	BlobStorageConfigured    bool
	BlobStorageS3Config      *blob_storage.BlobStorageS3Config
	CloudProvider            blob_storage.BlobStorageType
	AzureBlobConfig          *blob_storage.AzureBlobConfig
	GcpBlobConfig            *blob_storage.GcpBlobConfig
	CloudStorageKey          string
	PrePostDeploySteps       []*StepObject
	RefPlugins               []*RefPluginObject
	TerminationGracePeriod   int
	WorkflowType             string
	PodGCDeleteDelayDuration string
}

const (
	CI_WORKFLOW_NAME           = "ci"
	CI_WORKFLOW_WITH_STAGES    = "ci-stages-with-env"
	CiStage                    = "CI"
	JobStage                   = "JOB"
	CdStage                    = "CD"
	CD_WORKFLOW_NAME           = "cd"
	CD_WORKFLOW_WITH_STAGES    = "cd-stages-with-env"
	WorkflowGenerateNamePrefix = "devtron.ai/generate-name-prefix"
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

func (workflowTemplate *WorkflowTemplate) SetActiveDeadlineSeconds(timeout int64) {
	workflowTemplate.ActiveDeadlineSeconds = &timeout
}

func (workflowTemplate *WorkflowTemplate) CreateObjectMetadata() *v12.ObjectMeta {

	workflowLabels := map[string]string{WorkflowGenerateNamePrefix: workflowTemplate.WorkflowNamePrefix}
	switch workflowTemplate.WorkflowType {
	case CI_WORKFLOW_NAME:
		workflowLabels["devtron.ai/workflow-purpose"] = "ci"
		return &v12.ObjectMeta{
			GenerateName: workflowTemplate.WorkflowNamePrefix + "-",
			Labels:       workflowLabels,
		}
	case CD_WORKFLOW_NAME:
		workflowLabels["devtron.ai/workflow-purpose"] = "cd"
		return &v12.ObjectMeta{
			GenerateName: workflowTemplate.WorkflowNamePrefix + "-",
			Annotations:  map[string]string{"workflows.argoproj.io/controller-instanceid": workflowTemplate.WfControllerInstanceID},
			Labels:       workflowLabels,
		}
	default:
		return nil
	}
}
