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
