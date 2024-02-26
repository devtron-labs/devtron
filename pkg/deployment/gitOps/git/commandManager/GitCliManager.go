package commandManager

import (
	"fmt"
	"github.com/devtron-labs/devtron/util"
	"os/exec"

	"strings"
	"time"
)

type GitCliManagerImpl struct {
	*GitManagerBaseImpl
}

func (impl *GitCliManagerImpl) AddRepo(ctx GitContext, rootDir string, remoteUrl string, isBare bool) error {
	err := impl.gitInit(ctx, rootDir)
	if err != nil {
		return err
	}
	return impl.gitCreateRemote(ctx, rootDir, remoteUrl)
}

func (impl *GitCliManagerImpl) CommitAndPush(ctx GitContext, repoRoot, commitMsg, name, emailId string) (commitHash string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("CommitAndPushAllChanges", "GitService", start, err)
	}()
	err = LocateGitRepo(repoRoot)
	if err != nil {
		return "", err
	}
	impl.setConfig(ctx, repoRoot, emailId)
	_, _, err = impl.add(ctx, repoRoot)
	if err != nil {
		return "", err
	}
	_, _, err = impl.commit(ctx, repoRoot, commitMsg, name, emailId)
	if err != nil {
		return "", err
	}
	commit, _, err := impl.lastCommitHash(ctx, repoRoot)
	if err != nil {
		return "", err
	}
	impl.logger.Debugw("git hash", "repo", repoRoot, "hash", commit)

	_, _, err = impl.push(ctx, repoRoot)

	return commit, err
}

func (impl *GitCliManagerImpl) Pull(ctx GitContext, repoRoot string) (err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Pull", "GitService", start, err)
	}()

	err = LocateGitRepo(repoRoot)
	if err != nil {
		return err
	}
	response, errMsg, err := impl.PullCli(ctx, repoRoot, "origin/master")

	if strings.Contains(response, "already up-to-date") || strings.Contains(errMsg, "already up-to-date") {
		err = nil
		return nil
	}
	return err
}

func (impl *GitCliManagerImpl) gitInit(ctx GitContext, rootDir string) error {
	impl.logger.Debugw("git", "-C", rootDir, "init")
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "init")
	defer cancel()
	output, errMsg, err := impl.runCommandWithCred(cmd, ctx.auth)
	impl.logger.Debugw("root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return err
}

func (impl *GitCliManagerImpl) setConfig(ctx GitContext, rootDir string, email string) {
	impl.logger.Debugw("git config ", "location", rootDir)
	cmdUser := exec.CommandContext(ctx, "git", "-C", rootDir, "config", "user.name", ctx.auth.Username)
	cmdEmail := exec.CommandContext(ctx, "git", "-C", rootDir, "config", "user.email", email)
	impl.runCommand(cmdUser)
	impl.runCommand(cmdEmail)
}

func (impl *GitCliManagerImpl) commit(ctx GitContext, rootDir string, commitMsg string, user string, email string) (response, errMsg string, err error) {
	impl.logger.Debugw("git commit ", "location", rootDir)
	author := fmt.Sprintf("%s <%s>", user, email)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "commit", "--allow-empty", "-m", commitMsg, "--author", author)
	defer cancel()
	output, errMsg, err := impl.runCommandWithCred(cmd, ctx.auth)
	impl.logger.Debugw("git commit output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliManagerImpl) lastCommitHash(ctx GitContext, rootDir string) (response, errMsg string, err error) {
	impl.logger.Debugw("git log ", "location", rootDir)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "log", "--pretty=format:'%h'", "-n", "1")
	defer cancel()
	output, errMsg, err := impl.runCommandWithCred(cmd, ctx.auth)
	impl.logger.Debugw("git commit output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliManagerImpl) add(ctx GitContext, rootDir string) (response, errMsg string, err error) {
	impl.logger.Debugw("git add ", "location", rootDir)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "add", "-A")
	defer cancel()
	output, errMsg, err := impl.runCommandWithCred(cmd, ctx.auth)
	impl.logger.Debugw("git add output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliManagerImpl) push(ctx GitContext, rootDir string) (response, errMsg string, err error) {
	impl.logger.Debugw("git push ", "location", rootDir)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "push", "origin", "master", "--force")
	defer cancel()
	output, errMsg, err := impl.runCommandWithCred(cmd, ctx.auth)
	impl.logger.Debugw("git add output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliManagerImpl) gitCreateRemote(ctx GitContext, rootDir string, url string) error {
	impl.logger.Debugw("git", "-C", rootDir, "remote", "add", "origin", url)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "remote", "add", "origin", url)
	defer cancel()
	output, errMsg, err := impl.runCommandWithCred(cmd, ctx.auth)
	impl.logger.Debugw("url", url, "opt", output, "errMsg", errMsg, "error", err)
	return err
}
