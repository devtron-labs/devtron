/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
	"github.com/devtron-labs/devtron/api/bean"
	apiGitOpsBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/bean"
	git "github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/commandManager"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/sliceUtil"
	"go.opentelemetry.io/otel"
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
	tlsConfig         *bean.TLSConfig
	isTlsEnabled      bool
}

func NewGitOpsHelperImpl(auth *git.BasicAuth, logger *zap.SugaredLogger, tlsConfig *bean.TLSConfig, isTlsEnabled bool) (*GitOpsHelper, error) {
	gitOpsHelper := &GitOpsHelper{
		Auth:              auth,
		logger:            logger,
		gitCommandManager: git.NewGitCommandManager(logger),
		tlsConfig:         tlsConfig,
		isTlsEnabled:      isTlsEnabled,
	}
	return gitOpsHelper, nil
}

func (impl *GitOpsHelper) SetAuth(auth *git.BasicAuth) {
	impl.Auth = auth
}

func (impl *GitOpsHelper) GetCloneDirectory(targetDir string) (clonedDir string) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("GetCloneDirectory", "GitService", start, nil)
	}()
	clonedDir = filepath.Join(bean2.GIT_WORKING_DIR, targetDir)
	return clonedDir
}

func (impl *GitOpsHelper) Clone(url, targetDir, targetRevision string) (clonedDir string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Clone", "GitService", start, err)
	}()
	var gitCtx git.GitContext
	gitCtx, clonedDir, err = impl.cloneAndFetch(url, targetDir)
	if err != nil {
		impl.logger.Errorw("error in git clone and fetch", "url", url, "targetDir", targetDir, "err", err)
		return clonedDir, err
	}
	err = impl.pullFromTargetRevision(gitCtx, clonedDir, targetRevision)
	if err != nil {
		impl.logger.Errorw("error in git pull from target revision", "clonedDir", clonedDir, "targetRevision", targetRevision, "err", err)
		return clonedDir, err
	}
	return clonedDir, nil
}

func (impl *GitOpsHelper) cloneAndFetch(url, targetDir string) (ctx git.GitContext, clonedDir string, err error) {
	impl.logger.Debugw("git checkout ", "url", url, "dir", targetDir)
	clonedDir = filepath.Join(bean2.GIT_WORKING_DIR, targetDir)
	ctx = git.BuildGitContext(context.Background()).WithCredentials(impl.Auth).
		WithTLSData(impl.tlsConfig.CaData, impl.tlsConfig.TLSKeyData, impl.tlsConfig.TLSCertData, impl.isTlsEnabled)
	err = impl.init(ctx, clonedDir, url, false)
	if err != nil {
		return ctx, clonedDir, err
	}
	_, errMsg, err := impl.gitCommandManager.Fetch(ctx, clonedDir)
	if errMsg != "" {
		impl.logger.Errorw("error in git fetch command", "errMsg", errMsg, "err", err)
		return ctx, clonedDir, fmt.Errorf(errMsg)
	} else if err != nil {
		impl.logger.Errorw("error in git fetch command", "clonedDir", clonedDir, "url", url, "err", err)
		return ctx, clonedDir, fmt.Errorf("error in git fetch command: %v", err)
	}
	impl.logger.Debugw("git fetch completed, pulling master branch data from remote origin")
	return ctx, clonedDir, nil
}

func (impl *GitOpsHelper) pullFromTargetRevision(ctx git.GitContext, clonedDir, targetRevision string) (err error) {
	branch, isEmptyRepo, err := impl.getSanitisedTargetRevision(ctx, clonedDir, targetRevision)
	if err != nil {
		impl.logger.Errorw("no branch found in git repo", "clonedDir", clonedDir, "targetRevision", targetRevision, "err", err)
		return err
	}
	if !isEmptyRepo {
		_, errMsg, err := impl.pullFromBranch(ctx, clonedDir, branch)
		if err != nil {
			impl.logger.Errorw("error on git pull", "clonedDir", clonedDir, "branch", branch, "errMsg", errMsg, "err", err)
			return err
		}
	}
	return nil
}

func (impl *GitOpsHelper) Pull(repoRoot, targetRevision string) (err error) {
	start := time.Now()
	ctx := git.BuildGitContext(context.Background()).WithCredentials(impl.Auth).
		WithTLSData(impl.tlsConfig.CaData, impl.tlsConfig.TLSKeyData, impl.tlsConfig.TLSCertData, impl.isTlsEnabled)
	err = impl.gitCommandManager.Pull(ctx, targetRevision, repoRoot)
	if err != nil {
		util.TriggerGitOpsMetrics("Pull", "GitService", start, err)
		return err
	}
	util.TriggerGitOpsMetrics("Pull", "GitService", start, nil)
	return nil
}

const PushErrorMessage = "failed to push some refs"

func (impl *GitOpsHelper) CommitAndPushAllChanges(ctx context.Context, repoRoot, targetRevision, commitMsg, name, emailId string) (commitHash string, err error) {
	start := time.Now()
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "GitOpsHelper.CommitAndPushAllChanges")
	defer func() {
		util.TriggerGitOpsMetrics("CommitAndPushAllChanges", "GitService", start, err)
		span.End()
	}()
	gitCtx := git.BuildGitContext(newCtx).WithCredentials(impl.Auth).
		WithTLSData(impl.tlsConfig.CaData, impl.tlsConfig.TLSKeyData, impl.tlsConfig.TLSCertData, impl.isTlsEnabled)
	commitHash, err = impl.gitCommandManager.CommitAndPush(gitCtx, repoRoot, targetRevision, commitMsg, name, emailId)
	if err != nil && strings.Contains(err.Error(), PushErrorMessage) {
		return commitHash, fmt.Errorf("%s %v", "push failed due to conflicts", err)
	}
	return commitHash, nil
}

func (impl *GitOpsHelper) pullFromBranch(ctx git.GitContext, rootDir, branch string) (string, string, error) {
	start := time.Now()
	response, errMsg, err := impl.gitCommandManager.PullCli(ctx, rootDir, branch)
	if err != nil {
		util.TriggerGitOpsMetrics("Pull", "GitCli", start, err)
		impl.logger.Errorw("error on git pull", "branch", branch, "err", err)
		return response, errMsg, err
	}
	util.TriggerGitOpsMetrics("Pull", "GitCli", start, nil)
	return response, errMsg, err
}

func (impl *GitOpsHelper) init(ctx git.GitContext, rootDir string, remoteUrl string, isBare bool) error {
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

	return impl.gitCommandManager.AddRepo(ctx, rootDir, remoteUrl, isBare)
}

func (impl *GitOpsHelper) getRemoteBranchList(ctx git.GitContext, rootDir string) ([]string, error) {
	validBranches := make([]string, 0)
	response, errMsg, err := impl.gitCommandManager.ListBranch(ctx, rootDir)
	if err != nil {
		impl.logger.Errorw("error on git pull", "response", response, "errMsg", errMsg, "err", err)
		return validBranches, err
	}
	branches := strings.Split(response, "\n")
	validBranches = sliceUtil.Filter(validBranches, branches, func(item string) bool {
		return len(item) != 0
	})
	impl.logger.Infow("total branch available in git repo", "branch length", len(validBranches), "branches", validBranches)
	return validBranches, nil
}

func (impl *GitOpsHelper) getSanitisedTargetRevision(ctx git.GitContext, rootDir, targetRevision string) (string, bool, error) {
	validBranches, err := impl.getRemoteBranchList(ctx, rootDir)
	if err != nil {
		impl.logger.Errorw("error in getting remote branch list", "rootDir", rootDir, "targetRevision", targetRevision, "err", err)
		return "", false, err
	}
	branch, isEmptyRepo, err := impl.getTargetRevision(ctx, rootDir, validBranches, targetRevision)
	if err != nil {
		impl.logger.Errorw("error in getting default branch", "branches", validBranches, "targetRevision", targetRevision, "err", err)
		return "", isEmptyRepo, err
	}
	if strings.HasPrefix(branch, "origin/") {
		branch = strings.TrimPrefix(branch, "origin/")
	}
	return branch, isEmptyRepo, nil
}

// getDefaultBranch returns the default branch of the git repository.
// - CASE 1: If the git repository has some branches, it will return the default branch.
// - CASE 2: If the git repository has no branches, it will return the current branch.
func (impl *GitOpsHelper) getDefaultBranch(ctx git.GitContext, rootDir string, branches []string) (branch string, isEmptyRepo bool, err error) {
	// if git repo has some branch, take pull of the default branch.
	// but eventually commit will push into TargetRevision branch.
	if len(branches) != 0 {
		// if no branch found then try to get the default branch from git
		impl.logger.Debugw("no branch found in git repo, trying to get default branch", "branches", branches)
		response, errMsg, err := impl.gitCommandManager.GetDefaultBranch(ctx, rootDir)
		if err == nil && strings.TrimSpace(response) != "" {
			branch = strings.TrimSpace(response)
			impl.logger.Debugw("default branch found in git repo", "branch", branch)
			return branch, isEmptyRepo, nil
		}
		impl.logger.Errorw("error on git pull", "response", response, "errMsg", errMsg, "err", err)
		// if error is not nil, then it means no default branch found in git repo.
		// so we will return the first branch as default branch.

		// As the origin/HEAD is found in the branches list,
		// we will consider it as an empty repository.
		isEmptyRepo = true
		impl.logger.Warnw("no default branch found in git repo, returning first branch", "branches", branches)
		branch = strings.TrimSpace(branches[0])
		return branch, isEmptyRepo, nil
	}
	isEmptyRepo = true
	impl.logger.Warnw("no branch found in git repo, returning default branch", "branch", util.GetDefaultTargetRevision())
	branch = util.GetDefaultTargetRevision()
	return branch, isEmptyRepo, err
}

func (impl *GitOpsHelper) getTargetRevision(ctx git.GitContext, rootDir string, branches []string, targetRevision string) (branch string, isEmptyRepo bool, err error) {
	// the preferred branch is bean.TargetRevisionMaster
	for _, item := range branches {
		if len(targetRevision) != 0 && item == targetRevision {
			return targetRevision, isEmptyRepo, nil
		} else if util.IsDefaultTargetRevision(item) {
			return util.GetDefaultTargetRevision(), isEmptyRepo, nil
		}
	}
	branch, isEmptyRepo, err = impl.getDefaultBranch(ctx, rootDir, branches)
	if err != nil {
		impl.logger.Errorw("error in getting default branch", "branches", branches, "targetRevision", targetRevision, "err", err)
		return "", isEmptyRepo, err
	}
	return branch, isEmptyRepo, err
}

/*
SanitiseCustomGitRepoURL
- It will sanitise the user given repository url based on GitOps provider

Case BITBUCKET_PROVIDER:
  - The clone URL format https://<user-name>@bitbucket.org/<workspace-name>/<repo-name>.git
  - Here the <user-name> can differ from user to user. SanitiseCustomGitRepoURL will return the repo url in format : https://bitbucket.org/<workspace-name>/<repo-name>.git

Case AZURE_DEVOPS_PROVIDER:
  - The clone URL format https://<organisation-name>@dev.azure.com/<organisation-name>/<project-name>/_git/<repo-name>
  - Here the <user-name> can differ from user to user. SanitiseCustomGitRepoURL will return the repo url in format : https://dev.azure.com/<organisation-name>/<project-name>/_git/<repo-name>
*/
func SanitiseCustomGitRepoURL(activeGitOpsConfig *apiGitOpsBean.GitOpsConfigDto, gitRepoURL string) (sanitisedGitRepoURL string) {
	sanitisedGitRepoURL = gitRepoURL
	if activeGitOpsConfig.Provider == bean2.BITBUCKET_PROVIDER && strings.Contains(gitRepoURL, fmt.Sprintf("://%s@%s", activeGitOpsConfig.Username, "bitbucket.org/")) {
		sanitisedGitRepoURL = strings.ReplaceAll(gitRepoURL, fmt.Sprintf("://%s@%s", activeGitOpsConfig.Username, "bitbucket.org/"), "://bitbucket.org/")
	}
	if activeGitOpsConfig.Provider == bean2.AZURE_DEVOPS_PROVIDER {
		azureDevopsOrgName := activeGitOpsConfig.Host[strings.LastIndex(activeGitOpsConfig.Host, "/")+1:]
		invalidBaseUrlFormat := fmt.Sprintf("://%s@%s", azureDevopsOrgName, "dev.azure.com/")
		if invalidBaseUrlFormat != "" && strings.Contains(gitRepoURL, invalidBaseUrlFormat) {
			sanitisedGitRepoURL = strings.ReplaceAll(gitRepoURL, invalidBaseUrlFormat, "://dev.azure.com/")
		}
	}
	return sanitisedGitRepoURL
}
