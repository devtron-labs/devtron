/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pipeline

import (
	"encoding/json"
	"testing"

	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	historyRepository "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captures what would be written to ci_build_config
type capturingCiBuildConfigRepository struct {
	pipelineConfig.CiBuildConfigRepository
	saved   []*pipelineConfig.CiBuildConfig
	updated []*pipelineConfig.CiBuildConfig
}

func (r *capturingCiBuildConfigRepository) Save(c *pipelineConfig.CiBuildConfig) error {
	c.Id = 100 + len(r.saved)
	r.saved = append(r.saved, c)
	return nil
}

func (r *capturingCiBuildConfigRepository) Update(c *pipelineConfig.CiBuildConfig) error {
	r.updated = append(r.updated, c)
	return nil
}

type noopCiTemplateRepository struct {
	pipelineConfig.CiTemplateRepository
}

func (noopCiTemplateRepository) Save(*pipelineConfig.CiTemplate) error   { return nil }
func (noopCiTemplateRepository) Update(*pipelineConfig.CiTemplate) error { return nil }

type noopCiTemplateOverrideRepository struct {
	pipelineConfig.CiTemplateOverrideRepository
}

func (noopCiTemplateOverrideRepository) Save(o *pipelineConfig.CiTemplateOverride) (*pipelineConfig.CiTemplateOverride, error) {
	return o, nil
}
func (noopCiTemplateOverrideRepository) Update(o *pipelineConfig.CiTemplateOverride) (*pipelineConfig.CiTemplateOverride, error) {
	return o, nil
}

func newCiTemplateServiceForTest(t *testing.T) (*CiTemplateServiceImpl, *capturingCiBuildConfigRepository) {
	logger, err := util.NewSugardLogger()
	require.NoError(t, err)
	buildConfigRepo := &capturingCiBuildConfigRepository{}
	return NewCiTemplateServiceImpl(logger, NewCiBuildConfigServiceImpl(logger, buildConfigRepo),
		noopCiTemplateRepository{}, noopCiTemplateOverrideRepository{}), buildConfigRepo
}

// the #2983 request: the dashboard echoes the override's hidden args (and target platform) back on save
func echoedDockerOverride(buildType bean2.CiBuildType, id int) *bean2.CiBuildConfigBean {
	return &bean2.CiBuildConfigBean{
		Id:            id,
		GitMaterialId: 1, BuildContextGitMaterialId: 1,
		CiBuildType: buildType,
		DockerBuildConfig: &bean2.DockerBuildConfig{
			DockerfilePath:     "Dockerfile",
			Args:               map[string]string{"AWS_ACCESS_KEY_ID": "AKIA-stale", "AWS_SECRET_ACCESS_KEY": "stale-secret"},
			TargetPlatform:     "linux/arm64,linux/amd64",
			DockerBuildOptions: map[string]string{"network": "host"},
			BuildContext:       "ctx",
		},
	}
}

func decodeMetadata(t *testing.T, entity *pipelineConfig.CiBuildConfig) map[string]interface{} {
	m := map[string]interface{}{}
	require.NoError(t, json.Unmarshal([]byte(entity.BuildMetadata), &m))
	return m
}

func assertOverrideMetadataWithoutArgs(t *testing.T, entity *pipelineConfig.CiBuildConfig) {
	m := decodeMetadata(t, entity)
	assert.NotContains(t, m, "args")
	assert.NotContains(t, entity.BuildMetadata, "AWS_SECRET_ACCESS_KEY")
	// everything that IS part of the override is persisted as sent
	assert.Equal(t, "linux/arm64,linux/amd64", m["targetPlatform"])
	assert.Equal(t, "Dockerfile", m["dockerfileRelativePath"])
	assert.Equal(t, map[string]interface{}{"network": "host"}, m["dockerBuildOptions"])
	assert.Equal(t, "ctx", m["buildContext"])
}

func TestCiTemplateService_OverrideNeverPersistsDockerBuildArgs(t *testing.T) {
	for _, buildType := range []bean2.CiBuildType{bean2.SELF_DOCKERFILE_BUILD_TYPE, bean2.MANAGED_DOCKERFILE_BUILD_TYPE} {
		t.Run(string(buildType)+"/Save (pipeline create, clone, template, ci switch)", func(t *testing.T) {
			svc, repo := newCiTemplateServiceForTest(t)
			cfg := echoedDockerOverride(buildType, 0)
			override := &pipelineConfig.CiTemplateOverride{CiPipelineId: 5}
			require.NoError(t, svc.Save(&bean2.CiTemplateBean{CiTemplateOverride: override, CiBuildConfig: cfg, UserId: 1}))

			require.Len(t, repo.saved, 1)
			assertOverrideMetadataWithoutArgs(t, repo.saved[0])
			assert.Equal(t, string(buildType), repo.saved[0].Type)
			assert.Equal(t, repo.saved[0].Id, override.CiBuildConfigId)
			// the caller's bean is cleared too: PatchMaterialValue/CreateCiConf hand this same bean to the
			// pipeline history snapshot and return it in the API response
			assert.Nil(t, cfg.DockerBuildConfig.Args)
		})

		t.Run(string(buildType)+"/Update (Update Pipeline)", func(t *testing.T) {
			svc, repo := newCiTemplateServiceForTest(t)
			cfg := echoedDockerOverride(buildType, 41)
			require.NoError(t, svc.Update(&bean2.CiTemplateBean{CiTemplateOverride: &pipelineConfig.CiTemplateOverride{Id: 11, CiPipelineId: 5}, CiBuildConfig: cfg, UserId: 1}))

			require.Len(t, repo.updated, 1)
			assert.Empty(t, repo.saved)
			assert.Equal(t, 41, repo.updated[0].Id)
			assert.Equal(t, 11, repo.updated[0].CiTemplateOverrideId)
			assertOverrideMetadataWithoutArgs(t, repo.updated[0])
			assert.Nil(t, cfg.DockerBuildConfig.Args)
		})

		t.Run(string(buildType)+"/Update of an override without build config row saves a new one without args", func(t *testing.T) {
			svc, repo := newCiTemplateServiceForTest(t)
			cfg := echoedDockerOverride(buildType, 0)
			require.NoError(t, svc.Update(&bean2.CiTemplateBean{CiTemplateOverride: &pipelineConfig.CiTemplateOverride{Id: 11}, CiBuildConfig: cfg, UserId: 1}))
			require.Len(t, repo.saved, 1)
			assertOverrideMetadataWithoutArgs(t, repo.saved[0])
		})
	}

	t.Run("regression: buildpack override keeps its buildpack env args", func(t *testing.T) {
		for _, update := range []bool{false, true} {
			svc, repo := newCiTemplateServiceForTest(t)
			cfg := &bean2.CiBuildConfigBean{Id: 41, CiBuildType: bean2.BUILDPACK_BUILD_TYPE, BuildPackConfig: &bean2.BuildPackConfig{BuilderId: "b", Args: map[string]string{"BP_ENV": "x"}}}
			tb := &bean2.CiTemplateBean{CiTemplateOverride: &pipelineConfig.CiTemplateOverride{Id: 11}, CiBuildConfig: cfg, UserId: 1}
			var entity *pipelineConfig.CiBuildConfig
			if update {
				require.NoError(t, svc.Update(tb))
				entity = repo.updated[0]
			} else {
				require.NoError(t, svc.Save(tb))
				entity = repo.saved[0]
			}
			assert.Equal(t, map[string]interface{}{"BP_ENV": "x"}, decodeMetadata(t, entity)["args"])
		}
	})

	t.Run("regression: ci job (skip build) override", func(t *testing.T) {
		svc, repo := newCiTemplateServiceForTest(t)
		cfg := &bean2.CiBuildConfigBean{CiBuildType: bean2.SKIP_BUILD_TYPE}
		require.NoError(t, svc.Save(&bean2.CiTemplateBean{CiTemplateOverride: &pipelineConfig.CiTemplateOverride{}, CiBuildConfig: cfg, UserId: 1}))
		require.Len(t, repo.saved, 1)
		assert.Equal(t, "", repo.saved[0].BuildMetadata)
	})
}

// Regression: the app level build config (ci_template) owns the args, Save/Update must keep them.
func TestCiTemplateService_AppLevelKeepsDockerBuildArgs(t *testing.T) {
	appLevel := func(id int) *bean2.CiBuildConfigBean {
		return &bean2.CiBuildConfigBean{Id: id, CiBuildType: bean2.SELF_DOCKERFILE_BUILD_TYPE, DockerBuildConfig: &bean2.DockerBuildConfig{
			DockerfilePath: "Dockerfile", Args: map[string]string{"APP": "1"}, TargetPlatform: "linux/amd64",
		}}
	}
	t.Run("Save", func(t *testing.T) {
		svc, repo := newCiTemplateServiceForTest(t)
		cfg := appLevel(0)
		require.NoError(t, svc.Save(&bean2.CiTemplateBean{CiTemplate: &pipelineConfig.CiTemplate{AppId: 1}, CiBuildConfig: cfg, UserId: 1}))
		assert.Equal(t, map[string]interface{}{"APP": "1"}, decodeMetadata(t, repo.saved[0])["args"])
		assert.Equal(t, map[string]string{"APP": "1"}, cfg.DockerBuildConfig.Args)
	})
	t.Run("Update", func(t *testing.T) {
		svc, repo := newCiTemplateServiceForTest(t)
		cfg := appLevel(9)
		require.NoError(t, svc.Update(&bean2.CiTemplateBean{CiTemplate: &pipelineConfig.CiTemplate{Id: 3, AppId: 1}, CiBuildConfig: cfg, UserId: 1}))
		assert.Equal(t, 3, repo.updated[0].CiTemplateId)
		assert.Equal(t, map[string]interface{}{"APP": "1"}, decodeMetadata(t, repo.updated[0])["args"])
		assert.Equal(t, map[string]string{"APP": "1"}, cfg.DockerBuildConfig.Args)
	})
}

type capturingCiPipelineHistoryRepository struct {
	historyRepository.CiPipelineHistoryRepository
	saved []*historyRepository.CiPipelineHistory
}

func (r *capturingCiPipelineHistoryRepository) Save(h *historyRepository.CiPipelineHistory) error {
	r.saved = append(r.saved, h)
	return nil
}

type noCiEnvMappingRepository struct {
	pipelineConfig.CiPipelineRepository
}

func (noCiEnvMappingRepository) FindCiEnvMappingByCiPipelineId(int) (*pipelineConfig.CiEnvMapping, error) {
	return &pipelineConfig.CiEnvMapping{}, pg.ErrNoRows
}

// PatchMaterialValue saves the override and then snapshots the very same bean into ci_pipeline_history;
// the snapshot must not keep a copy of the args either.
func TestCiTemplateService_OverrideHistorySnapshotHasNoDockerBuildArgs(t *testing.T) {
	logger, err := util.NewSugardLogger()
	require.NoError(t, err)
	svc, _ := newCiTemplateServiceForTest(t)
	historyRepo := &capturingCiPipelineHistoryRepository{}
	historySvc := history.NewCiPipelineHistoryServiceImpl(historyRepo, logger, noCiEnvMappingRepository{})

	tb := &bean2.CiTemplateBean{
		CiTemplateOverride: &pipelineConfig.CiTemplateOverride{Id: 11, CiPipelineId: 5, Active: true},
		CiBuildConfig:      echoedDockerOverride(bean2.SELF_DOCKERFILE_BUILD_TYPE, 41),
		UserId:             1,
	}
	require.NoError(t, svc.Update(tb))
	require.NoError(t, historySvc.SaveHistory(&pipelineConfig.CiPipeline{Id: 5, IsDockerConfigOverridden: true}, nil, tb, historyRepository.TRIGGER_UPDATE))

	require.Len(t, historyRepo.saved, 1)
	snapshot := historyRepository.CiPipelineTemplateOverrideHistoryDTO{}
	require.NoError(t, json.Unmarshal([]byte(historyRepo.saved[0].CiTemplateOverrideHistory), &snapshot))
	metadata := map[string]interface{}{}
	require.NoError(t, json.Unmarshal([]byte(snapshot.BuildMetadata), &metadata))
	assert.NotContains(t, metadata, "args")
	assert.Equal(t, "linux/arm64,linux/amd64", metadata["targetPlatform"])
}
