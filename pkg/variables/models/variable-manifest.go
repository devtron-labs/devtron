package models

type VariableRequest struct {
	Manifest ScopedVariableManifest `json:"manifest"`
	UserId   int32                  `json:"-"`
}
type ScopedVariableManifest struct {
	ApiVersion string         `json:"apiVersion" validate:"oneof=devtron.ai/v1beta1"`
	Kind       string         `json:"kind" validate:"oneof=Variable"`
	Spec       []VariableSpec `json:"spec" validate:"required,dive"`
}

type VariableSpec struct {
	Description      string              `json:"description"`
	ShortDescription string              `json:"shortDescription" validate:"max=120"`
	IsSensitive      bool                `json:"isSensitive"`
	Name             string              `json:"name" validate:"required"`
	Values           []VariableValueSpec `json:"values" validate:"dive"`
}

type VariableValueSpec struct {
	Category  AttributeType `json:"category" validate:"oneof=Global"`
	Value     interface{}   `json:"value" validate:"required"`
	Selectors *Selector     `json:"selectors,omitempty"`
}

type Selector struct {
	AttributeSelectors map[IdentifierType]string `json:"attributeSelectors"`
}
