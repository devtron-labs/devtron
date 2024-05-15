package helper

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/xeipuuv/gojsonschema"
	"log"
	"net/http"
	"reflect"
	"strings"
)

func ValidateSchemaAndObjectData(schema, objectData string) (*gojsonschema.Result, error) {
	//validate user provided json with the schema
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(objectData)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return result, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.SchemaValidationFailedErrorUserMessage, err.Error())
	} else if !result.Valid() {
		errStr := ""
		for _, errResult := range result.Errors() {
			errStr += fmt.Sprintln(errResult.String())
		}
		return result, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.SchemaValidationFailedErrorUserMessage, errStr)
	}
	return result, nil
}

func GetReferencedPathsAndUpdatedSchema(schema string) (map[string]bool, string, error) {
	referencedPaths := make(map[string]bool)
	schemaJsonMap := make(map[string]interface{})
	schemaWithUpdatedRefData := ""
	err := json.Unmarshal([]byte(schema), &schemaJsonMap)
	if err != nil {
		return referencedPaths, schemaWithUpdatedRefData, err
	}
	getRefTypeInJsonAndAddRefKey(schemaJsonMap, referencedPaths)
	//marshaling new schema with $ref keys
	responseSchemaByte, err := json.Marshal(schemaJsonMap)
	if err != nil {
		return referencedPaths, schemaWithUpdatedRefData, err
	}
	schemaWithUpdatedRefData = string(responseSchemaByte)
	return referencedPaths, schemaWithUpdatedRefData, nil
}

func ExtractDiffPaths(json1, json2 []byte) ([]string, error) {
	pathList, err := compareJSON(json1, json2)
	if err != nil {
		log.Println("error in comparing json", "err", err, "json1", json1, "json2", json2)
		return nil, err
	}
	pathsToRemove := make([]string, 0)
	for _, path := range pathList {
		if len(path) > 0 {
			path = path[1:]
		}
		pathSplit := strings.Split(path, ",")
		if len(pathSplit) > 0 && (pathSplit[len(pathSplit)-1] == bean.Items || pathSplit[len(pathSplit)-1] == bean.AdditionalProperties) {
			pathSplit = pathSplit[:len(pathSplit)-1]
		}

		// remove properties attribute from path array
		idx := 0
		propertiesCount := 0
		for _, e := range pathSplit {
			if e != bean.Properties {
				times := propertiesCount / 2
				for time := 0; time < times; time++ {
					pathSplit[idx] = bean.Properties
					idx++
				}
				pathSplit[idx] = e
				propertiesCount = 0
				idx++
			} else {
				propertiesCount++
			}
		}

		pathSplit = pathSplit[:idx]

		path = strings.Join(pathSplit, ",")
		pathsToRemove = append(pathsToRemove, path)
	}
	return pathsToRemove, nil
}

func compareJSON(json1, json2 []byte) ([]string, error) {
	var m1 interface{}
	var m2 interface{}

	err := json.Unmarshal(json1, &m1)
	if err != nil {
		log.Println("error in unmarshalling json", "err", err, "json1", json1)
		return nil, err
	}
	err = json.Unmarshal(json2, &m2)
	if err != nil {
		log.Println("error in unmarshalling json", "err", err, "json2", json2)
		return nil, err
	}

	pathList := make([]string, 0)
	compareMaps(m1.(map[string]interface{}), m2.(map[string]interface{}), "", &pathList)
	return pathList, nil
}

func compareMaps(m1, m2 map[string]interface{}, currentPath string, pathList *[]string) {
	for k, v1 := range m1 {
		newPath := fmt.Sprintf("%s,%s", currentPath, k)
		if v2, ok := m2[k]; ok {
			switch v1.(type) {
			case []interface{}:
				if k == bean.Enum && !reflect.DeepEqual(v1, v2) {
					*pathList = append(*pathList, currentPath)
				}
			case map[string]interface{}:
				switch v2.(type) {
				case map[string]interface{}:
					compareMaps(v1.(map[string]interface{}), v2.(map[string]interface{}), newPath, pathList)
				default:
				}
			default:
				if !reflect.DeepEqual(v1, v2) {
					*pathList = append(*pathList, currentPath)
				}
			}
		} else {
			if k != bean.Required {
				*pathList = append(*pathList, newPath)
			}
		}
	}
}

func getRefTypeInJsonAndAddRefKey(schemaJsonMap map[string]interface{}, referencedObjects map[string]bool) {
	for key, value := range schemaJsonMap {
		if key == bean.RefTypeKey {
			valStr, ok := value.(string)
			if ok && strings.HasPrefix(valStr, bean.ReferencesPrefix) {
				schemaJsonMap[bean.RefKey] = valStr //adding $ref key for FE schema parsing
				delete(schemaJsonMap, bean.TypeKey) //deleting type because FE will be using $ref and thus type will be invalid
				referencedObjects[valStr] = true
			}
		} else {
			schemaUpdatedWithRef := resolveValForIteration(value, referencedObjects)
			schemaJsonMap[key] = schemaUpdatedWithRef
		}
	}
}

func resolveValForIteration(value interface{}, referencedObjects map[string]bool) interface{} {
	schemaUpdatedWithRef := value
	if valNew, ok := value.(map[string]interface{}); ok {
		getRefTypeInJsonAndAddRefKey(valNew, referencedObjects)
		schemaUpdatedWithRef = valNew
	} else if valArr, ok := value.([]interface{}); ok {
		for index, val := range valArr {
			schemaUpdatedWithRefNew := resolveValForIteration(val, referencedObjects)
			valArr[index] = schemaUpdatedWithRefNew
		}
		schemaUpdatedWithRef = valArr
	}
	return schemaUpdatedWithRef
}
