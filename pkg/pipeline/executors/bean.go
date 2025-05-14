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

package executors

import (
	k8sApiV1 "k8s.io/api/core/v1"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	STEP_NAME_REGEX              = "create-env-%s-gb-%d"
	TEMPLATE_NAME_REGEX          = "%s-gb-%d"
	WORKFLOW_MINIO_CRED          = "workflow-minio-cred"
	CRED_ACCESS_KEY              = "accessKey"
	CRED_SECRET_KEY              = "secretKey"
	S3_ENDPOINT_URL              = "s3.amazonaws.com"
	DEVTRON_WORKFLOW_LABEL_KEY   = "devtron.ai/workflow-purpose"
	DEVTRON_WORKFLOW_LABEL_VALUE = "cd"
	WORKFLOW_GENERATE_NAME_REGEX = "%s-"
	RESOURCE_CREATE_ACTION       = "create"
)

var ArgoWorkflowOwnerRef = k8sMetaV1.OwnerReference{APIVersion: "argoproj.io/v1alpha1", Kind: "Workflow", Name: "{{workflow.name}}", UID: "{{workflow.uid}}", BlockOwnerDeletion: &[]bool{true}[0]}

var AccessKeySelector = &k8sApiV1.SecretKeySelector{Key: CRED_ACCESS_KEY, LocalObjectReference: k8sApiV1.LocalObjectReference{Name: WORKFLOW_MINIO_CRED}}

var SecretKeySelector = &k8sApiV1.SecretKeySelector{Key: CRED_SECRET_KEY, LocalObjectReference: k8sApiV1.LocalObjectReference{Name: WORKFLOW_MINIO_CRED}}

const (
	WorkflowJobBackoffLimit = 0
	WorkflowJobFinalizer    = "foregroundDeletion"
)
