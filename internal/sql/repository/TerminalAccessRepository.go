package repository

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
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
}

func NewTerminalAccessRepositoryImpl(dbConnection *pg.DB) *TerminalAccessRepositoryImpl {
	return &TerminalAccessRepositoryImpl{
		dbConnection: dbConnection,
	}
}

func (impl TerminalAccessRepositoryImpl) FetchTerminalAccessTemplate(templateName string) (*models.TerminalAccessTemplates, error) {
	panic("implement me")
}

func (impl TerminalAccessRepositoryImpl) FetchAllTemplates() ([]*models.TerminalAccessTemplates, error) {
	panic("implement me")
}

func (impl TerminalAccessRepositoryImpl) GetUserTerminalAccessData(id int) (*models.UserTerminalAccessData, error) {
	panic("implement me")
}

// GetUserTerminalAccessDataByUser return empty array for no data and return only running instances
func (impl TerminalAccessRepositoryImpl) GetUserTerminalAccessDataByUser(userId int32) ([]*models.UserTerminalAccessData, error) {
	panic("implement me")
}

func (impl TerminalAccessRepositoryImpl) SaveUserTerminalAccessData(data *models.UserTerminalAccessData) error {
	panic("implement me")
}

func (impl TerminalAccessRepositoryImpl) UpdateUserTerminalAccessData(data *models.UserTerminalAccessData) error {
	panic("implement me")
}

func (impl TerminalAccessRepositoryImpl) UpdateUserTerminalStatus(id int, status string) error {
	panic("implement me")
}
