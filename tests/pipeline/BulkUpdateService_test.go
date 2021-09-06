package pipeline

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/pipeline"
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
		Payload                *pipeline.BulkUpdatePayload
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
		includes := &pipeline.NameIncludesExcludes{Names: namesIncludes}
		excludes := &pipeline.NameIncludesExcludes{Names: namesExcludes}
		deploymentTemplateSpec := &pipeline.DeploymentTemplateSpec{
			PatchJson: record[6]}
		deploymentTemplateTask := &pipeline.DeploymentTemplateTask{
			Spec: deploymentTemplateSpec,
		}
		configMapSpec := &pipeline.CmAndSecretSpec{
			Names:     strings.Fields(record[7]),
			PatchJson: record[8],
		}
		configMapTask := &pipeline.CmAndSecretTask{
			Spec: configMapSpec,
		}
		secretSpec := &pipeline.CmAndSecretSpec{
			Names:     strings.Fields(record[9]),
			PatchJson: record[10],
		}
		secretTask := &pipeline.CmAndSecretTask{
			Spec: secretSpec,
		}
		payload := &pipeline.BulkUpdatePayload{
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
