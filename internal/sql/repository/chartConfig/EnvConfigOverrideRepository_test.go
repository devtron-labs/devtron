/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package chartConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func getEcr() EnvConfigOverrideRepository {
	return nil
	//return NewEnvConfigOverrideRepository(models.GetDbTransaction())
}

func TestEnvConfigOverrideRepositoryImpl_Save(t *testing.T) {
	eco := &EnvConfigOverride{
		ChartId: 2,
		//	TargetEnvironment: "test",
		Status: models.CHARTSTATUS_NEW,
		//	EnvMergedValues:   `{"blue":{"enabled":true},"green":{"enabled":true},"image":{"tag":2},"ingress":{"enabled":true},"productionSlot":"blue","replicaCount":1}`,
		EnvOverrideValues: "{}",
		ManualReviewed:    false,
		Active:            true,
		AuditLog:          sql.AuditLog{CreatedBy: 1, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: 1},
	}
	repo := getEcr()
	err := repo.Save(eco)
	assert.NoError(t, err)
}

func TestEnvConfigOverrideRepositoryImpl_Get(t *testing.T) {
	eco, err := getEcr().Get(3)
	assert.NoError(t, err)
	assert.NotNil(t, eco)
}

func TestEnvConfigOverrideRepositoryImpl_ActiveEnvConfigOverride(t *testing.T) {
	/*eco, err := getEcr().ActiveEnvConfigOverride("nginx-ingress", "test")
	assert.NoError(t, err)
	assert.NotNil(t, eco)*/
}

func TestEnvConfigOverrideRepositoryImpl_GetByChartAndEnvironment(t *testing.T) {
	/*conf, err:=getEcr().GetByChartAndEnvironment(&Chart{Id:24},"test")

	b:=errors.IsNotFound(err)
	fmt.Println(b)
	assert.NoError(t, err)
	assert.NotNil(t, conf)*/
}
