package util

import (
	"fmt"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"os"
	"os/exec"
	"strings"
	"time"
)

type GitCliUtil struct {
	logger *zap.SugaredLogger
}

func NewGitCliUtil(logger *zap.SugaredLogger) *GitCliUtil {
	return &GitCliUtil{
		logger: logger,
	}
}

const GIT_ASK_PASS = "/git-ask-pass.sh"
const Branch_Master = "master"

func (impl *GitCliUtil) Fetch(rootDir string, username string, password string, allowInsecureTLS bool) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Fetch", "GitCli", start, err)
	}()
	impl.logger.Debugw("git fetch ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "fetch", "origin", "--tags", "--force")
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password, allowInsecureTLS)
	impl.logger.Debugw("fetch output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliUtil) Pull(rootDir string, username string, password string, branch string, allowInsecureTLS bool) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Pull", "GitCli", start, err)
	}()
	impl.logger.Debugw("git pull ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "pull", "origin", branch, "--force")
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password, allowInsecureTLS)
	impl.logger.Debugw("pull output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliUtil) Checkout(rootDir string, branch string) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Checkout", "GitCli", start, err)
	}()
	impl.logger.Debugw("git checkout ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "checkout", branch, "--force")
	output, errMsg, err := impl.runCommand(cmd)
	impl.logger.Debugw("checkout output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliUtil) ListBranch(rootDir string, username string, password string, allowInsecureTLS bool) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("ListBranch", "GitCli", start, err)
	}()
	impl.logger.Debugw("git branch ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "branch", "-r")
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password, allowInsecureTLS)
	impl.logger.Debugw("branch output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliUtil) runCommandWithCred(cmd *exec.Cmd, userName, password string, allowInsecureTLS bool) (response, errMsg string, err error) {
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_ASKPASS=%s", GIT_ASK_PASS),
		fmt.Sprintf("GIT_USERNAME=%s", userName),
		fmt.Sprintf("GIT_PASSWORD=%s", password),
	)
	if allowInsecureTLS {
		cmd.Env = append(cmd.Env, "GIT_SSL_NO_VERIFY=true")
	}
	return impl.runCommand(cmd)
}

func (impl *GitCliUtil) CommitAndPush(rootDir, commitMsg, name, emailId, username, password string, allowInsecureTLS bool) (commitHash string, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("CommitAndPush", "GitCli", start, err)
	}()
	impl.logger.Debugw("git commit and push", "location", rootDir)

	// Stage changes
	stageCmd := exec.Command("git", "-C", rootDir, "add", ".")
	_, stageErrMsg, stageErr := impl.runCommandWithCred(stageCmd, username, password, allowInsecureTLS)
	if stageErr != nil {
		return "", stageErrMsg, stageErr
	}

	// Commit changes
	commitCmd := exec.Command("git", "-C", rootDir, "commit", "-m", commitMsg, "--author", fmt.Sprintf("%s <%s>", name, emailId))
	commitOutput, commitErrMsg, commitErr := impl.runCommandWithCred(commitCmd, username, password, allowInsecureTLS)
	if commitErr != nil {
		return "", commitErrMsg, commitErr
	}

	// Push changes
	pushCmd := exec.Command("git", "-C", rootDir, "push", "origin", "master", "--force")
	pushOutput, pushErrMsg, pushErr := impl.runCommandWithCred(pushCmd, username, password, allowInsecureTLS)
	if pushErr != nil {
		return "", pushErrMsg, pushErr
	}

	impl.logger.Debugw("commit and push output", "root", rootDir, "commitOutput", commitOutput, "pushOutput", pushOutput)
	return commitOutput, "", nil
}

func (impl *GitCliUtil) runCommand(cmd *exec.Cmd) (response, errMsg string, err error) {
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

func (impl *GitCliUtil) Init(rootDir string, remoteUrl string, isBare bool) error {
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

func (impl *GitCliUtil) Clone(rootDir string, remoteUrl string, username string, password string, allowInsecureTLS bool) (response, errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Clone", "GitCli", start, err)
	}()
	impl.logger.Infow("git clone request", "rootDir", rootDir, "remoteUrl", remoteUrl, "username", username)
	err = impl.Init(rootDir, remoteUrl, false)
	if err != nil {
		return "", "", err
	}
	response, errMsg, err = impl.Fetch(rootDir, username, password, allowInsecureTLS)
	if err == nil && errMsg == "" {
		impl.logger.Warn("git fetch completed, pulling master branch data from remote origin")
		response, errMsg, err = impl.ListBranch(rootDir, username, password, allowInsecureTLS)
		if err != nil {
			impl.logger.Errorw("error on git pull", "response", response, "errMsg", errMsg, "err", err)
			return response, errMsg, err
		}
		branches := strings.Split(response, "\n")
		impl.logger.Infow("total branch available in git repo", "branches", branches)
		branch := ""
		for _, item := range branches {
			if strings.TrimSpace(item) == "origin/master" {
				branch = Branch_Master
			}
		}
		//if git repo has some branch take pull of the first branch, but eventually proxy chart will push into master branch
		if len(branch) == 0 && branches != nil && len(branches[0]) > 0 {
			branch = strings.ReplaceAll(branches[0], "origin/", "")
		}
		if branch == "" {
			impl.logger.Warnw("no branch found in git repo", "remoteUrl", remoteUrl, "response", response)
			return response, "", nil
		}
		response, errMsg, err = impl.Pull(rootDir, username, password, branch, allowInsecureTLS)
		if err != nil {
			impl.logger.Errorw("error on git pull", "branch", branch, "err", err)
			return response, errMsg, err
		}
	}
	return response, errMsg, err
}
