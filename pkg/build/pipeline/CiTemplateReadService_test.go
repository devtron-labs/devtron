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
	"testing"

	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCiTemplateOverrideRepository struct {
	pipelineConfig.CiTemplateOverrideRepository
	overrides []*pipelineConfig.CiTemplateOverride
}

func (f *fakeCiTemplateOverrideRepository) FindByAppId(int) ([]*pipelineConfig.CiTemplateOverride, error) {
	return f.overrides, nil
}

func (f *fakeCiTemplateOverrideRepository) FindByCiPipelineIds([]int) ([]*pipelineConfig.CiTemplateOverride, error) {
	return f.overrides, nil
}

func (f *fakeCiTemplateOverrideRepository) FindByCiPipelineId(int) (*pipelineConfig.CiTemplateOverride, error) {
	return f.overrides[0], nil
}

type fakeCiTemplateRepository struct {
	pipelineConfig.CiTemplateRepository
	template *pipelineConfig.CiTemplate
}

func (f *fakeCiTemplateRepository) FindByAppId(int) (*pipelineConfig.CiTemplate, error) {
	return f.template, nil
}

func (f *fakeCiTemplateRepository) FindByAppIds([]int) ([]*pipelineConfig.CiTemplate, error) {
	return []*pipelineConfig.CiTemplate{f.template}, nil
}

// the #2983 override row: args and target platform the editor showed as empty
const staleDockerOverrideMetadata = `{"dockerfileRelativePath":"Dockerfile","dockerfileContent":"","args":{"AWS_ACCESS_KEY_ID":"AKIA-stale","AWS_SECRET_ACCESS_KEY":"stale-secret"},"targetPlatform":"linux/arm64,linux/amd64","dockerBuildOptions":{"network":"host"},"buildContext":"ctx","useBuildx":false,"buildxProvenanceMode":"","buildxDriverImage":""}`

func overrideRow(buildType bean2.CiBuildType, metadata string) *pipelineConfig.CiTemplateOverride {
	return &pipelineConfig.CiTemplateOverride{
		Id: 11, CiPipelineId: 21, GitMaterialId: 31, BuildContextGitMaterialId: 32, DockerfilePath: "legacy/Dockerfile",
		DockerRegistryId: "reg", DockerRepository: "repo", Active: true, CiBuildConfigId: 41,
		CiBuildConfig: &pipelineConfig.CiBuildConfig{Id: 41, Type: string(buildType), BuildMetadata: metadata},
	}
}

func newReadService(t *testing.T, templates *fakeCiTemplateRepository, overrides *fakeCiTemplateOverrideRepository) *CiTemplateReadServiceImpl {
	logger, err := util.NewSugardLogger()
	require.NoError(t, err)
	return NewCiTemplateReadServiceImpl(logger, templates, overrides)
}

// every exported read that goes through extractBuildConfigBean
func readOverrideBuildConfigs(t *testing.T, svc *CiTemplateReadServiceImpl) map[string]*bean2.CiBuildConfigBean {
	byPipelineId, err := svc.FindTemplateOverrideByCiPipelineId(21)
	require.NoError(t, err)
	byAppId, err := svc.FindTemplateOverrideByAppId(1)
	require.NoError(t, err)
	require.Len(t, byAppId, 1)
	byPipelineIds, err := svc.FindTemplateOverrideByCiPipelineIds([]int{21})
	require.NoError(t, err)
	require.Len(t, byPipelineIds, 1)
	return map[string]*bean2.CiBuildConfigBean{
		"FindTemplateOverrideByCiPipelineId":  byPipelineId.CiBuildConfig,
		"FindTemplateOverrideByAppId":         byAppId[0].CiBuildConfig,
		"FindTemplateOverrideByCiPipelineIds": byPipelineIds[0].CiBuildConfig,
	}
}

func TestCiTemplateReadService_OverrideNeverServesDockerBuildArgs(t *testing.T) {
	for _, buildType := range []bean2.CiBuildType{bean2.SELF_DOCKERFILE_BUILD_TYPE, bean2.MANAGED_DOCKERFILE_BUILD_TYPE} {
		t.Run(string(buildType)+"/stored args are dropped, the rest of the override is served", func(t *testing.T) {
			svc := newReadService(t, nil, &fakeCiTemplateOverrideRepository{overrides: []*pipelineConfig.CiTemplateOverride{overrideRow(buildType, staleDockerOverrideMetadata)}})
			for read, cfg := range readOverrideBuildConfigs(t, svc) {
				require.NotNil(t, cfg.DockerBuildConfig, read)
				assert.Nil(t, cfg.DockerBuildConfig.Args, read)
				assert.Equal(t, buildType, cfg.CiBuildType, read)
				assert.Equal(t, 41, cfg.Id, read)
				assert.Equal(t, 31, cfg.GitMaterialId, read)
				assert.Equal(t, 32, cfg.BuildContextGitMaterialId, read)
				assert.Equal(t, "Dockerfile", cfg.DockerBuildConfig.DockerfilePath, read)
				assert.Equal(t, "linux/arm64,linux/amd64", cfg.DockerBuildConfig.TargetPlatform, read)
				assert.Equal(t, map[string]string{"network": "host"}, cfg.DockerBuildConfig.DockerBuildOptions, read)
				assert.Equal(t, "ctx", cfg.DockerBuildConfig.BuildContext, read)
			}
		})
	}

	t.Run("override without args is served unchanged", func(t *testing.T) {
		svc := newReadService(t, nil, &fakeCiTemplateOverrideRepository{overrides: []*pipelineConfig.CiTemplateOverride{
			overrideRow(bean2.SELF_DOCKERFILE_BUILD_TYPE, `{"dockerfileRelativePath":"Dockerfile","targetPlatform":"linux/amd64"}`)}})
		for read, cfg := range readOverrideBuildConfigs(t, svc) {
			assert.Nil(t, cfg.DockerBuildConfig.Args, read)
			assert.Equal(t, "linux/amd64", cfg.DockerBuildConfig.TargetPlatform, read)
		}
	})

	t.Run("override without a ci_build_config row falls back to a dockerfile config without args", func(t *testing.T) {
		row := overrideRow(bean2.SELF_DOCKERFILE_BUILD_TYPE, "")
		row.CiBuildConfig = nil
		svc := newReadService(t, nil, &fakeCiTemplateOverrideRepository{overrides: []*pipelineConfig.CiTemplateOverride{row}})
		for read, cfg := range readOverrideBuildConfigs(t, svc) {
			assert.Equal(t, bean2.SELF_DOCKERFILE_BUILD_TYPE, cfg.CiBuildType, read)
			assert.Equal(t, "legacy/Dockerfile", cfg.DockerBuildConfig.DockerfilePath, read)
			assert.Nil(t, cfg.DockerBuildConfig.Args, read)
		}
	})

	t.Run("regression: buildpack override keeps its buildpack env args", func(t *testing.T) {
		svc := newReadService(t, nil, &fakeCiTemplateOverrideRepository{overrides: []*pipelineConfig.CiTemplateOverride{
			overrideRow(bean2.BUILDPACK_BUILD_TYPE, `{"builderId":"b","args":{"BP_ENV":"x"}}`)}})
		for read, cfg := range readOverrideBuildConfigs(t, svc) {
			assert.Nil(t, cfg.DockerBuildConfig, read)
			assert.Equal(t, map[string]string{"BP_ENV": "x"}, cfg.BuildPackConfig.Args, read)
		}
	})

	t.Run("regression: ci job (skip build) override", func(t *testing.T) {
		svc := newReadService(t, nil, &fakeCiTemplateOverrideRepository{overrides: []*pipelineConfig.CiTemplateOverride{overrideRow(bean2.SKIP_BUILD_TYPE, "")}})
		for read, cfg := range readOverrideBuildConfigs(t, svc) {
			assert.Equal(t, bean2.SKIP_BUILD_TYPE, cfg.CiBuildType, read)
			assert.Nil(t, cfg.DockerBuildConfig, read)
		}
	})
}

// Regression: app level build config owns the args and must keep serving them.
func TestCiTemplateReadService_AppLevelKeepsDockerBuildArgs(t *testing.T) {
	t.Run("app level build config row", func(t *testing.T) {
		template := &pipelineConfig.CiTemplate{Id: 1, AppId: 1, CiBuildConfig: &pipelineConfig.CiBuildConfig{
			Type: string(bean2.SELF_DOCKERFILE_BUILD_TYPE), BuildMetadata: `{"dockerfileRelativePath":"Dockerfile","args":{"APP":"1"},"targetPlatform":"linux/amd64"}`,
		}}
		svc := newReadService(t, &fakeCiTemplateRepository{template: template}, nil)
		byAppId, err := svc.FindByAppId(1)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"APP": "1"}, byAppId.CiBuildConfig.DockerBuildConfig.Args)
		assert.Equal(t, "linux/amd64", byAppId.CiBuildConfig.DockerBuildConfig.TargetPlatform)
		byAppIds, err := svc.FindByAppIds([]int{1})
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"APP": "1"}, byAppIds[1].CiBuildConfig.DockerBuildConfig.Args)
	})
	t.Run("legacy app level template without build config row", func(t *testing.T) {
		template := &pipelineConfig.CiTemplate{Id: 1, AppId: 1, DockerfilePath: "Dockerfile", Args: `{"LEGACY":"x"}`, TargetPlatform: "linux/arm64"}
		svc := newReadService(t, &fakeCiTemplateRepository{template: template}, nil)
		byAppId, err := svc.FindByAppId(1)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"LEGACY": "x"}, byAppId.CiBuildConfig.DockerBuildConfig.Args)
		byAppIds, err := svc.FindByAppIds([]int{1})
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"LEGACY": "x"}, byAppIds[1].CiBuildConfig.DockerBuildConfig.Args)
	})
}
