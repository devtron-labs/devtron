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

package sql

import (
	"github.com/devtron-labs/devtron/internal/middleware"
	"go.uber.org/zap"
	"reflect"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-pg/pg"
)

type Config struct {
	Addr                   string `env:"PG_ADDR" envDefault:"127.0.0.1"`
	Port                   string `env:"PG_PORT" envDefault:"32080"`
	User                   string `env:"PG_USER" envDefault:""`
	Password               string `env:"PG_PASSWORD" envDefault:"" secretData:"-"`
	Database               string `env:"PG_DATABASE" envDefault:"orchestrator"`
	ApplicationName        string `env:"APP" envDefault:"orchestrator"`
	LogQuery               bool   `env:"PG_LOG_QUERY" envDefault:"true"`
	LogAllQuery            bool   `env:"PG_LOG_ALL_QUERY" envDefault:"false"`
	ExportPromMetrics      bool   `env:"PG_EXPORT_PROM_METRICS" envDefault:"false"`
	QueryDurationThreshold int64  `env:"PG_QUERY_DUR_THRESHOLD" envDefault:"5000"`
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
		logger.Errorw("error in connecting db ", "db", obfuscateSecretTags(cfg), "err", err)
		return nil, err
	} else {
		logger.Infow("connected with db", "db", obfuscateSecretTags(cfg))
	}

	//--------------
	dbConnection.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
		queryDuration := time.Since(event.StartTime)

		// Expose prom metrics
		if cfg.ExportPromMetrics {
			middleware.PgQueryDuration.WithLabelValues("value").Observe(queryDuration.Seconds())
		}

		query, err := event.FormattedQuery()
		if err != nil {
			logger.Errorw("Error formatting query",
				"err", err)
			return
		}

		// Log pg query if enabled
		if cfg.LogAllQuery || (cfg.LogQuery && queryDuration.Milliseconds() > cfg.QueryDurationThreshold) {
			logger.Debugw("query time",
				"duration", queryDuration.Seconds(),
				"query", query)
		}
	})
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
