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

package batch

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/apis/devtron/v1"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"strings"
)

type DataHolderAction interface {
	Execute(holder *v1.DataHolder, props v1.InheritedProps, dataType string) error
}

type DataHolderActionImpl struct {
	logger           *zap.SugaredLogger
	appRepo          pipelineConfig.AppRepository
	configMapService pipeline.ConfigMapService
	envService       cluster.EnvironmentService
}

func NewDataHolderActionImpl(appRepo pipelineConfig.AppRepository, configMapService pipeline.ConfigMapService, envService cluster.EnvironmentService, logger *zap.SugaredLogger) *DataHolderActionImpl {
	dh := &DataHolderActionImpl{
		logger:           logger,
		appRepo:          appRepo,
		configMapService: configMapService,
		envService:       envService,
	}
	return dh
}

var dataHolderExecutor = []func(impl DataHolderActionImpl, holder *v1.DataHolder, dataType string) error{executeDataHolderClone, executeDataHolderDelete, executeDataHolderCreate}

func (impl DataHolderActionImpl) Execute(holder *v1.DataHolder, props v1.InheritedProps, dataType string) error {
	if holder == nil {
		return nil
	}
	errs := make([]string, 0)
	for _, f := range dataHolderExecutor {
		err := holder.UpdateMissingProps(props)
		errs = util.AppendErrorString(errs, err)
		if err != nil {
			continue
		}
		errs = util.AppendErrorString(errs, f(impl, holder, dataType))
	}
	return util.GetErrorOrNil(errs)
}

func executeDataHolderClone(impl DataHolderActionImpl, holder *v1.DataHolder, dataType string) error {
	if holder.Operation != v1.Clone {
		return nil
	}
	//fetch from source and store in destination
	//if its source, then its part of clone and has to be skipped
	if holder.Source != nil && (holder.Source.App != nil || holder.Source.Secret != nil || holder.Source.ConfigMap != nil) {
		//skip insertion
		return nil
	}
	//if its destination only, then delete it from database
	//Get appId and envId and secret/configMap name
	if holder.Destination == nil || (holder.Destination != nil && holder.Destination.App == nil && holder.Destination.Secret == nil && holder.Destination.ConfigMap == nil) {
		return fmt.Errorf("destination not defined to clone %s", dataType)
	}
	appSrc, err := impl.appRepo.FindActiveByName(*holder.Source.App)
	if err != nil {
		return err
	}
	appDest, err := impl.appRepo.FindActiveByName(*holder.Destination.App)
	if err != nil {
		return err
	}
	var envSrc *cluster.EnvironmentBean
	var envDest *cluster.EnvironmentBean
	var configData *pipeline.ConfigDataRequest
	if holder.Source.Environment != nil {
		if envSrc, err = impl.envService.FindOne(*holder.Source.Environment); err != nil {
			return err
		}
	}

	if holder.Source.Environment != nil {
		if envSrc, err = impl.envService.FindOne(*holder.Source.Environment); err != nil {
			return err
		}
		if strings.ToLower(dataType) == v1.ConfigMap {
			if configData, err = impl.configMapService.CMEnvironmentFetch(appSrc.Id, envSrc.Id); err != nil {
				return err
			}

		} else {
			if configData, err = impl.configMapService.CSEnvironmentFetch(appSrc.Id, envSrc.Id); err != nil {
				return err
			}
		}
	} else {
		if strings.ToLower(dataType) == v1.ConfigMap {
			if configData, err = impl.configMapService.CMGlobalFetch(appSrc.Id); err != nil {
				return err
			}
		} else {
			if configData, err = impl.configMapService.CSGlobalFetch(appSrc.Id); err != nil {
				return err
			}
		}
	}

	if configData == nil {
		return fmt.Errorf("source not found to clone %s", dataType)
	}
	if holder.Destination.Environment != nil {
		if envDest, err = impl.envService.FindOne(*holder.Destination.Environment); err != nil {
			return err
		}
	}
	configData.AppId = appDest.Id
	if envDest != nil {
		configData.EnvironmentId = envDest.Id
	}
	err2 := updateKeys(dataType, holder, configData)
	if err2 != nil {
		return err2
	}

	if strings.ToLower(dataType) == v1.ConfigMap {
		if envDest != nil {
			if configData, err = impl.configMapService.CMEnvironmentAddUpdate(configData); err != nil {
				return err
			}
		} else {
			if configData, err = impl.configMapService.CMGlobalAddUpdate(configData); err != nil {
				return err
			}
		}
	} else {
		if envDest != nil {
			if configData, err = impl.configMapService.CSEnvironmentAddUpdate(configData); err != nil {
				return err
			}
		} else {
			if configData, err = impl.configMapService.CSGlobalAddUpdate(configData); err != nil {
				return err
			}
		}
	}
	return nil
}

func updateKeys(dataType string, holder *v1.DataHolder, configData *pipeline.ConfigDataRequest) error {
	var name string
	if dataType == v1.ConfigMap && holder.Destination.ConfigMap != nil {
		name = *holder.Destination.ConfigMap
	} else if holder.Destination.Secret != nil {
		name = *holder.Destination.Secret
	}
	//TODO: this logic is wrong, if name is specified then only that data is to be cloned and nothing else
	for i := range configData.ConfigData {
		if len(name) != 0 && configData.ConfigData[i].Name == name {
			d := make(map[string]interface{}, 0)
			err := json.Unmarshal(configData.ConfigData[i].Data, &d)
			if err != nil {
				return err
			}
			for k, v := range holder.Data {
				d[k] = v
				//if value is empty it means, this key is to be deleted
				if len(v.(string)) == 0 {
					delete(d, k)
				}
			}
			bd, err := json.Marshal(d)
			if err != nil {
				return err
			}
			configData.ConfigData[i].Data = bd
		}
	}
	return nil
}

func executeDataHolderDelete(impl DataHolderActionImpl, holder *v1.DataHolder, dataType string) error {
	if holder.Operation != v1.Delete {
		return nil
	}
	//if its source, then its part of clone and has to be skipped
	if holder.Source != nil && (holder.Source.App != nil || holder.Source.Secret != nil || holder.Source.ConfigMap != nil) {
		//skip insertion
		return nil
	}
	//if its destination only, then delete it from database
	//Get appId and enviId and secret/configmap name
	if holder.Destination == nil || holder.Destination.App == nil || (holder.Destination.Secret == nil && holder.Destination.ConfigMap == nil) {
		return fmt.Errorf("%s not uniquely identifiable", dataType)
	}
	app, err := impl.appRepo.FindActiveByName(*holder.Destination.App)
	if err != nil {
		return err
	}
	var env *cluster.EnvironmentBean
	if holder.Destination.Environment != nil {
		if env, err = impl.envService.FindOne(*holder.Destination.Environment); err != nil {
			return err
		}
		if strings.ToLower(dataType) == v1.ConfigMap {
			//TODO: pass userId
			if len(holder.Data) > 0 {
				err = deleteKeys(func() (request *pipeline.ConfigDataRequest, err error) {
					return impl.configMapService.CMEnvironmentFetch(app.Id, env.Id)
				}, impl.configMapService.CMEnvironmentAddUpdate, holder, dataType)
				if err != nil {
					return err
				}
			} else {
				deleted, err := impl.configMapService.CMEnvironmentDeleteByAppIdAndEnvId(*holder.Destination.ConfigMap, app.Id, env.Id, 1)
				if err != nil {
					return err
				}
				if !deleted {
					return fmt.Errorf("unable to delete %s named %s", v1.ConfigMap, *holder.Destination.ConfigMap)
				}
			}
		} else {
			if len(holder.Data) > 0 {
				err = deleteKeys(func() (request *pipeline.ConfigDataRequest, err error) {
					return impl.configMapService.CSEnvironmentFetch(app.Id, env.Id)
				}, impl.configMapService.CSEnvironmentAddUpdate, holder, dataType)
				if err != nil {
					return err
				}
			} else {
				deleted, err := impl.configMapService.CSEnvironmentDeleteByAppIdAndEnvId(*holder.Destination.Secret, app.Id, env.Id, 1)
				if err != nil {
					return err
				}
				if !deleted {
					return fmt.Errorf("unable to delete %s named %s", v1.Secret, *holder.Destination.Secret)
				}
			}
		}
	} else {
		if strings.ToLower(dataType) == v1.ConfigMap {
			//TODO: pass userId
			if len(holder.Data) > 0 {
				err = deleteKeys(func() (request *pipeline.ConfigDataRequest, err error) {
					return impl.configMapService.CMGlobalFetch(app.Id)
				}, impl.configMapService.CMGlobalAddUpdate, holder, dataType)
				if err != nil {
					return err
				}
			} else {
				deleted, err := impl.configMapService.CMGlobalDeleteByAppId(*holder.Destination.ConfigMap, app.Id, 1)
				if err != nil {
					return err
				}
				if !deleted {
					return fmt.Errorf("unable to delete %s named %s", v1.ConfigMap, *holder.Destination.ConfigMap)
				}
			}
		} else {
			if len(holder.Data) > 0 {
				err = deleteKeys(func() (request *pipeline.ConfigDataRequest, err error) {
					return impl.configMapService.CSGlobalFetch(app.Id)
				}, impl.configMapService.CSGlobalAddUpdate, holder, dataType)
				if err != nil {
					return err
				}
			} else {
				deleted, err := impl.configMapService.CSGlobalDeleteByAppId(*holder.Destination.Secret, app.Id, 1)
				if err != nil {
					return err
				}
				if !deleted {
					return fmt.Errorf("unable to delete %s named %s", v1.Secret, *holder.Destination.Secret)
				}
			}
		}
	}
	return nil
}

func deleteKeys(fetch func() (*pipeline.ConfigDataRequest, error), save func(request *pipeline.ConfigDataRequest) (*pipeline.ConfigDataRequest, error), holder *v1.DataHolder, dataType string) error {
	configData, err := fetch()
	if err != nil {
		return err
	}
	if configData != nil {
		err2 := deleteDataKeys(dataType, holder, configData)
		if err2 != nil {
			return err2
		}
	} else {
		return fmt.Errorf("configdata missing for %s", dataType)
	}
	_, err = save(configData)
	if err != nil {
		return err
	}
	return nil
}

func deleteDataKeys(dataType string, holder *v1.DataHolder, configData *pipeline.ConfigDataRequest) error {
	var name string
	if dataType == v1.ConfigMap {
		name = *holder.Destination.ConfigMap
	} else {
		name = *holder.Destination.Secret
	}
	//If secret and configMap name is missing then we need to clone all secrets and configMaps
	if len(name) == 0 {
		return nil
	}
	cda := make([]*pipeline.ConfigData, 0)
	for _, item := range configData.ConfigData {
		if item.Name == name {
			d := make(map[string]interface{}, 0)
			err := json.Unmarshal([]byte(item.Data), &d)
			if err != nil {
				return err
			}
			for k := range holder.Data {
				if _, ok := d[k]; ok {
					delete(d, k)
				}
			}
			bd, err := json.Marshal(d)
			if err != nil {
				return err
			}
			item.Data = bd
			cda = append(cda, item)
		}
	}
	configData.ConfigData = cda
	return nil
}

func executeDataHolderCreate(impl DataHolderActionImpl, holder *v1.DataHolder, dataType string) error {
	if holder.Operation != v1.Create {
		return nil
	}
	app, err := impl.appRepo.FindActiveByName(*holder.Destination.App)
	if err != nil {
		return err
	}
	var env *cluster.EnvironmentBean
	var name string
	if dataType == v1.ConfigMap {
		name = *holder.Destination.ConfigMap
	} else {
		name = *holder.Destination.Secret
	}
	d, err := json.Marshal(holder.Data)
	if err != nil {
		return err
	}
	envId := 0
	if holder.Destination.Environment != nil {
		if env, err = impl.envService.FindOne(*holder.Destination.Environment); err != nil {
			return err
		}
		envId = env.Id
	}

	if !holder.External && len(holder.Data) == 0 {
		return fmt.Errorf("data cannot be empty for %s", dataType)
	}
	external := false
	if len(holder.ExternalType) != 0 && !holder.External {
		external = true
	}
	var configDataArr []*pipeline.ConfigData
	configDataArr = append(configDataArr, &pipeline.ConfigData{
		Name:               name,
		Type:               holder.Type,
		External:           external,
		MountPath:          holder.MountPath,
		Data:               d,
		DefaultData:        nil,
		DefaultMountPath:   "",
		Global:             holder.Destination.Environment == nil,
		ExternalSecretType: holder.ExternalType,
	})
	//TODO: add User Id
	configData := &pipeline.ConfigDataRequest{
		AppId:         app.Id,
		EnvironmentId: envId,
		ConfigData: configDataArr,
		UserId:        1,
	}

	//TODO: get configData for app and env to populate Id
	if strings.ToLower(dataType) == v1.ConfigMap {
		if env != nil {
			d, err := impl.configMapService.CMEnvironmentFetch(app.Id, envId)
			if err == nil {
				configData.Id = d.Id
			}
			if configData, err = impl.configMapService.CMEnvironmentAddUpdate(configData); err != nil {
				return fmt.Errorf("error `%s` creating %s name %s", err.Error(), dataType, name)
			}
		} else {
			d, err := impl.configMapService.CMGlobalFetch(app.Id)
			if err == nil {
				configData.Id = d.Id
			}
			if configData, err = impl.configMapService.CMGlobalAddUpdate(configData); err != nil {
				return fmt.Errorf("error `%s` creating %s name %s", err.Error(), dataType, name)
			}
		}
	} else {
		if env != nil {
			d, err := impl.configMapService.CSEnvironmentFetch(app.Id, envId)
			if err == nil {
				configData.Id = d.Id
			}
			if configData, err = impl.configMapService.CSEnvironmentAddUpdate(configData); err != nil {
				return fmt.Errorf("error `%s` creating %s name %s", err.Error(), dataType, name)
			}
		} else {
			d, err := impl.configMapService.CSGlobalFetch(app.Id)
			if err == nil {
				configData.Id = d.Id
			}
			if configData, err = impl.configMapService.CSGlobalAddUpdate(configData); err != nil {
				return fmt.Errorf("error `%s` creating %s name %s", err.Error(), dataType, name)
			}
		}
	}
	return nil
}

func executeDataHolderEdit(impl DataHolderActionImpl, holder *v1.DataHolder, dataType string) error {
	if holder.Operation != v1.Edit {
		return nil
	}
	return nil
}
