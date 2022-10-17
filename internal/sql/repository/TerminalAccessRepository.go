package repository

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type TerminalAccessRepository interface {
	FetchTerminalAccessTemplate(templateName string) (*models.TerminalAccessTemplates, error)
	FetchAllTemplates() ([]*models.TerminalAccessTemplates, error)
	GetUserTerminalAccessData(id int) (*models.UserTerminalAccessData, error)
	GetUserTerminalAccessDataByUser(userId int32) ([]*models.UserTerminalAccessData, error)
	GetAllUserTerminalAccessData() ([]*models.UserTerminalAccessData, error)
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
		Model(&templates).
		Select()
	if err == pg.ErrNoRows {
		impl.Logger.Debug("no terminal access templates found")
		err = nil
	}
	//templates = append(templates, &models.TerminalAccessTemplates{
	//	TemplateName:     bean.TerminalAccessRoleTemplateName,
	//	TemplateKindData: "{\"group\":\"\", \"version\":\"\", \"kind\":\"\"}",
	//	TemplateData:     "",
	//})
	templates = append(templates, &models.TerminalAccessTemplates{
		TemplateName:     models.TerminalAccessServiceAccountTemplateName,
		TemplateKindData: "{\"version\":\"v1\", \"kind\":\"ServiceAccount\"}",
		TemplateData:     "{\"apiVersion\":\"v1\",\"kind\":\"ServiceAccount\",\"metadata\":{\"name\":\"terminal-access-service-account\",\"namespace\":\"default\"}}",
	})
	templates = append(templates, &models.TerminalAccessTemplates{
		TemplateName:     models.TerminalAccessRoleBindingTemplateName,
		TemplateKindData: "{\"group\":\"rbac.authorization.k8s.io\",\"version\":\"v1\",\"kind\":\"ClusterRoleBinding\"}",
		TemplateData:     "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRoleBinding\",\"metadata\":{\"name\":\"terminal-access-role-binding\"},\"subjects\":[{\"kind\":\"ServiceAccount\",\"name\":\"terminal-access-service-account\",\"namespace\":\"default\"}],\"roleRef\":{\"kind\":\"ClusterRole\",\"name\":\"cluster-admin\",\"apiGroup\":\"rbac.authorization.k8s.io\"}}",
	})
	templates = append(templates, &models.TerminalAccessTemplates{
		TemplateName:     models.TerminalAccessPodTemplateName,
		TemplateKindData: "{\"group\":\"\", \"version\":\"v1\", \"kind\":\"Pod\"}",
		TemplateData:     "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"name\":\"${pod_name}\"},\"spec\":{\"serviceAccountName\":\"terminal-access-service-account\",\"containers\":[{\"name\":\"internal-kubectl\",\"image\":\"${base_image}\"}]}}",
	})

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

// GetUserTerminalAccessDataByUser return empty array for no data and return only running/starting instances
func (impl TerminalAccessRepositoryImpl) GetUserTerminalAccessDataByUser(userId int32) ([]*models.UserTerminalAccessData, error) {
	var accessDataArray []*models.UserTerminalAccessData
	err := impl.dbConnection.Model(&accessDataArray).
		Where("user_id = ?", userId).
		WhereGroup(func(query *orm.Query) (*orm.Query, error) {
			query = query.WhereOr("status = ?", string(models.TerminalPodRunning)).WhereOr("status = ?", string(models.TerminalPodStarting))
			return query, nil
		}).
		Select()
	if err == pg.ErrNoRows {
		impl.Logger.Debug("no running/starting pods found", "userId", userId)
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

func (impl TerminalAccessRepositoryImpl) GetAllUserTerminalAccessData() ([]*models.UserTerminalAccessData, error) {
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
