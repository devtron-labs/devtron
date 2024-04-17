package bean

import "github.com/devtron-labs/devtron/pkg/devtronResource/bean"

type GetResourceQueryParams struct {
	Id         int                               `schema:"id"`
	Name       string                            `schema:"name"`
	Component  []bean.DevtronResourceUIComponent `schema:"component"`
	Identifier string                            `schema:"identifier"`
}

type GetResourceListQueryParams struct {
	IsLite     bool `schema:"lite"`
	FetchChild bool `schema:"fetchChild"`
}

const (
	PathParamKind                    = "kind"
	PathParamVersion                 = "version"
	QueryParamIsExposed              = "onlyIsExposed"
	QueryParamLite                   = "lite"
	QueryParamIdentifier             = "identifier"
	QueryParamFetchChild             = "fetchChild"
	QueryParamId                     = "id"
	QueryParamName                   = "name"
	QueryParamComponent              = "component"
	ResourceUpdateSuccessMessage     = "Resource object updated successfully."
	ResourceCreateSuccessMessage     = "Resource object created successfully."
	DependenciesUpdateSuccessMessage = "Resource dependencies updated successfully."
)
