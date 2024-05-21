package bean

import "github.com/devtron-labs/devtron/pkg/devtronResource/bean"

type GetQueryParams struct {
	Id         int    `schema:"id"`
	Identifier string `schema:"identifier"`
}

type QueryParams interface {
	GetQueryParams | GetResourceQueryParams | GetDependencyQueryParams | GetConfigOptionsQueryParams | GetTaskRunInfoQueryParams
}

type GetResourceQueryParams struct {
	GetQueryParams
	Component []bean.DevtronResourceUIComponent `schema:"component"`
}

type GetDependencyQueryParams struct {
	GetQueryParams
	IsLite           bool     `schema:"lite"`
	DependenciesInfo []string `schema:"dependencyInfo"`
}

type GetTaskRunInfoQueryParams struct {
	GetQueryParams
	IsLite     bool `schema:"lite"`
	LevelIndex int  `schema:"levelIndex"`
	ShowAll    bool `schema:"showAll"`
}

type ConfigOptionType = string

type GetConfigOptionsQueryParams struct {
	GetQueryParams
	DependenciesInfo []string         `schema:"dependencyInfo,required"`
	ConfigOption     ConfigOptionType `schema:"configOption"`
	FilterCriteria   []string         `schema:"filterCriteria"`
	SearchKey        string           `schema:"searchKey"`
	Limit            int              `schema:"limit"`
	Offset           int              `schema:"offset"`
}

const (
	ArtifactConfig ConfigOptionType = "artifact"
	CommitConfig   ConfigOptionType = "commit"
)

type GetResourceListQueryParams struct {
	IsLite         bool     `schema:"lite"`
	FetchChild     bool     `schema:"fetchChild"`
	FilterCriteria []string `schema:"filterCriteria"`
}

type GetHistoryQueryParams struct {
	FilterCriteria []string `schema:"filterCriteria"`
	OffSet         int      `schema:"offSet"`
	Limit          int      `schema:"limit"`
}

type GetHistoryConfigQueryParams struct {
	BaseConfigurationId  int      `schema:"baseConfigurationId"`
	HistoryComponent     string   `schema:"historyComponent"`
	HistoryComponentName string   `schema:"historyComponentName"`
	FilterCriteria       []string `schema:"filterCriteria"`
}

const (
	RequestInvalidKindVersionErrMessage = "Invalid kind and version! Implementation not supported."
	PathParamKind                       = "kind"
	PathParamVersion                    = "version"
	QueryParamIsExposed                 = "onlyIsExposed"
	QueryParamLite                      = "lite"
	QueryParamIdentifier                = "identifier"
	QueryParamFetchChild                = "fetchChild"
	QueryParamId                        = "id"
	QueryParamName                      = "name"
	QueryParamComponent                 = "component"
	ResourceUpdateSuccessMessage        = "Resource object updated successfully."
	ResourceCreateSuccessMessage        = "Resource object created successfully."
	ResourceCloneSuccessMessage         = "Resource object cloned successfully."
	DependenciesUpdateSuccessMessage    = "Resource dependencies updated successfully."
)
