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

package chartRepo

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/devtron-labs/devtron/pkg/sql"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AppSyncConfig struct {
	DbConfig                         sql.Config
	DockerImage                      string
	AppSyncJobResourcesObj           string
	ChartProviderConfig              *ChartProviderConfig
	AppSyncServiceAccount            string
	ParallelismLimitForTagProcessing int
	AppSyncJobShutDownWaitDuration   int
	Volumes                          []corev1.Volume
	VolumeMounts                     []corev1.VolumeMount
}

type ChartProviderConfig struct {
	ChartProviderId string
	IsOCIRegistry   bool
}

func manualAppSyncJobByteArr(
	dockerImage, appSyncJobResourcesObj, appSyncServiceAccount string,
	chartProviderConfig *ChartProviderConfig,
	parallelismLimitForTagProcessing, appSyncJobShutDownWaitDuration int,
	volumes, volumeMounts string,
) ([]byte, error) {
	if chartProviderConfig == nil {
		return nil, fmt.Errorf("chart sync provider config is required")
	}
	cfg, err := sql.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("get chart sync database config: %w", err)
	}
	configValues := AppSyncConfig{
		DbConfig:                         *cfg,
		DockerImage:                      dockerImage,
		AppSyncJobResourcesObj:           appSyncJobResourcesObj,
		ChartProviderConfig:              chartProviderConfig,
		AppSyncServiceAccount:            appSyncServiceAccount,
		ParallelismLimitForTagProcessing: parallelismLimitForTagProcessing,
		AppSyncJobShutDownWaitDuration:   appSyncJobShutDownWaitDuration,
	}
	container := corev1.Container{
		Name: "chart-sync", Image: configValues.DockerImage,
		Ports: []corev1.ContainerPort{{ContainerPort: 8080, Name: "metrics", Protocol: corev1.ProtocolTCP}},
		Env: []corev1.EnvVar{
			{Name: "PG_ADDR", Value: configValues.DbConfig.Addr},
			{Name: "PG_DATABASE", Value: configValues.DbConfig.Database},
			{Name: "PG_USER", Value: configValues.DbConfig.User},
			{Name: "PG_PASSWORD", Value: configValues.DbConfig.Password},
			{Name: "PG_SSL_MODE", Value: configValues.DbConfig.SslMode},
			{Name: "PG_SSL_ROOT_CERT", Value: configValues.DbConfig.SslRootCert},
			{Name: "CHART_PROVIDER_ID", Value: configValues.ChartProviderConfig.ChartProviderId},
			{Name: "IS_OCI_REGISTRY", Value: strconv.FormatBool(configValues.ChartProviderConfig.IsOCIRegistry)},
			{Name: "PARALLELISM_LIMIT_FOR_TAG_PROCESSING", Value: strconv.Itoa(configValues.ParallelismLimitForTagProcessing)},
			{Name: "APP_SYNC_SHUTDOWN_WAIT_DURATION", Value: strconv.Itoa(configValues.AppSyncJobShutDownWaitDuration)},
		},
	}
	for _, field := range []struct {
		name, value string
		target      interface{}
	}{
		{"APP_SYNC_JOB_RESOURCES_OBJ", configValues.AppSyncJobResourcesObj, &container.Resources},
		{"APP_SYNC_JOB_VOLUMES", volumes, &configValues.Volumes},
		{"APP_SYNC_JOB_VOLUME_MOUNTS", volumeMounts, &configValues.VolumeMounts},
	} {
		if field.value != "" {
			if err := json.Unmarshal([]byte(field.value), field.target); err != nil {
				return nil, fmt.Errorf("parse %s: %w", field.name, err)
			}
		}
	}
	for _, mount := range configValues.VolumeMounts {
		found := false
		for _, volume := range configValues.Volumes {
			if volume.Name == mount.Name {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("chart sync volume mount %q has no matching volume", mount.Name)
		}
	}
	container.VolumeMounts = configValues.VolumeMounts
	backoffLimit, activeDeadlineSeconds := int32(4), int64(15000)
	labels := map[string]string{"app": "app-manual-sync-job", "component": "devtron"}
	job := batchv1.Job{
		TypeMeta:   metav1.TypeMeta{APIVersion: "batch/v1", Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{Name: "app-manual-sync-job", Namespace: "devtroncd", Labels: labels},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					ServiceAccountName: configValues.AppSyncServiceAccount,
					Containers:         []corev1.Container{container}, Volumes: configValues.Volumes,
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
			},
			BackoffLimit: &backoffLimit, ActiveDeadlineSeconds: &activeDeadlineSeconds,
		},
	}
	return json.Marshal(job)
}
