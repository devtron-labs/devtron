package commandManager

import (
	"context"
	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

type GitCommandManager interface {
	GitCommandManagerBase
	AddRepo(ctx context.Context, rootDir string, remoteUrl string, isBare bool, auth *BasicAuth) error
	CommitAndPush(ctx context.Context, repoRoot, commitMsg, name, emailId string, auth *BasicAuth) (string, error)
	Pull(ctx context.Context, repoRoot string, auth *BasicAuth) (err error)
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

type Configuration struct {
	UseGitCli           bool `env:"USE_GIT_CLI" envDefault:"false"`
	CliCmdTimeoutGlobal int  `env:"CLI_CMD_TIMEOUT_GLOBAL_SECONDS" envDefault:"0"`
}

func ParseConfiguration() (*Configuration, error) {
	cfg := &Configuration{}
	err := env.Parse(cfg)
	return cfg, err
}

// BasicAuth represent a HTTP basic auth
type BasicAuth struct {
	Username, Password string
}

const GIT_ASK_PASS = "/git-ask-pass.sh"

const Branch_Master = "master"
