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

package sql

import (
	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/common-lib/utils/bean"
	"go.uber.org/zap"
	"reflect"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-pg/pg"
)

// CATEGORY=POSTGRES
type Config struct {
	Addr            string `env:"PG_ADDR" envDefault:"127.0.0.1"  description:"address of postgres service" example:"postgresql-postgresql.devtroncd" deprecated:"false"`
	Port            string `env:"PG_PORT" envDefault:"5432" description:"port of postgresql service" example:"5432"`
	User            string `env:"PG_USER" envDefault:""  description:"user for postgres" example:"postgres"`
	Password        string `env:"PG_PASSWORD" envDefault:"" secretData:"-"  description:"password for postgres, associated with PG_USER" example:"confidential ;)"`
	Database        string `env:"PG_DATABASE" envDefault:"orchestrator"  description:"postgres database to be made connection with" example:"orchestrator, casbin, git_sensor, lens"`
	CasbinDatabase  string `env:"CASBIN_DATABASE" envDefault:"casbin""`
	ApplicationName string `env:"APP" envDefault:"orchestrator" description:"Application name"`
	ReadTimeout     int64  `env:"PG_READ_TIMEOUT" envDefault:"30"`
	WriteTimeout    int64  `env:"PG_WRITE_TIMEOUT" envDefault:"30"`
	bean.PgQueryMonitoringConfig
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return cfg, err
	}
	monitoringCfg, err := bean.GetPgQueryMonitoringConfig(cfg.ApplicationName)
	if err != nil {
		return cfg, err
	}
	cfg.PgQueryMonitoringConfig = monitoringCfg
	return cfg, err
}

func NewDbConnection(cfg *Config, logger *zap.SugaredLogger) (*pg.DB, error) {
	options := pg.Options{
		Addr:            cfg.Addr + ":" + cfg.Port,
		User:            cfg.User,
		Password:        cfg.Password,
		Database:        cfg.Database,
		ApplicationName: cfg.ApplicationName,
		ReadTimeout:     time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(cfg.WriteTimeout) * time.Second,
	}
	dbConnection := pg.Connect(&options)
	// check db connection
	var test string
	_, err := dbConnection.QueryOne(&test, `SELECT 1`)

	if err != nil {
		logger.Errorw("error in connecting db ", "db", obfuscateSecretTags(cfg), "err", err)
		return nil, err
	} else {
		logger.Infow("connected with db", "db", obfuscateSecretTags(cfg))
	}

	// --------------
	dbConnection.OnQueryProcessed(utils.GetPGPostQueryProcessor(cfg.PgQueryMonitoringConfig))
	return dbConnection, err
}

func obfuscateSecretTags(cfg interface{}) interface{} {

	cfgDpl := reflect.New(reflect.ValueOf(cfg).Elem().Type()).Interface()
	cfgDplElm := reflect.ValueOf(cfgDpl).Elem()
	t := cfgDplElm.Type()
	for i := 0; i < t.NumField(); i++ {
		if _, ok := t.Field(i).Tag.Lookup("secretData"); ok {
			cfgDplElm.Field(i).SetString("********")
		} else {
			cfgDplElm.Field(i).Set(reflect.ValueOf(cfg).Elem().Field(i))
		}
	}
	return cfgDpl
}

/*type DbConnectionHolder struct {
	connection *pg.DB
}

func (holder *DbConnectionHolder) CloseConnection() error {
	return holder.connection.Close()
}*/
