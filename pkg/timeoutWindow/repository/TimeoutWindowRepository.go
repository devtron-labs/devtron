/*
 * Copyright (c) 2024. Devtron Inc.
 */

package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type TimeWindowRepository interface {
	Create(tx *pg.Tx, model *TimeoutWindowConfiguration) (*TimeoutWindowConfiguration, error)
	Update(tx *pg.Tx, model *TimeoutWindowConfiguration) (*TimeoutWindowConfiguration, error)
	CreateInBatch(tx *pg.Tx, models []*TimeoutWindowConfiguration) ([]*TimeoutWindowConfiguration, error)
	UpdateInBatch(models []*TimeoutWindowConfiguration) ([]*TimeoutWindowConfiguration, error)
	GetWithExpressionAndFormat(tx *pg.Tx, expression string, format bean.ExpressionFormat) (*TimeoutWindowConfiguration, error)
	GetWithIds(ids []int) ([]*TimeoutWindowConfiguration, error)
	UpdateTimeoutExpressionAndFormatForIds(tx *pg.Tx, expression string, ids []int, format bean.ExpressionFormat, loggedInUserId int32) error
}
type TimeoutWindowConfiguration struct {
	TableName               struct{}              `sql:"timeout_window_configuration" pg:",discard_unknown_columns"`
	Id                      int                   `sql:"id,pk"`
	TimeoutWindowExpression string                `sql:"timeout_window_expression,notnull"`
	ExpressionFormat        bean.ExpressionFormat `sql:"timeout_window_expression_format,notnull"` // '1=timestamp, 2=other format'
	sql.AuditLog
}

type TimeWindowRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewTimeWindowRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *TimeWindowRepositoryImpl {
	return &TimeWindowRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

// GetWithExpressionAndFormat takes expression and format as input and return corresponding db entry matching it else return pg no rows
func (impl TimeWindowRepositoryImpl) GetWithExpressionAndFormat(tx *pg.Tx, expression string, format bean.ExpressionFormat) (*TimeoutWindowConfiguration, error) {
	var model TimeoutWindowConfiguration
	//using transaction here as to get the just created timeoutWindowConfig
	err := tx.Model(&model).
		Where("timeout_window_expression = ?", expression).
		Where("timeout_window_expression_format = ?", format).
		Order("id desc").Limit(1).
		Select()
	if err != nil {
		impl.logger.Errorw("error in GetWithExpressionAndFormat", "err", err)
		return nil, err
	}
	return &model, nil
}

// GetWithIds takes in timeout window ids and results rows corresponding to that id in db.
func (impl TimeWindowRepositoryImpl) GetWithIds(ids []int) ([]*TimeoutWindowConfiguration, error) {
	var model []*TimeoutWindowConfiguration

	if len(ids) == 0 {
		return model, nil
	}

	err := impl.dbConnection.Model(&model).Where("id in (?)", pg.In(ids)).Select()
	if err != nil {
		impl.logger.Errorw("error in GetWithIds", "err", err, "ids", ids)
		return nil, err
	}
	return model, nil

}

// Create takes timeModel in input and create it in db.
func (impl TimeWindowRepositoryImpl) Create(tx *pg.Tx, model *TimeoutWindowConfiguration) (*TimeoutWindowConfiguration, error) {
	err := tx.Insert(model)
	if err != nil {
		impl.logger.Errorw("error in CreateInBatch time window", "err", err)
		return nil, err
	}
	return model, nil
}

// Update updates the time window model in db, returns the updated model
func (impl TimeWindowRepositoryImpl) Update(tx *pg.Tx, model *TimeoutWindowConfiguration) (*TimeoutWindowConfiguration, error) {
	_, err := tx.Model(&model).Update()
	if err != nil {
		impl.logger.Errorw("error in Update time window", "err", err)
		return nil, err
	}
	return model, nil
}

// CreateInBatch create takes timeModels in input and creates them in db in bulk
func (impl TimeWindowRepositoryImpl) CreateInBatch(tx *pg.Tx, models []*TimeoutWindowConfiguration) ([]*TimeoutWindowConfiguration, error) {
	err := tx.Insert(&models)
	if err != nil {
		impl.logger.Errorw("error in CreateInBatch time window", "err", err)
		return nil, err
	}
	return models, nil
}

// UpdateInBatch updates the time windows model in bulk in db, returns the updated models
func (impl TimeWindowRepositoryImpl) UpdateInBatch(models []*TimeoutWindowConfiguration) ([]*TimeoutWindowConfiguration, error) {
	_, err := impl.dbConnection.Model(&models).Update()
	if err != nil {
		impl.logger.Errorw("error in UpdateInBatch time window", "err", err)
		return nil, err
	}
	return models, nil
}

// UpdateTimeoutExpressionAndFormatForIds bulk updates expression and format for given user ids
func (impl TimeWindowRepositoryImpl) UpdateTimeoutExpressionAndFormatForIds(tx *pg.Tx, expression string, ids []int, format bean.ExpressionFormat, loggedInUserId int32) error {
	var model []*TimeoutWindowConfiguration
	_, err := tx.Model(&model).Set("timeout_window_expression = ?", expression).
		Set("timeout_window_expression_format = ?", format).
		Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", loggedInUserId).
		Where("id in (?)", pg.In(ids)).
		Update()
	if err != nil {
		impl.logger.Errorw("error in UpdateTimeoutExpressionForIds ", "err", err)
		return err
	}
	return nil

}
