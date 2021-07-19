package pipeline

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/bulkUpdate"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	jsonpatch "github.com/evanphx/json-patch"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
)

var bulkUpdateService *pipeline.BulkUpdateServiceImpl
var bulkUpdateRepository bulkUpdate.BulkUpdateRepositoryImpl

func setup() {
	config, _ := models.GetConfig()
	logger := util.NewSugardLogger()
	dbConnection, _ := models.NewDbConnection(config, logger)
	bulkUpdateRepository := bulkUpdate.NewBulkUpdateRepository(dbConnection, logger)
	bulkUpdateService = pipeline.NewBulkUpdateServiceImpl(bulkUpdateRepository, nil, nil, nil, nil, nil, "",
		pipeline.DefaultChart(""), util.MergeUtil{}, nil, nil, nil, nil, nil,
		nil, nil, nil, nil)
}

func TestBulkUpdateDeploymentTemplate(t *testing.T) {
	setup()
	type test struct {
		ApiVersion string
		Kind       string
		Payload    pipeline.BulkUpdatePayload
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
		includes := pipeline.NameIncludesExcludes{Names: namesIncludes}
		excludes := pipeline.NameIncludesExcludes{Names: namesExcludes}
		spec := pipeline.Spec{
			PatchJson: record[6]}
		task := pipeline.Tasks{
			Spec: spec,
		}
		payload := pipeline.BulkUpdatePayload{
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
			got, _ := bulkUpdateService.BulkUpdateDeploymentTemplate(tt.Payload)
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
