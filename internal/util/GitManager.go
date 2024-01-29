package util

import (
	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

type Configuration struct {
	UseGitCli           bool `env:"USE_GIT_CLI" envDefault:"false"`
	CliCmdTimeoutGlobal int  `env:"CLI_CMD_TIMEOUT_GLOBAL_SECONDS" envDefault:"0"`
}

func ParseConfiguration() (*Configuration, error) {
	cfg := &Configuration{}
	err := env.Parse(cfg)
	return cfg, err
}

const GIT_ASK_PASS = "/git-ask-pass.sh"
const Branch_Master = "master"

type GitManager interface {
	GitManagerBase
	AddRepo(rootDir string, remoteUrl string, isBare bool, auth *BasicAuth) error
	CommitAndPush(repoRoot, commitMsg, name, emailId string, auth *BasicAuth) (string, error)
	Pull(repoRoot string, auth *BasicAuth) (err error)
}

type GitManagerImpl struct {
	GitManager
	logger *zap.SugaredLogger
	cfg    *Configuration
}

func NewGitManagerImpl(logger *zap.SugaredLogger) *GitManagerImpl {

	cfg, err := ParseConfiguration()
	if err != nil {
		return nil
	}
	baseImpl := &GitManagerBaseImpl{
		logger: logger,
	}
	if cfg.UseGitCli {
		return &GitManagerImpl{GitManager: &GitCliManagerImpl{GitManagerBaseImpl: baseImpl}}
	}
	return &GitManagerImpl{GitManager: &GoGitSDKManagerImpl{GitManagerBaseImpl: baseImpl}}
}

func (impl *GitManagerImpl) Clone(rootDir string, remoteUrl string, auth *BasicAuth) (errMsg string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Clone", "GitCli", start, err)
	}()
	impl.logger.Infow("git clone request", "rootDir", rootDir, "remoteUrl", remoteUrl, "username", auth.username)
	err = impl.init(rootDir, remoteUrl, false, auth)
	if err != nil {
		return "", err
	}
	_, errMsg, err = impl.Fetch(rootDir, auth.username, auth.password)
	if err == nil && errMsg == "" {
		impl.logger.Warn("git fetch completed, pulling master branch data from remote origin")
		_, errMsg, err := impl.pullFromBranch(rootDir, auth.username, auth.password)
		if err != nil {
			impl.logger.Errorw("error on git pull", "err", err)
			return errMsg, err
		}
	}
	return errMsg, err
}

func (impl *GitManagerImpl) pullFromBranch(rootDir string, username string, password string) (string, string, error) {
	branch, err := impl.getBranch(rootDir, username, password)
	if err != nil || branch == "" {
		impl.logger.Warnw("no branch found in git repo", "rootDir", rootDir)
		return "", "", err
	}
	response, errMsg, err := impl.PullCli(rootDir, username, password, branch)
	if err != nil {
		impl.logger.Errorw("error on git pull", "branch", branch, "err", err)
		return response, errMsg, err
	}
	return response, errMsg, err
}

func (impl *GitManagerImpl) init(rootDir string, remoteUrl string, isBare bool, auth *BasicAuth) error {
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
	return impl.AddRepo(rootDir, remoteUrl, isBare, auth)
}

func (impl *GitManagerImpl) getBranch(rootDir string, username string, password string) (string, error) {
	response, errMsg, err := impl.ListBranch(rootDir, username, password)
	if err != nil {
		impl.logger.Errorw("error on git pull", "response", response, "errMsg", errMsg, "err", err)
		return response, err
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
	return branch, nil
}
