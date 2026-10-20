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

package bean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCiBuildConfigBean_ClearDockerBuildArgs(t *testing.T) {
	t.Run("nil receiver is a no-op", func(t *testing.T) {
		var b *CiBuildConfigBean
		assert.NotPanics(t, func() { b.ClearDockerBuildArgs() })
	})

	t.Run("nil docker build config (buildpack / skip build) is a no-op", func(t *testing.T) {
		for _, buildType := range []CiBuildType{BUILDPACK_BUILD_TYPE, SKIP_BUILD_TYPE} {
			b := &CiBuildConfigBean{
				CiBuildType:     buildType,
				BuildPackConfig: &BuildPackConfig{BuilderId: "builder", Args: map[string]string{"BP_KEY": "v"}},
			}
			assert.NotPanics(t, func() { b.ClearDockerBuildArgs() }, buildType)
			assert.Nil(t, b.DockerBuildConfig, buildType)
			// buildpack env args are part of the override ("build the container image" section), never touched
			assert.Equal(t, map[string]string{"BP_KEY": "v"}, b.BuildPackConfig.Args, buildType)
		}
	})

	argShapes := map[string]map[string]string{
		"nil args":       nil,
		"empty args":     {},
		"populated args": {"AWS_ACCESS_KEY_ID": "AKIA...", "AWS_SECRET_ACCESS_KEY": "secret", "FOO": "bar"},
	}
	for _, buildType := range []CiBuildType{SELF_DOCKERFILE_BUILD_TYPE, MANAGED_DOCKERFILE_BUILD_TYPE} {
		for name, args := range argShapes {
			t.Run(string(buildType)+"/"+name+" are cleared, every other field kept", func(t *testing.T) {
				b := &CiBuildConfigBean{
					Id:                        7,
					GitMaterialId:             3,
					BuildContextGitMaterialId: 4,
					UseRootBuildContext:       true,
					CiBuildType:               buildType,
					PipelineType:              "CI_BUILD",
					DockerBuildConfig: &DockerBuildConfig{
						DockerfilePath:     "Dockerfile",
						DockerfileContent:  "FROM alpine",
						Args:               args,
						TargetPlatform:     "linux/arm64,linux/amd64",
						Language:           "go",
						DockerBuildOptions: map[string]string{"network": "host"},
						BuildContext:       "./ctx",
						UseBuildx:          true,
					},
				}
				b.ClearDockerBuildArgs()

				assert.Nil(t, b.DockerBuildConfig.Args)
				assert.Equal(t, 7, b.Id)
				assert.Equal(t, 3, b.GitMaterialId)
				assert.Equal(t, 4, b.BuildContextGitMaterialId)
				assert.True(t, b.UseRootBuildContext)
				assert.Equal(t, buildType, b.CiBuildType)
				assert.Equal(t, "CI_BUILD", b.PipelineType)
				d := b.DockerBuildConfig
				assert.Equal(t, "Dockerfile", d.DockerfilePath)
				assert.Equal(t, "FROM alpine", d.DockerfileContent)
				// target platform IS part of the override and must survive
				assert.Equal(t, "linux/arm64,linux/amd64", d.TargetPlatform)
				assert.Equal(t, "go", d.Language)
				assert.Equal(t, map[string]string{"network": "host"}, d.DockerBuildOptions)
				assert.Equal(t, "./ctx", d.BuildContext)
				assert.True(t, d.UseBuildx)
			})
		}
	}

	t.Run("idempotent", func(t *testing.T) {
		b := &CiBuildConfigBean{CiBuildType: SELF_DOCKERFILE_BUILD_TYPE, DockerBuildConfig: &DockerBuildConfig{Args: map[string]string{"A": "1"}}}
		b.ClearDockerBuildArgs()
		b.ClearDockerBuildArgs()
		assert.Nil(t, b.DockerBuildConfig.Args)
	})
}
