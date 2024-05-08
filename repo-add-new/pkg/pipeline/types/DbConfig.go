package types

type DbMigrationConfigBean struct {
	Id            int    `json:"id"`
	DbConfigId    int    `json:"dbConfigId"`
	PipelineId    int    `json:"pipelineId"`
	GitMaterialId int    `json:"gitMaterialId"`
	ScriptSource  string `json:"scriptSource"` //location of file in git. relative to git root
	MigrationTool string `json:"migrationTool"`
	Active        bool   `json:"active"`
	UserId        int32  `json:"-"`
}

type DbConfigBean struct {
	Id       int    `json:"id,omitempty" validate:"number"`
	Name     string `json:"name,omitempty" validate:"required"` //name by which user identifies this db
	Type     string `json:"type,omitempty" validate:"required"` //type of db, PG, MYsql, MariaDb
	Host     string `json:"host,omitempty" validate:"host"`
	Port     string `json:"port,omitempty" validate:"max=4"`
	DbName   string `json:"dbName,omitempty" validate:"required"` //name of database inside PG
	UserName string `json:"userName,omitempty"`
	Password string `json:"password,omitempty"`
	Active   bool   `json:"active,omitempty"`
	UserId   int32  `json:"-"`
}
