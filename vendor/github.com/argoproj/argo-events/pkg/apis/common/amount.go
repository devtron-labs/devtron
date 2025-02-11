package common

import "strconv"

/**
This inspired by intstr.IntOrStr and json.Number.
*/

// Amount represent a numeric amount.
type Amount struct {
	Value []byte `json:"value" protobuf:"bytes,1,opt,name=value"`
}

func NewAmount(s string) Amount {
	return Amount{Value: []byte(s)}
}

func (a *Amount) UnmarshalJSON(value []byte) error {
	a.Value = value
	return nil
}

func (n Amount) MarshalJSON() ([]byte, error) {
	return n.Value, nil
}

func (n Amount) OpenAPISchemaType() []string {
	return []string{"number"}
}

func (n Amount) OpenAPISchemaFormat() string { return "" }

func (n *Amount) Float64() (float64, error) {
	return strconv.ParseFloat(string(n.Value), 64)
}
