package repository

import "github.com/devtron-labs/devtron/pkg/bean"

type CiPipelineHistory struct {
	tableName        struct{} `sql:"ci_template_history" pg:",discard_unknown_columns"`
	Id               int      `sql:"id"`
	CiPipelineId     int      `sql:"ci_pipeline_id"`
	DockerRegistryId string   `sql:"docker_registry_id"`
	DockerRepository string   `sql:"docker_repository"`
	DockerfilePath   string   `sql:"dockerfile_path"`
	Active           bool     `sql:"active,notnull"`
	CiBuildConfigId  int      `sql:"ci_build_config_id"`
	GitMaterialId    int      `sql:"git_material_id"` //id stored in db GitMaterial( foreign key)
	Path             string   `sql:"path"`            // defaults to root of git repo
	//depricated was used in gocd remove this
	CheckoutPath string          `sql:"checkout_path"` //path where code will be checked out for single source `./` default for multiSource configured by user
	Type         bean.SourceType `sql:"type"`
	Value        string          `sql:"value"`
	ScmId        string          `sql:"scm_id"`      //id of gocd object
	ScmName      string          `sql:"scm_name"`    //gocd scm name
	ScmVersion   string          `sql:"scm_version"` //gocd scm version
	Regex        string          `json:"regex"`
	GitTag       string          `sql:"-"`
}
