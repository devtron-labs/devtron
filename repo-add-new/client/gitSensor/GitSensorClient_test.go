package gitSensor

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGitSensorClientWithValidConfigAndGrpcEnabled(t *testing.T) {

	config := &ClientConfig{
		Url:      "127.0.0.1:7070",
		Protocol: "GRPC",
	}

	logger, err := util.NewSugardLogger()
	_, err = NewGitSensorClient(logger, config)

	assert.Nil(t, err)
}

func TestNewGitSensorClientWithValidConfigAndGrpcDisabled(t *testing.T) {

	config := &ClientConfig{
		Url:      "127.0.0.1:7070",
		Protocol: "REST",
	}

	logger, err := util.NewSugardLogger()
	_, err = NewGitSensorClient(logger, config)

	assert.Nil(t, err)
}
