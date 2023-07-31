package pipeline

import (
	"encoding/json"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	v1alpha12 "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/util"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"strconv"
)

var ArgoWorkflowOwnerRef = v1.OwnerReference{APIVersion: "argoproj.io/v1alpha1", Kind: "Workflow", Name: "{{workflow.name}}", UID: "{{workflow.uid}}", BlockOwnerDeletion: &[]bool{true}[0]}

type ConfigMapSecretDto struct {
	Name     string
	Data     map[string]string
	OwnerRef v1.OwnerReference
}

func ExtractVolumesFromCmCs(configMaps []bean2.ConfigSecretMap, secrets []bean2.ConfigSecretMap) []v12.Volume {
	var volumes []v12.Volume
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

func extractVolumesFromConfigSecretMaps(isCm bool, configSecretMaps []bean2.ConfigSecretMap) []v12.Volume {
	var volumes []v12.Volume
	for _, configSecretMap := range configSecretMaps {
		if configSecretMap.Type != util.ConfigMapSecretUsageTypeVolume {
			// not volume type so ignoring
			continue
		}
		var volumeSource v12.VolumeSource
		if isCm {
			volumeSource = v12.VolumeSource{
				ConfigMap: &v12.ConfigMapVolumeSource{
					LocalObjectReference: v12.LocalObjectReference{
						Name: configSecretMap.Name,
					},
				},
			}
		} else {
			volumeSource = v12.VolumeSource{
				Secret: &v12.SecretVolumeSource{
					SecretName: configSecretMap.Name,
				},
			}
		}
		volumes = append(volumes, v12.Volume{
			Name:         configSecretMap.Name + "-vol",
			VolumeSource: volumeSource,
		})
	}
	return volumes
}

func UpdateContainerEnvsFromCmCs(workflowMainContainer *v12.Container, configMaps []bean2.ConfigSecretMap, secrets []bean2.ConfigSecretMap) {
	for _, configMap := range configMaps {
		updateContainerEnvs(true, workflowMainContainer, configMap)
	}

	for _, secret := range secrets {
		updateContainerEnvs(false, workflowMainContainer, secret)
	}
}

func updateContainerEnvs(isCM bool, workflowMainContainer *v12.Container, configSecretMap bean2.ConfigSecretMap) {
	if configSecretMap.Type == repository.VOLUME_CONFIG {
		workflowMainContainer.VolumeMounts = append(workflowMainContainer.VolumeMounts, v12.VolumeMount{
			Name:      configSecretMap.Name + "-vol",
			MountPath: configSecretMap.MountPath,
		})
	} else if configSecretMap.Type == repository.ENVIRONMENT_CONFIG {
		var envFrom v12.EnvFromSource
		if isCM {
			envFrom = v12.EnvFromSource{
				ConfigMapRef: &v12.ConfigMapEnvSource{
					LocalObjectReference: v12.LocalObjectReference{
						Name: configSecretMap.Name,
					},
				},
			}
		} else {
			envFrom = v12.EnvFromSource{
				SecretRef: &v12.SecretEnvSource{
					LocalObjectReference: v12.LocalObjectReference{
						Name: configSecretMap.Name,
					},
				},
			}
		}
		workflowMainContainer.EnvFrom = append(workflowMainContainer.EnvFrom, envFrom)
	}
}

func GetConfigMapJson(configMapSecretDto ConfigMapSecretDto) (string, error) {
	configMapBody := GetConfigMapBody(configMapSecretDto)
	configMapJson, err := json.Marshal(configMapBody)
	if err != nil {
		return "", err
	}
	return string(configMapJson), err
}

func GetConfigMapBody(configMapSecretDto ConfigMapSecretDto) v12.ConfigMap {
	return v12.ConfigMap{
		TypeMeta: v1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:            configMapSecretDto.Name,
			OwnerReferences: []v1.OwnerReference{configMapSecretDto.OwnerRef},
		},
		Data: configMapSecretDto.Data,
	}
}

func GetSecretJson(configMapSecretDto ConfigMapSecretDto) (string, error) {
	secretBody := GetSecretBody(configMapSecretDto)
	secretJson, err := json.Marshal(secretBody)
	if err != nil {
		return "", err
	}
	return string(secretJson), err
}

func GetSecretBody(configMapSecretDto ConfigMapSecretDto) v12.Secret {
	secretDataMap := make(map[string][]byte)

	// adding handling to get base64 decoded value in map value
	cmsDataMarshaled, _ := json.Marshal(configMapSecretDto.Data)
	json.Unmarshal(cmsDataMarshaled, &secretDataMap)

	return v12.Secret{
		TypeMeta: v1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:            configMapSecretDto.Name,
			OwnerReferences: []v1.OwnerReference{configMapSecretDto.OwnerRef},
		},
		Data: secretDataMap,
		Type: "Opaque",
	}
}

func GetFromGlobalCmCsDtos(globalCmCsConfigs []*bean.GlobalCMCSDto) ([]bean2.ConfigSecretMap, []bean2.ConfigSecretMap, error) {
	workflowConfigMaps := make([]bean2.ConfigSecretMap, 0, len(globalCmCsConfigs))
	workflowSecrets := make([]bean2.ConfigSecretMap, 0, len(globalCmCsConfigs))

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

func AddTemplatesForGlobalSecretsInWorkflowTemplate(globalCmCsConfigs []*bean.GlobalCMCSDto, steps *[]v1alpha1.ParallelSteps, volumes *[]v12.Volume, templates *[]v1alpha1.Template) error {

	cmIndex := 0
	csIndex := 0
	for _, config := range globalCmCsConfigs {
		if config.ConfigType == repository.CM_TYPE_CONFIG {
			cmJson, err := GetConfigMapJson(ConfigMapSecretDto{Name: config.Name, Data: config.Data, OwnerRef: ArgoWorkflowOwnerRef})
			if err != nil {
				return err
			}
			if config.Type == repository.VOLUME_CONFIG {
				*volumes = append(*volumes, v12.Volume{
					Name: config.Name + "-vol",
					VolumeSource: v12.VolumeSource{
						ConfigMap: &v12.ConfigMapVolumeSource{
							LocalObjectReference: v12.LocalObjectReference{
								Name: config.Name,
							},
						},
					},
				})
			}
			*steps = append(*steps, v1alpha1.ParallelSteps{
				Steps: []v1alpha1.WorkflowStep{
					{
						Name:     "create-env-cm-gb-" + strconv.Itoa(cmIndex),
						Template: "cm-gb-" + strconv.Itoa(cmIndex),
					},
				},
			})
			*templates = append(*templates, v1alpha1.Template{
				Name: "cm-gb-" + strconv.Itoa(cmIndex),
				Resource: &v1alpha1.ResourceTemplate{
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

			secretJson, err := GetSecretJson(ConfigMapSecretDto{Name: config.Name, Data: encodedSecretDataMap, OwnerRef: ArgoWorkflowOwnerRef})
			if err != nil {
				return err
			}
			if config.Type == repository.VOLUME_CONFIG {
				*volumes = append(*volumes, v12.Volume{
					Name: config.Name + "-vol",
					VolumeSource: v12.VolumeSource{
						Secret: &v12.SecretVolumeSource{
							SecretName: config.Name,
						},
					},
				})
			}
			*steps = append(*steps, v1alpha1.ParallelSteps{
				Steps: []v1alpha1.WorkflowStep{
					{
						Name:     "create-env-sec-gb-" + strconv.Itoa(csIndex),
						Template: "sec-gb-" + strconv.Itoa(csIndex),
					},
				},
			})
			*templates = append(*templates, v1alpha1.Template{
				Name: "sec-gb-" + strconv.Itoa(csIndex),
				Resource: &v1alpha1.ResourceTemplate{
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

func GetClientInstance(config *rest.Config, namespace string) (v1alpha12.WorkflowInterface, error) {
	clientSet, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	wfClient := clientSet.ArgoprojV1alpha1().Workflows(namespace) // create the workflow client
	return wfClient, nil
}
