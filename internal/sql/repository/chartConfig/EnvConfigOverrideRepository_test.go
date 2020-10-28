/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package chartConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func getEcr() EnvConfigOverrideRepository {
	return nil
	//return NewEnvConfigOverrideRepository(models.GetDbConnection())
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
		AuditLog:          models.AuditLog{CreatedBy: 1, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: 1},
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
