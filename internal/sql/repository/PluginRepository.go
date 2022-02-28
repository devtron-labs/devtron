package repository

import (
	_ "github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type Plugin struct {
	tableName         struct{} `sql:"plugin_scripts" pg:",discard_unknown_columns"`
	Id                int      `sql:"id,pk"`
	PluginId          int      `sql:"plugin_id,notnull"`
	PluginName        string   `sql:"plugin_name,notnull"`
	PluginDescription string   `sql:"plugin_description,notnull"`
	PluginBody        string   `sql:"plugin_body,notnull"`
}

type PluginFields struct {
	tableName         struct{} `sql:"plugin_fields" pg:",discard_unknown_columns"`
	PluginId          int      `sql:"plugin_id,notnull"`
	StepId            int      `sql:"step_id,notnull"`
	FieldName         string   `sql:"key_name,notnull"`
	FieldDefaultValue string   `sql:"default_value,notnull"`
	FieldDescription  string   `sql:"plugin_key_description,notnull"`
	FieldType         string   `sql:"plugin_field_type,notnull"`
}

type Tags struct {
	tableName struct{} `sql:"plugin_tags" pg:",discard_unknown_columns"`
	TagId     int      `sql:"tag_id,pk"`
	TagName   string   `sql:"tag_name,notnull"`
}

type PluginTagsMap struct {
	tableName struct{} `sql:"plugin_tags_map" pg:",discard_unknown_columns"`
	TagId     int      `sql:"tag_id"`
	PluginId  int      `sql:"plugin_id"`
}

type PluginSteps struct {
	tableName             struct{} `sql:"plugin_steps" pg:",discard_unknown_columns"`
	StepId                int      `sql:"steps_id,pk"`
	StepName              string   `sql:"steps_name"`
	StepsTemplateLanguage string   `sql:"steps_template_language,notnull"`
	StepsTemplate         string   `sql:"steps_template,notnull"`
}

type PluginStepsSequence struct {
	tableName  struct{} `sql:"plugin_steps_sequence" pg:",discard_unknown_columns"`
	SequenceId int      `sql:"sequence_id,pk"`
	StepsId    int      `sql:"steps_id"`
	PluginId   int      `sql:"plugin_id"`
}

type PluginRepository interface {
	SavePlugin(plugin *Plugin) (*Plugin, error)
	SaveFields(inputs []*PluginFields) error
	SaveSteps(step *PluginSteps) (*PluginSteps, error)
	SaveTag(tagName *Tags) (*Tags, error)
	SaveStepsSequence(stepSeq *PluginStepsSequence) error
	SavePluginTagsMap(tagsMap *PluginTagsMap) error
	FindTagId(tagName string) (*Tags, error)
	FindByAppId(pluginId int) (*Plugin, []*PluginFields, []*PluginSteps, []string, error)
	Update(plugin *Plugin, inputs []*PluginFields) error
	Delete(pluginId int) error
}

type PluginRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPluginRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *PluginRepositoryImpl {
	return &PluginRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl *PluginRepositoryImpl) SaveFields(inputs []*PluginFields) error {
	for _, input := range inputs {
		err := impl.dbConnection.Insert(input)
		if err != nil {
			impl.logger.Errorw("Plugin couldn't be saved", "err", err)
			return err
		}
	}
	return nil
}

func (impl *PluginRepositoryImpl) SaveStepsSequence(stepSeq *PluginStepsSequence) error {
	return impl.dbConnection.Insert(stepSeq)
}

func (impl *PluginRepositoryImpl) SavePlugin(plugin *Plugin) (*Plugin, error) {
	err := impl.dbConnection.Insert(plugin)
	return plugin, err
}

func (impl *PluginRepositoryImpl) SaveSteps(step *PluginSteps) (*PluginSteps, error) {
	err := impl.dbConnection.Insert(step)
	return step, err
}

func (impl *PluginRepositoryImpl) SaveTag(tagName *Tags) (*Tags, error) {
	err := impl.dbConnection.Insert(tagName)
	return tagName, err
}

func (impl *PluginRepositoryImpl) SavePluginTagsMap(tagsMap *PluginTagsMap) error {
	return impl.dbConnection.Insert(tagsMap)
}

func (impl *PluginRepositoryImpl) FindByAppId(pluginId int) (*Plugin, []*PluginFields, []*PluginSteps, []string, error) {
	plugin := &Plugin{}
	err := impl.dbConnection.Model(plugin).Where("id = ? ", pluginId).Select()
	if err != nil {
		impl.logger.Errorw("Plugin not found for given Id", "err", err)
	}

	pluginfields, err := impl.FindPluginFieldsById(pluginId)
	if err != nil {
		impl.logger.Errorw("Plugin fields not found for given Id", "err", err)
	}

	var pluginsteps []*PluginSteps
	err = impl.dbConnection.Model(&pluginsteps).Where("plugin_id = ? ", pluginId).Select()
	if err != nil {
		impl.logger.Errorw("Plugin not found for given Id", "err", err)
	}

	tags, err := impl.FindTagsById(pluginId)

	return plugin, pluginfields, pluginsteps, tags, err
}

func (impl *PluginRepositoryImpl) FindTagsById(id int) ([]string, error) {
	var tagIdsMap []*PluginTagsMap
	err := impl.dbConnection.Model(&tagIdsMap).Where("plugin_id = ? ", id).Select()
	if err != nil {
		impl.logger.Errorw("Tag Ids couldn't be found", "err", err)
	}
	var tags []string
	for _, tagIdMap := range tagIdsMap {
		tag := &Tags{}
		err := impl.dbConnection.Model(tag).Where("tag_id = ? ", tagIdMap.TagId).Select()
		if err != nil {
			impl.logger.Errorw("Tags couldn't be found", "err", err)
		}
		tags = append(tags, tag.TagName)
	}
	return tags, err
}

func (impl *PluginRepositoryImpl) FindTagId(tagName string) (*Tags, error) {
	plugintag := &Tags{}
	err := impl.dbConnection.Model(plugintag).Where("tag_name = ? ", tagName).Select()
	return plugintag, err
}

func (impl *PluginRepositoryImpl) Update(plugin *Plugin, inputs []*PluginFields) error {
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
	plugin, _, pluginsteps, _, err := impl.FindByAppId(pluginId)
	if err != nil {
		impl.logger.Errorw("Plugin not found for given Id", "err", err)
	}
	err = impl.dbConnection.Delete(plugin)
	if err != nil {
		impl.logger.Errorw("Plugin couldn't be deleted", "err", err)
	}
	err = impl.DeleteFieldsByID(pluginId)
	if err != nil {
		impl.logger.Errorw("Plugin couldn't be deleted", "err", err)
	}
	err = impl.dbConnection.Delete(pluginsteps)
	if err != nil {
		impl.logger.Errorw("Plugin couldn't be deleted", "err", err)
	}

	for _, pluginstep := range pluginsteps {
		//stepfields, err := impl.FindPluginFieldsById(pluginstep.StepId)
		//if err != nil {
		//	impl.logger.Errorw("Plugin Steps fields couldn't be found", "err", err)
		//}
		err = impl.DeleteFieldsByID(pluginstep.StepId)
		if err != nil {
			impl.logger.Errorw("Plugin Steps fields couldn't be deleted", "err", err)
		}
	}

	return err
}

func (impl *PluginRepositoryImpl) FindPluginFieldsById(id int) ([]*PluginFields, error) {
	var pluginInputs []*PluginFields
	err := impl.dbConnection.Model(&pluginInputs).Where("plugin_id = ? ", id).Select()
	if err != nil {
		impl.logger.Errorw("Plugin not found for given Id", "err", err)
	}
	return pluginInputs, err
}

func (impl *PluginRepositoryImpl) DeleteFieldsByID(id int) error {
	plugin := &PluginFields{}
	_, err := impl.dbConnection.Model(plugin).Where("plugin_id = ? ", id).Delete()
	if err != nil {
		impl.logger.Errorw("Plugin couldn't be deleted", "err", err)
	}
	return err
}

func (impl *PluginRepositoryImpl) DeleteTagsMapByID(id int) error {
	plugin := &PluginTagsMap{}
	_, err := impl.dbConnection.Model(plugin).Where("plugin_id = ? ", id).Delete()
	if err != nil {
		impl.logger.Errorw("Plugin couldn't be deleted", "err", err)
	}
	return err
}
