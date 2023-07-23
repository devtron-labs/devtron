package chartProvider

import repository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"

type ChartProviderResponseDto struct {
	Id               string                  `json:"id" validate:"required"`
	Name             string                  `json:"name" validate:"required"`
	Active           bool                    `json:"active" validate:"required"`
	IsEditable       bool                    `json:"isEditable"`
	IsOCIRegistry    bool                    `json:"isOCIRegistry"`
	RegistryProvider repository.RegistryType `json:"registryProvider"`
	UserId           int32                   `json:"-"`
}

type ChartProviderRequestDto struct {
	Id            string `json:"id" validate:"required"`
	IsOCIRegistry bool   `json:"isOCIRegistry"`
	Active        bool   `json:"active,omitempty"`
	UserId        int32  `json:"-"`
}
