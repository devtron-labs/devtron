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

package git

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean/gitOps"
	"time"
)

type UnimplementedGitOpsClient struct{}

func (u *UnimplementedGitOpsClient) CreateRepository(ctx context.Context, config *gitOps.GitOpsConfigDto) (url string, isNew bool, isEmpty bool, detailedErrorGitOpsConfigActions DetailedErrorGitOpsConfigActions) {
	return "", false, false, DetailedErrorGitOpsConfigActions{
		StageErrorMap: map[string]error{
			fmt.Sprintf("GitOps will not work"): fmt.Errorf("no gitops config found, please configure gitops first and try again"),
		},
	}
}

func (u *UnimplementedGitOpsClient) CommitValues(ctx context.Context, config *ChartConfig, gitOpsConfig *gitOps.GitOpsConfigDto, publishStatusConflictError bool) (commitHash string, commitTime time.Time, err error) {
	return "", time.Time{}, fmt.Errorf("no gitops config found, please configure gitops first")
}

func (u *UnimplementedGitOpsClient) GetRepoUrl(config *gitOps.GitOpsConfigDto) (repoUrl string, isRepoEmpty bool, err error) {
	return "", false, fmt.Errorf("no gitops config found, please configure gitops first")
}

func (u *UnimplementedGitOpsClient) DeleteRepository(config *gitOps.GitOpsConfigDto) error {
	return fmt.Errorf("no gitops config found, please configure gitops first")
}

func (u *UnimplementedGitOpsClient) CreateReadme(ctx context.Context, config *gitOps.GitOpsConfigDto) (string, error) {
	return "", fmt.Errorf("no gitops config found, please configure gitops first")
}

func (u *UnimplementedGitOpsClient) CreateFirstCommitOnHead(ctx context.Context, config *gitOps.GitOpsConfigDto) (string, error) {
	return "", fmt.Errorf("no gitops config found, please configure gitops first")
}
