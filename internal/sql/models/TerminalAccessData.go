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

package models

import "github.com/devtron-labs/devtron/pkg/sql"

type TerminalAccessTemplates struct {
	tableName    struct{} `sql:"terminal_access_templates" pg:",discard_unknown_columns"`
	Id           int      `sql:"id,pk"`
	TemplateName string   `sql:"template_name"`
	TemplateData string   `sql:"template_data"`
	sql.AuditLog
}

type UserTerminalAccessData struct {
	tableName struct{} `sql:"user_terminal_access_data" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	UserId    int32    `sql:"user_id"`
	ClusterId int      `sql:"cluster_id"`
	NodeName  string   `sql:"node_name"`
	PodName   string   `sql:"pod_name"`
	Status    string   `sql:"status"`
	Metadata  string   `sql:"metadata"`
	sql.AuditLog
}
