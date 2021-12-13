package appstore

import (
	"bytes"
	"github.com/devtron-labs/devtron/pkg/sql"
	"text/template"
)

func manualAppSyncJobByteArr() []byte {
	cfg, _ := sql.GetConfig()
	configValues := sql.Config{Addr: cfg.Addr, Database: cfg.Database, User: cfg.User, Password: cfg.Password}

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
            "image": "quay.io/devtron/chart-sync:1227622d-132-3775",
            "env": [
              {
                "name": "PG_ADDR",
                "value": "{{.Addr}}"
              },
              {
                "name": "PG_DATABASE",
                "value": "{{.Database}}"
              },
              {
                "name": "PG_USER",
                "value": "{{.User}}"
              },
              {
                "name": "PG_PASSWORD",
                "value": "{{.Password}}"
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
