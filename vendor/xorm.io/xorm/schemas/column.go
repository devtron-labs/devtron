// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schemas

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	TWOSIDES = iota + 1
	ONLYTODB
	ONLYFROMDB
)

// Column defines database column
type Column struct {
	Name            string
	TableName       string
	FieldName       string // Avaiable only when parsed from a struct
	SQLType         SQLType
	IsJSON          bool
	Length          int
	Length2         int
	Nullable        bool
	Default         string
	Indexes         map[string]int
	IsPrimaryKey    bool
	IsAutoIncrement bool
	MapType         int
	IsCreated       bool
	IsUpdated       bool
	IsDeleted       bool
	IsCascade       bool
	IsVersion       bool
	DefaultIsEmpty  bool // false means column has no default set, but not default value is empty
	EnumOptions     map[string]int
	SetOptions      map[string]int
	DisableTimeZone bool
	TimeZone        *time.Location // column specified time zone
	Comment         string
}

// NewColumn creates a new column
func NewColumn(name, fieldName string, sqlType SQLType, len1, len2 int, nullable bool) *Column {
	return &Column{
		Name:            name,
		TableName:       "",
		FieldName:       fieldName,
		SQLType:         sqlType,
		Length:          len1,
		Length2:         len2,
		Nullable:        nullable,
		Default:         "",
		Indexes:         make(map[string]int),
		IsPrimaryKey:    false,
		IsAutoIncrement: false,
		MapType:         TWOSIDES,
		IsCreated:       false,
		IsUpdated:       false,
		IsDeleted:       false,
		IsCascade:       false,
		IsVersion:       false,
		DefaultIsEmpty:  true, // default should be no default
		EnumOptions:     make(map[string]int),
		Comment:         "",
	}
}

// ValueOf returns column's filed of struct's value
func (col *Column) ValueOf(bean interface{}) (*reflect.Value, error) {
	dataStruct := reflect.Indirect(reflect.ValueOf(bean))
	return col.ValueOfV(&dataStruct)
}

// ValueOfV returns column's filed of struct's value accept reflevt value
func (col *Column) ValueOfV(dataStruct *reflect.Value) (*reflect.Value, error) {
	var fieldValue reflect.Value
	fieldPath := strings.Split(col.FieldName, ".")

	if dataStruct.Type().Kind() == reflect.Map {
		keyValue := reflect.ValueOf(fieldPath[len(fieldPath)-1])
		fieldValue = dataStruct.MapIndex(keyValue)
		return &fieldValue, nil
	} else if dataStruct.Type().Kind() == reflect.Interface {
		structValue := reflect.ValueOf(dataStruct.Interface())
		dataStruct = &structValue
	}

	level := len(fieldPath)
	fieldValue = dataStruct.FieldByName(fieldPath[0])
	for i := 0; i < level-1; i++ {
		if !fieldValue.IsValid() {
			break
		}
		if fieldValue.Kind() == reflect.Struct {
			fieldValue = fieldValue.FieldByName(fieldPath[i+1])
		} else if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
			fieldValue = fieldValue.Elem().FieldByName(fieldPath[i+1])
		} else {
			return nil, fmt.Errorf("field %v is not valid", col.FieldName)
		}
	}

	if !fieldValue.IsValid() {
		return nil, fmt.Errorf("field %v is not valid", col.FieldName)
	}

	return &fieldValue, nil
}

// ConvertID converts id content to suitable type according column type
func (col *Column) ConvertID(sid string) (interface{}, error) {
	if col.SQLType.IsNumeric() {
		n, err := strconv.ParseInt(sid, 10, 64)
		if err != nil {
			return nil, err
		}
		return n, nil
	} else if col.SQLType.IsText() {
		return sid, nil
	}
	return nil, errors.New("not supported")
}
