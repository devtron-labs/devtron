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

package bean

import (
	"encoding/json"
)

type ConfigMapRootJson struct {
	ConfigMapJson ConfigMapJson `json:"ConfigMaps"`
}
type ConfigMapJson struct {
	Enabled bool  `json:"enabled"`
	Maps    []Map `json:"maps"`
}

type ConfigSecretRootJson struct {
	ConfigSecretJson ConfigSecretJson `json:"ConfigSecrets"`
}
type ConfigSecretJson struct {
	Enabled bool   `json:"enabled"`
	Secrets []*Map `json:"secrets"`
}

type ConfigMapAndSecretJson struct {
	ConfigMapJson    ConfigMapJson    `json:"configMapJson"`
	ConfigSecretJson ConfigSecretJson `json:"configSecretJson"`
}

type Map struct {
	Name         string          `json:"name"`
	Type         string          `json:"type"`
	External     bool            `json:"external"`
	MountPath    string          `json:"mountPath"`
	Data         json.RawMessage `json:"data,omitempty"`
	ExternalType string          `json:"externalType"`
	RoleARN      string          `json:"roleARN"`
	SecretData   json.RawMessage `json:"secretData,omitempty"`
}
