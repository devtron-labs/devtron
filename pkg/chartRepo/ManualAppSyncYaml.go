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
	"bytes"
	"github.com/devtron-labs/devtron/pkg/sql"
	"text/template"
)

type AppSyncConfig struct {
	DbConfig               sql.Config
	DockerImage            string
	AppSyncJobResourcesObj string
	ChartProviderConfig    *ChartProviderConfig
	AppSyncServiceAccount  string
}

type ChartProviderConfig struct {
	ChartProviderId string
	IsOCIRegistry   bool
}

func manualAppSyncJobByteArr(dockerImage string, appSyncJobResourcesObj string, appSyncServiceAccount string, chartProviderConfig *ChartProviderConfig) []byte {
	cfg, _ := sql.GetConfig()
	configValues := AppSyncConfig{
		DbConfig:               sql.Config{Addr: cfg.Addr, Database: cfg.Database, User: cfg.User, Password: cfg.Password},
		DockerImage:            dockerImage,
		AppSyncJobResourcesObj: appSyncJobResourcesObj,
		ChartProviderConfig:    chartProviderConfig,
		AppSyncServiceAccount:  appSyncServiceAccount,
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
		"serviceAccount": "{{.AppSyncServiceAccount}}",
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
