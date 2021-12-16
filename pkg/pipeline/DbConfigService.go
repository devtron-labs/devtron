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
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
	"time"
)

type DbConfigBean struct {
	Id       int    `json:"id,omitempty" validate:"number"`
	Name     string `json:"name,omitempty" validate:"required"` //name by which user identifies this db
	Type     string `json:"type,omitempty" validate:"required"` //type of db, PG, MYsql, MariaDb
	Host     string `json:"host,omitempty" validate:"host"`
	Port     string `json:"port,omitempty" validate:"max=4"`
	DbName   string `json:"dbName,omitempty" validate:"required"` //name of database inside PG
	UserName string `json:"userName,omitempty"`
	Password string `json:"password,omitempty"`
	Active   bool   `json:"active,omitempty"`
	UserId   int32  `json:"-"`
}

type DbConfigService interface {
	Save(dbConfigBean *DbConfigBean) (dbConfig *DbConfigBean, err error)
	GetAll() (dbConfigs []*DbConfigBean, err error)
	GetById(id int) (dbConfig *DbConfigBean, err error)
	Update(dbConfigBean *DbConfigBean) (dbConfig *DbConfigBean, err error)
	GetForAutocomplete() (dbConfigs []*DbConfigBean, err error)
}
type DbConfigServiceImpl struct {
	configRepo repository.DbConfigRepository
	logger     *zap.SugaredLogger
}

func NewDbConfigService(configRepo repository.DbConfigRepository,
	logger *zap.SugaredLogger) *DbConfigServiceImpl {
	return &DbConfigServiceImpl{
		configRepo: configRepo,
		logger:     logger,
	}
}
func (impl DbConfigServiceImpl) Save(dbConfigBean *DbConfigBean) (dbConfig *DbConfigBean, err error) {
	t := repository.DbType(dbConfigBean.Type)
	if valid := t.IsValid(); !valid {
		impl.logger.Errorw("invalid type", "dbType", dbConfigBean.Type)
		return nil, fmt.Errorf("invalid type %s ", dbConfigBean.Type)
	}
	config := &repository.DbConfig{
		Password: dbConfigBean.Password,
		Port:     dbConfigBean.Port,
		Host:     dbConfigBean.Host,
		UserName: dbConfigBean.UserName,
		Type:     t,
		Name:     dbConfigBean.Name,
		Active:   dbConfigBean.Active,
		DbName:   dbConfigBean.DbName,
		AuditLog: sql.AuditLog{
			CreatedBy: dbConfigBean.UserId,
			UpdatedBy: dbConfigBean.UserId,
			CreatedOn: time.Now(),
			UpdatedOn: time.Now(),
		},
	}
	err = impl.configRepo.Save(config)
	if err != nil {
		impl.logger.Errorw("error in saving db config", "err", err)
		return nil, err
	}
	dbConfigBean.Id = config.Id
	return dbConfigBean, nil
}

func (impl DbConfigServiceImpl) GetAll() (dbConfigs []*DbConfigBean, err error) {
	configs, err := impl.configRepo.GetAll()
	if err != nil {
		return nil, err
	}
	for _, cfg := range configs {
		bean := impl.modelToBeanAdaptor(cfg)
		dbConfigs = append(dbConfigs, bean)
	}
	return dbConfigs, err
}
func (impl DbConfigServiceImpl) GetById(id int) (dbConfig *DbConfigBean, err error) {
	cfg, err := impl.configRepo.GetById(id)
	if err != nil {
		return nil, err
	}
	dbConfig = impl.modelToBeanAdaptor(cfg)
	return dbConfig, nil
}

func (impl DbConfigServiceImpl) Update(dbConfigBean *DbConfigBean) (dbConfig *DbConfigBean, err error) {
	var t repository.DbType
	if dbConfigBean.Type != "" {
		t = repository.DbType(dbConfigBean.Type)
		if valid := t.IsValid(); !valid {
			impl.logger.Errorw("invalid type", "dbType", dbConfigBean.Type)
			return nil, fmt.Errorf("invalid type %s ", dbConfigBean.Type)
		}
	}

	config := &repository.DbConfig{
		Id:       dbConfigBean.Id,
		Password: dbConfigBean.Password,
		Port:     dbConfigBean.Port,
		Host:     dbConfigBean.Host,
		UserName: dbConfigBean.UserName,
		Type:     t,
		Name:     dbConfigBean.Name,
		Active:   dbConfigBean.Active,
		DbName:   dbConfigBean.DbName,
		AuditLog: sql.AuditLog{
			UpdatedBy: dbConfigBean.UserId,
			UpdatedOn: time.Now(),
		},
	}
	_, err = impl.configRepo.Update(config)
	return dbConfigBean, err
}

func (impl DbConfigServiceImpl) modelToBeanAdaptor(conf *repository.DbConfig) (bean *DbConfigBean) {
	bean = &DbConfigBean{
		DbName:   conf.DbName,
		Active:   conf.Active,
		Name:     conf.Name,
		Type:     string(conf.Type),
		UserName: conf.UserName,
		Host:     conf.Host,
		Port:     conf.Port,
		Password: conf.Password,
		Id:       conf.Id,
	}
	return bean
}

func (impl DbConfigServiceImpl) GetForAutocomplete() (dbConfigs []*DbConfigBean, err error) {
	dbConf, err := impl.configRepo.GetActiveForAutocomplete()
	if err != nil {
		return nil, err
	}
	for _, cfg := range dbConf {
		bean := impl.modelToBeanAdaptor(cfg)
		dbConfigs = append(dbConfigs, bean)
	}
	return dbConfigs, nil
}
