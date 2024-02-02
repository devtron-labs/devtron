package bean

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
)

type AppBean struct {
	Id      int            `json:"id"`
	Name    string         `json:"name,notnull"`
	TeamId  int            `json:"teamId,omitempty"`
	AppType helper.AppType `json:"-"` // used in specific case only
}

func InitFromAppEntity(appEntity *app.App) *AppBean {
	return &AppBean{
		Id:      appEntity.Id,
		Name:    appEntity.AppName,
		TeamId:  appEntity.TeamId,
		AppType: appEntity.AppType,
	}
}

type TeamAppBean struct {
	ProjectId   int        `json:"projectId"`
	ProjectName string     `json:"projectName"`
	AppList     []*AppBean `json:"appList"`
}
