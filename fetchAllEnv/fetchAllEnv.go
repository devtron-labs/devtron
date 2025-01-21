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
	"encoding/json"
	"errors"
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
	Env            string
	EnvType        string
	EnvValue       string
	EnvDescription string
	Example        string
	Deprecated     string
}

type CategoryField struct {
	Category string
	Fields   []EnvField
}

const (
	categoryCommentStructPrefix = "CATEGORY="
	defaultCategory             = "DEVTRON"
	deprecatedDefaultValue      = "false"

	envFieldTypeTag               = "env"
	envDefaultFieldTypeTag        = "envDefault"
	envDescriptionFieldTypeTag    = "description"
	envPossibleValuesFieldTypeTag = "example"
	envDeprecatedFieldTypeTag     = "deprecated"
	MARKDOWN_FILENAME             = "env_gen.md"
	MARKDOWN_JSON_FILENAME        = "env_gen.json"
)

const MarkdownTemplate = `
{{range . }}
## {{ .Category }} Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
{{range .Fields }} | {{ .Env }} | {{ .EnvType }} |{{ .EnvValue }} | {{ .EnvDescription }} | {{ .Example }} | {{ .Deprecated }} |
{{end}}
{{end}}`

func main() {
	WalkThroughProject()
	return
}

func WalkThroughProject() {
	categoryFieldsMap := make(map[string][]EnvField)
	uniqueKeys := make(map[string]bool)
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			err = processGoFile(path, categoryFieldsMap, uniqueKeys)
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
	writeToFile(categoryFieldsMap)
}

func processGoFile(filePath string, categoryFieldsMap map[string][]EnvField, uniqueKeys map[string]bool) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Println("error parsing file:", err)
		return err
	}
	ast.Inspect(node, func(n ast.Node) bool {
		if genDecl, ok := n.(*ast.GenDecl); ok {
			// checking if type declaration, one of [func, map, struct, array, channel, interface]
			if genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						// only checking struct type declarations
						if structType, ok2 := typeSpec.Type.(*ast.StructType); ok2 {
							allFields := make([]EnvField, 0, len(structType.Fields.List))
							for _, field := range structType.Fields.List {
								if field.Tag != nil {
									envField := getEnvKeyAndValue(field)
									envKey := envField.Env
									if len(envKey) == 0 || uniqueKeys[envKey] {
										continue
									}
									allFields = append(allFields, envField)
									uniqueKeys[envKey] = true
								}
							}
							if len(allFields) > 0 {
								category := getCategoryForAStruct(genDecl)
								categoryFieldsMap[category] = append(categoryFieldsMap[category], allFields...)
							}
						}
					}
				}
			}
		}
		return true
	})
	return nil
}

func getEnvKeyAndValue(field *ast.Field) EnvField {
	tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`")) // remove surrounding backticks

	envKey := addReadmeTableDelimiterEscapeChar(tag.Get(envFieldTypeTag))
	envValue := addReadmeTableDelimiterEscapeChar(tag.Get(envDefaultFieldTypeTag))
	envDescription := addReadmeTableDelimiterEscapeChar(tag.Get(envDescriptionFieldTypeTag))
	envPossibleValues := addReadmeTableDelimiterEscapeChar(tag.Get(envPossibleValuesFieldTypeTag))
	envDeprecated := addReadmeTableDelimiterEscapeChar(tag.Get(envDeprecatedFieldTypeTag))
	// check if there exist any value provided in env for this field
	if value, ok := os.LookupEnv(envKey); ok {
		envValue = value
	}
	env := EnvField{
		Env:            envKey,
		EnvValue:       envValue,
		EnvDescription: envDescription,
		Example:        envPossibleValues,
		Deprecated:     envDeprecated,
	}
	if indent, ok := field.Type.(*ast.Ident); ok && indent != nil {
		env.EnvType = indent.Name
	}
	if len(envDeprecated) == 0 {
		env.Deprecated = deprecatedDefaultValue
	}
	return env
}

func getCategoryForAStruct(genDecl *ast.GenDecl) string {
	category := defaultCategory
	if genDecl.Doc != nil {
		commentTexts := strings.Split(genDecl.Doc.Text(), "\n")
		for _, comment := range commentTexts {
			commentText := strings.TrimPrefix(strings.ReplaceAll(comment, " ", ""), "//") // this can happen if comment group is in /* */
			if strings.HasPrefix(commentText, categoryCommentStructPrefix) {
				categories := strings.Split(strings.TrimPrefix(commentText, categoryCommentStructPrefix), ",")
				if len(categories) > 0 && len(categories[0]) > 0 { //only supporting one category as of now
					category = categories[0] //overriding category
					break
				}
			}
		}
	}
	return category
}

func addReadmeTableDelimiterEscapeChar(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}

func writeToFile(categoryFieldsMap map[string][]EnvField) {
	cfs := make([]CategoryField, 0, len(categoryFieldsMap))
	for category, allFields := range categoryFieldsMap {
		sort.Slice(allFields, func(i, j int) bool {
			return allFields[i].Env < allFields[j].Env
		})

		cfs = append(cfs, CategoryField{
			Category: category,
			Fields:   allFields,
		})
	}
	sort.Slice(cfs, func(i, j int) bool {
		return cfs[i].Category < cfs[j].Category
	})
	file, err := os.Create(MARKDOWN_FILENAME)
	if err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}
	defer file.Close()
	tmpl, err := template.New("markdown").Parse(MarkdownTemplate)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(file, cfs)
	if err != nil {
		panic(err)
	}
	cfsMarshaled, err := json.Marshal(cfs)
	if err != nil {
		log.Println("error marshalling category fields:", err)
		panic(err)
	}
	err = os.WriteFile(MARKDOWN_JSON_FILENAME, cfsMarshaled, 0644)
	if err != nil {
		panic(err)
	}
}
