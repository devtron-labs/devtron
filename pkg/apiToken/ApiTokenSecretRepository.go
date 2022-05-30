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

package apiToken

import (
	"github.com/go-pg/pg"
	"time"
)

type ApiTokenSecret struct {
	tableName struct{}  `sql:"api_token_secret"`
	Id        int       `sql:"id,pk"`
	Secret    string    `sql:"secret,notnull"`
	CreatedOn time.Time `sql:"created_on,notnull"`
	UpdatedOn time.Time `sql:"updated_on"`
}

type ApiTokenSecretRepository interface {
	Get() (*ApiTokenSecret, error)
}

type ApiTokenSecretRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewApiTokenSecretRepositoryImpl(dbConnection *pg.DB) *ApiTokenSecretRepositoryImpl {
	return &ApiTokenSecretRepositoryImpl{dbConnection: dbConnection}
}

func (impl ApiTokenSecretRepositoryImpl) Get() (*ApiTokenSecret, error) {
	apiTokenSecret := &ApiTokenSecret{}
	err := impl.dbConnection.Model(apiTokenSecret).
		Select()
	return apiTokenSecret, err
}
