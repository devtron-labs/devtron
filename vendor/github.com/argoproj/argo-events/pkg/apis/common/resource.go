package common

import "encoding/json"

/**
This inspired by intstr.IntOrStr and json.RawMessage.
*/

// Resource represent arbitrary structured data.
type Resource struct {
	Value []byte `json:"value" protobuf:"bytes,1,opt,name=value"`
}

func NewResource(s interface{}) Resource {
	data, _ := json.Marshal(s)
	return Resource{Value: data}
}

func (a *Resource) UnmarshalJSON(value []byte) error {
	a.Value = value
	return nil
}

func (n Resource) MarshalJSON() ([]byte, error) {
	return n.Value, nil
}

func (n Resource) OpenAPISchemaType() []string {
	return []string{"object"}
}

func (n Resource) OpenAPISchemaFormat() string { return "" }
