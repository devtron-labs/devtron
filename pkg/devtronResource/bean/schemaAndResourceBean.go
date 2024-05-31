/*
 * Copyright (c) 2024. Devtron Inc.
 */

package bean

import "time"

type DevtronResourceBean struct {
	DisplayName          string                       `json:"displayName,omitempty"`
	Description          string                       `json:"description,omitempty"`
	DevtronResourceId    int                          `json:"devtronResourceId"`
	Kind                 string                       `json:"kind,omitempty"`
	VersionSchemaDetails []*DevtronResourceSchemaBean `json:"versionSchemaDetails,omitempty"`
	LastUpdatedOn        time.Time                    `json:"lastUpdatedOn,omitempty"`
}

type DevtronResourceSchemaBean struct {
	DevtronResourceSchemaId int    `json:"devtronResourceSchemaId"`
	Version                 string `json:"version,omitempty"`
	Schema                  string `json:"schema,omitempty"`
	SampleSchema            string `json:"sampleSchema,omitempty"`
}

type DevtronResourceSchemaRequestBean struct {
	DevtronResourceSchemaId int    `json:"devtronResourceSchemaId"`
	Schema                  string `json:"schema,omitempty"`
	DisplayName             string `json:"displayName,omitempty"`
	Description             string `json:"description,omitempty"`
	UserId                  int    `json:"-"`
}
