package commandManager

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"strings"
	"time"
)

type GitCommandManagerBase interface {
	Fetch(ctx GitContext, rootDir string) (response, errMsg string, err error)
	ListBranch(ctx GitContext, rootDir string) (response, errMsg string, err error)
	PullCli(ctx GitContext, rootDir string, branch string) (response, errMsg string, err error)
}

type GitManagerBaseImpl struct {
	logger *zap.SugaredLogger
	cfg    *configuration
}

func (impl *GitManagerBaseImpl) Fetch(ctx GitContext, rootDir string) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Fetch", "GitCli", start, err)
	}()
	impl.logger.Debugw("git fetch ", "location", rootDir)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "fetch", "origin", "--tags", "--force")
	defer cancel()
	output, errMsg, err := impl.runCommandWithCred(cmd, ctx.auth)
	impl.logger.Debugw("fetch output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitManagerBaseImpl) ListBranch(ctx GitContext, rootDir string) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("ListBranch", "GitCli", start, err)
	}()
	impl.logger.Debugw("git branch ", "location", rootDir)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "branch", "-r")
	defer cancel()
	output, errMsg, err := impl.runCommandWithCred(cmd, ctx.auth)
	impl.logger.Debugw("branch output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitManagerBaseImpl) PullCli(ctx GitContext, rootDir string, branch string) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Pull", "GitCli", start, err)
	}()
	impl.logger.Debugw("git pull ", "location", rootDir)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "pull", "origin", branch, "--force")
	defer cancel()
	output, errMsg, err := impl.runCommandWithCred(cmd, ctx.auth)
	impl.logger.Debugw("pull output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitManagerBaseImpl) runCommandWithCred(cmd *exec.Cmd, auth *BasicAuth) (response, errMsg string, err error) {
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_ASKPASS=%s", GIT_ASK_PASS),
		fmt.Sprintf("GIT_USERNAME=%s", auth.Username),
		fmt.Sprintf("GIT_PASSWORD=%s", auth.Password),
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

func (impl *GitManagerBaseImpl) createCmdWithContext(ctx GitContext, name string, arg ...string) (*exec.Cmd, context.CancelFunc) {
	newCtx := ctx
	cancel := func() {}

	timeout := impl.cfg.CliCmdTimeoutGlobal
	if _, ok := ctx.Deadline(); !ok && timeout > 0 {
		newCtx, cancel = ctx.WithTimeout(timeout)
	}
	cmd := exec.CommandContext(newCtx, name, arg...)
	return cmd, cancel
}
