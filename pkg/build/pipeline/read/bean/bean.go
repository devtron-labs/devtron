package bean

type CiPipelineMin struct {
	Id               int
	Name             string
	AppId            int
	TeamId           int
	ParentCiPipeline int
	CiPipelineType   string
}
