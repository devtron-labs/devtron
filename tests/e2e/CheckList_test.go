
/*
 * Copyright (c) 2020-2024. Devtron Inc.
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