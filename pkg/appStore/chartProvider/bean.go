package chartProvider

type ChartProviderResponseDto struct {
	Id               string `json:"id" validate:"required"`
	Name             string `json:"name" validate:"required"`
	Active           bool   `json:"active" validate:"required"`
	IsEditable       bool   `json:"isEditable"`
	IsOCIRegistry    bool   `json:"isOCIRegistry"`
	RegistryProvider string `json:"registryProvider"`
	UserId           int32  `json:"-"`
}

type ChartProviderToggleRequestDto struct {
	Id            string `json:"id" validate:"required"`
	Active        bool   `json:"active" validate:"required"`
	IsOCIRegistry bool   `json:"isOCIRegistry" validate:"required"`
	UserId        int32  `json:"-"`
}

type ChartProviderSyncRequestDto struct {
	Id            string `json:"id" validate:"required"`
	IsOCIRegistry bool   `json:"isOCIRegistry" validate:"required"`
	UserId        int32  `json:"-"`
}
