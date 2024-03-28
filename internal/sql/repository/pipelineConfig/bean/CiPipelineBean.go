package bean

type LinkedCIDetails struct {
	AppName         string `sql:"app_name"`
	EnvironmentName string `sql:"environment_name"`
	TriggerMode     string `sql:"trigger_mode"`
	PipelineId      int    `sql:"pipeline_id"`
	AppId           int    `sql:"app_id"`
	EnvironmentId   int    `sql:"environment_id"`
}

type CiPipelinesMap struct {
	Id               int `json:"id"`
	ParentCiPipeline int `json:"parentCiPipeline"`
}
