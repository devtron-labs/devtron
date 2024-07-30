package bean

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
)
