package tests

import (
	"github.com/devtron-labs/devtron/enterprise/pkg/scanningResultsParser"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func loadData(t *testing.T, fileName string) string {

	jsonBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Error(err)
	}

	return string(jsonBytes)

}

func TestParsing(t *testing.T) {
	t.Run("imageScan results", func(tt *testing.T) {
		json := loadData(t, "image_scan.json")
		vulns := scanningResultsParser.ParseImageScanResult(json)
		assert.NotNil(t, vulns)
	})

	t.Run("codeScan results", func(tt *testing.T) {
		json := loadData(t, "code_scan.json")
		misConfigurations := scanningResultsParser.ParseCodeScanResult(json)
		assert.NotNil(t, misConfigurations)
	})

	t.Run("ParseMisConfigurations", func(tt *testing.T) {
		json := loadData(t, "code_scan.json")
		exposedSecrets := scanningResultsParser.ParseK8sConfigScanResult(json)
		assert.NotNil(t, exposedSecrets)
	})
}
