package module

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/common-lib/utils"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"reflect"
	"testing"
)

func TestModuleServiceImpl_GetAllModuleInfo(t *testing.T) {
	type fields struct {
		logger                         *zap.SugaredLogger
		moduleRepository               moduleRepo.ModuleRepository
		dbConnection                   *pg.DB
		moduleResourceStatusRepository moduleRepo.ModuleResourceStatusRepository
	}
	cfg, _ := sql.GetConfig()
	logger, err := utils.NewSugardLogger()
	assert.Nil(t, err)
	dbConnection, err := sql.NewDbConnection(cfg, logger)
	assert.Nil(t, err)
	tests := []struct {
		name    string
		fields  fields
		want    []ModuleInfoDto
		wantErr bool
	}{
		{
			name: "test1_for_data_present_in_db",
			fields: fields{
				logger:                         logger,
				dbConnection:                   dbConnection,
				moduleRepository:               moduleRepo.NewModuleRepositoryImpl(dbConnection),
				moduleResourceStatusRepository: moduleRepo.NewModuleResourceStatusRepositoryImpl(dbConnection),
			},
			want:    getModuleDtoResponse1(),
			wantErr: false,
		},
		{
			name: "test2_for_empty_data_in_db",
			fields: fields{
				logger:                         logger,
				dbConnection:                   dbConnection,
				moduleRepository:               moduleRepo.NewModuleRepositoryImpl(dbConnection),
				moduleResourceStatusRepository: moduleRepo.NewModuleResourceStatusRepositoryImpl(dbConnection),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := ModuleServiceImpl{
				logger:                         tt.fields.logger,
				moduleRepository:               tt.fields.moduleRepository,
				moduleResourceStatusRepository: tt.fields.moduleResourceStatusRepository,
			}
			got, err := impl.GetAllModuleInfo()
			if (err != nil) == !tt.wantErr {
				got = nil
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetAllModuleInfo() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				d, _ := json.Marshal(got)
				fmt.Printf("%s\n", d)
				t.Errorf("GetAllModuleInfo() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func getModuleDtoResponse1() []ModuleInfoDto {
	return []ModuleInfoDto{
		{
			Name:                  "cicd",
			Status:                "installed",
			ModuleResourcesStatus: nil,
		},
		{
			Name:                  "argo-cd",
			Status:                "installed",
			ModuleResourcesStatus: nil,
		},
		{
			Name:                  "security.clair",
			Status:                "installed",
			ModuleResourcesStatus: nil,
		},
		{
			Name:                  "monitoring.grafana",
			Status:                "installed",
			ModuleResourcesStatus: nil,
		},
		{
			Name:   "notifier",
			Status: "installed",
			ModuleResourcesStatus: []*ModuleResourceStatusDto{
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Service",
					Name:          "dashboard-service",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Service",
					Name:          "devtron-service",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Service",
					Name:          "argocd-dex-server",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Service",
					Name:          "kubelink-service",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Service",
					Name:          "devtron-minio",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Service",
					Name:          "devtron-minio-svc",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Service",
					Name:          "postgresql-postgresql-metrics",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Service",
					Name:          "postgresql-postgresql-headless",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Service",
					Name:          "postgresql-postgresql",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Pod",
					Name:          "dashboard-d4854d794-d8tg7",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "ReplicaSet",
					Name:          "dashboard-d4854d794",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "Deployment",
					Name:          "dashboard",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "ReplicaSet",
					Name:          "devtron-58985997c6",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "ReplicaSet",
					Name:          "devtron-74c9f746b5",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "ReplicaSet",
					Name:          "devtron-548c489547",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Pod",
					Name:          "devtron-6477c56bff-9d895",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "ReplicaSet",
					Name:          "devtron-6477c56bff",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "Deployment",
					Name:          "devtron",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Pod",
					Name:          "argocd-dex-server-5f884659bd-psptj",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "ReplicaSet",
					Name:          "argocd-dex-server-5f884659bd",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "Deployment",
					Name:          "argocd-dex-server",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Pod",
					Name:          "inception-646dbc8ddb-762r6",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "ReplicaSet",
					Name:          "inception-646dbc8ddb",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "Deployment",
					Name:          "inception",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Pod",
					Name:          "kubelink-74dc4c4967-6n4f7",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "ReplicaSet",
					Name:          "kubelink-74dc4c4967",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "Deployment",
					Name:          "kubelink",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Pod",
					Name:          "workflow-controller-859654596b-sxjg4",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "ReplicaSet",
					Name:          "workflow-controller-859654596b",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "Deployment",
					Name:          "workflow-controller",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Pod",
					Name:          "minio-devtron-0",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "StatefulSet",
					Name:          "minio-devtron",
					HealthStatus:  "Healthy",
					HealthMessage: "statefulset rolling update complete 1 pods at revision minio-devtron-54cc7669f5...",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Pod",
					Name:          "postgresql-postgresql-0",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "apps",
					Version:       "v1",
					Kind:          "StatefulSet",
					Name:          "postgresql-postgresql",
					HealthStatus:  "Healthy",
					HealthMessage: "statefulset rolling update complete 1 pods at revision postgresql-postgresql-5d4cc658bd...",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Pod",
					Name:          "postgresql-migrate-devtron-otb8k-4bz9m",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "batch",
					Version:       "v1",
					Kind:          "Job",
					Name:          "postgresql-migrate-devtron-otb8k",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Pod",
					Name:          "postgresql-migrate-casbin-va3rv-jkj9w",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "batch",
					Version:       "v1",
					Kind:          "Job",
					Name:          "postgresql-migrate-casbin-va3rv",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Pod",
					Name:          "devtron-minio-make-bucket-job-6cvjj",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "batch",
					Version:       "v1",
					Kind:          "Job",
					Name:          "devtron-minio-make-bucket-job",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Pod",
					Name:          "app-sync-cronjob-27997620-rzkls",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "batch",
					Version:       "v1",
					Kind:          "Job",
					Name:          "app-sync-cronjob-27997620",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Pod",
					Name:          "app-sync-cronjob-27999060-krvsb",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "batch",
					Version:       "v1",
					Kind:          "Job",
					Name:          "app-sync-cronjob-27999060",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "",
					Version:       "v1",
					Kind:          "Pod",
					Name:          "app-sync-cronjob-28000500-96pbj",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
				{
					Group:         "batch",
					Version:       "v1",
					Kind:          "Job",
					Name:          "app-sync-cronjob-28000500",
					HealthStatus:  "Healthy",
					HealthMessage: "",
				},
			},
		},
	}
}
