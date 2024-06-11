/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package utils

import (
	"fmt"
	"github.com/caarlos0/env"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger is the defaut logger
	logger *zap.SugaredLogger
	//FIXME: remove this
	//defer Logger.Sync()
)

// Deprecated: instead calling this method inject utils from wire
func GetLogger() *zap.SugaredLogger {
	return logger
}

type LogConfig struct {
	Level int `env:"LOG_LEVEL" envDefault:"0"` // default info
}

func InitLogger() (*zap.SugaredLogger, error) {
	cfg := &LogConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse utils env config: " + err.Error())
		return nil, err
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.Level(cfg.Level))
	l, err := config.Build()
	if err != nil {
		fmt.Println("failed to create the default utils: " + err.Error())
		return nil, err
	}
	logger = l.Sugar()
	return logger, nil
}

func NewSugardLogger() (*zap.SugaredLogger, error) {
	return InitLogger()
}
