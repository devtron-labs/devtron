package types

import "github.com/devtron-labs/devtron/internal/sql/repository/imageTagging"

type ImageTaggingResponseDTO struct {
	ImageReleaseTags           []*repository.ImageTag   `json:"imageReleaseTags"`
	AppReleaseTags             []string                 `json:"appReleaseTags"`
	ImageComment               *repository.ImageComment `json:"imageComment"`
	ProdEnvExists              bool                     `json:"tagsEditable"`
	HideImageTaggingHardDelete bool                     `json:"hideImageTaggingHardDelete"`
}

type ImageTaggingRequestDTO struct {
	CreateTags     []*repository.ImageTag  `json:"createTags"`
	SoftDeleteTags []*repository.ImageTag  `json:"softDeleteTags"`
	ImageComment   repository.ImageComment `json:"imageComment"`
	HardDeleteTags []*repository.ImageTag  `json:"hardDeleteTags"`
	ExternalCi     bool                    `json:"-"`
}
