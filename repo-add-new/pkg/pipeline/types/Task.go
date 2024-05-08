package types

type TaskYaml struct {
	Version          string             `yaml:"version"`
	CdPipelineConfig []CdPipelineConfig `yaml:"cdPipelineConf"`
}

type CdPipelineConfig struct {
	BeforeTasks []*Task `yaml:"beforeStages"`
	AfterTasks  []*Task `yaml:"afterStages"`
}

type Task struct {
	Id             int    `json:"id,omitempty"`
	Index          int    `json:"index,omitempty"`
	Name           string `json:"name" yaml:"name"`
	Script         string `json:"script" yaml:"script"`
	OutputLocation string `json:"outputLocation" yaml:"outputLocation"` // file/dir
	RunStatus      bool   `json:"-,omitempty"`                          // task run was attempted or not
}
