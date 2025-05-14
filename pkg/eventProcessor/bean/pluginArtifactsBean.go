/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package bean

import (
	"slices"
	"time"
)

type Kind string
type CredentialSourceType string
type ArtifactType string

const (
	PluginArtifactsKind           Kind                 = "PluginArtifacts"
	GlobalContainerRegistrySource CredentialSourceType = "global_container_registry"
	ArtifactTypeContainer         ArtifactType         = "CONTAINER"
)

type PluginArtifacts struct {
	Kind      Kind       `json:"Kind"`
	Artifacts []Artifact `json:"Artifacts"`
}

func NewPluginArtifact() *PluginArtifacts {
	return &PluginArtifacts{
		Kind:      PluginArtifactsKind,
		Artifacts: make([]Artifact, 0),
	}
}

func (p *PluginArtifacts) MergePluginArtifact(pluginArtifact *PluginArtifacts) {
	if pluginArtifact == nil {
		return
	}
	p.Artifacts = append(p.Artifacts, pluginArtifact.Artifacts...)
}

func (p *PluginArtifacts) GetRegistryToUniqueContainerArtifactDataMapping() map[string][]string {
	registryToImageMapping := make(map[string][]string)
	for _, artifact := range p.Artifacts {
		if artifact.Type == ArtifactTypeContainer {
			if artifact.CredentialsSourceType == GlobalContainerRegistrySource {
				if _, ok := registryToImageMapping[artifact.CredentialSourceValue]; !ok {
					registryToImageMapping[artifact.CredentialSourceValue] = make([]string, 0)
				}
				registryToImageMapping[artifact.CredentialSourceValue] = append(registryToImageMapping[artifact.CredentialSourceValue], artifact.Data...)
				slices.Sort(registryToImageMapping[artifact.CredentialSourceValue])
				slices.Compact(registryToImageMapping[artifact.CredentialSourceValue])
			}
		}
	}
	return registryToImageMapping
}

type Artifact struct {
	Type                      ArtifactType         `json:"Type"`
	Data                      []string             `json:"Data"`
	CredentialsSourceType     CredentialSourceType `json:"CredentialsSourceType"`
	CredentialSourceValue     string               `json:"CredentialSourceValue"`
	CreatedByPluginIdentifier string               `json:"createdByPluginIdentifier"`
	CreatedOn                 time.Time            `json:"createdOn"`
}
