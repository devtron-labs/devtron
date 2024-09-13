package repository

import (
	"github.com/go-pg/pg"
)

type GenericNoteFileBasedRepositoryImpl struct {
}

func NewGenericNoteFileBasedRepository() *GenericNoteFileBasedRepositoryImpl {
	return &GenericNoteFileBasedRepositoryImpl{}
}

func (impl GenericNoteFileBasedRepositoryImpl) StartTx() (*pg.Tx, error) {
	//TODO implement me
	panic("implement me")
}

func (impl GenericNoteFileBasedRepositoryImpl) RollbackTx(tx *pg.Tx) error {
	//TODO implement me
	panic("implement me")
}

func (impl GenericNoteFileBasedRepositoryImpl) CommitTx(tx *pg.Tx) error {
	//TODO implement me
	panic("implement me")
}

func (impl GenericNoteFileBasedRepositoryImpl) Save(tx *pg.Tx, model *GenericNote) error {
	//TODO implement me
	panic("implement me")
}

func (impl GenericNoteFileBasedRepositoryImpl) FindByClusterId(id int) (*GenericNote, error) {
	//TODO implement me
	panic("implement me")
}

func (impl GenericNoteFileBasedRepositoryImpl) FindByAppId(id int) (*GenericNote, error) {
	//TODO implement me
	panic("implement me")
}

func (impl GenericNoteFileBasedRepositoryImpl) FindByIdentifier(identifier int, identifierType NoteType) (*GenericNote, error) {
	//TODO implement me
	panic("implement me")
}

func (impl GenericNoteFileBasedRepositoryImpl) Update(tx *pg.Tx, model *GenericNote) error {
	//TODO implement me
	panic("implement me")
}

func (impl GenericNoteFileBasedRepositoryImpl) GetGenericNotesForAppIds(appIds []int) ([]*GenericNote, error) {
	//TODO implement me
	panic("implement me")
}

func (impl GenericNoteFileBasedRepositoryImpl) GetDescriptionFromAppIds(appIds []int) ([]*GenericNote, error) {
	//TODO implement me
	panic("implement me")
}
