package repository

type ImageTag struct {
	tableName  struct{} `sql:"release_tags" json:",omitempty"  pg:",discard_unknown_columns"`
	Id         int      `sql:"id"`
	AppId      int      `sql:"app_id"`
	ArtifactId int      `sql:"artifact_id"`
	Active     bool     `sql:"active"`
}

type ImageComment struct {
	tableName  struct{} `sql:"image_comments" json:",omitempty"  pg:",discard_unknown_columns"`
	Id         int      `sql:"id"`
	Comment    int      `sql:"app_id"`
	ArtifactId int      `sql:"artifact_id"`
}

type ImageTagAudit struct {
	tableName struct{} `sql:"release_tags" json:",omitempty"  pg:",discard_unknown_columns"`
	Id        int      `sql:"id"`
	Data      string   `sql:"data"`

	ArtifactId int  `sql:"artifact_id"`
	Active     bool `sql:"active"`
}
