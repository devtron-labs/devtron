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

package bean

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"strings"
)

type ConfigDataRequest struct {
	Id            int           `json:"id"`
	AppId         int           `json:"appId"`
	EnvironmentId int           `json:"environmentId,omitempty"`
	ConfigData    []*ConfigData `json:"configData"`
	Deletable     bool          `json:"isDeletable"`
	UserId        int32         `json:"-"`

	// enterprise fields below
	IsExpressEdit bool `json:"isExpressEdit"`
}

type ESOSecretData struct {
	SecretStore     json.RawMessage `json:"secretStore,omitempty"`
	SecretStoreRef  json.RawMessage `json:"secretStoreRef,omitempty"`
	ESOData         []ESOData       `json:"esoData,omitempty"`
	RefreshInterval string          `json:"refreshInterval,omitempty"`
	ESODataFrom     json.RawMessage `json:"esoDataFrom,omitempty"`
	Template        json.RawMessage `json:"template,omitempty"`
}

type ESOData struct {
	SecretKey string `json:"secretKey"`
	Key       string `json:"key"`
	Property  string `json:"property,omitempty"`
}

// there is an adapter written in pkg/bean folder to convert below ConfigData struct to pkg/bean's ConfigData

type ConfigData struct {
	Name                  string               `json:"name"`
	Type                  string               `json:"type"`
	External              bool                 `json:"external"`
	MountPath             string               `json:"mountPath,omitempty"`
	Data                  json.RawMessage      `json:"data"`
	PatchData             json.RawMessage      `json:"patchData"`
	MergeStrategy         models.MergeStrategy `json:"mergeStrategy"`
	DefaultData           json.RawMessage      `json:"defaultData,omitempty"`
	DefaultMountPath      string               `json:"defaultMountPath,omitempty"`
	Global                bool                 `json:"global"`
	ExternalSecretType    string               `json:"externalType"`
	ESOSecretData         ESOSecretData        `json:"esoSecretData"`
	DefaultESOSecretData  ESOSecretData        `json:"defaultESOSecretData,omitempty"`
	ExternalSecret        []ExternalSecret     `json:"secretData"`
	DefaultExternalSecret []ExternalSecret     `json:"defaultSecretData,omitempty"`
	RoleARN               string               `json:"roleARN"`
	SubPath               bool                 `json:"subPath"`
	ESOSubPath            []string             `json:"esoSubPath"`
	FilePermission        string               `json:"filePermission"`
	Overridden            bool                 `json:"overridden"`
}

func (c *ConfigData) IsESOExternalSecretType() bool {
	return strings.HasPrefix(c.ExternalSecretType, "ESO")
}

type ExternalSecret struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Property string `json:"property,omitempty"`
	IsBinary bool   `json:"isBinary"`
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

type JobEnvOverrideResponse struct {
	Id              int    `json:"id"`
	AppId           int    `json:"appId"`
	EnvironmentId   int    `json:"environmentId,omitempty"`
	EnvironmentName string `json:"environmentName,omitempty"`
}

type CreateJobEnvOverridePayload struct {
	AppId  int   `json:"appId"`
	EnvId  int   `json:"envId"`
	UserId int32 `json:"-"`
}

type SecretsList struct {
	ConfigData []*ConfigData `json:"secrets"`
}

type ConfigsList struct {
	ConfigData []*ConfigData `json:"maps"`
}

type ConfigNameAndType struct {
	Id   int
	Name string
	Type ResourceType
}

type ResourceType string

const (
	CM                 ResourceType = "ConfigMap"
	CS                 ResourceType = "Secret"
	DeploymentTemplate ResourceType = "Deployment Template"
	PipelineStrategy   ResourceType = "Pipeline Strategy"
)

func (r ResourceType) ToString() string {
	return string(r)
}

type ResolvedCmCsRequest struct {
	Scope resourceQualifiers.Scope
	AppId int
	EnvId int
	IsJob bool
}

func NewResolvedCmCsRequest(scope resourceQualifiers.Scope) *ResolvedCmCsRequest {
	return &ResolvedCmCsRequest{
		Scope: scope,
	}
}

func (r *ResolvedCmCsRequest) WithAppId(appId int) *ResolvedCmCsRequest {
	r.AppId = appId
	return r
}

func (r *ResolvedCmCsRequest) WithEnvId(envId int) *ResolvedCmCsRequest {
	r.EnvId = envId
	return r
}

func (r *ResolvedCmCsRequest) ForJob(isJob bool) *ResolvedCmCsRequest {
	r.IsJob = isJob
	return r
}
