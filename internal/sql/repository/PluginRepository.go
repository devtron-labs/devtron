package repository

import (
	_ "github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type Plugin struct {
	tableName            struct{} `sql:"plugin_scripts" pg:",discard_unknown_columns"`
	Id                   int      `sql:"id,pk"`
	Name                 string   `sql:"name,notnull"`
	Description          string   `sql:"description,notnull"`
	Body                 string   `sql:"body,notnull"`
	StepTemplateLanguage string   `sql:"step_template_language,notnull"`
	StepTemplate         string   `sql:"step_template,notnull"`
}

type PluginInputs struct {
	tableName    struct{} `sql:"plugin_inputs" pg:",discard_unknown_columns"`
	Id           int      `sql:"plugin_id,notnull"`
	Name         string   `sql:"key_name,notnull"`
	DefaultValue string   `sql:"default_value,notnull"`
	Description  string   `sql:"plugin_key_description,notnull"`
}

type PluginRepository interface {
	Save(plugin *Plugin, inputs []*PluginInputs) error
	FindByAppId(pluginId int) (*Plugin, []*PluginInputs, error)
	Update(plugin *Plugin, inputs []*PluginInputs) error
	Delete(pluginId int) error
}

type PluginRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPluginRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *PluginRepositoryImpl {
	return &PluginRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl *PluginRepositoryImpl) Save(plugin *Plugin, inputs []*PluginInputs) error {
	check := impl.dbConnection.Insert(plugin)
	for _, input := range inputs {
		err := impl.dbConnection.Insert(input)
		if err != nil {
			impl.logger.Errorw("Plugin couldn't be saved", "err", err)
			return err
		}
	}
	return check
}

func (impl *PluginRepositoryImpl) FindByAppId(pluginId int) (*Plugin, []*PluginInputs, error) {
	plugin := &Plugin{}
	err := impl.dbConnection.Model(plugin).Where("id = ? ", pluginId).Select()
	if err != nil {
		impl.logger.Errorw("Plugin not found for given Id", "err", err)
	}
	var pluginInputs []*PluginInputs
	err = impl.dbConnection.Model(&pluginInputs).Where("plugin_id = ? ", pluginId).Select()
	if err != nil {
		impl.logger.Errorw("Plugin not found for given Id", "err", err)
	}
	return plugin, pluginInputs, err
}

func (impl *PluginRepositoryImpl) Update(plugin *Plugin, inputs []*PluginInputs) error {
	err := impl.dbConnection.Update(plugin)
	for _, input := range inputs {
		err := impl.dbConnection.Update(input)
		if err != nil {
			impl.logger.Errorw("Plugin couldn't be saved", "err", err)
		}
	}
	return err
}

func (impl *PluginRepositoryImpl) Delete(pluginId int) error {
	plugin, pluginInputs, err := impl.FindByAppId(pluginId)
	if err != nil {
		impl.logger.Errorw("Plugin not found for given Id", "err", err)
	}
	err = impl.dbConnection.Delete(plugin)
	if err != nil {
		impl.logger.Errorw("Plugin couldn't be deleted", "err", err)
	}
	err = impl.dbConnection.Delete(pluginInputs)
	if err != nil {
		impl.logger.Errorw("Plugin couldn't be deleted", "err", err)
	}
	return err
}
