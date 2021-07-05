package pipeline

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"go.uber.org/zap"
	"io"
	"log"
	"os"
	"strconv"
	"testing"
)

var chartService *pipeline.ChartServiceImpl
var chartRepository chartConfig.ChartRepositoryImpl

func setup() {
	config, _ := models.GetConfig()
	dbConnection, _ := models.NewDbConnection(config, &zap.SugaredLogger{})
	chartRepository := chartConfig.NewChartRepository(dbConnection)
	chartService = pipeline.NewChartServiceImpl(chartRepository, nil, nil, nil, nil, "",
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
		includes := pipeline.NameIncludesExcludes{Name: record[2]}
		excludes := pipeline.NameIncludesExcludes{Name: record[3]}
		spec := pipeline.Specs{
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
			got, _ := chartService.BulkUpdateDeploymentTemplate(tt.Payload)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
