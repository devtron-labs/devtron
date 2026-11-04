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

package adapter

import (
	"testing"

	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean2 "github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// staleOverrideArgs mimics the #2983 data: args saved into the override's build_metadata that the
// pipeline editor never shows, including credentials.
func staleOverrideArgs() map[string]string {
	return map[string]string{"AWS_ACCESS_KEY_ID": "AKIA-stale", "AWS_SECRET_ACCESS_KEY": "stale-secret", "SHARED": "from-override"}
}

func dockerOverride(buildType bean2.CiBuildType, args map[string]string) *bean2.CiBuildConfigBean {
	return &bean2.CiBuildConfigBean{
		CiBuildType: buildType,
		DockerBuildConfig: &bean2.DockerBuildConfig{
			DockerfilePath:     "Dockerfile.override",
			Args:               args,
			TargetPlatform:     "linux/arm64,linux/amd64",
			DockerBuildOptions: map[string]string{"network": "host"},
		},
	}
}

func appLevelDocker(args map[string]string) *bean2.CiBuildConfigBean {
	return &bean2.CiBuildConfigBean{
		CiBuildType:       bean2.SELF_DOCKERFILE_BUILD_TYPE,
		DockerBuildConfig: &bean2.DockerBuildConfig{DockerfilePath: "Dockerfile", Args: args, TargetPlatform: "linux/amd64"},
	}
}

func appLevelBuildpack() *bean2.CiBuildConfigBean {
	return &bean2.CiBuildConfigBean{
		CiBuildType:     bean2.BUILDPACK_BUILD_TYPE,
		BuildPackConfig: &bean2.BuildPackConfig{BuilderId: "b", Args: map[string]string{"BP_ENV": "x"}},
	}
}

func TestInheritAppLevelDockerBuildArgs(t *testing.T) {
	type tc struct {
		name         string
		ciTemplate   *pipelineConfig.CiTemplate
		appLevel     *bean2.CiBuildConfigBean
		overrideArgs map[string]string
		want         map[string]string
		wantErr      bool
	}
	cases := []tc{
		// app level build config is dockerfile based
		{name: "app level docker args replace stale override args", appLevel: appLevelDocker(map[string]string{"APP": "1", "SHARED": "from-app"}), overrideArgs: staleOverrideArgs(),
			want: map[string]string{"APP": "1", "SHARED": "from-app"}},
		{name: "app level docker without args drops stale override args", appLevel: appLevelDocker(nil), overrideArgs: staleOverrideArgs(), want: nil},
		{name: "app level docker with empty args drops stale override args", appLevel: appLevelDocker(map[string]string{}), overrideArgs: staleOverrideArgs(), want: map[string]string{}},
		{name: "app level docker args, override already clean", appLevel: appLevelDocker(map[string]string{"APP": "1"}), overrideArgs: nil, want: map[string]string{"APP": "1"}},
		{name: "app level docker config wins over legacy ci_template.args", ciTemplate: &pipelineConfig.CiTemplate{Args: `{"LEGACY":"x"}`},
			appLevel: appLevelDocker(map[string]string{"APP": "1"}), overrideArgs: staleOverrideArgs(), want: map[string]string{"APP": "1"}},
		// app level build config is not dockerfile based - previously the override args leaked into the build here
		{name: "app level buildpack drops stale override args", appLevel: appLevelBuildpack(), overrideArgs: staleOverrideArgs(), want: nil},
		{name: "app level buildpack ignores legacy ci_template.args", ciTemplate: &pipelineConfig.CiTemplate{Args: `{"LEGACY":"x"}`},
			appLevel: appLevelBuildpack(), overrideArgs: staleOverrideArgs(), want: nil},
		{name: "app level skip build drops stale override args", appLevel: &bean2.CiBuildConfigBean{CiBuildType: bean2.SKIP_BUILD_TYPE}, overrideArgs: staleOverrideArgs(), want: nil},
		// legacy template without a ci_build_config row - previously the override args leaked into the build here too
		{name: "legacy template args replace stale override args", ciTemplate: &pipelineConfig.CiTemplate{Args: `{"LEGACY":"x","SHARED":"from-legacy"}`},
			overrideArgs: staleOverrideArgs(), want: map[string]string{"LEGACY": "x", "SHARED": "from-legacy"}},
		{name: "legacy template with empty args string", ciTemplate: &pipelineConfig.CiTemplate{Args: ""}, overrideArgs: staleOverrideArgs(), want: map[string]string{}},
		{name: "legacy template with {} args", ciTemplate: &pipelineConfig.CiTemplate{Args: "{}"}, overrideArgs: staleOverrideArgs(), want: map[string]string{}},
		{name: "legacy template with null args", ciTemplate: &pipelineConfig.CiTemplate{Args: "null"}, overrideArgs: staleOverrideArgs(), want: nil},
		{name: "nil template and nil app level config", ciTemplate: nil, overrideArgs: staleOverrideArgs(), want: map[string]string{}},
		{name: "legacy template with malformed args errors", ciTemplate: &pipelineConfig.CiTemplate{Args: `{"broken"`}, overrideArgs: staleOverrideArgs(), wantErr: true},
	}
	for _, overrideType := range []bean2.CiBuildType{bean2.SELF_DOCKERFILE_BUILD_TYPE, bean2.MANAGED_DOCKERFILE_BUILD_TYPE} {
		for _, c := range cases {
			t.Run(string(overrideType)+"/"+c.name, func(t *testing.T) {
				override := dockerOverride(overrideType, c.overrideArgs)
				err := InheritAppLevelDockerBuildArgs(override, c.ciTemplate, c.appLevel)
				if c.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, c.want, override.DockerBuildConfig.Args)
				for k := range staleOverrideArgs() {
					if k == "SHARED" {
						continue
					}
					assert.NotContains(t, override.DockerBuildConfig.Args, k)
				}
				// nothing but args is inherited: the rest of the override stays the override's
				assert.Equal(t, "Dockerfile.override", override.DockerBuildConfig.DockerfilePath)
				assert.Equal(t, "linux/arm64,linux/amd64", override.DockerBuildConfig.TargetPlatform)
				assert.Equal(t, map[string]string{"network": "host"}, override.DockerBuildConfig.DockerBuildOptions)
				assert.Equal(t, overrideType, override.CiBuildType)
			})
		}
	}

	t.Run("buildpack override is untouched", func(t *testing.T) {
		override := &bean2.CiBuildConfigBean{CiBuildType: bean2.BUILDPACK_BUILD_TYPE, BuildPackConfig: &bean2.BuildPackConfig{Args: map[string]string{"BP": "1"}}}
		require.NoError(t, InheritAppLevelDockerBuildArgs(override, &pipelineConfig.CiTemplate{Args: `{"LEGACY":"x"}`}, appLevelDocker(map[string]string{"APP": "1"})))
		assert.Nil(t, override.DockerBuildConfig)
		assert.Equal(t, map[string]string{"BP": "1"}, override.BuildPackConfig.Args)
	})
	t.Run("skip build (ci job) override is untouched", func(t *testing.T) {
		override := &bean2.CiBuildConfigBean{CiBuildType: bean2.SKIP_BUILD_TYPE}
		require.NoError(t, InheritAppLevelDockerBuildArgs(override, nil, appLevelDocker(map[string]string{"APP": "1"})))
		assert.Nil(t, override.DockerBuildConfig)
	})
	t.Run("nil override is a no-op", func(t *testing.T) {
		assert.NoError(t, InheritAppLevelDockerBuildArgs(nil, &pipelineConfig.CiTemplate{Args: `{"broken"`}, nil))
	})
	t.Run("malformed legacy args are not parsed when the override has no docker config", func(t *testing.T) {
		override := &bean2.CiBuildConfigBean{CiBuildType: bean2.BUILDPACK_BUILD_TYPE, BuildPackConfig: &bean2.BuildPackConfig{}}
		assert.NoError(t, InheritAppLevelDockerBuildArgs(override, &pipelineConfig.CiTemplate{Args: `{"broken"`}, nil))
	})
}

// TestOverriddenPipelineEffectiveBuildArgs runs the two adapter calls the CI trigger makes for an overridden
// pipeline (InheritAppLevelDockerBuildArgs, then OverrideCiBuildConfig) and checks what reaches ci-runner.
func TestOverriddenPipelineEffectiveBuildArgs(t *testing.T) {
	type tc struct {
		name           string
		ciTemplate     *pipelineConfig.CiTemplate
		appLevel       *bean2.CiBuildConfigBean
		pipelineArgs   string // ci_pipeline.docker_args, "{}" when empty (as the trigger defaults it)
		wantArgs       map[string]string
		wantTargetPlat string
	}
	cases := []tc{
		{
			// #2983: pipeline args deleted in the editor, override metadata still holds the old args
			name: "issue 2983: pipeline args cleared, nothing at app level", ciTemplate: &pipelineConfig.CiTemplate{},
			appLevel: appLevelDocker(nil), pipelineArgs: "{}", wantArgs: map[string]string{}, wantTargetPlat: "linux/arm64,linux/amd64",
		},
		{
			name: "app level args only", ciTemplate: &pipelineConfig.CiTemplate{},
			appLevel: appLevelDocker(map[string]string{"APP": "1"}), pipelineArgs: "{}", wantArgs: map[string]string{"APP": "1"}, wantTargetPlat: "linux/arm64,linux/amd64",
		},
		{
			name: "pipeline args only", ciTemplate: &pipelineConfig.CiTemplate{},
			appLevel: appLevelDocker(nil), pipelineArgs: `{"PIPE":"p"}`, wantArgs: map[string]string{"PIPE": "p"}, wantTargetPlat: "linux/arm64,linux/amd64",
		},
		{
			name: "pipeline args win over app level args on the same key", ciTemplate: &pipelineConfig.CiTemplate{},
			appLevel: appLevelDocker(map[string]string{"APP": "1", "SHARED": "from-app"}), pipelineArgs: `{"SHARED":"from-pipeline","PIPE":"p"}`,
			wantArgs: map[string]string{"APP": "1", "SHARED": "from-pipeline", "PIPE": "p"}, wantTargetPlat: "linux/arm64,linux/amd64",
		},
		{
			name: "buildpack app level: only pipeline args", ciTemplate: &pipelineConfig.CiTemplate{Args: `{"LEGACY":"x"}`},
			appLevel: appLevelBuildpack(), pipelineArgs: `{"PIPE":"p"}`, wantArgs: map[string]string{"PIPE": "p"}, wantTargetPlat: "linux/arm64,linux/amd64",
		},
		{
			name: "legacy app level: legacy args + pipeline args", ciTemplate: &pipelineConfig.CiTemplate{Args: `{"LEGACY":"x","SHARED":"from-legacy"}`},
			appLevel: nil, pipelineArgs: `{"SHARED":"from-pipeline"}`, wantArgs: map[string]string{"LEGACY": "x", "SHARED": "from-pipeline"}, wantTargetPlat: "linux/arm64,linux/amd64",
		},
		{
			name: "legacy app level, nothing configured", ciTemplate: &pipelineConfig.CiTemplate{},
			appLevel: nil, pipelineArgs: "{}", wantArgs: map[string]string{}, wantTargetPlat: "linux/arm64,linux/amd64",
		},
	}
	for _, c := range cases {
		// both the read-stripped bean (normal path) and a raw bean still carrying args must give the same result
		for _, overrideArgs := range []map[string]string{nil, staleOverrideArgs()} {
			name := c.name + "/override args stripped"
			if overrideArgs != nil {
				name = c.name + "/override args present"
			}
			t.Run(name, func(t *testing.T) {
				override := dockerOverride(bean2.SELF_DOCKERFILE_BUILD_TYPE, overrideArgs)
				require.NoError(t, InheritAppLevelDockerBuildArgs(override, c.ciTemplate, c.appLevel))
				got, err := OverrideCiBuildConfig("Dockerfile", c.ciTemplate.Args, c.pipelineArgs, "{}", c.ciTemplate.TargetPlatform, override)
				require.NoError(t, err)
				assert.Equal(t, c.wantArgs, got.DockerBuildConfig.Args)
				assert.NotContains(t, got.DockerBuildConfig.Args, "AWS_ACCESS_KEY_ID")
				assert.NotContains(t, got.DockerBuildConfig.Args, "AWS_SECRET_ACCESS_KEY")
				assert.Equal(t, c.wantTargetPlat, got.DockerBuildConfig.TargetPlatform)
			})
		}
	}
}

// Regression: the non-overridden trigger path (same OverrideCiBuildConfig, fed the app level bean) is unchanged.
func TestOverrideCiBuildConfig_NonOverriddenPipelineRegression(t *testing.T) {
	t.Run("app level args merged with pipeline args, pipeline wins", func(t *testing.T) {
		got, err := OverrideCiBuildConfig("Dockerfile", "", `{"SHARED":"from-pipeline","PIPE":"p"}`, "{}", "",
			appLevelDocker(map[string]string{"APP": "1", "SHARED": "from-app"}))
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"APP": "1", "SHARED": "from-pipeline", "PIPE": "p"}, got.DockerBuildConfig.Args)
		assert.Equal(t, "linux/amd64", got.DockerBuildConfig.TargetPlatform)
	})
	t.Run("legacy template without build config row", func(t *testing.T) {
		got, err := OverrideCiBuildConfig("Dockerfile", `{"LEGACY":"x","SHARED":"from-legacy"}`, `{"SHARED":"from-pipeline"}`, `{"network":"host"}`, "linux/arm64", nil)
		require.NoError(t, err)
		assert.Equal(t, bean2.SELF_DOCKERFILE_BUILD_TYPE, got.CiBuildType)
		assert.Equal(t, map[string]string{"LEGACY": "x", "SHARED": "from-pipeline"}, got.DockerBuildConfig.Args)
		assert.Equal(t, "linux/arm64", got.DockerBuildConfig.TargetPlatform)
		assert.Equal(t, map[string]string{"network": "host"}, got.DockerBuildConfig.DockerBuildOptions)
	})
	t.Run("buildpack app level is left alone", func(t *testing.T) {
		got, err := OverrideCiBuildConfig("Dockerfile", "", `{"PIPE":"p"}`, "{}", "", appLevelBuildpack())
		require.NoError(t, err)
		assert.Nil(t, got.DockerBuildConfig)
		assert.Equal(t, map[string]string{"BP_ENV": "x"}, got.BuildPackConfig.Args)
	})
}
