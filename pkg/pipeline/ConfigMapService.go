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

package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/commonService"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"regexp"
	"time"
)

type ConfigMapRequest struct {
	Id            int             `json:"id"`
	AppId         int             `json:"app_id"`
	EnvironmentId int             `json:"environment_id"`
	PipelineId    int             `json:"pipeline_id"`
	ConfigMapData json.RawMessage `json:"config_map_data"`
	SecretData    json.RawMessage `json:"secret_data"`
	UserId        int32           `json:"-"`
}

type ConfigDataRequest struct {
	Id            int           `json:"id"`
	AppId         int           `json:"appId"`
	EnvironmentId int           `json:"environmentId,omitempty"`
	ConfigData    []*ConfigData `json:"configData"`
	UserId        int32         `json:"-"`
}

type BulkPatchRequest struct {
	Payload     []*BulkPatchPayload `json:"payload"`
	Filter      *BulkPatchFilter    `json:"filter,omitempty"`
	ProjectId   int                 `json:"projectId"`
	Global      bool                `json:"global"`
	Type        string              `json:"type"`
	Name        string              `json:"name"`
	Key         string              `json:"key"`
	Value       string              `json:"value"`
	PatchAction int                 `json:"patchAction"` // 1=add, 2=update, 0=delete
	UserId      int32               `json:"-"`
}

type BulkPatchPayload struct {
	AppId int `json:"appId"`
	EnvId int `json:"envId"`
}

type BulkPatchFilter struct {
	AppNameIncludes string `json:"appNameIncludes,omitempty"`
	AppNameExcludes string `json:"appNameExcludes,omitempty"`
	EnvId           int    `json:"envId,omitempty"`
}

type ExternalSecret struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Property string `json:"property,omitempty"`
	IsBinary bool   `json:"isBinary"`
}

type ConfigData struct {
	Name                  string           `json:"name"`
	Type                  string           `json:"type"`
	External              bool             `json:"external"`
	MountPath             string           `json:"mountPath,omitempty"`
	Data                  json.RawMessage  `json:"data"`
	DefaultData           json.RawMessage  `json:"defaultData,omitempty"`
	DefaultMountPath      string           `json:"defaultMountPath,omitempty"`
	Global                bool             `json:"global"`
	ExternalSecretType    string           `json:"externalType"`
	ExternalSecret        []ExternalSecret `json:"secretData"`
	DefaultExternalSecret []ExternalSecret `json:"defaultSecretData,omitempty"`
	RoleARN               string           `json:"roleARN"`
	SubPath               bool             `json:"subPath"`
	FilePermission        string           `json:"filePermission"`
}

const (
	KubernetesSecret  string = "KubernetesSecret"
	AWSSecretsManager string = "AWSSecretsManager"
	AWSSystemManager  string = "AWSSystemManager"
	HashiCorpVault    string = "HashiCorpVault"
)

type ConfigsList struct {
	ConfigData []*ConfigData `json:"maps"`
}
type SecretsList struct {
	ConfigData []*ConfigData `json:"secrets"`
}

type ConfigMapService interface {
	CMGlobalAddUpdate(configMapRequest *ConfigDataRequest) (*ConfigDataRequest, error)
	CMGlobalFetch(appId int) (*ConfigDataRequest, error)
	CMEnvironmentAddUpdate(configMapRequest *ConfigDataRequest) (*ConfigDataRequest, error)
	CMEnvironmentFetch(appId int, envId int) (*ConfigDataRequest, error)

	CSGlobalAddUpdate(configMapRequest *ConfigDataRequest) (*ConfigDataRequest, error)
	CSGlobalFetch(appId int) (*ConfigDataRequest, error)
	CSEnvironmentAddUpdate(configMapRequest *ConfigDataRequest) (*ConfigDataRequest, error)
	CSEnvironmentFetch(appId int, envId int) (*ConfigDataRequest, error)

	CMGlobalDelete(name string, id int, userId int32) (bool, error)
	CMEnvironmentDelete(name string, id int, userId int32) (bool, error)
	CSGlobalDelete(name string, id int, userId int32) (bool, error)
	CSEnvironmentDelete(name string, id int, userId int32) (bool, error)

	CMGlobalDeleteByAppId(name string, appId int, userId int32) (bool, error)
	CMEnvironmentDeleteByAppIdAndEnvId(name string, appId int, envId int, userId int32) (bool, error)
	CSGlobalDeleteByAppId(name string, appId int, userId int32) (bool, error)
	CSEnvironmentDeleteByAppIdAndEnvId(name string, appId int, envId int, userId int32) (bool, error)

	CSGlobalFetchForEdit(name string, id int, userId int32) (*ConfigDataRequest, error)
	CSEnvironmentFetchForEdit(name string, id int, appId int, envId int, userId int32) (*ConfigDataRequest, error)
	ConfigSecretGlobalBulkPatch(bulkPatchRequest *BulkPatchRequest) (*BulkPatchRequest, error)
	ConfigSecretEnvironmentBulkPatch(bulkPatchRequest *BulkPatchRequest) (*BulkPatchRequest, error)
}

type ConfigMapServiceImpl struct {
	chartRepository             chartConfig.ChartRepository
	logger                      *zap.SugaredLogger
	repoRepository              chartConfig.ChartRepoRepository
	mergeUtil                   util.MergeUtil
	pipelineConfigRepository    chartConfig.PipelineConfigRepository
	configMapRepository         chartConfig.ConfigMapRepository
	environmentConfigRepository chartConfig.EnvConfigOverrideRepository
	commonService               commonService.CommonService
	appRepository               pipelineConfig.AppRepository
}

func NewConfigMapServiceImpl(chartRepository chartConfig.ChartRepository,
	logger *zap.SugaredLogger,
	repoRepository chartConfig.ChartRepoRepository,
	mergeUtil util.MergeUtil,
	pipelineConfigRepository chartConfig.PipelineConfigRepository,
	configMapRepository chartConfig.ConfigMapRepository, environmentConfigRepository chartConfig.EnvConfigOverrideRepository,
	commonService commonService.CommonService, appRepository pipelineConfig.AppRepository) *ConfigMapServiceImpl {
	return &ConfigMapServiceImpl{
		chartRepository:             chartRepository,
		logger:                      logger,
		repoRepository:              repoRepository,
		mergeUtil:                   mergeUtil,
		pipelineConfigRepository:    pipelineConfigRepository,
		configMapRepository:         configMapRepository,
		environmentConfigRepository: environmentConfigRepository,
		commonService:               commonService,
		appRepository:               appRepository,
	}
}

func (impl ConfigMapServiceImpl) adapter(model *chartConfig.ConfigMapAppModel) (*ConfigMapRequest, error) {
	configMapRequest := &ConfigMapRequest{
		Id:            model.Id,
		AppId:         model.AppId,
		ConfigMapData: []byte(model.ConfigMapData),
		SecretData:    []byte(model.SecretData),
	}
	return configMapRequest, nil
}
func (impl ConfigMapServiceImpl) adapterEnv(model *chartConfig.ConfigMapEnvModel) (*ConfigMapRequest, error) {
	configMapRequest := &ConfigMapRequest{
		Id:            model.Id,
		AppId:         model.AppId,
		EnvironmentId: model.EnvironmentId,
		ConfigMapData: []byte(model.ConfigMapData),
		SecretData:    []byte(model.SecretData),
	}
	return configMapRequest, nil
}

func (impl ConfigMapServiceImpl) CMGlobalAddUpdate(configMapRequest *ConfigDataRequest) (*ConfigDataRequest, error) {
	if len(configMapRequest.ConfigData) != 1 {
		return nil, fmt.Errorf("invalid request multiple config found for add or update")
	}
	configData := configMapRequest.ConfigData[0]
	valid, err := impl.validateConfigData(configData)
	if err != nil && !valid {
		impl.logger.Errorw("error in validating", "error", err)
		return configMapRequest, err
	}
	if configMapRequest.Id > 0 {
		model, err := impl.configMapRepository.GetByIdAppLevel(configMapRequest.Id)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
		configsList := &ConfigsList{}
		found := false
		var configs []*ConfigData
		if len(model.ConfigMapData) > 0 {
			err = json.Unmarshal([]byte(model.ConfigMapData), configsList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "error", err)
			}
		}
		for _, item := range configsList.ConfigData {
			if item.Name == configData.Name {
				item.Data = configData.Data
				item.MountPath = configData.MountPath
				item.Type = configData.Type
				item.External = configData.External
				item.ExternalSecretType = configData.ExternalSecretType
				found = true
				item.SubPath = configData.SubPath
				item.FilePermission = configData.FilePermission
			}
			configs = append(configs, item)
		}

		if !found {
			configs = append(configs, configData)
		}
		configsList.ConfigData = configs
		configDataByte, err := json.Marshal(configsList)
		if err != nil {
			return nil, err
		}
		model.ConfigMapData = string(configDataByte)
		model.UpdatedBy = configMapRequest.UserId
		model.UpdatedOn = time.Now()

		configMap, err := impl.configMapRepository.UpdateAppLevel(model)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
		configMapRequest.Id = configMap.Id

	} else {
		//creating config map record for first time
		configsList := &ConfigsList{
			ConfigData: configMapRequest.ConfigData,
		}
		configDataByte, err := json.Marshal(configsList)
		if err != nil {
			return nil, err
		}
		model := &chartConfig.ConfigMapAppModel{
			AppId:         configMapRequest.AppId,
			ConfigMapData: string(configDataByte),
		}
		model.CreatedBy = configMapRequest.UserId
		model.UpdatedBy = configMapRequest.UserId
		model.CreatedOn = time.Now()
		model.UpdatedOn = time.Now()

		configMap, err := impl.configMapRepository.CreateAppLevel(model)
		if err != nil {
			impl.logger.Errorw("error while creating app level", "error", err)
			return nil, err
		}
		configMapRequest.Id = configMap.Id
	}

	return configMapRequest, nil
}

func (impl ConfigMapServiceImpl) CMGlobalFetch(appId int) (*ConfigDataRequest, error) {
	configMapGlobal, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
	if err != nil && pg.ErrNoRows != err {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	if pg.ErrNoRows == err {
		impl.logger.Debugw("no config map data found for this request", "appId", appId)
	}

	configMapGlobalList := &ConfigsList{}
	if len(configMapGlobal.ConfigMapData) > 0 {
		err = json.Unmarshal([]byte(configMapGlobal.ConfigMapData), configMapGlobalList)
		if err != nil {
			impl.logger.Debugw("error while Unmarshal", "error", err)
		}
	}
	configDataRequest := &ConfigDataRequest{}
	configDataRequest.Id = configMapGlobal.Id
	configDataRequest.AppId = appId
	//configDataRequest.ConfigData = configMapGlobalList.ConfigData
	for _, item := range configMapGlobalList.ConfigData {
		item.Global = true
		configDataRequest.ConfigData = append(configDataRequest.ConfigData, item)
	}
	if configDataRequest.ConfigData == nil {
		list := []*ConfigData{}
		configDataRequest.ConfigData = list
	} else {
		//configDataRequest.ConfigData = configMapGlobalList.ConfigData
	}

	return configDataRequest, nil
}

func (impl ConfigMapServiceImpl) CMEnvironmentAddUpdate(configMapRequest *ConfigDataRequest) (*ConfigDataRequest, error) {

	if len(configMapRequest.ConfigData) != 1 {
		return nil, fmt.Errorf("invalid request multiple config found for add or update")
	}
	configData := configMapRequest.ConfigData[0]
	valid, err := impl.validateConfigData(configData)
	if err != nil && !valid {
		impl.logger.Errorw("error in validating", "error", err)
		return configMapRequest, err
	}
	if configMapRequest.Id > 0 {
		model, err := impl.configMapRepository.GetByIdEnvLevel(configMapRequest.Id)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
		configsList := &ConfigsList{}
		found := false
		var configs []*ConfigData
		if len(model.ConfigMapData) > 0 {
			err = json.Unmarshal([]byte(model.ConfigMapData), configsList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "error", err)
			}
		}
		for _, item := range configsList.ConfigData {
			if item.Name == configData.Name {
				item.Data = configData.Data
				item.MountPath = configData.MountPath
				item.Type = configData.Type
				item.External = configData.External
				item.ExternalSecretType = configData.ExternalSecretType
				item.SubPath = configData.SubPath
				item.FilePermission = configData.FilePermission
				found = true
			}
			configs = append(configs, item)
		}

		if !found {
			configs = append(configs, configData)
		}
		configsList.ConfigData = configs
		configDataByte, err := json.Marshal(configsList)
		if err != nil {
			return nil, err
		}
		model.ConfigMapData = string(configDataByte)
		model.UpdatedBy = configMapRequest.UserId
		model.UpdatedOn = time.Now()

		configMap, err := impl.configMapRepository.UpdateEnvLevel(model)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
		configMapRequest.Id = configMap.Id

	} else {
		//creating config map record for first time
		configsList := &ConfigsList{
			ConfigData: configMapRequest.ConfigData,
		}
		configDataByte, err := json.Marshal(configsList)
		if err != nil {
			return nil, err
		}
		model := &chartConfig.ConfigMapEnvModel{
			AppId:         configMapRequest.AppId,
			EnvironmentId: configMapRequest.EnvironmentId,
			ConfigMapData: string(configDataByte),
		}
		model.CreatedBy = configMapRequest.UserId
		model.UpdatedBy = configMapRequest.UserId
		model.CreatedOn = time.Now()
		model.UpdatedOn = time.Now()

		configMap, err := impl.configMapRepository.CreateEnvLevel(model)
		if err != nil {
			impl.logger.Errorw("error while creating app level", "error", err)
			return nil, err
		}
		configMapRequest.Id = configMap.Id
	}

	return configMapRequest, nil
}

func (impl ConfigMapServiceImpl) CMEnvironmentFetch(appId int, envId int) (*ConfigDataRequest, error) {
	configMapGlobal, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
	if err != nil && pg.ErrNoRows != err {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	if pg.ErrNoRows == err {
		impl.logger.Debugw("no config map data found for this request", "appId", appId)
	}
	configMapGlobalList := &ConfigsList{}
	if len(configMapGlobal.ConfigMapData) > 0 {
		err = json.Unmarshal([]byte(configMapGlobal.ConfigMapData), configMapGlobalList)
		if err != nil {
			impl.logger.Errorw("error while Unmarshal", "error", err)
		}
	}
	configMapEnvLevel, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
	if err != nil && pg.ErrNoRows != err {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	if pg.ErrNoRows == err {
		impl.logger.Debugw("no config map data found for this request", "appId", appId)
	}
	configsListEnvLevel := &ConfigsList{}
	if len(configMapEnvLevel.ConfigMapData) > 0 {
		err = json.Unmarshal([]byte(configMapEnvLevel.ConfigMapData), configsListEnvLevel)
		if err != nil {
			impl.logger.Debugw("error while Unmarshal", "error", err)
		}
	}
	configDataRequest := &ConfigDataRequest{}
	configDataRequest.Id = configMapEnvLevel.Id
	configDataRequest.AppId = appId
	configDataRequest.EnvironmentId = envId
	//configDataRequest.ConfigData = configsListEnvLevel.ConfigData

	kv1 := make(map[string]json.RawMessage)
	kv11 := make(map[string]string)
	kv2 := make(map[string]json.RawMessage)
	for _, item := range configMapGlobalList.ConfigData {
		kv1[item.Name] = item.Data
		kv11[item.Name] = item.MountPath
	}
	for _, item := range configsListEnvLevel.ConfigData {
		kv2[item.Name] = item.Data
	}

	//add those items which are in global only
	for _, item := range configMapGlobalList.ConfigData {
		if _, ok := kv2[item.Name]; !ok {
			item.Global = true
			item.DefaultData = item.Data
			item.DefaultMountPath = item.MountPath
			item.Data = nil
			item.MountPath = ""
			item.SubPath = item.SubPath
			item.FilePermission = item.FilePermission
			configDataRequest.ConfigData = append(configDataRequest.ConfigData, item)
		}
	}

	//add all the items from environment level, add default data to items which are override from global
	for _, item := range configsListEnvLevel.ConfigData {
		if val, ok := kv1[item.Name]; ok {
			item.DefaultData = val
			item.DefaultMountPath = kv11[item.Name]
			item.Global = true
			configDataRequest.ConfigData = append(configDataRequest.ConfigData, item)
		} else {
			configDataRequest.ConfigData = append(configDataRequest.ConfigData, item)
		}
	}

	if configDataRequest.ConfigData == nil {
		list := []*ConfigData{}
		configDataRequest.ConfigData = list
	} else {
		//configDataRequest.ConfigData = configMapGlobalList.ConfigData
	}

	return configDataRequest, nil
}

// ---------------------------------------------------------------------------------------------

func (impl ConfigMapServiceImpl) CSGlobalAddUpdate(configMapRequest *ConfigDataRequest) (*ConfigDataRequest, error) {
	if len(configMapRequest.ConfigData) != 1 {
		return nil, fmt.Errorf("invalid request multiple config found for add or update")
	}
	configData := configMapRequest.ConfigData[0]
	valid, err := impl.validateConfigData(configData)
	if err != nil && !valid {
		impl.logger.Errorw("error in validating", "error", err)
		return configMapRequest, err
	}

	valid, err = impl.validateExternalSecretChartCompatibility(configMapRequest.AppId, configMapRequest.EnvironmentId, configData)
	if err != nil && !valid {
		impl.logger.Errorw("error in validating", "error", err)
		return configMapRequest, err
	}

	if configMapRequest.Id > 0 {
		model, err := impl.configMapRepository.GetByIdAppLevel(configMapRequest.Id)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
		secretsList := &SecretsList{}
		found := false
		var configs []*ConfigData
		if len(model.SecretData) > 0 {
			err = json.Unmarshal([]byte(model.SecretData), secretsList)
			if err != nil {
				impl.logger.Debugw("error while Unmarshal", "error", err)
			}
		}
		for _, item := range secretsList.ConfigData {
			if item.Name == configData.Name {
				item.Data = configData.Data
				item.MountPath = configData.MountPath
				item.Type = configData.Type
				item.External = configData.External
				item.ExternalSecretType = configData.ExternalSecretType
				item.ExternalSecret = configData.ExternalSecret
				item.RoleARN = configData.RoleARN
				found = true
				item.SubPath = configData.SubPath
				item.FilePermission = configData.FilePermission
			}
			configs = append(configs, item)
		}

		if !found {
			configs = append(configs, configData)
		}
		secretsList.ConfigData = configs
		configDataByte, err := json.Marshal(secretsList)
		if err != nil {
			return nil, err
		}
		model.SecretData = string(configDataByte)
		model.UpdatedBy = configMapRequest.UserId
		model.UpdatedOn = time.Now()

		secret, err := impl.configMapRepository.UpdateAppLevel(model)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
		configMapRequest.Id = secret.Id

	} else {
		//creating config map record for first time
		secretsList := &SecretsList{
			ConfigData: configMapRequest.ConfigData,
		}
		secretDataByte, err := json.Marshal(secretsList)
		if err != nil {
			return nil, err
		}
		model := &chartConfig.ConfigMapAppModel{
			AppId:      configMapRequest.AppId,
			SecretData: string(secretDataByte),
		}
		model.CreatedBy = configMapRequest.UserId
		model.UpdatedBy = configMapRequest.UserId
		model.CreatedOn = time.Now()
		model.UpdatedOn = time.Now()

		secret, err := impl.configMapRepository.CreateAppLevel(model)
		if err != nil {
			impl.logger.Errorw("error while creating app level", "error", err)
			return nil, err
		}
		configMapRequest.Id = secret.Id
	}

	return configMapRequest, nil
}

func (impl ConfigMapServiceImpl) CSGlobalFetch(appId int) (*ConfigDataRequest, error) {
	configMapGlobal, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
	if err != nil && pg.ErrNoRows != err {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	if pg.ErrNoRows == err {
		impl.logger.Warnw("no app level secret found for this request", "appId", appId)
	}

	configMapGlobalList := &SecretsList{}
	if len(configMapGlobal.SecretData) > 0 {
		err = json.Unmarshal([]byte(configMapGlobal.SecretData), configMapGlobalList)
		if err != nil {
			impl.logger.Warnw("error while Unmarshal", "error", err)
		}
	}
	configDataRequest := &ConfigDataRequest{}
	configDataRequest.Id = configMapGlobal.Id
	configDataRequest.AppId = appId
	//configDataRequest.ConfigData = configMapGlobalList.ConfigData

	for _, item := range configMapGlobalList.ConfigData {
		item.Global = true
		configDataRequest.ConfigData = append(configDataRequest.ConfigData, item)
	}

	//removing actual values
	var configs []*ConfigData
	for _, item := range configDataRequest.ConfigData {
		resultMap := make(map[string]string)
		resultMapFinal := make(map[string]string)

		if item.Data != nil {
			err = json.Unmarshal(item.Data, &resultMap)
			if err != nil {
				impl.logger.Warnw("unmarshal failed: ", "error", err)
				configs = append(configs, item)
				continue
			}
			for k := range resultMap {
				resultMapFinal[k] = ""
			}
			resultByte, err := json.Marshal(resultMapFinal)
			if err != nil {
				impl.logger.Errorw("error while marshaling request ", "err", err)
				return nil, err
			}
			item.Data = resultByte
		}

		var externalSecret []ExternalSecret
		if item.ExternalSecret != nil && len(item.ExternalSecret) > 0 {
			for _, es := range item.ExternalSecret {
				externalSecret = append(externalSecret, ExternalSecret{Key: es.Key, Name: es.Name, Property: es.Property, IsBinary: es.IsBinary})
			}
		}
		item.ExternalSecret = externalSecret
		configs = append(configs, item)
	}
	configDataRequest.ConfigData = configs

	if configDataRequest.ConfigData == nil {
		list := []*ConfigData{}
		configDataRequest.ConfigData = list
	} else {
		//configDataRequest.ConfigData = configMapGlobalList.ConfigData
	}

	return configDataRequest, nil
}

func (impl ConfigMapServiceImpl) CSEnvironmentAddUpdate(configMapRequest *ConfigDataRequest) (*ConfigDataRequest, error) {
	if len(configMapRequest.ConfigData) != 1 {
		return nil, fmt.Errorf("invalid request multiple config found for add or update")
	}

	configData := configMapRequest.ConfigData[0]
	valid, err := impl.validateConfigData(configData)
	if err != nil && !valid {
		impl.logger.Errorw("error in validating", "error", err)
		return configMapRequest, err
	}

	valid, err = impl.validateExternalSecretChartCompatibility(configMapRequest.AppId, configMapRequest.EnvironmentId, configData)
	if err != nil && !valid {
		impl.logger.Errorw("error in validating", "error", err)
		return configMapRequest, err
	}

	if configMapRequest.Id > 0 {
		model, err := impl.configMapRepository.GetByIdEnvLevel(configMapRequest.Id)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
		configsList := &SecretsList{}
		found := false
		var configs []*ConfigData
		if len(model.SecretData) > 0 {
			err = json.Unmarshal([]byte(model.SecretData), configsList)
			if err != nil {
				impl.logger.Warnw("error while Unmarshal", "error", err)
			}
		}
		for _, item := range configsList.ConfigData {
			if item.Name == configData.Name {
				item.Data = configData.Data
				item.MountPath = configData.MountPath
				item.Type = configData.Type
				item.External = configData.External
				item.ExternalSecretType = configData.ExternalSecretType
				item.ExternalSecret = configData.ExternalSecret
				item.RoleARN = configData.RoleARN
				item.SubPath = configData.SubPath
				item.FilePermission = configData.FilePermission
				found = true
			}
			configs = append(configs, item)
		}

		if !found {
			configs = append(configs, configData)
		}
		configsList.ConfigData = configs
		secretDataByte, err := json.Marshal(configsList)
		if err != nil {
			return nil, err
		}
		model.SecretData = string(secretDataByte)
		model.UpdatedBy = configMapRequest.UserId
		model.UpdatedOn = time.Now()

		configMap, err := impl.configMapRepository.UpdateEnvLevel(model)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
		configMapRequest.Id = configMap.Id

	} else {
		//creating config map record for first time
		secretsList := &SecretsList{
			ConfigData: configMapRequest.ConfigData,
		}
		secretDataByte, err := json.Marshal(secretsList)
		if err != nil {
			return nil, err
		}
		model := &chartConfig.ConfigMapEnvModel{
			AppId:         configMapRequest.AppId,
			EnvironmentId: configMapRequest.EnvironmentId,
			SecretData:    string(secretDataByte),
		}
		model.CreatedBy = configMapRequest.UserId
		model.UpdatedBy = configMapRequest.UserId
		model.CreatedOn = time.Now()
		model.UpdatedOn = time.Now()

		configMap, err := impl.configMapRepository.CreateEnvLevel(model)
		if err != nil {
			impl.logger.Errorw("error while creating app level", "error", err)
			return nil, err
		}
		configMapRequest.Id = configMap.Id
	}

	return configMapRequest, nil
}

func (impl ConfigMapServiceImpl) CSEnvironmentFetch(appId int, envId int) (*ConfigDataRequest, error) {
	configMapGlobal, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
	if err != nil && pg.ErrNoRows != err {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	if pg.ErrNoRows == err {
		impl.logger.Warnw("no app level secret found for this request", "appId", appId)
	}
	configMapGlobalList := &SecretsList{}
	if len(configMapGlobal.SecretData) > 0 {
		err = json.Unmarshal([]byte(configMapGlobal.SecretData), configMapGlobalList)
		if err != nil {
			impl.logger.Warnw("error while Unmarshal", "error", err)
		}
	}

	configMapEnvLevel, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
	if err != nil && pg.ErrNoRows != err {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	if pg.ErrNoRows == err {
		impl.logger.Warnw("no env level secret found for this request", "appId", appId)
	}
	configsListEnvLevel := &SecretsList{}
	if len(configMapEnvLevel.SecretData) > 0 {
		err = json.Unmarshal([]byte(configMapEnvLevel.SecretData), configsListEnvLevel)
		if err != nil {
			impl.logger.Warnw("error while Unmarshal", "error", err)
		}
	}
	configDataRequest := &ConfigDataRequest{}
	configDataRequest.Id = configMapEnvLevel.Id
	configDataRequest.AppId = appId
	configDataRequest.EnvironmentId = envId

	//configDataRequest.ConfigData = configsListEnvLevel.ConfigData
	//var configs []ConfigData
	kv1 := make(map[string]json.RawMessage)
	kv11 := make(map[string]string)
	kv2 := make(map[string]json.RawMessage)

	kv1External := make(map[string][]ExternalSecret)
	kv2External := make(map[string][]ExternalSecret)

	for _, item := range configMapGlobalList.ConfigData {
		kv1[item.Name] = item.Data
		kv11[item.Name] = item.MountPath
		kv1External[item.Name] = item.ExternalSecret
	}
	for _, item := range configsListEnvLevel.ConfigData {
		kv2[item.Name] = item.Data
		kv2External[item.Name] = item.ExternalSecret
	}

	//add those items which are in global only
	for _, item := range configMapGlobalList.ConfigData {
		if _, ok := kv2[item.Name]; !ok {
			item.Global = true
			if item.Data != nil && item.ExternalSecret == nil {
				item.DefaultData = item.Data
			} else if item.ExternalSecret != nil {
				/*bytes, err := json.Marshal(item.ExternalSecret)
				if err != nil {

				}
				item.DefaultData = bytes*/
				item.DefaultExternalSecret = item.ExternalSecret
			}
			item.DefaultMountPath = item.MountPath
			item.Data = nil
			item.ExternalSecret = nil
			item.MountPath = ""
			item.SubPath = item.SubPath
			item.FilePermission = item.FilePermission
			configDataRequest.ConfigData = append(configDataRequest.ConfigData, item)
		}
	}

	//add all the items from environment level, add default data to items which are override from global
	for _, item := range configsListEnvLevel.ConfigData {
		if val, ok := kv1[item.Name]; ok {
			item.DefaultData = val
			item.DefaultExternalSecret = kv1External[item.Name]
			item.DefaultMountPath = kv11[item.Name]
			item.Global = true
			configDataRequest.ConfigData = append(configDataRequest.ConfigData, item)
		} else {
			configDataRequest.ConfigData = append(configDataRequest.ConfigData, item)
		}
	}

	//removing actual values
	var configs []*ConfigData
	for _, item := range configDataRequest.ConfigData {

		if item.Data != nil {
			resultMap := make(map[string]string)
			resultMapFinal := make(map[string]string)
			err = json.Unmarshal(item.Data, &resultMap)
			if err != nil {
				impl.logger.Warnw("unmarshal failed: ", "error", err)
				//item.Data = []byte("[]")
				configs = append(configs, item)
				continue
				//return nil, err
			}
			for k := range resultMap {
				resultMapFinal[k] = ""
			}
			resultByte, err := json.Marshal(resultMapFinal)
			if err != nil {
				impl.logger.Errorw("error while marshaling request ", "err", err)
				return nil, err
			}
			item.Data = resultByte
		}

		var externalSecret []ExternalSecret
		if item.ExternalSecret != nil && len(item.ExternalSecret) > 0 {
			for _, es := range item.ExternalSecret {
				externalSecret = append(externalSecret, ExternalSecret{Key: es.Key, Name: es.Name, Property: es.Property, IsBinary: es.IsBinary})
			}
		}
		item.ExternalSecret = externalSecret

		if item.DefaultData != nil {
			resultMap := make(map[string]string)
			resultMapFinal := make(map[string]string)
			err = json.Unmarshal(item.DefaultData, &resultMap)
			if err != nil {
				impl.logger.Warnw("unmarshal failed: ", "error", err)
				//item.Data = []byte("[]")
				configs = append(configs, item)
				continue
				//return nil, err
			}
			for k := range resultMap {
				resultMapFinal[k] = ""
			}
			resultByte, err := json.Marshal(resultMapFinal)
			if err != nil {
				impl.logger.Errorw("error while marshaling request ", "err", err)
				return nil, err
			}
			item.DefaultData = resultByte
		}

		if item.DefaultExternalSecret != nil {
			var externalSecret []ExternalSecret
			if item.DefaultExternalSecret != nil && len(item.DefaultExternalSecret) > 0 {
				for _, es := range item.DefaultExternalSecret {
					externalSecret = append(externalSecret, ExternalSecret{Key: es.Key, Name: es.Name, Property: es.Property, IsBinary: es.IsBinary})
				}
			}
			item.DefaultExternalSecret = externalSecret
		}
		configs = append(configs, item)
	}
	configDataRequest.ConfigData = configs

	if configDataRequest.ConfigData == nil {
		list := []*ConfigData{}
		configDataRequest.ConfigData = list
	} else {
		//configDataRequest.ConfigData = configMapGlobalList.ConfigData
	}
	return configDataRequest, nil
}

func (impl ConfigMapServiceImpl) CMGlobalDelete(name string, id int, userId int32) (bool, error) {

	model, err := impl.configMapRepository.GetByIdAppLevel(id)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return false, err
	}
	configsList := &ConfigsList{}
	found := false
	var configs []*ConfigData
	if len(model.ConfigMapData) > 0 {
		err = json.Unmarshal([]byte(model.ConfigMapData), configsList)
		if err != nil {
			impl.logger.Warnw("error while Unmarshal", "error", err)
		}
	}
	for _, item := range configsList.ConfigData {
		if item.Name == name {
			found = true
		} else {
			configs = append(configs, item)
		}
	}

	if found {
		configsList.ConfigData = configs
		configDataByte, err := json.Marshal(configsList)
		if err != nil {
			return false, err
		}
		model.ConfigMapData = string(configDataByte)
		model.UpdatedBy = userId
		model.UpdatedOn = time.Now()

		_, err = impl.configMapRepository.UpdateAppLevel(model)
		if err != nil {
			impl.logger.Errorw("error while updating at app level", "error", err)
			return false, err
		}
	} else {
		impl.logger.Debugw("no config map found for delete with this name", "name", name)

	}

	return true, nil
}

func (impl ConfigMapServiceImpl) CMEnvironmentDelete(name string, id int, userId int32) (bool, error) {

	model, err := impl.configMapRepository.GetByIdEnvLevel(id)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return false, err
	}
	configsList := &ConfigsList{}
	found := false
	var configs []*ConfigData
	if len(model.ConfigMapData) > 0 {
		err = json.Unmarshal([]byte(model.ConfigMapData), configsList)
		if err != nil {
			impl.logger.Warnw("error while Unmarshal", "error", err)
		}
	}
	for _, item := range configsList.ConfigData {
		if item.Name == name {
			found = true
		} else {
			configs = append(configs, item)
		}
	}

	if found {
		configsList.ConfigData = configs
		configDataByte, err := json.Marshal(configsList)
		if err != nil {
			return false, err
		}
		model.ConfigMapData = string(configDataByte)
		model.UpdatedBy = userId
		model.UpdatedOn = time.Now()

		_, err = impl.configMapRepository.UpdateEnvLevel(model)
		if err != nil {
			impl.logger.Errorw("error while updating at env level", "error", err)
			return false, err
		}
	} else {
		impl.logger.Debugw("no config map found for delete with this name", "name", name)
	}

	return true, nil
}

func (impl ConfigMapServiceImpl) CSGlobalDelete(name string, id int, userId int32) (bool, error) {

	model, err := impl.configMapRepository.GetByIdAppLevel(id)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return false, err
	}
	configsList := &SecretsList{}
	found := false
	var configs []*ConfigData
	if len(model.SecretData) > 0 {
		err = json.Unmarshal([]byte(model.SecretData), configsList)
		if err != nil {
			impl.logger.Debugw("error while Unmarshal", "error", err)
		}
	}
	for _, item := range configsList.ConfigData {
		if item.Name == name {
			found = true
		} else {
			configs = append(configs, item)
		}
	}

	if found {
		configsList.ConfigData = configs
		configDataByte, err := json.Marshal(configsList)
		if err != nil {
			return false, err
		}
		model.SecretData = string(configDataByte)
		model.UpdatedBy = userId
		model.UpdatedOn = time.Now()

		_, err = impl.configMapRepository.UpdateAppLevel(model)
		if err != nil {
			impl.logger.Errorw("error while updating at app level", "error", err)
			return false, err
		}
	} else {
		impl.logger.Debugw("no config map found for delete with this name", "name", name)

	}

	return true, nil
}

func (impl ConfigMapServiceImpl) CSEnvironmentDelete(name string, id int, userId int32) (bool, error) {

	model, err := impl.configMapRepository.GetByIdEnvLevel(id)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return false, err
	}
	configsList := &SecretsList{}
	found := false
	var configs []*ConfigData
	if len(model.SecretData) > 0 {
		err = json.Unmarshal([]byte(model.SecretData), configsList)
		if err != nil {
			impl.logger.Warnw("error while Unmarshal", "error", err)
		}
	}
	for _, item := range configsList.ConfigData {
		if item.Name == name {
			found = true
		} else {
			configs = append(configs, item)
		}
	}

	if found {
		configsList.ConfigData = configs
		configDataByte, err := json.Marshal(configsList)
		if err != nil {
			return false, err
		}
		model.SecretData = string(configDataByte)
		model.UpdatedBy = userId
		model.UpdatedOn = time.Now()

		_, err = impl.configMapRepository.UpdateEnvLevel(model)
		if err != nil {
			impl.logger.Errorw("error while updating at env level ", "error", err)
			return false, err
		}
	} else {
		impl.logger.Debugw("no config map found for delete with this name", "name", name)
	}

	return true, nil
}

/////

func (impl ConfigMapServiceImpl) CMGlobalDeleteByAppId(name string, appId int, userId int32) (bool, error) {

	model, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return false, err
	}
	configsList := &ConfigsList{}
	found := false
	var configs []*ConfigData
	if len(model.ConfigMapData) > 0 {
		err = json.Unmarshal([]byte(model.ConfigMapData), configsList)
		if err != nil {
			impl.logger.Warnw("error while Unmarshal", "error", err)
		}
	}
	for _, item := range configsList.ConfigData {
		if item.Name == name {
			found = true
		} else {
			configs = append(configs, item)
		}
	}

	if found {
		configsList.ConfigData = configs
		configDataByte, err := json.Marshal(configsList)
		if err != nil {
			return false, err
		}
		model.ConfigMapData = string(configDataByte)
		model.UpdatedBy = userId
		model.UpdatedOn = time.Now()

		_, err = impl.configMapRepository.UpdateAppLevel(model)
		if err != nil {
			impl.logger.Errorw("error while updating at app level", "error", err)
			return false, err
		}
	} else {
		impl.logger.Debugw("no config map found for delete with this name", "name", name)

	}

	return true, nil
}

func (impl ConfigMapServiceImpl) CMEnvironmentDeleteByAppIdAndEnvId(name string, appId int, envId int, userId int32) (bool, error) {

	model, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return false, err
	}
	configsList := &ConfigsList{}
	found := false
	var configs []*ConfigData
	if len(model.ConfigMapData) > 0 {
		err = json.Unmarshal([]byte(model.ConfigMapData), configsList)
		if err != nil {
			impl.logger.Warnw("error while Unmarshal", "error", err)
		}
	}
	for _, item := range configsList.ConfigData {
		if item.Name == name {
			found = true
		} else {
			configs = append(configs, item)
		}
	}

	if found {
		configsList.ConfigData = configs
		configDataByte, err := json.Marshal(configsList)
		if err != nil {
			return false, err
		}
		model.ConfigMapData = string(configDataByte)
		model.UpdatedBy = userId
		model.UpdatedOn = time.Now()

		_, err = impl.configMapRepository.UpdateEnvLevel(model)
		if err != nil {
			impl.logger.Errorw("error while updating at env level", "error", err)
			return false, err
		}
	} else {
		impl.logger.Debugw("no config map found for delete with this name", "name", name)
	}

	return true, nil
}

func (impl ConfigMapServiceImpl) CSGlobalDeleteByAppId(name string, appId int, userId int32) (bool, error) {

	model, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return false, err
	}
	configsList := &SecretsList{}
	found := false
	var configs []*ConfigData
	if len(model.SecretData) > 0 {
		err = json.Unmarshal([]byte(model.SecretData), configsList)
		if err != nil {
			impl.logger.Warnw("error while Unmarshal", "error", err)
		}
	}
	for _, item := range configsList.ConfigData {
		if item.Name == name {
			found = true
		} else {
			configs = append(configs, item)
		}
	}

	if found {
		configsList.ConfigData = configs
		configDataByte, err := json.Marshal(configsList)
		if err != nil {
			return false, err
		}
		model.SecretData = string(configDataByte)
		model.UpdatedBy = userId
		model.UpdatedOn = time.Now()

		_, err = impl.configMapRepository.UpdateAppLevel(model)
		if err != nil {
			impl.logger.Errorw("error while updating at app level", "error", err)
			return false, err
		}
	} else {
		impl.logger.Debugw("no config map found for delete with this name", "name", name)

	}

	return true, nil
}

func (impl ConfigMapServiceImpl) CSEnvironmentDeleteByAppIdAndEnvId(name string, appId int, envId int, userId int32) (bool, error) {

	model, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return false, err
	}
	configsList := &SecretsList{}
	found := false
	var configs []*ConfigData
	if len(model.SecretData) > 0 {
		err = json.Unmarshal([]byte(model.SecretData), configsList)
		if err != nil {
			impl.logger.Warnw("error while Unmarshal", "error", err)
		}
	}
	for _, item := range configsList.ConfigData {
		if item.Name == name {
			found = true
		} else {
			configs = append(configs, item)
		}
	}

	if found {
		configsList.ConfigData = configs
		configDataByte, err := json.Marshal(configsList)
		if err != nil {
			return false, err
		}
		model.SecretData = string(configDataByte)
		model.UpdatedBy = userId
		model.UpdatedOn = time.Now()

		_, err = impl.configMapRepository.UpdateEnvLevel(model)
		if err != nil {
			impl.logger.Errorw("error while updating at env level ", "error", err)
			return false, err
		}
	} else {
		impl.logger.Debugw("no config map found for delete with this name", "name", name)
	}

	return true, nil
}

////

func (impl ConfigMapServiceImpl) CSGlobalFetchForEdit(name string, id int, userId int32) (*ConfigDataRequest, error) {
	configMapEnvLevel, err := impl.configMapRepository.GetByIdAppLevel(id)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}

	configsList := &SecretsList{}
	var configs []*ConfigData
	if len(configMapEnvLevel.SecretData) > 0 {
		err = json.Unmarshal([]byte(configMapEnvLevel.SecretData), configsList)
		if err != nil {
			impl.logger.Warnw("error while Unmarshal", "error", err)
		}
	}
	for _, item := range configsList.ConfigData {
		if item.Name == name {
			configs = append(configs, item)
			break
		}
	}

	configDataRequest := &ConfigDataRequest{}
	configDataRequest.Id = configMapEnvLevel.Id
	configDataRequest.AppId = configMapEnvLevel.AppId
	configDataRequest.ConfigData = configs
	return configDataRequest, nil
}

func (impl ConfigMapServiceImpl) CSEnvironmentFetchForEdit(name string, id int, appId int, envId int, userId int32) (*ConfigDataRequest, error) {
	configDataRequest := &ConfigDataRequest{}
	configsList := &SecretsList{}
	var configs []*ConfigData
	if id > 0 {
		configMapEnvLevel, err := impl.configMapRepository.GetByIdEnvLevel(id)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
		if len(configMapEnvLevel.SecretData) > 0 {
			err = json.Unmarshal([]byte(configMapEnvLevel.SecretData), configsList)
			if err != nil {
				impl.logger.Warnw("error while Unmarshal", "error", err)
			}
		}
		for _, item := range configsList.ConfigData {
			if item.Name == name {
				configs = append(configs, item)
				break
			}
		}
	}
	if len(configs) == 0 {
		configMapGlobal, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
		if err != nil && pg.ErrNoRows != err {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
		if pg.ErrNoRows == err {
			impl.logger.Warnw("no app level secret found for this request", "appId", appId)
		}
		configMapGlobalList := &SecretsList{}
		if len(configMapGlobal.SecretData) > 0 {
			err = json.Unmarshal([]byte(configMapGlobal.SecretData), configMapGlobalList)
			if err != nil {
				impl.logger.Warnw("error while Unmarshal", "error", err)
			}
		}
		for _, item := range configMapGlobalList.ConfigData {
			if item.Name == name {
				configs = append(configs, item)
				break
			}
		}
	}
	configDataRequest.Id = id
	configDataRequest.AppId = appId
	configDataRequest.EnvironmentId = envId
	configDataRequest.ConfigData = configs
	return configDataRequest, nil
}

func (impl ConfigMapServiceImpl) validateConfigData(configData *ConfigData) (bool, error) {
	dataMap := make(map[string]string)
	if configData.Data != nil {
		err := json.Unmarshal(configData.Data, &dataMap)
		if err != nil {
			impl.logger.Errorw("error while Unmarshal", "error", err)
			return false, fmt.Errorf("unmarshal failed for data, please provide valid json")
		}
	}
	re := regexp.MustCompile("[-._a-zA-Z0-9]+") //^[A-ZA-Z0-9_]+$
	for key := range dataMap {
		if !re.MatchString(key) {
			return false, fmt.Errorf("invalid key : %s", key)
		}
	}
	return true, nil
}

func (impl ConfigMapServiceImpl) updateConfigData(configData *ConfigData, syncRequest *BulkPatchRequest) (*ConfigData, error) {
	dataMap := make(map[string]string)
	var updatedData json.RawMessage
	if configData.Data != nil {
		err := json.Unmarshal(configData.Data, &dataMap)
		if err != nil {
			impl.logger.Errorw("error while Unmarshal", "error", err)
			return configData, fmt.Errorf("unmarshal failed for data")
		}
		if syncRequest.PatchAction == 1 {
			dataMap[syncRequest.Key] = syncRequest.Value
		} else if syncRequest.PatchAction == 2 {
			if _, ok := dataMap[syncRequest.Key]; ok {
				dataMap[syncRequest.Key] = syncRequest.Value
			}
		} else if syncRequest.PatchAction == 3 {
			if _, ok := dataMap[syncRequest.Key]; ok {
				delete(dataMap, syncRequest.Key)
			}
		}
		updatedData, err = json.Marshal(dataMap)
		if err != nil {
			impl.logger.Errorw("error while marshal", "error", err)
			return configData, fmt.Errorf("marshal failed for data")
		}
		configData.Data = updatedData
	} else if syncRequest.PatchAction == 1 {
		err := json.Unmarshal(configData.Data, &dataMap)
		if err != nil {
			impl.logger.Errorw("error while Unmarshal", "error", err)
			return configData, fmt.Errorf("unmarshal failed for data")
		}
		dataMap[syncRequest.Key] = syncRequest.Value
		updatedData, err = json.Marshal(dataMap)
		if err != nil {
			impl.logger.Errorw("error while marshal", "error", err)
			return configData, fmt.Errorf("marshal failed for data")
		}
		configData.Data = updatedData
	}
	return configData, nil
}

func (impl ConfigMapServiceImpl) ConfigSecretGlobalBulkPatch(bulkPatchRequest *BulkPatchRequest) (*BulkPatchRequest, error) {
	_, err := impl.buildBulkPayload(bulkPatchRequest)
	if err != nil {
		impl.logger.Errorw("service err, ConfigSecretGlobalBulkPatch", "err", err, "payload", bulkPatchRequest)
		return nil, fmt.Errorf("")
	}
	if len(bulkPatchRequest.Payload) == 0 {
		return nil, fmt.Errorf("invalid request no payload found for sync")
	}
	for _, payload := range bulkPatchRequest.Payload {
		model, err := impl.configMapRepository.GetByAppIdAppLevel(payload.AppId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
		if err == pg.ErrNoRows {
			continue
		}
		if bulkPatchRequest.Type == "CM" {
			configsList := &ConfigsList{}
			var configs []*ConfigData
			if len(model.ConfigMapData) > 0 {
				err = json.Unmarshal([]byte(model.ConfigMapData), configsList)
				if err != nil {
					impl.logger.Warnw("error while Unmarshal", "error", err)
				}
			}
			for _, item := range configsList.ConfigData {
				if item.Name == bulkPatchRequest.Name {
					updatedConfigData, err := impl.updateConfigData(item, bulkPatchRequest)
					if err != nil {
						impl.logger.Warnw("error while updating data", "error", err)
					}
					item.Data = updatedConfigData.Data
				}
				configs = append(configs, item)
			}
			configsList.ConfigData = configs
			configDataByte, err := json.Marshal(configsList)
			if err != nil {
				return nil, err
			}
			model.ConfigMapData = string(configDataByte)
		} else if bulkPatchRequest.Type == "CS" {
			secretsList := &SecretsList{}
			var configs []*ConfigData
			if len(model.SecretData) > 0 {
				err = json.Unmarshal([]byte(model.SecretData), secretsList)
				if err != nil {
					impl.logger.Warnw("error while Unmarshal", "error", err)
				}
			}
			for _, item := range secretsList.ConfigData {
				if item.Name == bulkPatchRequest.Name {
					updatedConfigData, err := impl.updateConfigData(item, bulkPatchRequest)
					if err != nil {
						impl.logger.Warnw("error while updating data", "error", err)
					}
					item.Data = updatedConfigData.Data
				}
				configs = append(configs, item)
			}
			secretsList.ConfigData = configs
			configDataByte, err := json.Marshal(secretsList)
			if err != nil {
				return nil, err
			}
			model.SecretData = string(configDataByte)
		}
		model.UpdatedBy = bulkPatchRequest.UserId
		model.UpdatedOn = time.Now()
		_, err = impl.configMapRepository.UpdateAppLevel(model)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
	}
	return bulkPatchRequest, nil
}

func (impl ConfigMapServiceImpl) ConfigSecretEnvironmentBulkPatch(bulkPatchRequest *BulkPatchRequest) (*BulkPatchRequest, error) {
	_, err := impl.buildBulkPayload(bulkPatchRequest)
	if err != nil {
		impl.logger.Errorw("service err, ConfigSecretEnvironmentBulkPatch", "err", err, "payload", bulkPatchRequest)
		return nil, fmt.Errorf("")
	}
	if len(bulkPatchRequest.Payload) == 0 {
		return nil, fmt.Errorf("invalid request no payload found for sync")
	}
	for _, payload := range bulkPatchRequest.Payload {
		if payload.AppId == 0 || payload.EnvId == 0 {
			return nil, fmt.Errorf("invalid request payload not complete for env level patch")
		}
	}
	for _, payload := range bulkPatchRequest.Payload {
		model, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(payload.AppId, payload.EnvId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
		if err == pg.ErrNoRows {
			continue
		}
		if bulkPatchRequest.Type == "CM" {
			configsList := &ConfigsList{}
			var configs []*ConfigData
			if len(model.ConfigMapData) > 0 {
				err = json.Unmarshal([]byte(model.ConfigMapData), configsList)
				if err != nil {
					impl.logger.Warnw("error while Unmarshal", "error", err)
				}
			}
			for _, item := range configsList.ConfigData {
				if item.Name == bulkPatchRequest.Name {
					updatedConfigData, err := impl.updateConfigData(item, bulkPatchRequest)
					if err != nil {
						impl.logger.Warnw("error while updating data", "error", err)
					}
					item.Data = updatedConfigData.Data
				}
				configs = append(configs, item)
			}
			configsList.ConfigData = configs
			configDataByte, err := json.Marshal(configsList)
			if err != nil {
				return nil, err
			}
			model.ConfigMapData = string(configDataByte)
		} else if bulkPatchRequest.Type == "CS" {
			secretsList := &SecretsList{}
			var configs []*ConfigData
			if len(model.SecretData) > 0 {
				err = json.Unmarshal([]byte(model.SecretData), secretsList)
				if err != nil {
					impl.logger.Warnw("error while Unmarshal", "error", err)
				}
			}
			for _, item := range secretsList.ConfigData {
				if item.Name == bulkPatchRequest.Name {
					updatedConfigData, err := impl.updateConfigData(item, bulkPatchRequest)
					if err != nil {
						impl.logger.Debugw("error while updating data", "error", err)
					}
					item.Data = updatedConfigData.Data
				}
				configs = append(configs, item)
			}
			secretsList.ConfigData = configs
			configDataByte, err := json.Marshal(secretsList)
			if err != nil {
				return nil, err
			}
			model.SecretData = string(configDataByte)
		}
		model.UpdatedBy = bulkPatchRequest.UserId
		model.UpdatedOn = time.Now()
		_, err = impl.configMapRepository.UpdateEnvLevel(model)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
	}
	return bulkPatchRequest, nil
}

func (impl ConfigMapServiceImpl) validateExternalSecretChartCompatibility(appId int, envId int, configData *ConfigData) (bool, error) {

	if configData.ExternalSecret != nil && len(configData.ExternalSecret) > 0 {
		for _, es := range configData.ExternalSecret {
			if len(es.Property) > 0 || es.IsBinary == true {
				chart, err := impl.commonService.FetchLatestChart(appId, envId)
				if err != nil {
					return false, err
				}
				chartVersion := chart.ChartVersion
				chartMajorVersion, chartMinorVersion, err := util2.ExtractChartVersion(chartVersion)
				if err != nil {
					impl.logger.Errorw("chart version parsing", "err", err)
					return false, err
				}
				if chartMajorVersion <= 3 && chartMinorVersion < 8 {
					return false, fmt.Errorf("this chart version dosent support property and isBinary keys, please upgrade chart: %s", configData.Name)
				}
			}
		}
	}
	return true, nil
}

func (impl ConfigMapServiceImpl) buildBulkPayload(bulkPatchRequest *BulkPatchRequest) (*BulkPatchRequest, error) {
	var payload []*BulkPatchPayload
	if bulkPatchRequest.Filter != nil {
		apps, err := impl.appRepository.FetchAppsByFilterV2(bulkPatchRequest.Filter.AppNameIncludes, bulkPatchRequest.Filter.AppNameExcludes, bulkPatchRequest.Filter.EnvId)
		if err != nil {
			impl.logger.Errorw("chart version parsing", "err", err)
			return bulkPatchRequest, err
		}
		for _, item := range apps {
			if bulkPatchRequest.Global {
				payload = append(payload, &BulkPatchPayload{AppId: item.Id})
			} else {
				payload = append(payload, &BulkPatchPayload{AppId: item.Id, EnvId: bulkPatchRequest.Filter.EnvId})
			}
		}
		bulkPatchRequest.Payload = payload
	} else if bulkPatchRequest.ProjectId > 0 && bulkPatchRequest.Global {
		//backward compatibility
		apps, err := impl.appRepository.FindAppsByTeamId(bulkPatchRequest.ProjectId)
		if err != nil {
			impl.logger.Errorw("service err, buildBulkPayload", "err", err, "payload", bulkPatchRequest)
			return bulkPatchRequest, err
		}
		var payload []*BulkPatchPayload
		for _, app := range apps {
			payload = append(payload, &BulkPatchPayload{AppId: app.Id})
		}
		bulkPatchRequest.Payload = payload
	}
	return bulkPatchRequest, nil
}
