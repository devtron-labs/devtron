// SPDX-License-Identifier: Apache-2.0

package chartRepo

import (
	"encoding/json"
	"strings"
	"testing"

	serverconfig "github.com/devtron-labs/devtron/pkg/server/config"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func manualSyncTestEnvironment(t *testing.T) {
	t.Helper()
	for key, value := range map[string]string{
		"PG_ADDR": "postgresql.example.internal", "PG_DATABASE": "orchestrator",
		"PG_USER": "sync-user", "PG_PASSWORD": "test-only-password",
		"PG_SSL_MODE": "verify-full", "PG_SSL_ROOT_CERT": "/etc/postgresql/ca.crt",
	} {
		t.Setenv(key, value)
	}
}

func decodeManualSyncTestJob(t *testing.T, resources, volumes, mounts string) batchv1.Job {
	t.Helper()
	data, err := manualAppSyncJobByteArr("example/chart-sync:test", resources, "chart-sync", &ChartProviderConfig{ChartProviderId: "17", IsOCIRegistry: true}, 3, 120, volumes, mounts)
	if err != nil {
		t.Fatal(err)
	}
	var job batchv1.Job
	if err := json.Unmarshal(data, &job); err != nil {
		t.Fatal(err)
	}
	return job
}

func TestManualAppSyncJobDefaults(t *testing.T) {
	manualSyncTestEnvironment(t)
	job := decodeManualSyncTestJob(t, "", "", "")
	if job.APIVersion != "batch/v1" || job.Kind != "Job" || job.Name != "app-manual-sync-job" || job.Namespace != "devtroncd" {
		t.Fatal("default Job identity changed")
	}
	pod := job.Spec.Template.Spec
	if pod.ServiceAccountName != "chart-sync" || pod.RestartPolicy != corev1.RestartPolicyOnFailure || len(pod.Containers) != 1 || len(pod.Volumes) != 0 {
		t.Fatal("default Pod configuration changed")
	}
	container := pod.Containers[0]
	if container.Name != "chart-sync" || container.Image != "example/chart-sync:test" || len(container.VolumeMounts) != 0 {
		t.Fatal("default container configuration changed")
	}
	if *job.Spec.BackoffLimit != 4 || *job.Spec.ActiveDeadlineSeconds != 15000 {
		t.Fatal("Job retry/deadline changed")
	}
	env := map[string]string{}
	for _, variable := range container.Env {
		env[variable.Name] = variable.Value
	}
	for key, want := range map[string]string{"PG_ADDR": "postgresql.example.internal", "PG_DATABASE": "orchestrator", "PG_USER": "sync-user", "PG_PASSWORD": "test-only-password", "PG_SSL_MODE": "verify-full", "PG_SSL_ROOT_CERT": "/etc/postgresql/ca.crt", "CHART_PROVIDER_ID": "17", "IS_OCI_REGISTRY": "true", "PARALLELISM_LIMIT_FOR_TAG_PROCESSING": "3", "APP_SYNC_SHUTDOWN_WAIT_DURATION": "120"} {
		if env[key] != want {
			t.Errorf("environment %s did not retain its configured value", key)
		}
	}
}

func TestManualAppSyncJobProjectsDatabaseCA(t *testing.T) {
	manualSyncTestEnvironment(t)
	volumes := `[{"name":"database-ca","secret":{"secretName":"postgresql-root-ca","items":[{"key":"ca.crt","path":"ca.crt"}]}}]`
	mounts := `[{"name":"database-ca","mountPath":"/etc/postgresql","readOnly":true}]`
	t.Setenv("APP_SYNC_JOB_VOLUMES", volumes)
	t.Setenv("APP_SYNC_JOB_VOLUME_MOUNTS", mounts)
	config, err := serverconfig.ParseServerEnvConfig()
	if err != nil {
		t.Fatal(err)
	}
	job := decodeManualSyncTestJob(t, `{"requests":{"cpu":"100m"},"limits":{"memory":"128Mi"}}`, config.AppSyncJobVolumes, config.AppSyncJobVolumeMounts)
	pod := job.Spec.Template.Spec
	if len(pod.Volumes) != 1 || pod.Volumes[0].Secret == nil || pod.Volumes[0].Secret.SecretName != "postgresql-root-ca" {
		t.Fatal("the CA Secret was not projected")
	}
	if got := pod.Volumes[0].Secret.Items; len(got) != 1 || got[0].Key != "ca.crt" || got[0].Path != "ca.crt" {
		t.Fatal("the CA Secret item selection changed")
	}
	container := pod.Containers[0]
	if len(container.VolumeMounts) != 1 || container.VolumeMounts[0].Name != "database-ca" || container.VolumeMounts[0].MountPath != "/etc/postgresql" || !container.VolumeMounts[0].ReadOnly {
		t.Fatal("the read-only CA mount was not preserved")
	}
	if container.Resources.Requests.Cpu().String() != "100m" || container.Resources.Limits.Memory().String() != "128Mi" {
		t.Fatal("existing resources were not preserved")
	}
}

func TestManualAppSyncJobRejectsInvalidConfiguration(t *testing.T) {
	manualSyncTestEnvironment(t)
	for _, test := range []struct{ name, resources, volumes, mounts, want string }{
		{"invalid resources", `{`, "", "", "APP_SYNC_JOB_RESOURCES_OBJ"},
		{"invalid volumes", "", `{`, "", "APP_SYNC_JOB_VOLUMES"},
		{"invalid mounts", "", "", `{`, "APP_SYNC_JOB_VOLUME_MOUNTS"},
		{"wrong volumes type", "", `{}`, "", "APP_SYNC_JOB_VOLUMES"},
		{"missing volume", "", "", `[{"name":"missing","mountPath":"/etc/postgresql"}]`, "no matching volume"},
	} {
		t.Run(test.name, func(t *testing.T) {
			data, err := manualAppSyncJobByteArr("example/chart-sync:test", test.resources, "chart-sync", &ChartProviderConfig{ChartProviderId: "17"}, 3, 120, test.volumes, test.mounts)
			if err == nil || !strings.Contains(err.Error(), test.want) || len(data) != 0 {
				t.Fatal("invalid configuration must return an error and no Job")
			}
		})
	}
}

func TestManualAppSyncJobEscapesCredentials(t *testing.T) {
	manualSyncTestEnvironment(t)
	password := "test-only-quote\"-backslash\\-newline\n"
	t.Setenv("PG_PASSWORD", password)
	job := decodeManualSyncTestJob(t, "", "[]", "[]")
	for _, variable := range job.Spec.Template.Spec.Containers[0].Env {
		if variable.Name == "PG_PASSWORD" {
			if variable.Value != password {
				t.Fatal("password did not survive the JSON round trip")
			}
			return
		}
	}
	t.Fatal("password environment variable missing")
}

func TestManualAppSyncJobRejectsInvalidDatabaseConfig(t *testing.T) {
	manualSyncTestEnvironment(t)
	t.Setenv("PG_READ_TIMEOUT", "not-a-number")
	data, err := manualAppSyncJobByteArr("example/chart-sync:test", "", "chart-sync", &ChartProviderConfig{ChartProviderId: "17"}, 3, 120, "", "")
	if err == nil || len(data) != 0 {
		t.Fatal("invalid database config must not produce a Job")
	}
}
