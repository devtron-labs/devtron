/*
 * Copyright (c) 2024. Devtron Inc.
 */

package pipelineConfig

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindAppAndEnvDetailsByListFilter(t *testing.T) {
	dbConfig, err := sql.GetConfig()
	if err != nil {
		t.Fail()
	}
	logger, _ := util.NewSugardLogger()
	dbConnection, err := sql.NewDbConnection(dbConfig, logger)
	if err != nil {
		t.SkipNow()
	}
	pipelineRepository := NewPipelineRepositoryImpl(dbConnection, logger)

	t.Run("empty list filter", func(tt *testing.T) {
		filter := CdPipelineListFilter{}
		_, err = pipelineRepository.FindAppAndEnvDetailsByListFilter(filter)
		if err != nil {
			logger.Infow("log : ", "filter", filter, "query", findAppAndEnvDetailsByListFilterQuery(filter))
		}
		assert.Nil(tt, err)
	})

	t.Run("with IncludeAppEnvIds only", func(tt *testing.T) {
		filter := CdPipelineListFilter{
			IncludeAppEnvIds: []string{"1,2", "1,3"},
		}

		_, err = pipelineRepository.FindAppAndEnvDetailsByListFilter(filter)
		if err != nil {
			logger.Infow("log : ", "filter", filter, "query", findAppAndEnvDetailsByListFilterQuery(filter))
		}
		assert.Nil(tt, err)
	})

	t.Run("with ExcludeAppEnvIds", func(tt *testing.T) {
		filter := CdPipelineListFilter{
			ExcludeAppEnvIds: []string{"1,2", "1,3"},
		}

		_, err = pipelineRepository.FindAppAndEnvDetailsByListFilter(filter)
		if err != nil {
			logger.Infow("log : ", "filter", filter, "query", findAppAndEnvDetailsByListFilterQuery(filter))
		}
		assert.Nil(tt, err)
	})

	t.Run("with both ExcludeAppEnvIds and IncludeAppEnvIds", func(tt *testing.T) {
		filter := CdPipelineListFilter{
			ExcludeAppEnvIds: []string{"1,2", "1,3"},
			IncludeAppEnvIds: []string{"1,2", "1,3"},
		}

		_, err = pipelineRepository.FindAppAndEnvDetailsByListFilter(filter)
		if err != nil {
			logger.Infow("log : ", "filter", filter, "query", findAppAndEnvDetailsByListFilterQuery(filter))
		}
		assert.Nil(tt, err)
	})

	t.Run("with AppNames", func(tt *testing.T) {
		filter := CdPipelineListFilter{
			AppNames: []string{"app1"},
		}

		_, err = pipelineRepository.FindAppAndEnvDetailsByListFilter(filter)
		if err != nil {
			logger.Infow("log : ", "filter", filter, "query", findAppAndEnvDetailsByListFilterQuery(filter))
		}
		assert.Nil(tt, err)
	})

	t.Run("with EnvNames", func(tt *testing.T) {
		filter := CdPipelineListFilter{
			EnvNames: []string{"env1", "env2"},
		}

		_, err = pipelineRepository.FindAppAndEnvDetailsByListFilter(filter)
		if err != nil {
			logger.Infow("log : ", "filter", filter, "query", findAppAndEnvDetailsByListFilterQuery(filter))
		}
		assert.Nil(tt, err)
	})

	t.Run("with AppNames,EnvNames", func(tt *testing.T) {
		filter := CdPipelineListFilter{
			AppNames: []string{"app1"},
			EnvNames: []string{"env1", "env2"},
		}

		_, err = pipelineRepository.FindAppAndEnvDetailsByListFilter(filter)
		if err != nil {
			logger.Infow("log : ", "filter", filter, "query", findAppAndEnvDetailsByListFilterQuery(filter))
		}
		assert.Nil(tt, err)
	})

	t.Run("with AppNames,EnvNames,IncludeAppEnvIds,ExcludeAppEnvIds ", func(tt *testing.T) {
		filter := CdPipelineListFilter{
			AppNames:         []string{"app1"},
			EnvNames:         []string{"env1", "env2"},
			ExcludeAppEnvIds: []string{"1,2", "1,3"},
			IncludeAppEnvIds: []string{"1,2", "1,3"},
		}

		_, err = pipelineRepository.FindAppAndEnvDetailsByListFilter(filter)
		if err != nil {
			logger.Infow("log : ", "filter", filter, "query", findAppAndEnvDetailsByListFilterQuery(filter))
		}
		assert.Nil(tt, err)
	})

	t.Run("filter with all options", func(tt *testing.T) {
		filter := CdPipelineListFilter{
			AppNames:         []string{"app1"},
			EnvNames:         []string{"env1", "env2"},
			ExcludeAppEnvIds: []string{"1,2", "1,3"},
			IncludeAppEnvIds: []string{"1,2", "1,3"},
			ListingFilterOptions: util2.ListingFilterOptions{
				Order:  "DESC",
				SortBy: "envName",
				Limit:  10,
				Offset: 1,
			},
		}

		_, err = pipelineRepository.FindAppAndEnvDetailsByListFilter(filter)
		if err != nil {
			logger.Infow("log : ", "filter", filter, "query", findAppAndEnvDetailsByListFilterQuery(filter))
		}
		assert.Nil(tt, err)
	})

}
