
/*
 * Copyright (c) 2020 Devtron Labs
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
 *
 */
package e2e

import (
    //"fmt"
    //"github.com/go-resty/resty/v2"
    "testing"
    "encoding/json"
    "net/http"
    "github.com/stretchr/testify/assert"
    "github.com/devtron-labs/devtron/internal/util"
)

//TODO : Ask if we should reuse struct from other package or create new. What's better?
type AppCheckList struct {
    GitOps int `json:"gitOps"`
    Project int `json:"project"`
    Git int `json:"git"`
    Environment int `json:"environment"`
    Docker int `json:"docker"`
    HostUrl int `json:"hostUrl"`
}

type ChartCheckList struct {
       GitOps int `json:"gitOps"`
       Project int `json:"project"`
       Environment int `json:"environment"`
}

type CheckListResult struct {
    AppCheckList AppCheckList `json:"appChecklist"`
    ChartCheckList ChartCheckList `json:"chartChecklist"`
    IsAppCreated bool `json:"isAppCreated"`
}

type CheckListResponse struct {
    Code int `json:"code"`
    Status string `json:"status"`
    Result CheckListResult `json:"result"`
}

func TestCheckList(t *testing.T) {
    resp, err := makeApiCall("/orchestrator/global/checklist", http.MethodGet, "", true)
    if(handleError(err, "TestCheckList")) {
        return
    }
    assert.Equal(t, 200, resp.StatusCode())
    var checkListResponse CheckListResponse
    json.Unmarshal([]byte(resp.Body()), &checkListResponse)
    util.GetLogger().Infow("Printing response from test TestCheckList : ", "CheckListResponse object:", checkListResponse)
}