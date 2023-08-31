package models

import (
	"reflect"
	"strconv"
)

type Payload struct {
	Variables []*Variables `json:"variables" validate:"required,dive"`
	UserId    int32        `json:"-"`
}
type Variables struct {
	Definition      Definition       `json:"definition" validate:"required,dive"`
	AttributeValues []AttributeValue `json:"attributeValue" validate:"dive" `
}
type AttributeValue struct {
	VariableValue   VariableValue             `json:"variableValue" validate:"required,dive"`
	AttributeType   AttributeType             `json:"attributeType" validate:"oneof=ApplicationEnv Application Env Cluster Global"`
	AttributeParams map[IdentifierType]string `json:"attributeParams"`
}

type Definition struct {
	VarName     string `json:"varName" validate:"required"`
	DataType    string `json:"dataType" validate:"oneof=json yaml primitive"`
	VarType     string `json:"varType" validate:"oneof=private public"`
	Description string `json:"description" validate:"max=300"`
}

type AttributeType string

const (
	ApplicationEnv AttributeType = "ApplicationEnv"
	Application    AttributeType = "Application"
	Env            AttributeType = "Env"
	Cluster        AttributeType = "Cluster"
	Global         AttributeType = "Global"
)

type IdentifierType string

const (
	EnvName         IdentifierType = "EnvName"
	ClusterName     IdentifierType = "ClusterName"
	ApplicationName IdentifierType = "ApplicationName"
)

var IdentifiersList = []IdentifierType{ApplicationName, EnvName, ClusterName}

type VariableValue struct {
	Value interface{} `json:"value" validate:"required"`
}

func (value VariableValue) StringValue() string {
	switch reflect.TypeOf(value.Value).Kind() {
	case reflect.Int:
		return strconv.Itoa(value.Value.(int))
	case reflect.Float64:
		return strconv.FormatFloat(value.Value.(float64), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(value.Value.(bool))
	}
	return value.Value.(string)
}

type Scope struct {
	AppId     int `json:"appId"`
	EnvId     int `json:"env_id"`
	ClusterId int `json:"clusterId"`
}
