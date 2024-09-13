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
