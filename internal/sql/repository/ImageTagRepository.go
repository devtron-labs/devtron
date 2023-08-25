package repository

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ImageTagRepository interface {
}

type ImageTagRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewImageTagRepository(dbConnection *pg.DB, logger *zap.SugaredLogger) *ImageTagRepositoryImpl {
	return &ImageTagRepositoryImpl{dbConnection: dbConnection, logger: logger}
}
