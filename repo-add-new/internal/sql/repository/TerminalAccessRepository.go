package repository

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
	"time"
)

type TerminalAccessRepository interface {
	FetchTerminalAccessTemplate(templateName string) (*models.TerminalAccessTemplates, error)
	FetchAllTemplates() ([]*models.TerminalAccessTemplates, error)
	GetUserTerminalAccessData(id int) (*models.UserTerminalAccessData, error)
	GetAllRunningUserTerminalData() ([]*models.UserTerminalAccessData, error)
	SaveUserTerminalAccessData(data *models.UserTerminalAccessData) error
	UpdateUserTerminalAccessData(data *models.UserTerminalAccessData) error
	UpdateUserTerminalStatus(id int, status string) error
}

type TerminalAccessRepositoryImpl struct {
	dbConnection   *pg.DB
	Logger         *zap.SugaredLogger
	templatesCache []*models.TerminalAccessTemplates
}

func NewTerminalAccessRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *TerminalAccessRepositoryImpl {

	repoImpl := &TerminalAccessRepositoryImpl{
		dbConnection: dbConnection,
		Logger:       logger,
	}
	go repoImpl.FetchAllTemplates()
	return repoImpl
}

func (impl TerminalAccessRepositoryImpl) FetchTerminalAccessTemplate(templateName string) (*models.TerminalAccessTemplates, error) {
	accessTemplates, err := impl.FetchAllTemplates()
	if err != nil {
		return nil, err
	}
	for _, accessTemplate := range accessTemplates {
		if accessTemplate.TemplateName == templateName {
			return accessTemplate, nil
		}
	}
	return nil, err
}

func (impl TerminalAccessRepositoryImpl) FetchAllTemplates() ([]*models.TerminalAccessTemplates, error) {

	if impl.templatesCache != nil && len(impl.templatesCache) != 0 {
		return impl.templatesCache, nil
	}

	var templates []*models.TerminalAccessTemplates
	err := impl.dbConnection.
		Model(&templates).
		Select()
	if err == pg.ErrNoRows {
		impl.Logger.Error("no terminal access templates found")
		err = nil
	}
	impl.templatesCache = templates
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

func (impl TerminalAccessRepositoryImpl) SaveUserTerminalAccessData(data *models.UserTerminalAccessData) error {
	data.CreatedBy = data.UserId
	data.UpdatedBy = data.UserId
	data.CreatedOn = time.Now()
	data.UpdatedOn = time.Now()
	return impl.dbConnection.Insert(data)
}

func (impl TerminalAccessRepositoryImpl) UpdateUserTerminalAccessData(data *models.UserTerminalAccessData) error {
	data.UpdatedBy = data.UserId
	data.UpdatedOn = time.Now()
	return impl.dbConnection.Update(data)
}

func (impl TerminalAccessRepositoryImpl) UpdateUserTerminalStatus(id int, status string) error {
	accessData := &models.UserTerminalAccessData{
		Id:     id,
		Status: status,
	}
	accessData.UpdatedOn = time.Now()
	_, err := impl.dbConnection.Model(accessData).WherePK().UpdateNotNull()
	return err
}

func (impl TerminalAccessRepositoryImpl) GetAllRunningUserTerminalData() ([]*models.UserTerminalAccessData, error) {
	var accessDataArray []*models.UserTerminalAccessData
	err := impl.dbConnection.Model(&accessDataArray).
		WhereGroup(func(query *orm.Query) (*orm.Query, error) {
			query = query.WhereOr("status = ?", string(models.TerminalPodRunning)).WhereOr("status = ?", string(models.TerminalPodStarting))
			return query, nil
		}).
		Select()

	if err == pg.ErrNoRows {
		impl.Logger.Debug("no running/starting pods found")
		err = nil
	}
	return accessDataArray, err
}
