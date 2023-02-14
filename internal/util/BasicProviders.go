/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package util

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"net/http"
	"path"
)

var (
	// Logger is the defaut logger
	logger *zap.SugaredLogger
	//FIXME: remove this
	//defer Logger.Sync()
)

// Deprecated: instead calling this method inject logger from wire
func GetLogger() *zap.SugaredLogger {
	return logger
}

type LogConfig struct {
	Level      int `env:"LOG_LEVEL" envDefault:"0"` // default info
	MaxSize    int `env:"LOG_MAX_SIZE" envDefault:"100"`
	MaxBackups int `env:"LOG_MAX_BACKEUPS" envDefault:"2"`
	MaxAge     int `env:"LOG_MAX_AGE" envDefault:"7"`
}

func InitLogger() (*zap.SugaredLogger, error) {
	cfg := &LogConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse logger env config: " + err.Error())
		return nil, err
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.Level(cfg.Level))
	l, err := config.Build()
	if err != nil {
		fmt.Println("failed to create the default logger: " + err.Error())
		return nil, err
	}
	logger = l.Sugar()
	return logger, nil
}

func InitFileBasedLogger() (*zap.SugaredLogger, error) {
	cfg := &LogConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse logger env config: " + err.Error())
		return nil, err
	}

	jsonEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	logWriter := getLogWriter(cfg)

	core := zapcore.NewCore(jsonEncoder, logWriter, zap.NewAtomicLevelAt(zapcore.Level(cfg.Level)))

	l := zap.New(core, zap.AddCaller())
	logger = l.Sugar()
	return logger, nil

}

func getLogWriter(cfg *LogConfig) zapcore.WriteSyncer {
	// lumberjack.Logger is already safe for concurrent use, so we don't need to
	// lock it.
	err, devtronDirPath := util.CheckOrCreateDevtronDir()
	if err != nil {
		devtronDirPath = "/tmp/"
	}
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path.Join(devtronDirPath, "./out.log"),
		MaxSize:    cfg.MaxSize, // megabytes
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge, // days
		Compress:   true,
	})

	return zapcore.AddSync(w)
}

func NewSugardLogger() (*zap.SugaredLogger, error) {
	return InitLogger()
}

func NewFileBaseSugaredLogger() (*zap.SugaredLogger, error) {
	return InitFileBasedLogger()
}

func NewHttpClient() *http.Client {
	return http.DefaultClient
}
