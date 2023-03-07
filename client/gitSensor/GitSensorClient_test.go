package gitSensor

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGitSensorClientWithValidConfig(t *testing.T) {

	config := &ClientConfig{
		Url:     "127.0.0.1:7070",
		UseGrpc: true,
	}

	logger, err := util.NewSugardLogger()
	_, err = NewGitSensorClient(logger, config)

	assert.Nil(t, err)
}
