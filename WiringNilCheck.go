package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"unsafe"
)

func CheckIfNilInWire() {
	app, err := InitializeApp()
	if err != nil {
		log.Panic(err)
	}
	nilFieldsMap := make(map[string]bool)
	checkNilFields(app, nilFieldsMap)
	fmt.Println("NIL Fields present in impls are: ", nilFieldsMap)
	//writing length of nilFieldsMap in a file (eg. output.env) so that we can export this data in pre ci pipeline bash and can failed the pre ci pipeline if nil nilFieldsMap's length is greater than zero
	err = writeResultToFile(len(nilFieldsMap))
	if err != nil {
		return
	}
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
		if skipUnnecessaryFiledsForCheck(fieldName, valName) { // skip unnecessary fileds and values
			continue
		}
		if !canFieldTypeBeNil(field) { // field can not be nil, skip
			continue
		} else if field.IsNil() { // check if the field is nil
			mapEntry := fmt.Sprintf("%s.%s", valName, fieldName)
			nilObjMap[mapEntry] = true
			continue
		}
		if canSkipFieldStructCheck(fieldName, valName) {
			continue
		}
		if !isExported(fieldName) && !field.CanInterface() {
			unexportedField := GetUnexportedField(field)
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

func canSkipFieldStructCheck(fieldName, valName string) bool {
	fieldName = strings.ToLower(fieldName)
	valName = strings.ToLower(valName)
	if valName == "githubclient" && (fieldName == "client" || fieldName == "gitopshelper") {
		return true
	}
	for _, str := range []string{"logger", "dbconnection", "syncedenforcer"} {
		if fieldName == str {
			return true
		}
	}
	return false
}

func skipUnnecessaryFiledsForCheck(fieldName, valName string) bool {
	fieldName = strings.ToLower(fieldName)
	valName = strings.ToLower(valName)
	if valName == "cicdconfig" {
		return true
	}
	fieldAndValName := map[string][]string{
		"app":                          {"enforcerv2", "server"},
		"gitfactory":                   {"client"},
		"argocdconnectionmanagerimpl":  {"argocdsettings"},
		"enforcerimpl":                 {"cache", "enforcerv2"},
		"helmappclientimpl":            {"applicationserviceclient"},
		"modulecronserviceimpl":        {"cron"},
		"oteltracingserviceimpl":       {"traceprovider"},
		"terminalaccessrepositoryimpl": {"templatescache"},
	}
	if _, ok := fieldAndValName[valName]; ok {
		for _, ignoreFieldName := range fieldAndValName[valName] {
			if ignoreFieldName == fieldName {
				return true
			}
		}
	}
	return false
}
func GetUnexportedField(field reflect.Value) interface{} {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}

func isExported(fieldName string) bool {
	return strings.ToUpper(fieldName[0:1]) == fieldName[0:1]
}

func writeResultToFile(data int) error {
	file, err := os.Create("/test/output.env")
	if err != nil {
		log.Println("Failed to create file:", err)
		return err
	}
	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf("OUTPUT=%d", data))
	if err != nil {
		log.Println("Failed to write to file:", err)
		return err
	}
	return nil
}
