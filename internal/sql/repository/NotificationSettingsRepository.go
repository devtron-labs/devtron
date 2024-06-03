/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package repository

import (
	"context"
	"github.com/devtron-labs/devtron/pkg/notifier/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.opentelemetry.io/otel"
	"strconv"
)

type NotificationSettingsRepository interface {
	FindNSViewCount() (int, error)
	SaveNotificationSettingsConfig(notificationSettingsView *NotificationSettingsView, tx *pg.Tx) (*NotificationSettingsView, error)
	FindNotificationSettingsViewById(id int) (*NotificationSettingsView, error)
	FindNotificationSettingsViewByIds(id []*int) ([]*NotificationSettingsView, error)
	UpdateNotificationSettingsView(notificationSettingsView *NotificationSettingsView, tx *pg.Tx) (*NotificationSettingsView, error)
	UpdateNotificationSettings(notificationSettings *NotificationSettings, tx *pg.Tx) (*NotificationSettings, error)
	UpdateNotificationRules(notificationRules []*NotificationRule, tx *pg.Tx) (int, error)
	FindNotificationRulesBySettingIds(settingIds []int) ([]*NotificationRule, error)
	FindNotificationSettingsByViewId(viewId int) ([]NotificationSettings, error)
	FindAllNotificationRuleIds(viewId int) ([]int, error)
	SaveAllNotificationSettings(notificationSettings []NotificationSettings, tx *pg.Tx) (int, error)
	SaveNotificationRule(notificationRule *NotificationRule, tx *pg.Tx) (*NotificationRule, error)
	DeleteNotificationSettingsByConfigId(viewId int, tx *pg.Tx) (int, error)
	DeleteNotificationRulesByViewId(notificationRuleIds []int, tx *pg.Tx) (int, error)
	FindAll(offset int, size int, internalOnly bool) ([]*NotificationSettingsView, error)
	DeleteNotificationSettingsViewById(id int, tx *pg.Tx) (int, error)

	FindNotificationSettingDeploymentOptions(settingRequest *SearchRequest) ([]*SettingOptionDTO, error)
	FindNotificationSettingBuildOptions(settingRequest *SearchRequest) ([]*SettingOptionDTO, error)
	FetchNotificationSettingGroupBy(viewId int) ([]NotificationSettings, error)
	FetchNotificationSettingByViewId(viewId int) ([]NotificationSettings, error)
	FindNotificationSettingsByConfigIdAndConfigType(configId int, configType string) ([]*NotificationSettings, error)
	FindNotificationSettingsWithRules(ctx context.Context, eventId int, req GetRulesRequest) ([]NotificationSettings, error)
}

type NotificationSettingsRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewNotificationSettingsRepositoryImpl(dbConnection *pg.DB) *NotificationSettingsRepositoryImpl {
	return &NotificationSettingsRepositoryImpl{dbConnection: dbConnection}
}

type NotificationSettingsView struct {
	tableName struct{} `sql:"notification_settings_view" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	Config    string   `sql:"config"`
	Internal  bool     `sql:"internal"`
	sql.AuditLog
}

type NotificationSettingsViewWithAppEnv struct {
	Id              int    `json:"id"`
	AppId           *int   `json:"app_id"`
	EnvId           *int   `json:"env_id"`
	ConfigName      string `sql:"config_name"`
	Config          string `sql:"config"`
	AppName         string `json:"app_name"`
	EnvironmentName string `json:"env_name"`
	sql.AuditLog
}

type NotificationSettings struct {
	tableName            struct{} `sql:"notification_settings" pg:",discard_unknown_columns"`
	Id                   int      `sql:"id,pk"`
	TeamId               *int     `sql:"team_id"`
	AppId                *int     `sql:"app_id"`
	EnvId                *int     `sql:"env_id"`
	PipelineId           *int     `sql:"pipeline_id"`
	PipelineType         string   `sql:"pipeline_type"`
	EventTypeId          int      `sql:"event_type_id"`
	Config               string   `sql:"config"`
	ViewId               int      `sql:"view_id"`
	NotificationRuleId   int      `sql:"notification_rule_id"`
	AdditionalConfigJson string   `sql:"additional_config_json"` // user defined config json;
	NotificationRule     *NotificationRule
}

type ImageScanFilterFlags struct {
	DiffViewEnabled bool `json:"diffViewEnabled,omitempty"`
}

type ImageScanFilterConditions struct {
	Severity         []string `json:"severity,omitempty"`
	PolicyPermission []string `json:"policyPermission,omitempty"`
}

type NotificationRule struct {
	tableName     struct{}           `sql:"notification_rule" pg:",discard_unknown_columns"`
	Id            int                `sql:"id,pk"`
	ConditionType bean.ConditionType `sql:"condition_type"` // 0 -> FAIL; 1 -> PASS;
	Expression    string             `sql:"expression"`     // CEL expression; conditions based on event variables
	sql.AuditLog
}

type SettingOptionDTO struct {
	//TeamId       int    `json:"-"`
	//AppId        int    `json:"-"`
	//EnvId        int    `json:"-"`
	PipelineId      int    `json:"pipelineId"`
	PipelineName    string `json:"pipelineName"`
	PipelineType    string `json:"pipelineType"`
	AppName         string `json:"appName"`
	EnvironmentName string `json:"environmentName,omitempty"`
}

func (impl *NotificationSettingsRepositoryImpl) FindNSViewCount() (int, error) {
	count, err := impl.dbConnection.Model(&NotificationSettingsView{}).
		WhereGroup(func(query *orm.Query) (*orm.Query, error) {
			query = query.
				WhereOr("internal IS NULL").
				WhereOr("internal = ?", false)
			return query, nil
		}).
		Count()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (impl *NotificationSettingsRepositoryImpl) FindAll(offset int, size int, internalOnly bool) ([]*NotificationSettingsView, error) {
	var ns []*NotificationSettingsView
	query := impl.dbConnection.Model(&ns)
	if internalOnly {
		query = query.Where("internal = ?", true)
	} else {
		query = query.
			WhereGroup(func(query *orm.Query) (*orm.Query, error) {
				query = query.
					WhereOr("internal IS NULL").
					WhereOr("internal = ?", false)
				return query, nil
			})
	}
	err := query.Order("created_on desc").
		Offset(offset).
		Limit(size).
		Select()
	if err != nil {
		return nil, err
	}
	return ns, err
}

func (impl *NotificationSettingsRepositoryImpl) SaveNotificationSettingsConfig(notificationSettingsView *NotificationSettingsView, tx *pg.Tx) (*NotificationSettingsView, error) {
	err := tx.Insert(notificationSettingsView)
	if err != nil {
		return nil, err
	}
	return notificationSettingsView, nil
}

func (impl *NotificationSettingsRepositoryImpl) FindNotificationSettingsViewById(id int) (*NotificationSettingsView, error) {
	notificationSettingsView := &NotificationSettingsView{}
	err := impl.dbConnection.Model(notificationSettingsView).Where("id = ?", id).Select()
	if err != nil {
		return nil, err
	}
	return notificationSettingsView, nil
}

func (impl *NotificationSettingsRepositoryImpl) FindNotificationSettingsViewByIds(ids []*int) ([]*NotificationSettingsView, error) {
	var notificationSettingsView []*NotificationSettingsView
	if len(ids) == 0 {
		return notificationSettingsView, nil
	}
	err := impl.dbConnection.Model(&notificationSettingsView).Where("id in (?)", pg.In(ids)).Select()
	if err != nil {
		return nil, err
	}
	return notificationSettingsView, nil
}

func (impl *NotificationSettingsRepositoryImpl) UpdateNotificationSettingsView(notificationSettingsView *NotificationSettingsView, tx *pg.Tx) (*NotificationSettingsView, error) {
	err := tx.Update(notificationSettingsView)
	if err != nil {
		return nil, err
	}
	return notificationSettingsView, nil
}

func (impl *NotificationSettingsRepositoryImpl) SaveAllNotificationSettings(notificationSettings []NotificationSettings, tx *pg.Tx) (int, error) {
	res, err := tx.Model(&notificationSettings).Insert()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (impl *NotificationSettingsRepositoryImpl) SaveNotificationRule(notificationRule *NotificationRule, tx *pg.Tx) (*NotificationRule, error) {
	err := tx.Insert(notificationRule)
	if err != nil {
		return nil, err
	}
	return notificationRule, nil
}

func (impl *NotificationSettingsRepositoryImpl) UpdateNotificationSettings(notificationSettings *NotificationSettings, tx *pg.Tx) (*NotificationSettings, error) {
	err := tx.Update(notificationSettings)
	if err != nil {
		return nil, err
	}
	return notificationSettings, nil
}

func (impl *NotificationSettingsRepositoryImpl) UpdateNotificationRules(notificationRules []*NotificationRule, tx *pg.Tx) (int, error) {
	res, err := tx.Model(&notificationRules).Update()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (impl *NotificationSettingsRepositoryImpl) FindNotificationRulesBySettingIds(settingIds []int) ([]*NotificationRule, error) {
	if settingIds == nil || len(settingIds) == 0 {
		return nil, pg.ErrNoRows
	}
	var notificationRules []*NotificationRule
	err := impl.dbConnection.Model(&notificationRules).
		Where("notification_setting_id in (?)", pg.In(settingIds)).
		Select()
	if err != nil {
		return nil, err
	}
	return notificationRules, nil
}

func (impl *NotificationSettingsRepositoryImpl) FindNotificationSettingsByViewId(viewId int) ([]NotificationSettings, error) {
	var notificationSettings []NotificationSettings
	err := impl.dbConnection.Model(&notificationSettings).Where("view_id = ?", viewId).Select()
	if err != nil {
		return nil, err
	}
	return notificationSettings, nil
}

func (impl *NotificationSettingsRepositoryImpl) FindAllNotificationRuleIds(viewId int) ([]int, error) {
	var notificationRuleIds []int
	err := impl.dbConnection.
		Model((*NotificationSettings)(nil)).
		Column("notification_rule_id").
		Where("view_id = ?", viewId).
		Where("notification_rule_id IS NOT NULL").
		Select(&notificationRuleIds)
	return notificationRuleIds, err
}

func (impl *NotificationSettingsRepositoryImpl) DeleteNotificationSettingsByConfigId(viewId int, tx *pg.Tx) (int, error) {
	var notificationSettings *NotificationSettings
	res, err := tx.Model(notificationSettings).Where("view_id = ?", viewId).Delete()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (impl *NotificationSettingsRepositoryImpl) DeleteNotificationRulesByViewId(notificationRuleIds []int, tx *pg.Tx) (int, error) {
	if notificationRuleIds == nil {
		return 0, pg.ErrNoRows
	}
	var notificationRules *NotificationRule
	res, err := tx.
		Model(notificationRules).
		Where("id in (?)", pg.In(notificationRuleIds)).
		Delete()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (impl *NotificationSettingsRepositoryImpl) DeleteNotificationSettingsViewById(id int, tx *pg.Tx) (int, error) {
	var notificationSettingsView *NotificationSettingsView
	res, err := tx.Model(notificationSettingsView).Where("id = ?", id).Delete()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (impl *NotificationSettingsRepositoryImpl) FindNotificationSettingDeploymentOptions(settingRequest *SearchRequest) ([]*SettingOptionDTO, error) {
	var settingOption []*SettingOptionDTO
	query := "SELECT p.id as pipeline_id,p.pipeline_name, env.environment_name, a.app_name FROM pipeline p" +
		" INNER JOIN app a on a.id=p.app_id" +
		" INNER JOIN environment env on env.id = p.environment_id"
	query = query + " WHERE p.deleted = false"

	if len(settingRequest.TeamId) > 0 {
		query = query + " AND a.team_id in (?)"
	} else if len(settingRequest.EnvId) > 0 {
		query = query + " AND p.environment_id in (?)"
	} else if len(settingRequest.AppId) > 0 {
		query = query + " AND p.app_id in (?)"
	} else if len(settingRequest.PipelineName) > 0 {
		query = query + " AND p.pipeline_name like (?)"
	}
	query = query + " GROUP BY 1,2,3,4;"

	var err error
	if len(settingRequest.TeamId) > 0 && len(settingRequest.EnvId) == 0 && len(settingRequest.AppId) == 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, pg.In(settingRequest.TeamId))
	} else if len(settingRequest.TeamId) == 0 && len(settingRequest.EnvId) > 0 && len(settingRequest.AppId) == 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, pg.In(settingRequest.EnvId))
	} else if len(settingRequest.TeamId) == 0 && len(settingRequest.EnvId) == 0 && len(settingRequest.AppId) > 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, pg.In(settingRequest.AppId))
	} else if len(settingRequest.TeamId) > 0 && len(settingRequest.EnvId) > 0 && len(settingRequest.AppId) == 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, pg.In(settingRequest.TeamId), pg.In(settingRequest.EnvId))
	} else if len(settingRequest.TeamId) > 0 && len(settingRequest.EnvId) == 0 && len(settingRequest.AppId) > 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, pg.In(settingRequest.TeamId), pg.In(settingRequest.AppId))
	} else if len(settingRequest.TeamId) == 0 && len(settingRequest.EnvId) > 0 && len(settingRequest.AppId) > 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, pg.In(settingRequest.EnvId), pg.In(settingRequest.AppId))
	} else if len(settingRequest.TeamId) > 0 && len(settingRequest.EnvId) > 0 && len(settingRequest.AppId) > 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, pg.In(settingRequest.TeamId), pg.In(settingRequest.EnvId), pg.In(settingRequest.AppId))
	} else if len(settingRequest.PipelineName) > 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, settingRequest.PipelineName)
	}
	if err != nil {
		return nil, err
	}
	return settingOption, err
}

func (impl *NotificationSettingsRepositoryImpl) FindNotificationSettingBuildOptions(settingRequest *SearchRequest) ([]*SettingOptionDTO, error) {
	var settingOption []*SettingOptionDTO
	query := "SELECT cip.id as pipeline_id,cip.name as pipeline_name, a.app_name from ci_pipeline cip" +
		" INNER JOIN app a on a.id = cip.app_id" +
		" INNER JOIN team t on t.id= a.team_id"
	if len(settingRequest.EnvId) > 0 {
		query = query + " INNER JOIN ci_artifact cia on cia.pipeline_id = cip.id"
		query = query + " INNER JOIN cd_workflow wf on wf.ci_artifact_id = cia.id"
		query = query + " INNER JOIN pipeline p on p.id = wf.pipeline_id"
	}
	query = query + " WHERE cip.deleted = false"

	if len(settingRequest.TeamId) > 0 {
		query = query + " AND a.team_id in (?)"
	} else if len(settingRequest.EnvId) > 0 {
		query = query + " AND p.environment_id in (?)"
	} else if len(settingRequest.AppId) > 0 {
		query = query + " AND cip.app_id in (?)"
	} else if len(settingRequest.PipelineName) > 0 {
		query = query + " AND cip.name like (?)"
	}
	query = query + " GROUP BY 1,2,3;"
	var err error
	if len(settingRequest.TeamId) > 0 && len(settingRequest.EnvId) == 0 && len(settingRequest.AppId) == 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, pg.In(settingRequest.TeamId))
	} else if len(settingRequest.TeamId) == 0 && len(settingRequest.EnvId) > 0 && len(settingRequest.AppId) == 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, pg.In(settingRequest.EnvId))
	} else if len(settingRequest.TeamId) == 0 && len(settingRequest.EnvId) == 0 && len(settingRequest.AppId) > 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, pg.In(settingRequest.AppId))
	} else if len(settingRequest.TeamId) > 0 && len(settingRequest.EnvId) > 0 && len(settingRequest.AppId) == 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, pg.In(settingRequest.TeamId), pg.In(settingRequest.EnvId))
	} else if len(settingRequest.TeamId) > 0 && len(settingRequest.EnvId) == 0 && len(settingRequest.AppId) > 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, pg.In(settingRequest.TeamId), pg.In(settingRequest.AppId))
	} else if len(settingRequest.TeamId) == 0 && len(settingRequest.EnvId) > 0 && len(settingRequest.AppId) > 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, pg.In(settingRequest.EnvId), pg.In(settingRequest.AppId))
	} else if len(settingRequest.TeamId) > 0 && len(settingRequest.EnvId) > 0 && len(settingRequest.AppId) > 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, pg.In(settingRequest.TeamId), pg.In(settingRequest.EnvId), pg.In(settingRequest.AppId))
	} else if len(settingRequest.PipelineName) > 0 {
		_, err = impl.dbConnection.Query(&settingOption, query, settingRequest.PipelineName)
	}
	if err != nil {
		return nil, err
	}
	return settingOption, err
}

type SearchRequest struct {
	TeamId       []*int `json:"teamId" validate:"number"`
	EnvId        []*int `json:"envId" validate:"number"`
	AppId        []*int `json:"appId" validate:"number"`
	PipelineName string `json:"pipelineName"`
	UserId       int32  `json:"-"`
}

type GetRulesRequest struct {
	TeamId       int
	EnvId        int
	AppId        int
	PipelineId   int
	PipelineType string
}

func (impl *NotificationSettingsRepositoryImpl) FetchNotificationSettingGroupBy(viewId int) ([]NotificationSettings, error) {
	var ns []NotificationSettings
	queryTemp := "select ns.team_id,ns.env_id,ns.app_id,ns.pipeline_id,ns.pipeline_type from notification_settings ns" +
		" where ns.view_id=? group by 1,2,3,4,5;"
	_, err := impl.dbConnection.Query(&ns, queryTemp, viewId)
	if err != nil {
		return nil, err
	}
	return ns, err
}

func (impl *NotificationSettingsRepositoryImpl) FetchNotificationSettingByViewId(viewId int) ([]NotificationSettings, error) {
	var notificationSettings []NotificationSettings
	err := impl.dbConnection.
		Model(&notificationSettings).
		Column("notification_settings.*", "NotificationRule").
		Where("view_id = ?", viewId).
		Select()
	if err != nil {
		return nil, err
	}
	return notificationSettings, err
}

func (impl *NotificationSettingsRepositoryImpl) FindNotificationSettingsByConfigIdAndConfigType(configId int, configType string) ([]*NotificationSettings, error) {
	var notificationSettings []*NotificationSettings
	err := impl.dbConnection.Model(&notificationSettings).Where("config::text like ?", "%dest\":\""+configType+"%").
		Where("config::text like ?", "%configId\":"+strconv.Itoa(configId)+"%").Select()
	if err != nil {
		return nil, err
	}
	return notificationSettings, nil
}

func (impl *NotificationSettingsRepositoryImpl) FindNotificationSettingsWithRules(ctx context.Context, eventId int, req GetRulesRequest) ([]NotificationSettings, error) {
	_, span := otel.Tracer("NotificationSettingsRepository").Start(ctx, "FindNotificationSettingsWithRules")
	defer span.End()
	if len(req.PipelineType) == 0 || eventId == 0 {
		return nil, pg.ErrNoRows
	}
	var notificationSettings []NotificationSettings
	err := impl.dbConnection.
		Model(&notificationSettings).
		Column("notification_settings.*", "NotificationRule").
		Where("notification_settings.pipeline_type = ?", req.PipelineType).
		Where("notification_settings.event_type_id = ?", eventId).
		WhereGroup(func(query *orm.Query) (*orm.Query, error) {
			query = query.
				Where("notification_settings.app_id = ?", req.AppId).
				WhereOr("notification_settings.env_id IS NULL").
				WhereOr("notification_settings.team_id IS NULL").
				WhereOr("notification_settings.pipeline_id IS NULL").
				WhereOrGroup(func(query *orm.Query) (*orm.Query, error) {
					query = query.
						Where("notification_settings.app_id IS NULL").
						WhereOr("notification_settings.env_id = ?", req.EnvId).
						WhereOr("notification_settings.team_id IS NULL").
						WhereOr("notification_settings.pipeline_id IS NULL").
						WhereOrGroup(func(query *orm.Query) (*orm.Query, error) {
							query = query.
								Where("notification_settings.app_id IS NULL").
								WhereOr("notification_settings.env_id IS NULL").
								WhereOr("notification_settings.team_id = ?", req.TeamId).
								WhereOr("notification_settings.pipeline_id IS NULL").
								WhereOrGroup(func(query *orm.Query) (*orm.Query, error) {
									query = query.
										Where("notification_settings.app_id IS NULL").
										WhereOr("notification_settings.env_id IS NULL").
										WhereOr("notification_settings.team_id IS NULL").
										WhereOr("notification_settings.pipeline_id = ?", req.PipelineId).
										WhereOrGroup(func(query *orm.Query) (*orm.Query, error) {
											query = query.
												Where("notification_settings.app_id IS NULL").
												WhereOr("notification_settings.env_id = ?", req.EnvId).
												WhereOr("notification_settings.team_id = ?", req.TeamId).
												WhereOr("notification_settings.pipeline_id IS NULL").
												WhereOrGroup(func(query *orm.Query) (*orm.Query, error) {
													query = query.
														Where("notification_settings.app_id = ?", req.AppId).
														WhereOr("notification_settings.env_id IS NULL").
														WhereOr("notification_settings.team_id = ?", req.TeamId).
														WhereOr("notification_settings.pipeline_id IS NULL").
														WhereOrGroup(func(query *orm.Query) (*orm.Query, error) {
															query = query.
																Where("notification_settings.app_id = ?", req.AppId).
																WhereOr("notification_settings.env_id = ?", req.EnvId).
																WhereOr("notification_settings.team_id = ?", req.TeamId).
																WhereOr("notification_settings.pipeline_id = ?", req.PipelineId)
															return query, nil
														})
													return query, nil
												})
											return query, nil
										})
									return query, nil
								})
							return query, nil
						})
					return query, nil
				})
			return query, nil
		}).
		Select()
	if err != nil {
		return nil, err
	}
	return notificationSettings, nil
}
