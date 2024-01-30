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

package git

import (
	"context"
	"fmt"
	git "github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/commandManager"
	"github.com/devtron-labs/devtron/util"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

type GitService interface {
	Clone(url, targetDir string) (clonedDir string, err error)
	CommitAndPushAllChanges(repoRoot, commitMsg, name, emailId string) (commitHash string, err error)
	GetCloneDirectory(targetDir string) (clonedDir string)
	Pull(repoRoot string) (err error)
	SetAuth(auth *git.BasicAuth)
}

type GitServiceImpl struct {
	Auth       *git.BasicAuth
	logger     *zap.SugaredLogger
	gitManager *git.GitManagerImpl
}

func NewGitServiceImpl(auth *git.BasicAuth, logger *zap.SugaredLogger, gitManager *git.GitManagerImpl) *GitServiceImpl {
	return &GitServiceImpl{
		Auth:       auth,
		logger:     logger,
		gitManager: gitManager,
	}
}

func (impl *GitServiceImpl) SetAuth(auth *git.BasicAuth) {
	impl.SetAuth(auth)
}

func (impl *GitServiceImpl) GetCloneDirectory(targetDir string) (clonedDir string) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("GetCloneDirectory", "GitService", start, nil)
	}()
	clonedDir = filepath.Join(GIT_WORKING_DIR, targetDir)
	return clonedDir
}

func (impl *GitServiceImpl) Clone(url, targetDir string) (clonedDir string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Clone", "GitService", start, err)
	}()
	impl.logger.Debugw("git checkout ", "url", url, "dir", targetDir)
	clonedDir = filepath.Join(GIT_WORKING_DIR, targetDir)
	errorMsg, err := impl.gitManager.Clone(context.Background(), clonedDir, url, impl.Auth)
	if err != nil {
		impl.logger.Errorw("error in git checkout", "url", url, "targetDir", targetDir, "err", err)
		return "", err
	}
	if errorMsg != "" {
		return "", fmt.Errorf(errorMsg)
	}
	return clonedDir, nil
}

func (impl *GitServiceImpl) Pull(repoRoot string) (err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Pull", "GitService", start, err)
	}()
	return impl.gitManager.Pull(context.Background(), repoRoot, impl.Auth)
}

func (impl GitServiceImpl) CommitAndPushAllChanges(repoRoot, commitMsg, name, emailId string) (commitHash string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("CommitAndPushAllChanges", "GitService", start, err)
	}()
	return impl.gitManager.CommitAndPush(context.Background(), repoRoot, commitMsg, name, emailId, impl.Auth)
}
