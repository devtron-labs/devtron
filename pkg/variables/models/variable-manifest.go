package models

type ScopedVariableManifest struct {
	ApiVersion string         `json:"apiVersion" validate:"oneof=devtron.ai/v1beta1"`
	Kind       string         `json:"kind" validate:"oneof=Variable"`
	Spec       []VariableSpec `json:"spec" validate:"required,dive"`
	UserId     int32          `json:"-"`
}

type VariableSpec struct {
	Description string              `json:"description" validate:"required, max=300"`
	Name        string              `json:"name" validate:"required"`
	Values      []VariableValueSpec `json:"values" validate:"dive"`
}

type VariableValueSpec struct {
	Category  AttributeType `json:"category" validate:"oneof=ApplicationEnv Application Env Cluster Global"`
	Value     interface{}   `json:"value" validate:"required"`
	Selectors Selector      `json:"selectors" validate:"dive"`
}

type Selector struct {
	AttributeSelectors map[IdentifierType]string `json:"attributeSelectors"`
}
