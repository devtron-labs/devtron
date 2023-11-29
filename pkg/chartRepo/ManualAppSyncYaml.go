package chartRepo

import (
	"bytes"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sql"
	"text/template"
)

type AppSyncConfig struct {
	DbConfig               sql.Config
	DockerImage            string
	AppSyncJobResourcesObj string
	ChartProviderConfig    *ChartProviderConfig
	JobName                string
}

type ChartProviderConfig struct {
	ChartProviderId string
	IsOCIRegistry   bool
}

const (
	MANUAL_APP_SYNC_JOB_PREFIX            = "app-manual-sync-job"
	MANUAL_APP_SYNC_JOB_OCI_PREFIX        = "oci-registry"
	MANUAL_APP_SYNC_JOB_CHART_REPO_PREFIX = "chart-repo"
)

func manualAppSyncJobByteArr(dockerImage string, appSyncJobResourcesObj string, chartProviderConfig *ChartProviderConfig) []byte {
	cfg, _ := sql.GetConfig()
	configValues := AppSyncConfig{
		DbConfig:               sql.Config{Addr: cfg.Addr, Database: cfg.Database, User: cfg.User, Password: cfg.Password},
		DockerImage:            dockerImage,
		AppSyncJobResourcesObj: appSyncJobResourcesObj,
		ChartProviderConfig:    chartProviderConfig,
		JobName:                GetUniqueIdentifierForManualAppSyncJob(chartProviderConfig),
	}
	temp := template.New("manualAppSyncJobByteArr")
	temp, _ = temp.Parse(`{"apiVersion": "batch/v1",
  "kind": "Job",
  "metadata": {
    "name": "{{.JobName}}",
    "namespace": "devtroncd"
  },
  "spec": {
    "template": {
      "spec": {
        "containers": [
          {
            "name": "chart-sync",
            "image": "{{.DockerImage}}",
			{{if .AppSyncJobResourcesObj}}
			"resources": {{.AppSyncJobResourcesObj}},
            {{end}}
            "env": [
              {
                "name": "PG_ADDR",
                "value": "postgresql-postgresql.devtroncd"
              },
              {
                "name": "PG_DATABASE",
                "value": "{{.DbConfig.Database}}"
              },
              {
                "name": "PG_USER",
                "value": "{{.DbConfig.User}}"
              },
              {
                "name": "PG_PASSWORD",
                "value": "{{.DbConfig.Password}}"
              },
			  {
                "name": "CHART_PROVIDER_ID",
                "value": "{{.ChartProviderConfig.ChartProviderId}}"
			  },
			  {
                "name": "IS_OCI_REGISTRY",
                "value": "{{.ChartProviderConfig.IsOCIRegistry}}"
			  }
            ]
          }
        ],
        "restartPolicy": "OnFailure"
      }
    },
    "backoffLimit": 4,
    "activeDeadlineSeconds": 15000
  }
}`)

	var manualAppSyncJobBufferBytes bytes.Buffer
	if err := temp.Execute(&manualAppSyncJobBufferBytes, configValues); err != nil {
		return nil
	}
	manualAppSyncJobByteArr := []byte(manualAppSyncJobBufferBytes.String())
	return manualAppSyncJobByteArr
}

func GetUniqueIdentifierForManualAppSyncJob(ChartProviderConfig *ChartProviderConfig) string {
	var uniqueJobIdentifier string
	if ChartProviderConfig.IsOCIRegistry {
		uniqueJobIdentifier = fmt.Sprintf("%s-%s", MANUAL_APP_SYNC_JOB_OCI_PREFIX, ChartProviderConfig.ChartProviderId)
	} else {
		uniqueJobIdentifier = fmt.Sprintf("%s-%s", MANUAL_APP_SYNC_JOB_CHART_REPO_PREFIX, ChartProviderConfig.ChartProviderId)

	}
	return fmt.Sprintf("%s-%s", MANUAL_APP_SYNC_JOB_PREFIX, uniqueJobIdentifier)
}
