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
