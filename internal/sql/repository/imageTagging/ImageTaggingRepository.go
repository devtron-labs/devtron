package repository

import (
	"github.com/go-pg/pg"
	"time"
)

type ImageTaggingAction int

//this action is only allowed for imageComments
const ActionEdit ImageTaggingAction = 1

const ActionSave ImageTaggingAction = 0
const ActionHardDelete ImageTaggingAction = 3
const ActionSoftDelete = 2

type ImageTag struct {
	TableName  struct{} `sql:"release_tags" json:",omitempty"  pg:",discard_unknown_columns"`
	Id         int      `sql:"id" json:"id"`
	AppId      int      `sql:"app_id" json:"appId"`
	ArtifactId int      `sql:"artifact_id" json:"artifactId"`
	Active     bool     `sql:"active" json:"active"`
}

type ImageComment struct {
	TableName  struct{} `sql:"image_comments" json:",omitempty"  pg:",discard_unknown_columns"`
	Id         int      `sql:"id" json:"id"`
	Comment    int      `sql:"app_id" json:"comment"`
	ArtifactId int      `sql:"artifact_id" json:"artifactId"`
	UserId     int      `sql:"user_id" json:"-"` //currently not sending userId in json response
}

type ImageTaggingAudit struct {
	TableName  struct{}           `sql:"release_tags" json:",omitempty"  pg:",discard_unknown_columns"`
	Id         int                `sql:"id"`
	Data       string             `sql:"data"`
	DataType   int                `sql:"data_type"`
	ArtifactId int                `sql:"artifact_id"`
	UpdatedOn  time.Time          `sql:"updated_on"`
	UpdatedBy  time.Time          `sql:"updated_by"`
	Action     ImageTaggingAction `sql:"action"`
}

type ImageTaggingRepository interface {
	SaveReleaseTag(tx *pg.Tx, imageTag *ImageTag) error
	SaveImageComment(tx *pg.Tx, imageComment *ImageComment) error
	GetTagsByAppId(appId int) ([]ImageTag, error)
	GetTagsByArtifactId(artifactId int) ([]ImageTag, error)
	GetImageComment(artifactId int) (ImageComment, error)
	UpdateReleaseTag(tx *pg.Tx, imageTag *ImageTag) error
	UpdateImageComment(tx *pg.Tx, imageComment *ImageComment) error
	DeleteReleaseTag(tx *pg.Tx, imageTag *ImageTag) error
}

type ImageTaggingRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewImageTaggingRepositoryImpl(db *pg.DB) *ImageTaggingRepositoryImpl {
	return &ImageTaggingRepositoryImpl{
		dbConnection: db,
	}
}

func (impl *ImageTaggingRepositoryImpl) SaveReleaseTag(tx *pg.Tx, imageTag *ImageTag) error {
	err := tx.Insert(imageTag)
	if err != nil {
		err = tx.Insert(tx, &ImageTaggingAudit{
			Action: ActionSave,
		})
	}
	return err
}

func (impl *ImageTaggingRepositoryImpl) SaveImageComment(tx *pg.Tx, imageComment *ImageComment) error {
	err := tx.Insert(imageComment)
	if err != nil {
		err = tx.Insert(tx, &ImageTaggingAudit{})
	}
	return err
}

func (impl *ImageTaggingRepositoryImpl) GetTagsByAppId(appId int) ([]ImageTag, error) {
	res := make([]ImageTag, 0)
	err := impl.dbConnection.Model(&res).
		Where("app_id=?", appId).
		Select()
	return res, err
}

func (impl *ImageTaggingRepositoryImpl) GetTagsByArtifactId(artifactId int) ([]ImageTag, error) {
	res := make([]ImageTag, 0)
	err := impl.dbConnection.Model(&res).
		Where("artifact_id=?", artifactId).
		Select()
	return res, err
}

func (impl *ImageTaggingRepositoryImpl) GetImageComment(artifactId int) (ImageComment, error) {
	res := ImageComment{}
	err := impl.dbConnection.Model(&res).
		Where("artifact_id=?", artifactId).
		Select()
	return res, err
}

//this will update the provided release tag
func (impl *ImageTaggingRepositoryImpl) UpdateReleaseTag(tx *pg.Tx, imageTag *ImageTag) error {
	//currently tags are not editable, can only be soft deleted or hard delete
	err := tx.Update(imageTag)
	if err != nil {
		err = tx.Insert(tx, &ImageTaggingAudit{
			Action: ActionSoftDelete,
		})
	}
	return err
}
func (impl *ImageTaggingRepositoryImpl) UpdateImageComment(tx *pg.Tx, imageComment *ImageComment) error {
	err := tx.Update(imageComment)
	if err != nil {
		err = tx.Insert(tx, &ImageTaggingAudit{
			Action: ActionEdit,
		})
	}
	return err
}

func (impl *ImageTaggingRepositoryImpl) DeleteReleaseTag(tx *pg.Tx, imageTag *ImageTag) error {
	err := tx.Delete(imageTag)
	if err != nil {
		err = tx.Insert(tx, &ImageTaggingAudit{
			Action: ActionHardDelete,
		})
	}
	return err
}
