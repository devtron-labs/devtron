/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"fmt"
	_ "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	_ "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"unsafe"
)

func main() {

	app, err := InitializeApp()
	if err != nil {
		log.Panic(err)
	}
	nilFieldsMap := make(map[string]bool)
	checkNilFields(app, nilFieldsMap)
	fmt.Println(nilFieldsMap)
	//     gracefulStop start
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v", sig)
		app.Stop()
		os.Exit(0)
	}()
	//      gracefulStop end

	app.Start()

}

func checkNilFields(obj interface{}, nilObjMap map[string]bool) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return
	}
	valName := val.Type().Name()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := val.Type().Field(i).Name
		pkgPath := val.Type().PkgPath()
		if pkgPath != "main" && !strings.Contains(pkgPath, "devtron-labs/devtron") {
			//package not from this repo, ignoring
			continue
		}
		if !canFieldTypeBeNil(field) { // field can not be nil, skip
			continue
		} else if field.IsNil() { // check if the field is nil
			mapEntry := fmt.Sprintf("%s.%s", valName, fieldName)
			nilObjMap[mapEntry] = true
			continue
		}
		if canSkipFieldStructCheck(fieldName) {
			continue
		}
		if !isExported(fieldName) && !field.CanInterface() {
			unexportedField := GetUnexportedField(field, fieldName)
			checkNilFields(unexportedField, nilObjMap)
		} else {
			// Recurse
			checkNilFields(field.Interface(), nilObjMap)
		}
	}
}

func canFieldTypeBeNil(field reflect.Value) bool {
	kind := field.Kind()
	switch kind {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer,
		reflect.Interface, reflect.Slice:
		return true
	default: //other types can not be nil
		return false
	}
}

func canSkipFieldStructCheck(fieldName string) bool {
	fieldName = strings.ToLower(fieldName)
	for _, str := range []string{"logger", "dbconnection", "syncedenforcer"} {
		if fieldName == str {
			return true
		}
	}
	return false
}

func GetUnexportedField(field reflect.Value, fieldName string) interface{} {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}

func isExported(fieldName string) bool {
	return strings.ToUpper(fieldName[0:1]) == fieldName[0:1]
}
