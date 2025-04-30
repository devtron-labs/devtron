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

package executors

import (
	"encoding/json"
	argoWfApiV1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	argoWfClientV1 "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/devtron-labs/common-lib/utils"
	apiBean "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors/adapter"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/util"
	k8sApiV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"strconv"
)

func ExtractVolumes(configMaps []apiBean.ConfigSecretMap, secrets []apiBean.ConfigSecretMap) []k8sApiV1.Volume {
	var volumes []k8sApiV1.Volume
	configMapVolumes := extractVolumesFromCmCs(configMaps, secrets)
	volumes = append(volumes, configMapVolumes...)

	// Add downwardAPI volume
	downwardAPIVolume := createDownwardAPIVolume()
	volumes = append(volumes, downwardAPIVolume)
	return volumes
}

func extractVolumesFromCmCs(configMaps []apiBean.ConfigSecretMap, secrets []apiBean.ConfigSecretMap) []k8sApiV1.Volume {
	var volumes []k8sApiV1.Volume
	configMapVolumes := extractVolumesFromConfigSecretMaps(true, configMaps)
	secretVolumes := extractVolumesFromConfigSecretMaps(false, secrets)

	for _, volume := range configMapVolumes {
		volumes = append(volumes, volume)
	}
	for _, volume := range secretVolumes {
		volumes = append(volumes, volume)
	}

	return volumes
}

func createDownwardAPIVolume() k8sApiV1.Volume {
	return k8sApiV1.Volume{
		Name: utils.DEVTRON_SELF_DOWNWARD_API_VOLUME,
		VolumeSource: k8sApiV1.VolumeSource{
			DownwardAPI: &k8sApiV1.DownwardAPIVolumeSource{
				Items: []k8sApiV1.DownwardAPIVolumeFile{
					{
						Path: utils.POD_LABELS,
						FieldRef: &k8sApiV1.ObjectFieldSelector{
							FieldPath: "metadata." + utils.POD_LABELS,
						},
					},
					{
						Path: utils.POD_ANNOTATIONS,
						FieldRef: &k8sApiV1.ObjectFieldSelector{
							FieldPath: "metadata." + utils.POD_ANNOTATIONS,
						},
					},
				},
			},
		},
	}
}

func extractVolumesFromConfigSecretMaps(isCm bool, configSecretMaps []apiBean.ConfigSecretMap) []k8sApiV1.Volume {
	var volumes []k8sApiV1.Volume
	for _, configSecretMap := range configSecretMaps {
		if configSecretMap.Type != util.ConfigMapSecretUsageTypeVolume {
			// not volume type so ignoring
			continue
		}
		var volumeSource k8sApiV1.VolumeSource
		if isCm {
			volumeSource = k8sApiV1.VolumeSource{
				ConfigMap: &k8sApiV1.ConfigMapVolumeSource{
					LocalObjectReference: k8sApiV1.LocalObjectReference{
						Name: configSecretMap.Name,
					},
				},
			}
		} else {
			volumeSource = k8sApiV1.VolumeSource{
				Secret: &k8sApiV1.SecretVolumeSource{
					SecretName: configSecretMap.Name,
				},
			}
		}
		volumes = append(volumes, k8sApiV1.Volume{
			Name:         configSecretMap.Name + "-vol",
			VolumeSource: volumeSource,
		})
	}
	return volumes
}

func GetFromGlobalCmCsDtos(globalCmCsConfigs []*bean.GlobalCMCSDto) ([]apiBean.ConfigSecretMap, []apiBean.ConfigSecretMap, error) {
	workflowConfigMaps := make([]apiBean.ConfigSecretMap, 0, len(globalCmCsConfigs))
	workflowSecrets := make([]apiBean.ConfigSecretMap, 0, len(globalCmCsConfigs))

	for _, config := range globalCmCsConfigs {
		configSecretMap, err := config.ConvertToConfigSecretMap()
		if err != nil {
			return workflowConfigMaps, workflowSecrets, err
		}
		if config.ConfigType == repository.CM_TYPE_CONFIG {
			workflowConfigMaps = append(workflowConfigMaps, configSecretMap)
		} else {
			workflowSecrets = append(workflowSecrets, configSecretMap)
		}
	}
	return workflowConfigMaps, workflowSecrets, nil
}

func AddTemplatesForGlobalSecretsInWorkflowTemplate(globalCmCsConfigs []*bean.GlobalCMCSDto, steps *[]argoWfApiV1.ParallelSteps, volumes *[]k8sApiV1.Volume, templates *[]argoWfApiV1.Template) error {

	cmIndex := 0
	csIndex := 0
	for _, config := range globalCmCsConfigs {
		if config.ConfigType == repository.CM_TYPE_CONFIG {
			cmJson, err := adapter.GetConfigMapJson(types.ConfigMapSecretDto{Name: config.Name, Data: config.Data, OwnerRef: ArgoWorkflowOwnerRef})
			if err != nil {
				return err
			}
			if config.Type == repository.VOLUME_CONFIG {
				*volumes = append(*volumes, k8sApiV1.Volume{
					Name: config.Name + "-vol",
					VolumeSource: k8sApiV1.VolumeSource{
						ConfigMap: &k8sApiV1.ConfigMapVolumeSource{
							LocalObjectReference: k8sApiV1.LocalObjectReference{
								Name: config.Name,
							},
						},
					},
				})
			}
			*steps = append(*steps, argoWfApiV1.ParallelSteps{
				Steps: []argoWfApiV1.WorkflowStep{
					{
						Name:     "create-env-cm-gb-" + strconv.Itoa(cmIndex),
						Template: "cm-gb-" + strconv.Itoa(cmIndex),
					},
				},
			})
			*templates = append(*templates, argoWfApiV1.Template{
				Name: "cm-gb-" + strconv.Itoa(cmIndex),
				Resource: &argoWfApiV1.ResourceTemplate{
					Action:            "create",
					SetOwnerReference: true,
					Manifest:          string(cmJson),
				},
			})
			cmIndex++
		} else if config.ConfigType == repository.CS_TYPE_CONFIG {

			// special handling for secret data since GetSecretJson expects encoded values in data map
			encodedSecretData, err := bean.ConvertToEncodedForm(config.Data)
			if err != nil {
				return err
			}
			var encodedSecretDataMap = make(map[string]string)
			err = json.Unmarshal(encodedSecretData, &encodedSecretDataMap)
			if err != nil {
				return err
			}

			secretJson, err := adapter.GetSecretJson(types.ConfigMapSecretDto{Name: config.Name, Data: encodedSecretDataMap, OwnerRef: ArgoWorkflowOwnerRef})
			if err != nil {
				return err
			}
			if config.Type == repository.VOLUME_CONFIG {
				*volumes = append(*volumes, k8sApiV1.Volume{
					Name: config.Name + "-vol",
					VolumeSource: k8sApiV1.VolumeSource{
						Secret: &k8sApiV1.SecretVolumeSource{
							SecretName: config.Name,
						},
					},
				})
			}
			*steps = append(*steps, argoWfApiV1.ParallelSteps{
				Steps: []argoWfApiV1.WorkflowStep{
					{
						Name:     "create-env-sec-gb-" + strconv.Itoa(csIndex),
						Template: "sec-gb-" + strconv.Itoa(csIndex),
					},
				},
			})
			*templates = append(*templates, argoWfApiV1.Template{
				Name: "sec-gb-" + strconv.Itoa(csIndex),
				Resource: &argoWfApiV1.ResourceTemplate{
					Action:            "create",
					SetOwnerReference: true,
					Manifest:          string(secretJson),
				},
			})
			csIndex++
		}
	}

	return nil
}

func GetClientInstance(config *rest.Config, namespace string) (argoWfClientV1.WorkflowInterface, error) {
	clientSet, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	wfClient := clientSet.ArgoprojV1alpha1().Workflows(namespace) // create the workflow client
	return wfClient, nil
}

func CheckIfReTriggerRequired(status, message, workflowRunnerStatus string) bool {
	return ((status == string(argoWfApiV1.NodeError) || status == string(argoWfApiV1.NodeFailed)) &&
		message == cdWorkflow.POD_DELETED_MESSAGE) && (workflowRunnerStatus != cdWorkflow.WorkflowCancel && workflowRunnerStatus != cdWorkflow.WorkflowAborted)

}

func getWorkflowLabelsForSystemExecutor(workflowTemplate bean.WorkflowTemplate) map[string]string {
	return map[string]string{
		DEVTRON_WORKFLOW_LABEL_KEY:      DEVTRON_WORKFLOW_LABEL_VALUE,
		"devtron.ai/purpose":            "workflow",
		"workflowType":                  workflowTemplate.WorkflowType,
		bean.WorkflowGenerateNamePrefix: workflowTemplate.WorkflowNamePrefix,
	}
}
