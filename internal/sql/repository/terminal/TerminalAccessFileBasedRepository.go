package terminal

import (
	"errors"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type TerminalAccessFileBasedRepository struct {
	logger       *zap.SugaredLogger
	dbConnection *gorm.DB
}

func NewTerminalAccessFileBasedRepository(connection *sql.SqliteConnection, logger *zap.SugaredLogger) *TerminalAccessFileBasedRepository {
	terminalAccessData := &models.UserTerminalAccessData{}
	terminalAccessTemplates := &models.TerminalAccessTemplates{}
	connection.Migrator.MigrateEntities(terminalAccessData, terminalAccessTemplates)
	logger.Debugw("cluster terminal access file based repository initialized")
	return &TerminalAccessFileBasedRepository{logger: logger, dbConnection: connection.DbConnection}
}

func (impl TerminalAccessFileBasedRepository) FetchTerminalAccessTemplate(templateName string) (*models.TerminalAccessTemplates, error) {
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

func (impl TerminalAccessFileBasedRepository) FetchAllTemplates() ([]*models.TerminalAccessTemplates, error) {
	accessTemplates, err := impl.fetchAllTemplates()
	if err != nil {
		return nil, err
	}
	if len(accessTemplates) == 0 {
		impl.createDefaultAccessTemplates()
		accessTemplates, err = impl.fetchAllTemplates()
	} else {
		nodeDebugTemplateExists := false
		for _, template := range accessTemplates {
			if template.TemplateName == "terminal-node-debug-pod" {
				nodeDebugTemplateExists = true
			}
		}
		if !nodeDebugTemplateExists {
			impl.createNodeDebugTemplate()
		}
	}
	return accessTemplates, err
}

func (impl TerminalAccessFileBasedRepository) fetchAllTemplates() ([]*models.TerminalAccessTemplates, error) {
	var accessTemplates []*models.TerminalAccessTemplates
	result := impl.dbConnection.
		Find(&accessTemplates)
	err := result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while finding all terminal access templates", "err", err)
		return nil, errors.New("failed to fetch access templates")
	}
	return accessTemplates, nil
}

func (impl TerminalAccessFileBasedRepository) GetUserTerminalAccessData(id int) (*models.UserTerminalAccessData, error) {
	accessData := &models.UserTerminalAccessData{}
	result := impl.dbConnection.
		Where("Id = ?", id).
		Find(accessData)
	err := result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while fetching access data", "id", id, "err", err)
		return nil, errors.New("failed to fetch access data")
	}
	return accessData, nil
}

func (impl TerminalAccessFileBasedRepository) GetAllRunningUserTerminalData() ([]*models.UserTerminalAccessData, error) {
	var accessData []*models.UserTerminalAccessData
	result := impl.dbConnection.
		Where("status = ?", string(models.TerminalPodRunning)).Or("status = ?", string(models.TerminalPodStarting)).
		Find(&accessData)
	err := result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while finding all running/starting terminal access data", "err", err)
		return nil, errors.New("failed to fetch access data")
	}
	return accessData, nil
}

func (impl TerminalAccessFileBasedRepository) SaveUserTerminalAccessData(data *models.UserTerminalAccessData) error {
	result := impl.dbConnection.Create(data)
	err := result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while saving terminal access data", "data", data, "err", err)
		return errors.New("error while saving terminal data")
	}
	return nil
}

func (impl TerminalAccessFileBasedRepository) UpdateUserTerminalAccessData(data *models.UserTerminalAccessData) error {
	result := impl.dbConnection.Model(data).Updates(data)
	err := result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while updating terminal access data", "data", data, "error", err)
		return errors.New("failed to update terminal access data")
	}
	return nil
}

func (impl TerminalAccessFileBasedRepository) UpdateUserTerminalStatus(id int, status string) error {
	result := impl.dbConnection.Model(&models.UserTerminalAccessData{}).Where("Id = ?", id).Update("status", status)
	err := result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while updating cluster connection status", "id", id, "status", status, "err", err)
		return errors.New("failed to update terminal access status")
	}
	return nil
}

func (impl TerminalAccessFileBasedRepository) createDefaultAccessTemplates() {

	var defaultTemplates []*models.TerminalAccessTemplates
	defaultTemplates = append(defaultTemplates, &models.TerminalAccessTemplates{
		TemplateName: "terminal-access-service-account",
		TemplateData: GetDefaultTerminalAccessServiceAccount(),
		AuditLog: sql.AuditLog{
			CreatedBy: 1,
			UpdatedBy: 1,
			CreatedOn: time.Now(),
			UpdatedOn: time.Now(),
		},
	})
	defaultTemplates = append(defaultTemplates, &models.TerminalAccessTemplates{
		TemplateName: "terminal-access-role-binding",
		TemplateData: GetDefaultTerminalAccessRoleBindingTemplate(),
		AuditLog: sql.AuditLog{
			CreatedBy: 1,
			UpdatedBy: 1,
			CreatedOn: time.Now(),
			UpdatedOn: time.Now(),
		},
	})
	defaultTemplates = append(defaultTemplates, &models.TerminalAccessTemplates{
		TemplateName: "terminal-access-pod",
		TemplateData: GetDefaultTerminalAccessPodTemplate(),
		AuditLog: sql.AuditLog{
			CreatedBy: 1,
			UpdatedBy: 1,
			CreatedOn: time.Now(),
			UpdatedOn: time.Now(),
		},
	})
	defaultTemplates = append(defaultTemplates, impl.getNodeDebugTemplate())
	result := impl.dbConnection.Create(&defaultTemplates)
	err := result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while creating default access templates", "err", err)
	}
}

func (impl TerminalAccessFileBasedRepository) createNodeDebugTemplate() {
	var defaultTemplates []*models.TerminalAccessTemplates
	defaultTemplates = append(defaultTemplates, impl.getNodeDebugTemplate())
	result := impl.dbConnection.Create(&defaultTemplates)
	err := result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while creating default access templates", "err", err)
	}
}

func (impl TerminalAccessFileBasedRepository) getNodeDebugTemplate() *models.TerminalAccessTemplates {
	return &models.TerminalAccessTemplates{
		TemplateName: "terminal-node-debug-pod",
		TemplateData: GetDefaultTerminalAccessNodeDebugTemplate(),
		AuditLog: sql.AuditLog{
			CreatedBy: 1,
			UpdatedBy: 1,
			CreatedOn: time.Now(),
			UpdatedOn: time.Now(),
		},
	}
}
