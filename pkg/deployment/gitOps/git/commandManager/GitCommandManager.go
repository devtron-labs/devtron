package commandManager

import (
	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

type GitCommandManager interface {
	GitCommandManagerBase
	AddRepo(ctx GitContext, rootDir string, remoteUrl string, isBare bool) error
	CommitAndPush(ctx GitContext, repoRoot, commitMsg, name, emailId string) (string, error)
	Pull(ctx GitContext, repoRoot string) (err error)
}

func NewGitCommandManager(logger *zap.SugaredLogger) GitCommandManager {

	cfg, err := ParseConfiguration()
	if err != nil {
		return nil
	}
	baseImpl := &GitManagerBaseImpl{
		logger: logger,
		cfg:    cfg,
	}
	if cfg.UseGitCli {
		return &GitCliManagerImpl{GitManagerBaseImpl: baseImpl}
	}
	return &GoGitSDKManagerImpl{GitManagerBaseImpl: baseImpl}
}

type configuration struct {
	UseGitCli           bool `env:"USE_GIT_CLI" envDefault:"false"`
	CliCmdTimeoutGlobal int  `env:"CLI_CMD_TIMEOUT_GLOBAL_SECONDS" envDefault:"0"`
}

func ParseConfiguration() (*configuration, error) {
	cfg := &configuration{}
	err := env.Parse(cfg)
	return cfg, err
}

const GIT_ASK_PASS = "/git-ask-pass.sh"

const Branch_Master = "master"
const ORIGIN_MASTER = "origin/master"
