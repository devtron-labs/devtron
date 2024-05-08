package chartRepo

import (
	"bytes"
	"github.com/devtron-labs/devtron/pkg/sql"
	"text/template"
)

type AppSyncConfig struct {
	DbConfig               sql.Config
	DockerImage            string
	AppSyncJobResourcesObj string
	ChartProviderConfig    *ChartProviderConfig
}

type ChartProviderConfig struct {
	ChartProviderId string
	IsOCIRegistry   bool
}

func manualAppSyncJobByteArr(dockerImage string, appSyncJobResourcesObj string, chartProviderConfig *ChartProviderConfig) []byte {
	cfg, _ := sql.GetConfig()
	configValues := AppSyncConfig{
		DbConfig:               sql.Config{Addr: cfg.Addr, Database: cfg.Database, User: cfg.User, Password: cfg.Password},
		DockerImage:            dockerImage,
		AppSyncJobResourcesObj: appSyncJobResourcesObj,
		ChartProviderConfig:    chartProviderConfig,
	}
	temp := template.New("manualAppSyncJobByteArr")
	temp, _ = temp.Parse(`{"apiVersion": "batch/v1",
  "kind": "Job",
  "metadata": {
    "name": "app-manual-sync-job",
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
                "value": "{{.DbConfig.Addr}}"
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
