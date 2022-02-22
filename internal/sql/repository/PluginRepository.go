package repository

import (
	_ "github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type Plugin struct {
	tableName   struct{} `sql:"plugin_scripts" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	Name        string   `sql:"name,notnull"`
	Description string   `sql:"description,notnull"`
}

type PluginRepository interface {
	Save(plugin *Plugin) error
	FindByAppId(pluginId int) (*Plugin, error)
	Update(plugin *Plugin) error
	Delete(plugin *Plugin) error
}

type PluginRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPluginRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *PluginRepositoryImpl {
	return &PluginRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl *PluginRepositoryImpl) Save(plugin *Plugin) error {
	return impl.dbConnection.Insert(plugin)
}

func (impl *PluginRepositoryImpl) FindByAppId(pluginId int) (*Plugin, error) {
	plugin := &Plugin{}
	err := impl.dbConnection.Model(plugin).Where("id = ? ", pluginId).Select()
	if err != nil {
		println(err)
	}
	return plugin, err
}

func (impl *PluginRepositoryImpl) Update(plugin *Plugin) error {
	err := impl.dbConnection.Update(plugin)
	return err
}

func (impl *PluginRepositoryImpl) Delete(plugin *Plugin) error {
	err := impl.dbConnection.Delete(plugin)
	return err
}
