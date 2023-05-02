package pipeline

import (
	"encoding/json"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

var ArgoWorkflowOwnerRef = v1.OwnerReference{APIVersion: "argoproj.io/v1alpha1", Kind: "Workflow", Name: "{{workflow.name}}", UID: "{{workflow.uid}}", BlockOwnerDeletion: &[]bool{true}[0]}

type ConfigMapSecretDto struct {
	Name     string
	Data     map[string]string
	OwnerRef v1.OwnerReference
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
	for key, value := range configMapSecretDto.Data {
		secretDataMap[key] = []byte(value)
	}
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
			secretJson, err := GetSecretJson(ConfigMapSecretDto{Name: config.Name, Data: config.Data, OwnerRef: ArgoWorkflowOwnerRef})
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
