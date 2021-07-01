package pipeline

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	Pipeline "github.com/devtron-labs/devtron/pkg/pipeline"
	"io"
	"log"
	"os"
	"strconv"
	"testing"
)

func TestBulkUpdateDeploymentTemplate(t *testing.T) {
	type test struct {
		ApiVersion string
		Kind       string
		Payload    Pipeline.BulkUpdateInput
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
		includes := Pipeline.NameIncludesExcludes{Name: record[2]}
		excludes := Pipeline.NameIncludesExcludes{Name: record[3]}
		spec := Pipeline.Specs{
			PatchJson: record[6]}
		task := Pipeline.Tasks{
			Spec: spec,
		}
		payload := Pipeline.BulkUpdateInput{
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
			got, _ := Pipeline.ChartService.BulkUpdateDeploymentTemplate(Pipeline.ChartServiceImpl{}, tt.Payload)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
