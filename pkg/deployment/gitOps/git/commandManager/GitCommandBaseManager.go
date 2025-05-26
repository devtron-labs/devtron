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
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib/git-manager"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"strings"
	"time"
)

type GitCommandManagerBase interface {
	Clone(ctx GitContext, rootDir, repoUrl string) (response, errMsg string, err error)
	Fetch(ctx GitContext, rootDir string) (response, errMsg string, err error)
	ListBranch(ctx GitContext, rootDir string) (response, errMsg string, err error)
	// GetCurrentBranch returns the current branch of the git repository directory.
	//	command: git -C <rootDir> branch --show-current
	GetCurrentBranch(ctx GitContext, rootDir string) (response, errMsg string, err error)
	// GetDefaultBranch returns the default branch of the git repository directory.
	//	command: git -C <rootDir> rev-parse --abbrev-ref origin/HEAD
	// If the repository is empty, it returns an error.
	// In that case, use GetCurrentBranch to get the default branch.
	GetDefaultBranch(ctx GitContext, rootDir string) (response, errMsg string, err error)
	PullCli(ctx GitContext, rootDir string, branch string) (response, errMsg string, err error)
}

type GitManagerBaseImpl struct {
	logger *zap.SugaredLogger
	cfg    *configuration
}

func (impl *GitManagerBaseImpl) Clone(ctx GitContext, rootDir, repoUrl string) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Clone", "GitCli", start, err)
	}()
	impl.logger.Debugw("git clone --depth=1 --single-branch --no-tags", "location", rootDir, "repoUrl", repoUrl)
	args := []string{"clone", "--depth=1", "--single-branch", "--no-tags", repoUrl, rootDir}
	cmd, cancel := impl.createCmdWithContext(ctx, "git", args...)
	defer cancel()
	tlsPathInfo, err := git_manager.CreateFilesForTlsData(git_manager.BuildTlsData(ctx.TLSKey, ctx.TLSCertificate, ctx.CACert, ctx.TLSVerificationEnabled), TLS_FOLDER)
	if err != nil {
		//making it non-blocking
		impl.logger.Errorw("error encountered in createFilesForTlsData", "err", err)
	}
	defer git_manager.DeleteTlsFiles(tlsPathInfo)
	output, errMsg, err := impl.runCommandWithCred(cmd, ctx.auth, tlsPathInfo)
	impl.logger.Debugw("git clone --depth=1 --single-branch --no-tags output", "root", rootDir, "repoUrl", repoUrl, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitManagerBaseImpl) Fetch(ctx GitContext, rootDir string) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Fetch", "GitCli", start, err)
	}()
	impl.logger.Debugw("git fetch ", "location", rootDir)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "fetch", "origin", "--tags", "--force")
	defer cancel()
	tlsPathInfo, err := git_manager.CreateFilesForTlsData(git_manager.BuildTlsData(ctx.TLSKey, ctx.TLSCertificate, ctx.CACert, ctx.TLSVerificationEnabled), TLS_FOLDER)
	if err != nil {
		//making it non-blocking
		impl.logger.Errorw("error encountered in createFilesForTlsData", "err", err)
	}
	defer git_manager.DeleteTlsFiles(tlsPathInfo)
	output, errMsg, err := impl.runCommandWithCred(cmd, ctx.auth, tlsPathInfo)
	impl.logger.Debugw("fetch output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitManagerBaseImpl) ListBranch(ctx GitContext, rootDir string) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("ListBranch", "GitCli", start, err)
	}()
	impl.logger.Debugw("git branch -r", "location", rootDir)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "branch", "-r")
	defer cancel()
	tlsPathInfo, err := git_manager.CreateFilesForTlsData(git_manager.BuildTlsData(ctx.TLSKey, ctx.TLSCertificate, ctx.CACert, ctx.TLSVerificationEnabled), TLS_FOLDER)
	if err != nil {
		//making it non-blocking
		impl.logger.Errorw("error encountered in createFilesForTlsData", "err", err)
	}
	defer git_manager.DeleteTlsFiles(tlsPathInfo)
	output, errMsg, err := impl.runCommandWithCred(cmd, ctx.auth, tlsPathInfo)
	impl.logger.Debugw("git branch -r output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitManagerBaseImpl) GetCurrentBranch(ctx GitContext, rootDir string) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("GetCurrentBranch", "GitCli", start, err)
	}()
	impl.logger.Debugw("git branch --show-current", "location", rootDir)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "branch", "--show-current")
	defer cancel()
	tlsPathInfo, err := git_manager.CreateFilesForTlsData(git_manager.BuildTlsData(ctx.TLSKey, ctx.TLSCertificate, ctx.CACert, ctx.TLSVerificationEnabled), TLS_FOLDER)
	if err != nil {
		//making it non-blocking
		impl.logger.Errorw("error encountered in createFilesForTlsData", "err", err)
	}
	defer git_manager.DeleteTlsFiles(tlsPathInfo)
	output, errMsg, err := impl.runCommandWithCred(cmd, ctx.auth, tlsPathInfo)
	impl.logger.Debugw("git branch --show-current output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitManagerBaseImpl) GetDefaultBranch(ctx GitContext, rootDir string) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("GetDefaultBranch", "GitCli", start, err)
	}()
	impl.logger.Debugw("git rev-parse --abbrev-ref origin/HEAD", "location", rootDir)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "rev-parse", "--abbrev-ref", "origin/HEAD")
	defer cancel()
	tlsPathInfo, err := git_manager.CreateFilesForTlsData(git_manager.BuildTlsData(ctx.TLSKey, ctx.TLSCertificate, ctx.CACert, ctx.TLSVerificationEnabled), TLS_FOLDER)
	if err != nil {
		//making it non-blocking
		impl.logger.Errorw("error encountered in createFilesForTlsData", "err", err)
	}
	defer git_manager.DeleteTlsFiles(tlsPathInfo)
	output, errMsg, err := impl.runCommandWithCred(cmd, ctx.auth, tlsPathInfo)
	impl.logger.Debugw("git rev-parse --abbrev-ref origin/HEAD output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitManagerBaseImpl) PullCli(ctx GitContext, rootDir string, branch string) (response, errMsg string, err error) {
	start := time.Now()
	impl.logger.Debugw("git pull ", "location", rootDir)
	cmd, cancel := impl.createCmdWithContext(ctx, "git", "-C", rootDir, "pull", "origin", branch, "--force")
	defer cancel()
	tlsPathInfo, err := git_manager.CreateFilesForTlsData(git_manager.BuildTlsData(ctx.TLSKey, ctx.TLSCertificate, ctx.CACert, ctx.TLSVerificationEnabled), TLS_FOLDER)
	if err != nil {
		//making it non-blocking
		impl.logger.Errorw("error encountered in createFilesForTlsData", "err", err)
	}
	defer git_manager.DeleteTlsFiles(tlsPathInfo)
	output, errMsg, err := impl.runCommandWithCred(cmd, ctx.auth, tlsPathInfo)
	if err != nil {
		if !IsAlreadyUpToDateError(response, errMsg) {
			util.TriggerGitOpsMetrics("Pull", "GitCli", start, err)
		}
	}
	util.TriggerGitOpsMetrics("Pull", "GitCli", start, nil)
	impl.logger.Debugw("pull output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitManagerBaseImpl) runCommandWithCred(cmd *exec.Cmd, auth *BasicAuth, tlsPathInfo *git_manager.TlsPathInfo) (response, errMsg string, err error) {
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_ASKPASS=%s", GIT_ASK_PASS),
		fmt.Sprintf("GIT_USERNAME=%s", auth.Username),
		fmt.Sprintf("GIT_PASSWORD=%s", auth.Password),
	)
	if tlsPathInfo != nil {
		if tlsPathInfo.TlsKeyPath != "" && tlsPathInfo.TlsCertPath != "" {
			cmd.Env = append(cmd.Env,
				fmt.Sprintf("GIT_SSL_KEY=%s", tlsPathInfo.TlsKeyPath),
				fmt.Sprintf("GIT_SSL_CERT=%s", tlsPathInfo.TlsCertPath))
		}
		if tlsPathInfo.CaCertPath != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_SSL_CAINFO=%s", tlsPathInfo.CaCertPath))
		}
	}
	return impl.runCommand(cmd)
}

func (impl *GitManagerBaseImpl) runCommand(cmd *exec.Cmd) (response, errMsg string, err error) {
	cmd.Env = append(cmd.Env, "HOME=/dev/null")
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		exErr, ok := err.(*exec.ExitError)
		if !ok {
			return "", "", fmt.Errorf("%s %v", outBytes, err)
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
