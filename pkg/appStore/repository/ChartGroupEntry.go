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

package appStoreRepository

import (
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ChartGroupEntry struct {
	TableName                    struct{} `sql:"chart_group_entry" pg:",discard_unknown_columns"`
	Id                           int      `sql:"id,pk"`
	AppStoreValuesVersionId      int      `sql:"app_store_values_version_id"`      //AppStoreVersionValuesId
	AppStoreApplicationVersionId int      `sql:"app_store_application_version_id"` //AppStoreApplicationVersionId
	ChartGroupId                 int      `sql:"chart_group_id"`
	Deleted                      bool     `sql:"deleted,notnull"`
	sql.AuditLog
	AppStoreApplicationVersion *appStoreDiscoverRepository.AppStoreApplicationVersion
	AppStoreValuesVersion      *AppStoreVersionValues
}

type ChartGroupEntriesRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewChartGroupEntriesRepositoryImpl(dbConnection *pg.DB, Logger *zap.SugaredLogger) *ChartGroupEntriesRepositoryImpl {
	return &ChartGroupEntriesRepositoryImpl{
		dbConnection: dbConnection,
		Logger:       Logger,
	}
}

type ChartGroupEntriesRepository interface {
	Save(model *ChartGroupEntry) (*ChartGroupEntry, error)
	SaveAndUpdateInTransaction(saveEntry []*ChartGroupEntry, updateEntry []*ChartGroupEntry) ([]*ChartGroupEntry, error)
	FindEntriesWithChartMetaByChartGroupId(chartGroupId []int) ([]*ChartGroupEntry, error)
	MarkChartGroupEntriesDeleted(chartGroupId []int, tx *pg.Tx) ([]*ChartGroupEntry, error)
}

func (impl *ChartGroupEntriesRepositoryImpl) Save(model *ChartGroupEntry) (*ChartGroupEntry, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl *ChartGroupEntriesRepositoryImpl) SaveAndUpdateInTransaction(saveEntry []*ChartGroupEntry, updateEntry []*ChartGroupEntry) ([]*ChartGroupEntry, error) {
	tx, err := impl.dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	if len(saveEntry) > 0 {
		err = tx.Insert(&saveEntry)
		if err != nil {
			return nil, err
		}
	}

	for _, entry := range updateEntry {
		err := tx.Update(entry)
		if err != nil {
			impl.Logger.Errorw("error in updating", "entry", entry, "err", err)
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	allEntries := append(saveEntry, saveEntry...)
	return allEntries, nil

}

func (impl *ChartGroupEntriesRepositoryImpl) FindEntriesWithChartMetaByChartGroupId(chartGroupId []int) ([]*ChartGroupEntry, error) {
	var chartGroupEntries []*ChartGroupEntry
	err := impl.dbConnection.Model(&chartGroupEntries).
		Column("chart_group_entry.*", "AppStoreApplicationVersion.AppStore.ChartRepo", "AppStoreValuesVersion.name").
		Where("chart_group_id in (?)", pg.In(chartGroupId)).
		Where("chart_group_entry.deleted = false").
		Select()
	return chartGroupEntries, err
}

func (impl *ChartGroupEntriesRepositoryImpl) MarkChartGroupEntriesDeleted(chartGroupId []int, tx *pg.Tx) ([]*ChartGroupEntry, error) {
	var chartGroupEntries []*ChartGroupEntry
	_, err := tx.Model(&chartGroupEntries).
		Where("id in (?)", pg.In(chartGroupId)).
		Set("deleted = ", true).
		Update()
	return chartGroupEntries, err
}

