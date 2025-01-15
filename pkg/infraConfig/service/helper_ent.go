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

package service

import (
	"github.com/caarlos0/env"
	v1 "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
)

func getDefaultInfraConfigFromEnv(envConfig *types.CiConfig) (*v1.InfraConfig, error) {
	infraConfiguration := &v1.InfraConfig{}
	err := env.Parse(infraConfiguration)
	if err != nil {
		return infraConfiguration, err
	}
	return infraConfiguration, nil
}
