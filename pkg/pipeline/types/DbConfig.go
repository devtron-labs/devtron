/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

type DbMigrationConfigBean struct {
	Id            int    `json:"id"`
	DbConfigId    int    `json:"dbConfigId"`
	PipelineId    int    `json:"pipelineId"`
	GitMaterialId int    `json:"gitMaterialId"`
	ScriptSource  string `json:"scriptSource"` //location of file in git. relative to git root
	MigrationTool string `json:"migrationTool"`
	Active        bool   `json:"active"`
	UserId        int32  `json:"-"`
}

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
