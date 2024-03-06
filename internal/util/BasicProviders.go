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
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"net/http"
	"reflect"
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

type HideSensitiveFieldsEncoder struct {
	zapcore.Encoder
	cfg zapcore.EncoderConfig
}

func redactField(ref *reflect.Value, i int) {
	refField := ref.Field(i)
	newValue := reflect.New(refField.Type()).Elem()
	fieldType := ref.Field(i).Type().Kind()
	// more cases can be added if required
	switch fieldType {
	case reflect.String:
		newValue.SetString("[REDACTED]")
	}
	ref.Field(i).Set(newValue)
}

func hideSensitiveData(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	ptrRef := reflect.ValueOf(v)
	if ptrRef.Kind() != reflect.Ptr {
		ptrRef = reflect.New(reflect.TypeOf(v))
		ptrRef.Elem().Set(reflect.ValueOf(v))
	}
	ref := ptrRef.Elem()
	refType := ref.Type()
	for i := 0; i < refType.NumField(); i++ {
		tag := refType.Field(i).Tag.Get("log")
		if tag == "hide" || tag == "false" {
			if ref.Field(i).CanSet() {
				redactField(&ref, i)
			}
		}
		fieldType := ref.Field(i).Type().Kind()
		// checks if the field is struct
		if fieldType == reflect.Struct {
			hideSensitiveData(ref.Field(i).Addr().Interface())
		} else if fieldType == reflect.Ptr && ref.Field(i).Elem().Kind() == reflect.Struct { // in case it is a ptr and points to a struct
			// making a copy so that original data do not get changed
			newCopy := reflect.New(ref.Field(i).Type().Elem()).Elem()
			newCopy.Set(ref.Field(i).Elem())
			hideSensitiveData(newCopy.Addr().Interface())
			ref.Field(i).Set(newCopy.Addr())
		}
	}
	return ref.Interface()
}

func (e *HideSensitiveFieldsEncoder) EncodeEntry(
	entry zapcore.Entry,
	fields []zapcore.Field,
) (*buffer.Buffer, error) {
	for idx, field := range fields {
		if field.Type == 23 {
			value := reflect.ValueOf(field.Interface)
			kind := value.Kind()
			if kind == reflect.Struct {
				fields[idx].Interface = hideSensitiveData(field.Interface)
			} else if value.Elem().Kind() == reflect.Struct { // in case ptr is passed in the log
				// passes only value so that original struct do not get changed
				fields[idx].Interface = hideSensitiveData(value.Elem().Interface())
			}
		}
	}
	return e.Encoder.EncodeEntry(entry, fields)
}

func newHideSensitiveFieldsEncoder(config zapcore.EncoderConfig) zapcore.Encoder {
	encoder := zapcore.NewConsoleEncoder(config)
	return &HideSensitiveFieldsEncoder{encoder, config}
}

func InitLogger() (*zap.SugaredLogger, error) {
	cfg := &LogConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse logger env config: " + err.Error())
		return nil, err
	}
	_ = zap.RegisterEncoder("hideSensitiveData", func(config zapcore.EncoderConfig) (zapcore.Encoder, error) {
		return newHideSensitiveFieldsEncoder(config), nil
	})

	config := zap.NewProductionConfig()
	if cfg.DevMode {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	} else {
		config.Encoding = "hideSensitiveData"
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
