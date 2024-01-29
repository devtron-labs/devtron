package util

import (
	"context"
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

func (impl *GoGitSDKManagerImpl) AddRepo(rootDir string, remoteUrl string, isBare bool, auth *BasicAuth) error {
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

func (impl GoGitSDKManagerImpl) Pull(repoRoot string, auth *BasicAuth) (err error) {

	_, workTree, err := impl.getRepoAndWorktree(repoRoot)
	if err != nil {
		return err
	}
	//-----------pull
	err = workTree.PullContext(context.Background(), &git.PullOptions{
		Auth: auth.toBasicAuth(),
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

func (impl GoGitSDKManagerImpl) CommitAndPush(repoRoot, commitMsg, name, emailId string, auth *BasicAuth) (string, error) {
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
	err = repo.Push(&git.PushOptions{
		Auth: auth.toBasicAuth(),
	})
	return commit.String(), err
}
func (auth *BasicAuth) toBasicAuth() *http.BasicAuth {
	return &http.BasicAuth{
		Username: auth.username,
		Password: auth.password,
	}
}
