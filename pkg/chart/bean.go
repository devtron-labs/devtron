package chart

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/models"
)

var ReservedChartRefNamesList *[]ReservedChartList

type ReservedChartList struct {
	LocationPrefix string
	Name           string
	hjbjhjk
}

type TemplateRequest struct {
	Id                      int                         `json:"id"  validate:"number"`
	AppId                   int                         `json:"appId,omitempty"  validate:"number,required"`
	RefChartTemplate        string                      `json:"refChartTemplate,omitempty"`
	RefChartTemplateVersion string                      `json:"refChartTemplateVersion,omitempty"`
	ChartRepositoryId       int                         `json:"chartRepositoryId,omitempty"`
	ValuesOverride          json.RawMessage             `json:"valuesOverride,omitempty" validate:"required"` //json format user value
	DefaultAppOverride      json.RawMessage             `json:"defaultAppOverride,omitempty"`                 //override values available
	ChartRefId              int                         `json:"chartRefId,omitempty"  validate:"number"`
	Latest                  bool                        `json:"latest"`
	IsAppMetricsEnabled     bool                        `json:"isAppMetricsEnabled"`
	Schema                  json.RawMessage             `json:"schema"`
	Readme                  string                      `json:"readme"`
	IsBasicViewLocked       bool                        `json:"isBasicViewLocked"`
	CurrentViewEditor       models.ChartsViewEditorType `json:"currentViewEditor"` //default "UNDEFINED" in db
	UserId                  int32                       `json:"-"`
}

type ChartUpgradeRequest struct {
	ChartRefId int   `json:"chartRefId"  validate:"number"`
	All        bool  `json:"all"`
	AppIds     []int `json:"appIds"`
	UserId     int32 `json:"-"`
}

type ChartRefChangeRequest struct {
	AppId            int `json:"appId" validate:"required"`
	EnvId            int `json:"envId" validate:"required"`
	TargetChartRefId int `json:"targetChartRefId" validate:"required"`
}

type PipelineConfigRequest struct {
	Id                   int             `json:"id"  validate:"number"`
	AppId                int             `json:"appId,omitempty"  validate:"number,required"`
	EnvConfigOverrideId  int             `json:"envConfigOverrideId,omitempty"`
	PipelineConfigValues json.RawMessage `json:"pipelineConfigValues,omitempty" validate:"required"` //json format user value
	PipelineId           int             `json:"PipelineId,omitempty"`
	Latest               bool            `json:"latest"`
	Previous             bool            `json:"previous"`
	EnvId                int             `json:"envId,omitempty"`
	ManualReviewed       bool            `json:"manualReviewed" validate:"required"`
	UserId               int32           `json:"-"`
}

type PipelineConfigRequestResponse struct {
	LatestPipelineConfigRequest    PipelineConfigRequest `json:"latestPipelineConfigRequest"`
	PreviousPipelineConfigRequest  PipelineConfigRequest `json:"previousPipelineConfigRequest"`
	PreviousPipelineConfigRequests PipelineConfigRequest `json:"previousPipelineConfigRequests"`
}

type AppConfigResponse struct {
	//DefaultAppConfig  json.RawMessage `json:"defaultAppConfig"`
	//AppConfig         TemplateRequest            `json:"appConfig"`
	LatestAppConfig   TemplateRequest `json:"latestAppConfig"`
	PreviousAppConfig TemplateRequest `json:"previousAppConfig"`
}

type DefaultChart string
