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

package models

import (
	"go.uber.org/zap"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-pg/pg"
)

type Config struct {
	Addr            string `env:"PG_ADDR" envDefault:"127.0.0.1"`
	Port            string `env:"PG_PORT" envDefault:"5432"`
	User            string `env:"PG_USER" envDefault:""`
	Password        string `env:"PG_PASSWORD" envDefault:""`
	Database        string `env:"PG_DATABASE" envDefault:"orchestrator"`
	ApplicationName string `env:"APP" envDefault:"orchestrator"`
	LogQuery        bool   `env:"PG_LOG_QUERY" envDefault:"true"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}

func NewDbConnection(cfg *Config, logger *zap.SugaredLogger) (*pg.DB, error) {
	options := pg.Options{
		Addr:            cfg.Addr + ":" + cfg.Port,
		User:            cfg.User,
		Password:        cfg.Password,
		Database:        cfg.Database,
		ApplicationName: cfg.ApplicationName,
	}
	dbConnection := pg.Connect(&options)
	//check db connection
	var test string
	_, err := dbConnection.QueryOne(&test, `SELECT 1`)

	if err != nil {
		logger.Errorw("error in connecting db ", "db", cfg, "err", err)
		return nil, err
	} else {
		logger.Infow("connected with db", "db", cfg)
	}
	//--------------
	if cfg.LogQuery {
		dbConnection.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
			query, err := event.FormattedQuery()
			if err != nil {
				panic(err)
			}
			logger.Infow("query time",
				"duration", time.Since(event.StartTime),
				"query", query)
		})
	}
	return dbConnection, err
}

/*type DbConnectionHolder struct {
	connection *pg.DB
}

func (holder *DbConnectionHolder) CloseConnection() error {
	return holder.connection.Close()
}*/
