/*
 * Copyright (c) 2026. Devtron Inc.
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
	"context"
	"testing"

	"github.com/devtron-labs/devtron/client/gitSensor"
	mock_gitSensor "github.com/devtron-labs/devtron/client/gitSensor/mocks"
	"github.com/devtron-labs/devtron/pkg/bean"
	gitMaterialRepository "github.com/devtron-labs/devtron/pkg/build/git/gitMaterial/repository"
	"github.com/golang/mock/gomock"
)

func TestValidateGitMaterialCloningMode(t *testing.T) {
	for _, cloningMode := range []string{bean.GitMaterialCloningModeFull, bean.GitMaterialCloningModeShallow} {
		if err := validateGitMaterialCloningMode(cloningMode); err != nil {
			t.Fatalf("expected cloning mode %q to be valid: %v", cloningMode, err)
		}
	}
	if err := validateGitMaterialCloningMode("DEPTH_10"); err == nil {
		t.Fatal("expected an unsupported cloning mode to fail validation")
	}
}

func TestGitMaterialIdentityNormalizesRetryValues(t *testing.T) {
	first := gitMaterialIdentity(" https://github.com/devtron-labs/devtron.git/ ", 1, "")
	second := gitMaterialIdentity("https://github.com/devtron-labs/devtron.git", 1, "./")
	if first != second {
		t.Fatalf("expected equivalent material identities, got %q and %q", first, second)
	}
}

func TestGitMaterialOptionsMatch(t *testing.T) {
	existingMaterial := &gitMaterialRepository.GitMaterial{
		FetchSubmodules: true,
		CloningMode:     bean.GitMaterialCloningModeShallow,
		FilterPattern:   []string{"services/api/**"},
	}
	requestedMaterial := &bean.GitMaterial{
		FetchSubmodules: true,
		CloningMode:     bean.GitMaterialCloningModeShallow,
		FilterPattern:   []string{"services/api/**"},
	}
	if !gitMaterialOptionsMatch(existingMaterial, requestedMaterial) {
		t.Fatal("expected identical retry options to match")
	}

	requestedMaterial.CloningMode = bean.GitMaterialCloningModeFull
	if gitMaterialOptionsMatch(existingMaterial, requestedMaterial) {
		t.Fatal("expected a changed cloning mode not to match")
	}
}

func TestAddRepositoryToGitSensorGroupsMaterialsByCloningMode(t *testing.T) {
	controller := gomock.NewController(t)
	gitSensorClient := mock_gitSensor.NewMockClient(controller)
	orchestrator := &CiCdPipelineOrchestratorImpl{GitSensorClient: gitSensorClient}

	gomock.InOrder(
		gitSensorClient.EXPECT().AddRepo(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, materials []*gitSensor.GitMaterial) error {
				assertMaterialsUseCloningMode(t, materials, bean.GitMaterialCloningModeFull)
				return nil
			},
		),
		gitSensorClient.EXPECT().AddRepo(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, materials []*gitSensor.GitMaterial) error {
				assertMaterialsUseCloningMode(t, materials, bean.GitMaterialCloningModeShallow)
				return nil
			},
		),
	)

	err := orchestrator.addRepositoryToGitSensor([]*bean.GitMaterial{
		{Id: 1, CloningMode: bean.GitMaterialCloningModeShallow},
		{Id: 2, CloningMode: bean.GitMaterialCloningModeFull},
		{Id: 3, CloningMode: bean.GitMaterialCloningModeShallow},
	})
	if err != nil {
		t.Fatalf("unexpected add repository error: %v", err)
	}
}

func assertMaterialsUseCloningMode(t *testing.T, materials []*gitSensor.GitMaterial, cloningMode string) {
	t.Helper()
	if len(materials) == 0 {
		t.Fatal("expected a non-empty git material batch")
	}
	for _, material := range materials {
		if material.CloningMode != cloningMode {
			t.Fatalf("expected a homogeneous %q batch, got material %d with %q", cloningMode, material.Id, material.CloningMode)
		}
	}
}
