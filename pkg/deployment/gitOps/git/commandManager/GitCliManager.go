package commandManager

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/util"
	"gopkg.in/src-d/go-billy.v4/osfs"

	"os"
	"path/filepath"
	"strings"
	"time"
)

type GitCliManagerImpl struct {
	*GitManagerBaseImpl
}

func (impl *GitCliManagerImpl) AddRepo(ctx context.Context, rootDir string, remoteUrl string, isBare bool, auth *BasicAuth) error {
	err := impl.gitInit(ctx, rootDir, auth.Username, auth.Password)
	if err != nil {
		return err
	}
	return impl.gitCreateRemote(ctx, rootDir, remoteUrl, auth.Username, auth.Password)
}

func (impl *GitCliManagerImpl) openRepoPlain(path string) error {

	if _, err := filepath.Abs(path); err != nil {
		return err
	}
	fst := osfs.New(path)
	_, err := fst.Stat(".git")
	if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (impl GitCliManagerImpl) CommitAndPush(ctx context.Context, repoRoot, commitMsg, name, emailId string, auth *BasicAuth) (commitHash string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("CommitAndPushAllChanges", "GitService", start, err)
	}()
	err = impl.openRepoPlain(repoRoot)
	if err != nil {
		return "", err
	}
	_, _, err = impl.add(ctx, repoRoot, auth.Username, auth.Password)
	if err != nil {
		return "", err
	}
	_, _, err = impl.commit(ctx, repoRoot, auth.Username, auth.Password, commitMsg, name, emailId)
	if err != nil {
		return "", err
	}
	commit, _, err := impl.lastCommitHash(ctx, repoRoot, auth.Username, auth.Password)
	if err != nil {
		return "", err
	}
	impl.logger.Debugw("git hash", "repo", repoRoot, "hash", commit)

	_, _, err = impl.push(ctx, repoRoot, auth.Username, auth.Password)

	return commit, err
}

func (impl *GitCliManagerImpl) Pull(ctx context.Context, repoRoot string, auth *BasicAuth) (err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Pull", "GitService", start, err)
	}()

	err = impl.openRepoPlain(repoRoot)
	if err != nil {
		return err
	}
	response, errMsg, err := impl.PullCli(ctx, repoRoot, auth.Username, auth.Password, "origin/master")

	if strings.Contains(response, "already up-to-date") || strings.Contains(errMsg, "already up-to-date") {
		err = nil
		return nil
	}
	return err
}

func (impl *GitCliManagerImpl) gitInit(ctx context.Context, rootDir string, username string, password string) error {
	impl.logger.Debugw("git", "-C", rootDir, "init")
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "init")
	defer cancel()
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return err
}

func (impl *GitCliManagerImpl) commit(ctx context.Context, rootDir string, username string, password string, commitMsg string, user string, email string) (response, errMsg string, err error) {
	impl.logger.Debugw("git commit ", "location", rootDir)
	author := fmt.Sprintf("%s <%s>", user, email)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "commit", "-m", commitMsg, "--author", author)
	defer cancel()
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("git commit output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliManagerImpl) lastCommitHash(ctx context.Context, rootDir string, username string, password string) (response, errMsg string, err error) {
	impl.logger.Debugw("git log ", "location", rootDir)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "log", "--pretty=format:'%h'", "-n", "1")
	defer cancel()
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("git commit output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliManagerImpl) add(ctx context.Context, rootDir string, username string, password string) (response, errMsg string, err error) {
	impl.logger.Debugw("git add ", "location", rootDir)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "add", "-A")
	defer cancel()
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("git add output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliManagerImpl) push(ctx context.Context, rootDir string, username string, password string) (response, errMsg string, err error) {
	impl.logger.Debugw("git push ", "location", rootDir)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "push", "--force")
	defer cancel()
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("git add output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliManagerImpl) gitCreateRemote(ctx context.Context, rootDir string, url string, username string, password string) error {
	impl.logger.Debugw("git", "-C", rootDir, "remote", "add", "origin", url)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "remote", "add", "origin", url)
	defer cancel()
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("url", url, "opt", output, "errMsg", errMsg, "error", err)
	return err
}
