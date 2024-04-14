package tests

import (
	"github.com/devtron-labs/devtron/enterprise/pkg/scanningResultsParser"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

var json = ""

func loadData(t *testing.T) {
	if json != "" {
		return
	}
	jsonBytes, err := ioutil.ReadFile("dashboard_code_scan.json")
	if err != nil {
		t.Error(err)
	}

	json = string(jsonBytes)

}

func TestParsing(t *testing.T) {
	loadData(t)
	t.Run("ParseVulnerabilities", func(tt *testing.T) {
		vulns := scanningResultsParser.ParseVulnerabilities(json)
		assert.NotNil(t, vulns)
	})

	t.Run("ParseMisConfigurations", func(tt *testing.T) {
		misConfigurations := scanningResultsParser.ParseMisConfigurations(json)
		assert.NotNil(t, misConfigurations)
	})

	t.Run("ParseMisConfigurations", func(tt *testing.T) {
		exposedSecrets := scanningResultsParser.ParseExposedSecrets(json)
		assert.NotNil(t, exposedSecrets)
	})
}
