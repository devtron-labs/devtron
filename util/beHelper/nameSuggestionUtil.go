/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package beHelper

import (
	"fmt"
	"github.com/devtron-labs/common-lib/git-manager/util"
	util2 "github.com/devtron-labs/devtron/util"
)

func GetCIPipelineName(appId int) string {
	return fmt.Sprintf("ci-%d-%s", appId, util2.Generate(4))
}

func GetCDPipelineName(appId int) string {
	return fmt.Sprintf("cd-%d-%s", appId, util2.Generate(4))
}

func GetAppWorkflowName(appId int) string {
	return fmt.Sprintf("wf-%d-%s", appId, util2.Generate(4))
}

func GetPipelineNameByPipelineType(pipelineType string, appId int) string {
	return fmt.Sprintf("%s-%d-%s", pipelineType, appId, util.Generate(4))
}
