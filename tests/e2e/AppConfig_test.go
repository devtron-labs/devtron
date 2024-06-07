/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package e2e

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOverrideConfig(t *testing.T) {

	/*msg := json.RawMessage([]byte(`{
	   		 "image": {
				"tag": "1.2.3"
	    	}
	  	}`))*/

	fmt.Println("test========================")
	/*overrideRequest := &bean.ValuesOverrideRequest{
		ChartName:         "demo-app-4",
		TargetEnvironment: "prod-env",
	}

	b, _ := json.Marshal(overrideRequest)

	req, _ := http.NewRequest("POST", "/helm/values", bytes.NewBuffer(b))

	response := executeRequest(req)
	fmt.Println(response)
	assert.True(t, true)*/
}

func clearTable() {

}
func InsertDefaults() {

}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	/*rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr*/
	return nil
}
