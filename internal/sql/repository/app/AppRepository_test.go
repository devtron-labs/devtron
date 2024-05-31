/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package app

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func getPipelineGroupRepo() *AppRepositoryImpl {
	return nil
	//return NewAppRepositoryImpl(models.NewDbConnection())
}

func TestPipelineGroupRepositoryImpl_FindActiveByName(t *testing.T) {
	pg, err := getPipelineGroupRepo().FindActiveByName("ke")
	assert.NoError(t, err)
	assert.NotNil(t, pg)
}
