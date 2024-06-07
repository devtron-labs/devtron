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
