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
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// GitOpsHelper GitOps Helper maintains the auth creds in state and is used by implementation of
// git client implementations and GitFactory
type GitOpsHelper struct {
	Auth              *git.BasicAuth
	logger            *zap.SugaredLogger
	gitCommandManager git.GitCommandManager
}

func NewGitOpsHelperImpl(auth *git.BasicAuth, logger *zap.SugaredLogger) *GitOpsHelper {
	return &GitOpsHelper{
		Auth:              auth,
		logger:            logger,
		gitCommandManager: git.NewGitCommandManager(logger),
	}
}

func (impl *GitOpsHelper) SetAuth(auth *git.BasicAuth) {
	impl.Auth = auth
}

func (impl *GitOpsHelper) GetCloneDirectory(targetDir string) (clonedDir string) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("GetCloneDirectory", "GitService", start, nil)
	}()
	clonedDir = filepath.Join(GIT_WORKING_DIR, targetDir)
	return clonedDir
}

func (impl *GitOpsHelper) Clone(url, targetDir string) (clonedDir string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Clone", "GitService", start, err)
	}()
	impl.logger.Debugw("git checkout ", "url", url, "dir", targetDir)
	clonedDir = filepath.Join(GIT_WORKING_DIR, targetDir)
	//errorMsg, err := impl.Clone2(context.Background(), clonedDir, clonedDir, impl.Auth)
	ctx := context.Background()
	err = impl.init(ctx, clonedDir, url, false, impl.Auth)
	if err != nil {
		return "", err
	}
	_, errMsg, err := impl.gitCommandManager.Fetch(ctx, clonedDir, impl.Auth.Username, impl.Auth.Password)
	if err == nil && errMsg == "" {
		impl.logger.Warn("git fetch completed, pulling master branch data from remote origin")
		_, errMsg, err := impl.pullFromBranch(ctx, clonedDir, impl.Auth.Username, impl.Auth.Password)
		if err != nil {
			impl.logger.Errorw("error on git pull", "err", err)
			return errMsg, err
		}
	}
	if errMsg != "" {
		return "", fmt.Errorf(errMsg)
	}
	return clonedDir, nil
}

func (impl *GitOpsHelper) Pull(repoRoot string) (err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Pull", "GitService", start, err)
	}()
	return impl.gitCommandManager.Pull(context.Background(), repoRoot, impl.Auth)
}

func (impl GitOpsHelper) CommitAndPushAllChanges(repoRoot, commitMsg, name, emailId string) (commitHash string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("CommitAndPushAllChanges", "GitService", start, err)
	}()
	return impl.gitCommandManager.CommitAndPush(context.Background(), repoRoot, commitMsg, name, emailId, impl.Auth)
}

func (impl *GitOpsHelper) pullFromBranch(ctx context.Context, rootDir string, username string, password string) (string, string, error) {
	branch, err := impl.getBranch(ctx, rootDir, username, password)
	if err != nil || branch == "" {
		impl.logger.Warnw("no branch found in git repo", "rootDir", rootDir)
		return "", "", err
	}
	response, errMsg, err := impl.gitCommandManager.PullCli(ctx, rootDir, username, password, branch)
	if err != nil {
		impl.logger.Errorw("error on git pull", "branch", branch, "err", err)
		return response, errMsg, err
	}
	return response, errMsg, err
}

func (impl *GitOpsHelper) init(ctx context.Context, rootDir string, remoteUrl string, isBare bool, auth *git.BasicAuth) error {
	//-----------------
	start := time.Now()
	var err error
	defer func() {
		util.TriggerGitOpsMetrics("Init", "GitCli", start, err)
	}()
	err = os.RemoveAll(rootDir)
	if err != nil {
		impl.logger.Errorw("error in cleaning rootDir", "err", err)
		return err
	}
	err = os.MkdirAll(rootDir, 0755)
	if err != nil {
		return err
	}
	return impl.gitCommandManager.AddRepo(ctx, rootDir, remoteUrl, isBare, auth)
}

func (impl *GitOpsHelper) getBranch(ctx context.Context, rootDir string, username string, password string) (string, error) {
	response, errMsg, err := impl.gitCommandManager.ListBranch(ctx, rootDir, username, password)
	if err != nil {
		impl.logger.Errorw("error on git pull", "response", response, "errMsg", errMsg, "err", err)
		return response, err
	}
	branches := strings.Split(response, "\n")
	impl.logger.Infow("total branch available in git repo", "branches", branches)
	branch := ""
	for _, item := range branches {
		if strings.TrimSpace(item) == git.ORIGIN_MASTER {
			branch = git.Branch_Master
		}
	}
	//if git repo has some branch take pull of the first branch, but eventually proxy chart will push into master branch
	if len(branch) == 0 && branches != nil {
		branch = strings.ReplaceAll(branches[0], "origin/", "")
	}
	return branch, nil
}
