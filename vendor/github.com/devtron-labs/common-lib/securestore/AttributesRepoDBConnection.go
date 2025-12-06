package securestore

import (
	"github.com/caarlos0/env"
	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/common-lib/utils/bean"
	"github.com/go-pg/pg"
	"log"
)

func NewAttributesRepositoryImplForDatabase(databaseName string) (*AttributesRepositoryImpl, error) {
	dbConn, err := newDbConnection(databaseName)
	if err != nil {
		return nil, err
	}
	return NewAttributesRepositoryImpl(dbConn), nil
}

type config struct {
	Addr            string `env:"PG_ADDR" envDefault:"127.0.0.1"`
	Port            string `env:"PG_PORT" envDefault:"5432"`
	User            string `env:"PG_USER" envDefault:""`
	Password        string `env:"PG_PASSWORD" envDefault:"" secretData:"-"`
	Database        string `env:"PG_DATABASE" envDefault:"orchestrator"`
	ApplicationName string `env:"APP" envDefault:"orchestrator"`
	bean.PgQueryMonitoringConfig
	LocalDev bool `env:"RUNTIME_CONFIG_LOCAL_DEV" envDefault:"false"`
}

func getDbConfig(databaseName string) (*config, error) {
	cfg := &config{}
	err := env.Parse(cfg)
	if err != nil {
		return cfg, err
	}
	monitoringCfg, err := bean.GetPgQueryMonitoringConfig(cfg.ApplicationName)
	if err != nil {
		return cfg, err
	}
	cfg.PgQueryMonitoringConfig = monitoringCfg
	if !cfg.LocalDev {
		cfg.Database = databaseName //overriding database
	}
	return cfg, err
}

func newDbConnection(databaseName string) (*pg.DB, error) {
	cfg, err := getDbConfig(databaseName)
	if err != nil {
		return nil, err
	}
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
	_, err = dbConnection.QueryOne(&test, `SELECT 1`)

	if err != nil {
		log.Println("error in connecting orchestrator db ", "err", err)
		return nil, err
	} else {
		log.Println("connected with orchestrator db")
	}
	//--------------
	if cfg.LogSlowQuery {
		dbConnection.OnQueryProcessed(utils.GetPGPostQueryProcessor(cfg.PgQueryMonitoringConfig))
	}
	return dbConnection, err
}
