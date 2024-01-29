package util

import (
	"fmt"
	"github.com/devtron-labs/devtron/util"
	"gopkg.in/src-d/go-billy.v4/osfs"

	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type GitCliManagerImpl struct {
	*GitManagerBaseImpl
}

func (impl *GitCliManagerImpl) gitInit(rootDir string, username string, password string) error {
	impl.logger.Debugw("git", "-C", rootDir, "init")
	cmd := exec.Command("git", "-C", rootDir, "init")
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return err
}

func (impl *GitCliManagerImpl) commit(rootDir string, username string, password string, commitMsg string, user string, email string) (response, errMsg string, err error) {
	impl.logger.Debugw("git commit ", "location", rootDir)
	author := fmt.Sprintf("%s <%s>", user, email)
	cmd := exec.Command("git", "-C", rootDir, "commit", "-m", commitMsg, "--author", author)
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("git commit output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliManagerImpl) lastCommitHash(rootDir string, username string, password string) (response, errMsg string, err error) {
	impl.logger.Debugw("git log ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "--pretty=format:'%h' -n 1")
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("git commit output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliManagerImpl) add(rootDir string, username string, password string) (response, errMsg string, err error) {
	impl.logger.Debugw("git add ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "add", "-A")
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("git add output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliManagerImpl) push(rootDir string, username string, password string) (response, errMsg string, err error) {
	impl.logger.Debugw("git push ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "push", "--force")
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("git add output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliManagerImpl) gitCreateRemote(rootDir string, url string, username string, password string) error {
	impl.logger.Debugw("git", "-C", rootDir, "remote", "add", "origin", url)
	cmd := exec.Command("git", "-C", rootDir, "remote", "add", "origin", url)
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("url", url, "opt", output, "errMsg", errMsg, "error", err)
	return err
}

func (impl *GitCliManagerImpl) AddRepo(rootDir string, remoteUrl string, isBare bool, auth *BasicAuth) error {
	err := impl.gitInit(rootDir, auth.username, auth.password)
	if err != nil {
		return err
	}
	return impl.gitCreateRemote(rootDir, remoteUrl, auth.username, auth.password)
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

func (impl GitCliManagerImpl) CommitAndPush(repoRoot, commitMsg, name, emailId string, auth *BasicAuth) (commitHash string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("CommitAndPushAllChanges", "GitService", start, err)
	}()
	err = impl.openRepoPlain(repoRoot)
	if err != nil {
		return "", err
	}
	_, _, err = impl.add(repoRoot, auth.username, auth.password)
	if err != nil {
		return "", err
	}
	_, _, err = impl.commit(repoRoot, auth.username, auth.password, commitMsg, name, emailId)
	if err != nil {
		return "", err
	}
	commit, _, err := impl.lastCommitHash(repoRoot, auth.username, auth.password)
	if err != nil {
		return "", err
	}
	impl.logger.Debugw("git hash", "repo", repoRoot, "hash", commit)

	_, _, err = impl.push(repoRoot, auth.username, auth.password)

	return commit, err
}

func (impl *GitCliManagerImpl) Pull(repoRoot string, auth *BasicAuth) (err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Pull", "GitService", start, err)
	}()

	err = impl.openRepoPlain(repoRoot)
	if err != nil {
		return err
	}
	response, errMsg, err := impl.PullCli(repoRoot, auth.username, auth.password, "origin/master")

	if strings.Contains(response, "already up-to-date") || strings.Contains(errMsg, "already up-to-date") {
		err = nil
		return nil
	}
	return err
}
