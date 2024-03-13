package linkedCIView

type SortOrder string

const (
	Asc  SortOrder = "ASC"
	Desc SortOrder = "DESC"
)

type LinkedCIDetailsReq struct {
	SourceCIPipeline int       `json:"sourceCIPipeline"`
	Order            SortOrder `json:"order"`
	Search           string    `json:"search"`
	Offset           int       `json:"offset"`
	Size             int       `json:"size"`
	EnvName          string    `json:"envName"`
}
