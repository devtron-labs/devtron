package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type GitMaterialHistory struct {
	tableName       struct{} `sql:"git_material_history" pg:",discard_unknown_columns"`
	Id              int      `sql:"id,pk"`
	GitMaterialId   int      `sql:"git_material_id"`
	AppId           int      `sql:"app_id,notnull"`
	GitProviderId   int      `sql:"git_provider_id,notnull"`
	Active          bool     `sql:"active,notnull"`
	Url             string   `sql:"url,omitempty"`
	Name            string   `sql:"name, omitempty"`
	CheckoutPath    string   `sql:"checkout_path, omitempty"`
	FetchSubmodules bool     `sql:"fetch_submodules,notnull"`
	FilterPattern   []string `sql:"filter_pattern"`
	sql.AuditLog
}

type GitMaterialHistoryRepository interface {
	SaveGitMaterialHistory(material *GitMaterialHistory) error
	SaveDeleteMaterialHistory(materials []*GitMaterialHistory) error
}

type GitMaterialHistoryRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewGitMaterialHistoryRepositoyImpl(dbConnection *pg.DB) *GitMaterialHistoryRepositoryImpl {
	return &GitMaterialHistoryRepositoryImpl{
		dbConnection: dbConnection,
	}
}

func (repo GitMaterialHistoryRepositoryImpl) SaveGitMaterialHistory(material *GitMaterialHistory) error {
	return repo.dbConnection.Insert(material)
}

func (repo GitMaterialHistoryRepositoryImpl) SaveDeleteMaterialHistory(materials []*GitMaterialHistory) error {

	err := repo.dbConnection.RunInTransaction(func(tx *pg.Tx) error {
		for _, material := range materials {
			_, err := tx.Model(material).WherePK().Insert()
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
