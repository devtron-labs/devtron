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

package pipeline

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/bulkUpdate"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bulkAction"
	"github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/sql"
	jsonpatch "github.com/evanphx/json-patch"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
)

var bulkUpdateService *bulkAction.BulkUpdateServiceImpl
var bulkUpdateRepository bulkUpdate.BulkUpdateRepositoryImpl

func setup() {
	config, _ := sql.GetConfig()
	logger, _ := util.NewSugardLogger()
	dbConnection, _ := sql.NewDbConnection(config, logger)
	bulkUpdateRepository := bulkUpdate.NewBulkUpdateRepository(dbConnection, logger)
	bulkUpdateService = bulkAction.NewBulkUpdateServiceImpl(bulkUpdateRepository, nil, nil, nil, nil, "",
		chart.DefaultChart(""), util.MergeUtil{}, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil)
}

func TestBulkUpdateDeploymentTemplate(t *testing.T) {
	setup()
	type test struct {
		ApiVersion string
		Kind       string
		Payload    *bulkAction.BulkUpdatePayload
		want       string
	}
	TestsCsvFile, err := os.Open("ChartService_test.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	r := csv.NewReader(TestsCsvFile)
	r.LazyQuotes = true
	var tests []test
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(record[6])
		var envId []int
		if err := json.Unmarshal([]byte(record[4]), &envId); err != nil {
			panic(err)
		}
		global, err := strconv.ParseBool(record[5])
		if err != nil {
			panic(err)
		}
		namesIncludes := strings.Fields(record[2])
		namesExcludes := strings.Fields(record[3])
		includes := &bulkAction.NameIncludesExcludes{Names: namesIncludes}
		excludes := &bulkAction.NameIncludesExcludes{Names: namesExcludes}
		spec := &bulkAction.DeploymentTemplateSpec{
			PatchJson: record[6]}
		task := &bulkAction.DeploymentTemplateTask{
			Spec: spec,
		}
		payload := &bulkAction.BulkUpdatePayload{
			Includes:           includes,
			Excludes:           excludes,
			EnvIds:             envId,
			Global:             global,
			DeploymentTemplate: task,
		}
		Input := test{
			ApiVersion: record[0],
			Kind:       record[1],
			Payload:    payload,
			want:       record[7],
		}
		tests = append(tests, Input)
		fmt.Println(Input)
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s,%s", tt.Payload.Includes, tt.Payload.Excludes)
		t.Run(testname, func(t *testing.T) {
			got := bulkUpdateService.BulkUpdateDeploymentTemplate(tt.Payload)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestUnitBulkUpdateDeploymentTemplate(t *testing.T) {
	setup()
	type test struct {
		patch  jsonpatch.Patch
		target string
		want   string
	}
	TestsCsvFile, err := os.Open("ChartService_UnitTest.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	r := csv.NewReader(TestsCsvFile)
	r.LazyQuotes = true
	var tests []test
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(record[0])
		fmt.Println(record[1])
		fmt.Println(record[2])
		patchJson, err := jsonpatch.DecodePatch([]byte(record[0]))
		if err != nil {
			log.Fatal(err)
		}
		Input := test{
			patch:  patchJson,
			target: record[1],
			want:   record[2],
		}
		tests = append(tests, Input)
		fmt.Println(Input)
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.target)
		t.Run(testname, func(t *testing.T) {
			got, _ := bulkUpdateService.ApplyJsonPatch(tt.patch, tt.target)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
