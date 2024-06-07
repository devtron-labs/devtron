/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package util

import (
	"fmt"
	"github.com/caarlos0/env"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
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
	Level int `env:"LOG_LEVEL" envDefault:"0"` // default info

	DevMode bool `env:"LOGGER_DEV_MODE" envDefault:"false"`
}

func InitLogger() (*zap.SugaredLogger, error) {
	cfg := &LogConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse logger env config: " + err.Error())
		return nil, err
	}

	config := zap.NewProductionConfig()
	if cfg.DevMode {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	config.Level = zap.NewAtomicLevelAt(zapcore.Level(cfg.Level))
	l, err := config.Build()
	if err != nil {
		fmt.Println("failed to create the default logger: " + err.Error())
		return nil, err
	}
	logger = l.Sugar()
	return logger, nil
}

func NewSugardLogger() (*zap.SugaredLogger, error) {
	return InitLogger()
}

func NewHttpClient() *http.Client {
	return http.DefaultClient
}
