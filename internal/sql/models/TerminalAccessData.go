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
