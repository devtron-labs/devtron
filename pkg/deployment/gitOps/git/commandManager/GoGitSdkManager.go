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

package commandManager

import (
	"github.com/devtron-labs/devtron/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"time"
)

type GoGitSDKManagerImpl struct {
	*GitManagerBaseImpl
}

func (impl *GoGitSDKManagerImpl) AddRepo(ctx GitContext, rootDir string, remoteUrl string, isBare bool) error {
	repo, err := git.PlainInit(rootDir, isBare)
	if err != nil {
		return err
	}
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{remoteUrl},
	})
	return err
}

func (impl GoGitSDKManagerImpl) Pull(ctx GitContext, repoRoot string) (err error) {

	_, workTree, err := impl.getRepoAndWorktree(repoRoot)
	if err != nil {
		return err
	}
	//-----------pull
	err = workTree.PullContext(ctx, &git.PullOptions{
		Auth: ctx.auth.ToBasicAuth(),
	})
	if err != nil && err.Error() == "already up-to-date" {
		err = nil
		return nil
	}
	return err
}

func (impl GoGitSDKManagerImpl) getRepoAndWorktree(repoRoot string) (*git.Repository, *git.Worktree, error) {
	var err error
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("getRepoAndWorktree", "GitService", start, err)
	}()
	r, err := git.PlainOpen(repoRoot)
	if err != nil {
		return nil, nil, err
	}
	w, err := r.Worktree()
	return r, w, err
}

func (impl GoGitSDKManagerImpl) CommitAndPush(ctx GitContext, repoRoot, commitMsg, name, emailId string) (string, error) {
	repo, workTree, err := impl.getRepoAndWorktree(repoRoot)
	if err != nil {
		return "", err
	}
	err = workTree.AddGlob("")
	if err != nil {
		return "", err
	}
	//--  commit
	commit, err := workTree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  name,
			Email: emailId,
			When:  time.Now(),
		},
		Committer: &object.Signature{
			Name:  name,
			Email: emailId,
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", err
	}
	impl.logger.Debugw("git hash", "repo", repoRoot, "hash", commit.String())
	//-----------push
	err = repo.PushContext(ctx, &git.PushOptions{
		Auth: ctx.auth.ToBasicAuth(),
	})
	return commit.String(), err
}
func (auth *BasicAuth) ToBasicAuth() *http.BasicAuth {
	return &http.BasicAuth{
		Username: auth.Username,
		Password: auth.Password,
	}
}
