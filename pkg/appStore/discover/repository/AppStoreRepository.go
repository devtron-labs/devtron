/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package appStoreDiscoverRepository

import (
	dockerArtifactStoreRegistry "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type AppStoreRepository interface {
	GetAppStoreApplications() ([]*AppStore, error)
	Delete(appStores []*AppStore) error
}

type AppStoreRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewAppStoreRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *AppStoreRepositoryImpl {
	return &AppStoreRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type AppStore struct {
	TableName             struct{}  `sql:"app_store" pg:",discard_unknown_columns"`
	Id                    int       `sql:"id,pk"`
	Name                  string    `sql:"name,notnull"`
	ChartRepoId           int       `sql:"chart_repo_id"`
	Active                bool      `sql:"active,notnull"`
	DockerArtifactStoreId string    `sql:"docker_artifact_store_id"`
	ChartGitLocation      string    `sql:"chart_git_location"`
	CreatedOn             time.Time `sql:"created_on,notnull"`
	UpdatedOn             time.Time `sql:"updated_on,notnull"`
	ChartRepo             *chartRepoRepository.ChartRepo
	DockerArtifactStore   *dockerArtifactStoreRegistry.DockerArtifactStore
}

func (impl AppStoreRepositoryImpl) GetAppStoreApplications() ([]*AppStore, error) {
	var models []*AppStore
	err := impl.dbConnection.Model(&models).Where("active = ? ", true).Select()
	if err != nil {
		return models, err
	}
	return models, nil
}

func (impl *AppStoreRepositoryImpl) Delete(appStores []*AppStore) error {
	err := impl.dbConnection.RunInTransaction(func(tx *pg.Tx) error {
		for _, appStore := range appStores {
			//appStoreApplicationVersionDeleteQuery := "delete from app_store_application_version where app_store_id = ?"
			//_, err := impl.dbConnection.Exec(appStoreApplicationVersionDeleteQuery, appStore.Id)
			//if err != nil {
			//	impl.Logger.Errorw("error in deleting app store application version by app store id", "err", err)
			//	return err
			//}
			appStoreDeleteQuery := "delete from app_store where id = ?"
			appStoreDeleteQuery = "update app_store set active=false where id = ?"
			_, err := impl.dbConnection.Exec(appStoreDeleteQuery, appStore.Id)
			if err != nil {
				impl.Logger.Errorw("error in deleting app store ", "err", err, "app_store_id", appStore.Id)
				return err
			}
		}
		return nil
	})
	return err
}
