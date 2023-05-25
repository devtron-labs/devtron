package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type K8sResourceHistory struct {
	tableName         struct{} `sql:"kubernetes_resource_history" pg:",discard_unknown_columns"`
	Id                int      `sql:"id,pk"`
	AppId             int      `sql:"app_id"`
	AppName           string   `sql:"app_name"`
	EnvId             int      `sql:"env_id"`
	Namespace         string   `sql:"namespace,omitempty"`
	ResourceName      string   `sql:"resource_name,notnull"`
	Kind              string   `sql:"kind,notnull"`
	Group             string   `sql:"group"`
	ForceDelete       bool     `sql:"force_delete, omitempty"`
	ActionType        string   `sql:"action_type"`
	DeploymentAppType string   `sql:"deployment_app_type"`
	sql.AuditLog
}

type K8sResourceHistoryRepository interface {
	SaveK8sResourceHistory(history *K8sResourceHistory) error
}

type K8sResourceHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewK8sResourceHistoryRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *K8sResourceHistoryRepositoryImpl {
	return &K8sResourceHistoryRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (repo K8sResourceHistoryRepositoryImpl) SaveK8sResourceHistory(k8sResourceHistory *K8sResourceHistory) error {
	return repo.dbConnection.Insert(k8sResourceHistory)
}
