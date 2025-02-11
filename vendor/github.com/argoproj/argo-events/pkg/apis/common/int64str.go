package common

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Int64OrString struct {
	Type     Type   `json:"type" protobuf:"varint,1,opt,name=type,casttype=Type"`
	Int64Val int64  `json:"int64Val,omitempty" protobuf:"varint,2,opt,name=int64Val"`
	StrVal   string `json:"strVal,omitempty" protobuf:"bytes,3,opt,name=strVal"`
}

// Type represents the stored type of Int64OrString.
type Type int64

const (
	Int64  Type = iota // The Int64OrString holds an int64.
	String             // The Int64OrString holds a string.
)

// FromString creates an Int64OrString object with a string value.
func FromString(val string) Int64OrString {
	return Int64OrString{Type: String, StrVal: val}
}

// FromInt64 creates an Int64OrString object with an int64 value.
func FromInt64(val int64) Int64OrString {
	return Int64OrString{Type: Int64, Int64Val: val}
}

// Parse the given string and try to convert it to an int64 before
// setting it as a string value.
func Parse(val string) Int64OrString {
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return FromString(val)
	}
	return FromInt64(i)
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (int64str *Int64OrString) UnmarshalJSON(value []byte) error {
	if value[0] == '"' {
		int64str.Type = String
		return json.Unmarshal(value, &int64str.StrVal)
	}
	int64str.Type = Int64
	return json.Unmarshal(value, &int64str.Int64Val)
}

// Int64Value returns the Int64Val if type Int64, or if
// it is a String, will attempt a conversion to int64,
// returning 0 if a parsing error occurs.
func (int64str *Int64OrString) Int64Value() int64 {
	if int64str.Type == String {
		i, _ := strconv.ParseInt(int64str.StrVal, 10, 64)
		return i
	}
	return int64str.Int64Val
}

// MarshalJSON implements the json.Marshaller interface.
func (int64str Int64OrString) MarshalJSON() ([]byte, error) {
	switch int64str.Type {
	case Int64:
		return json.Marshal(int64str.Int64Val)
	case String:
		return json.Marshal(int64str.StrVal)
	default:
		return []byte{}, fmt.Errorf("impossible Int64OrString.Type")
	}
}

// OpenAPISchemaType is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
//
// See: https://github.com/kubernetes/kube-openapi/tree/master/pkg/generators
func (Int64OrString) OpenAPISchemaType() []string { return []string{"string"} }

// OpenAPISchemaFormat is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
func (Int64OrString) OpenAPISchemaFormat() string { return "int64-or-string" }
