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
	"github.com/devtron-labs/devtron/pkg/bulkAction"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestBulkUpdate(t *testing.T) {
	setup()
	type test struct {
		ApiVersion             string
		Kind                   string
		Payload                *bulkAction.BulkUpdatePayload
		deploymentTemplateWant string
		configMapWant          string
		secretWant             string
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
		deploymentTemplateSpec := &bulkAction.DeploymentTemplateSpec{
			PatchJson: record[6]}
		deploymentTemplateTask := &bulkAction.DeploymentTemplateTask{
			Spec: deploymentTemplateSpec,
		}
		configMapSpec := &bulkAction.CmAndSecretSpec{
			Names:     strings.Fields(record[7]),
			PatchJson: record[8],
		}
		configMapTask := &bulkAction.CmAndSecretTask{
			Spec: configMapSpec,
		}
		secretSpec := &bulkAction.CmAndSecretSpec{
			Names:     strings.Fields(record[9]),
			PatchJson: record[10],
		}
		secretTask := &bulkAction.CmAndSecretTask{
			Spec: secretSpec,
		}
		payload := &bulkAction.BulkUpdatePayload{
			Includes:           includes,
			Excludes:           excludes,
			EnvIds:             envId,
			Global:             global,
			DeploymentTemplate: deploymentTemplateTask,
			ConfigMap:          configMapTask,
			Secret:             secretTask,
		}
		Input := test{
			ApiVersion:             record[0],
			Kind:                   record[1],
			Payload:                payload,
			deploymentTemplateWant: record[11],
			configMapWant:          record[12],
			secretWant:             record[13],
		}
		tests = append(tests, Input)
		fmt.Println(Input)
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s,%s", tt.Payload.Includes, tt.Payload.Excludes)
		t.Run(testname, func(t *testing.T) {
			got := bulkUpdateService.BulkUpdate(tt.Payload)
			if got.DeploymentTemplate.Message[len(got.DeploymentTemplate.Message)-1] != tt.deploymentTemplateWant {
				t.Errorf("got %s, want %s", got, tt.deploymentTemplateWant)
			}
			if got.ConfigMap.Message[len(got.ConfigMap.Message)-1] != tt.configMapWant {
				t.Errorf("got %s, want %s", got, tt.configMapWant)
			}
			if got.Secret.Message[len(got.Secret.Message)-1] != tt.secretWant {
				t.Errorf("got %s, want %s", got, tt.secretWant)
			}
		})
	}
}
