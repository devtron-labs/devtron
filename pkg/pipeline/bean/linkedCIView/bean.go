package linkedCIView

type LinkedCiInfoFilters struct {
	PaginationQueryParams
	EnvName string `json:"envName"`
}

type LinkedCIDetailsRes struct {
	AppName          string `json:"appName"`
	AppId            int    `json:"appId"`
	EnvironmentName  string `json:"environmentName"`
	EnvironmentId    int    `json:"environmentId"`
	TriggerMode      string `json:"triggerMode"`
	DeploymentStatus string `json:"deploymentStatus"`
}

type SortOrder string

const (
	Asc  SortOrder = "ASC"
	Desc SortOrder = "DESC"
)

type PaginationQueryParams struct {
	Order     SortOrder `json:"order"`
	Offset    int       `json:"offset"`
	Size      int       `json:"size"`
	SearchKey string    `json:"searchKey"`
}

type PaginatedResponse[T any] struct {
	TotalCount int `json:"totalCount"` // Total results count
	Offset     int `json:"offset"`     // Current page number
	Size       int `json:"size"`       // Current page size
	Data       []T `json:"data"`
}

// PushData will append item to the PaginatedResponse.Data
func (m *PaginatedResponse[T]) PushData(item ...T) {
	m.Data = append(m.Data, item...)
}

// UpdateTotalCount will update the TotalCount in PaginatedResponse
func (m *PaginatedResponse[_]) UpdateTotalCount(totalCount int) { // not using the type param in this method
	m.TotalCount = totalCount
}

// UpdateOffset will update the Offset in PaginatedResponse
func (m *PaginatedResponse[_]) UpdateOffset(offset int) { // not using the type param in this method
	m.Offset = offset
}

// UpdateSize will update the Size in PaginatedResponse
func (m *PaginatedResponse[_]) UpdateSize(size int) { // not using the type param in this method
	m.Size = size
}
