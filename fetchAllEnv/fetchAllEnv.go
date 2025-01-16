/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"
)

type EnvField struct {
	Env               string
	EnvValue          string
	EnvDescription    string
	EnvPossibleValues string
}

const (
	envFieldTypeTag               = "env"
	envDefaultFieldTypeTag        = "envDefault"
	envDescriptionFieldTypeTag    = "envDescription"
	envPossibleValuesFieldTypeTag = "possibleValues"
	MARKDOWN_FILENAME             = "env_gen.md"
)

const MarkdownTemplate = `
## Devtron Environment Variables
| Key   | Value        | Description       | Possible values       |
|-------|--------------|-------------------|-----------------------|
{{range .}} | {{ .Env }} | {{ .EnvValue }} | {{ .EnvDescription }} | {{ .EnvPossibleValues }} | 
{{end}}`

func writeToFile(allFields []EnvField) {
	sort.Slice(allFields, func(i, j int) bool {
		return allFields[i].Env < allFields[j].Env
	})

	file, err := os.Create(MARKDOWN_FILENAME)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	tmpl, err := template.New("markdown").Parse(MarkdownTemplate)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(file, allFields)
	if err != nil {
		panic(err)
	}
}

func WalkThroughProject() {
	var allFields []EnvField
	uniqueKeys := make(map[string]bool)

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			err = processGoFile(path, &allFields, &uniqueKeys)
			if err != nil {
				log.Println("error in processing go file", err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		return
	}
	writeToFile(allFields)
}

func convertTagToStructTag(tag string) reflect.StructTag {
	return reflect.StructTag(strings.Split(tag, "`")[1])
}

func getEnvKeyAndValue(tag reflect.StructTag) EnvField {
	envKey := addReadmeTableDelimiterEscapeChar(tag.Get(envFieldTypeTag))
	envValue := addReadmeTableDelimiterEscapeChar(tag.Get(envDefaultFieldTypeTag))
	envDescription := addReadmeTableDelimiterEscapeChar(tag.Get(envDescriptionFieldTypeTag))
	envPossibleValues := addReadmeTableDelimiterEscapeChar(tag.Get(envPossibleValuesFieldTypeTag))
	// check if there exist any value provided in env for this field
	if value, ok := os.LookupEnv(envKey); ok {
		envValue = value
	}
	return EnvField{
		Env:               envKey,
		EnvValue:          envValue,
		EnvDescription:    envDescription,
		EnvPossibleValues: envPossibleValues,
	}
}

func addReadmeTableDelimiterEscapeChar(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}

func processGoFile(filePath string, allFields *[]EnvField, uniqueKeys *map[string]bool) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Println("error parsing file:", err)
		return err
	}

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if structType, ok := x.Type.(*ast.StructType); ok {
				for _, field := range structType.Fields.List {
					if field.Tag != nil {
						strippedTags := convertTagToStructTag(field.Tag.Value)
						envField := getEnvKeyAndValue(strippedTags)
						envKey := envField.Env
						if len(envKey) == 0 || (*uniqueKeys)[envKey] {
							continue
						}
						*allFields = append(*allFields, envField)
						(*uniqueKeys)[envKey] = true
					}
				}
			}
		}
		return true
	})
	return nil
}

func main() {
	WalkThroughProject()
	return
}
