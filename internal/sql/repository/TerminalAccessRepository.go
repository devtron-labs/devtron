package repository

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type TerminalAccessRepository interface {
	FetchTerminalAccessTemplate(templateName string) (*models.TerminalAccessTemplates, error)
	FetchAllTemplates() ([]*models.TerminalAccessTemplates, error)
	GetUserTerminalAccessData(id int) (*models.UserTerminalAccessData, error)
	GetUserTerminalAccessDataByUser(userId int32) ([]*models.UserTerminalAccessData, error)
	SaveUserTerminalAccessData(data *models.UserTerminalAccessData) error
	UpdateUserTerminalAccessData(data *models.UserTerminalAccessData) error
	UpdateUserTerminalStatus(id int, status string) error
}

type TerminalAccessRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewTerminalAccessRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *TerminalAccessRepositoryImpl {
	return &TerminalAccessRepositoryImpl{
		dbConnection: dbConnection,
		Logger:       logger,
	}
}

func (impl TerminalAccessRepositoryImpl) FetchTerminalAccessTemplate(templateName string) (*models.TerminalAccessTemplates, error) {
	template := &models.TerminalAccessTemplates{}
	err := impl.dbConnection.
		Model(template).
		Where("template_name = ?", templateName).
		Select()
	return template, err
}

func (impl TerminalAccessRepositoryImpl) FetchAllTemplates() ([]*models.TerminalAccessTemplates, error) {
	var templates []*models.TerminalAccessTemplates
	err := impl.dbConnection.
		Model(templates).
		Select()
	if err == pg.ErrNoRows {
		impl.Logger.Errorw("no terminal access templates found")
		err = nil
	}
	return templates, err
}

func (impl TerminalAccessRepositoryImpl) GetUserTerminalAccessData(id int) (*models.UserTerminalAccessData, error) {
	terminalAccessData := &models.UserTerminalAccessData{Id: id}
	err := impl.dbConnection.
		Model(terminalAccessData).
		WherePK().
		Select()
	return terminalAccessData, err
}

// GetUserTerminalAccessDataByUser return empty array for no data and return only running instances
func (impl TerminalAccessRepositoryImpl) GetUserTerminalAccessDataByUser(userId int32) ([]*models.UserTerminalAccessData, error) {
	var accessDataArray []*models.UserTerminalAccessData
	err := impl.dbConnection.Model(accessDataArray).
		Where("user_id = ?", userId).
		WhereGroup(func(query *orm.Query) (*orm.Query, error) {
			query = query.WhereOr("status = ?", string(bean.TerminalPodRunning)).WhereOr("status = ?", string(bean.TerminalPodStarting))
			return query, nil
		}).
		Select()
	if err == pg.ErrNoRows {
		impl.Logger.Errorw("no running/starting pods found", "userId", userId)
		err = nil
	}
	return accessDataArray, err
}

func (impl TerminalAccessRepositoryImpl) SaveUserTerminalAccessData(data *models.UserTerminalAccessData) error {
	return impl.dbConnection.Insert(data)
}

func (impl TerminalAccessRepositoryImpl) UpdateUserTerminalAccessData(data *models.UserTerminalAccessData) error {
	return impl.dbConnection.Update(data)
}

func (impl TerminalAccessRepositoryImpl) UpdateUserTerminalStatus(id int, status string) error {
	accessDataArray := &models.UserTerminalAccessData{
		Id:     id,
		Status: status,
	}
	_, err := impl.dbConnection.Model(accessDataArray).WherePK().UpdateNotNull()
	return err
}
