/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package models

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type HelmValues struct {
	tableName         struct{}  `sql:"helm_values"`
	AppName           string    `sql:"app_name,pk"`
	TargetEnvironment string    `sql:"environment,pk"` //target environment
	Values            string    `sql:"values_yaml"`
	Active            bool      `sql:"active,notnull"`
	CreatedOn         time.Time `sql:"created_on"`
	CreatedBy         int32     `sql:"created_by"`
	UpdatedOn         time.Time `sql:"updated_on"`
	UpdatedBy         int32     `sql:"updated_by"`
}

type HelmValuesService interface {
	AddHelmValues(manifest *HelmValues) error
	GetHelmValues(appName, targetEnvironment string) (*HelmValues, error)
}

type HelmValuesServiceImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewHelmValuesServiceImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *HelmValuesServiceImpl {
	return &HelmValuesServiceImpl{dbConnection: dbConnection, Logger: Logger}
}

func (impl HelmValuesServiceImpl) AddHelmValues(manifest *HelmValues) error {
	err := impl.dbConnection.Insert(manifest)
	if err != nil {
		impl.Logger.Errorw("error in db insert", "err", err)
	}
	return err
}

func (impl HelmValuesServiceImpl) GetHelmValues(appName, targetEnvironment string) (*HelmValues, error) {
	hv := &HelmValues{
		AppName:           appName,
		TargetEnvironment: targetEnvironment,
	}
	err := impl.dbConnection.Select(hv)
	return hv, err
}
