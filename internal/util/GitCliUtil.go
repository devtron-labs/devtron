package util

import (
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"os"
	"os/exec"
	"strings"
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

func (impl *GitCliUtil) Fetch(rootDir string, username string, password string) (response, errMsg string, err error) {
	impl.logger.Debugw("git fetch ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "fetch", "origin", "--tags", "--force")
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("fetch output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliUtil) Pull(rootDir string, username string, password string, branch string) (response, errMsg string, err error) {
	impl.logger.Debugw("git pull ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "pull", "origin", branch, "--force")
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("pull output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}
func (impl *GitCliUtil) Checkout(rootDir string, branch string) (response, errMsg string, err error) {
	impl.logger.Debugw("git checkout ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "checkout", branch, "--force")
	output, errMsg, err := impl.runCommand(cmd)
	impl.logger.Debugw("checkout output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliUtil) ListBranch(rootDir string, username string, password string) (response, errMsg string, err error) {
	impl.logger.Debugw("git branch ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "branch", "-r")
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("branch output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliUtil) runCommandWithCred(cmd *exec.Cmd, userName, password string) (response, errMsg string, err error) {
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_ASKPASS=%s", GIT_ASK_PASS),
		fmt.Sprintf("GIT_USERNAME=%s", userName),
		fmt.Sprintf("GIT_PASSWORD=%s", password),
	)
	return impl.runCommand(cmd)
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
	err := os.RemoveAll(rootDir)
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

func (impl *GitCliUtil) Clone(rootDir string, remoteUrl string, username string, password string) (response, errMsg string, err error) {
	impl.logger.Infow("input", rootDir, remoteUrl, username)
	err = impl.Init(rootDir, remoteUrl, false)
	if err != nil {
		return "", "", err
	}
	response, errMsg, err = impl.Fetch(rootDir, username, password)
	if err == nil && errMsg == "" {
		impl.logger.Warn("git fetch completed, pulling master branch data from remote origin")
		response, errMsg, err = impl.ListBranch(rootDir, username, password)
		branches := strings.Split(response, "\n")
		impl.logger.Info(branches)
		branch := ""
		for _, item := range branches {
			if strings.TrimSpace(item) == "origin/master" {
				branch = Branch_Master
			}
		}
		if len(branch) == 0 && len(branches) > 0 {
			branch = strings.ReplaceAll(branches[0], "origin/", "")
		} else if len(branch) == 0 {
			// only fetch will work, as we don't have any branch for pull
			return "", "", nil
		}
		response, errMsg, err = impl.Pull(rootDir, username, password, branch)
		if err != nil {
			impl.logger.Errorw("error on git pull", "branch", branch, "err", err)
			return "", "", err
		}
	}
	return response, errMsg, err
}
