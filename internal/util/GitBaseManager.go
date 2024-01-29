package util

import (
	"fmt"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"strings"
	"time"
)

type GitManagerBase interface {
	Fetch(rootDir string, username string, password string) (response, errMsg string, err error)
	ListBranch(rootDir string, username string, password string) (response, errMsg string, err error)
	PullCli(rootDir string, username string, password string, branch string) (response, errMsg string, err error)
}

type GitManagerBaseImpl struct {
	logger *zap.SugaredLogger
}

func (impl *GitManagerBaseImpl) Fetch(rootDir string, username string, password string) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Fetch", "GitCli", start, err)
	}()
	impl.logger.Debugw("git fetch ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "fetch", "origin", "--tags", "--force")
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("fetch output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitManagerBaseImpl) runCommandWithCred(cmd *exec.Cmd, userName, password string) (response, errMsg string, err error) {
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_ASKPASS=%s", GIT_ASK_PASS),
		fmt.Sprintf("GIT_USERNAME=%s", userName),
		fmt.Sprintf("GIT_PASSWORD=%s", password),
	)
	return impl.runCommand(cmd)
}

func (impl *GitManagerBaseImpl) runCommand(cmd *exec.Cmd) (response, errMsg string, err error) {
	cmd.Env = append(cmd.Env, "HOME=/dev/null")
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		exErr, ok := err.(*exec.ExitError)
		if !ok {
			return "", "", err
		}
		errOutput := string(exErr.Stderr)
		return "", errOutput, err
	}
	output := string(outBytes)
	output = strings.TrimSpace(output)
	return output, "", nil
}

func (impl *GitManagerBaseImpl) ListBranch(rootDir string, username string, password string) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("ListBranch", "GitCli", start, err)
	}()
	impl.logger.Debugw("git branch ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "branch", "-r")
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("branch output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitManagerBaseImpl) PullCli(rootDir string, username string, password string, branch string) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Pull", "GitCli", start, err)
	}()
	impl.logger.Debugw("git pull ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "pull", "origin", branch, "--force")
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("pull output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}
