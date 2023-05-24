
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
    "github.com/go-resty/resty/v2"
    "encoding/json"
    "net/http"
    "github.com/devtron-labs/devtron/internal/util"
    "io/ioutil"
    "os"
    "fmt"
)

type LogInResult struct {
    Token string `json:"token"`
}

type LogInResponse struct {
    Code int `json:"code"`
    Status string `json:"status"`
    Result LogInResult `json:"result"`
}

type EnvironmentConf struct {
    TestEnv string `json.testEnv`
    EnvList []Environment `json:"environment"`
}

type Environment struct {
    EnvironmentName string `json:"envName"`
    BaseServerUrl string `json:baseServerUrl`
    LogInUserName string `json:"loginUserName"`
    LogInUserPwd string `json:loginUserPwd`
}

//Give resty client along with some basic settings like cookie etc
func getRestyClient(setCookie bool) *resty.Client {
     baseServerUrl, _, _ := getBaseServerDetails()
     client := resty.New()
     client.SetBaseURL(baseServerUrl)
     if(setCookie) {
         client.SetCookie(&http.Cookie{Name:"argocd.token",
                 Value: getAuthToken(),
                 Path: "/",
                 Domain: baseServerUrl})
     }
     return client
}

//this function will read testEnvironmentConfig.json file and get baseServerUrl, username and pwd depending on environment
func getBaseServerDetails() (string, string, string) {
    var envToTest, baseServerUrl, loginUserName, loginUserPwd string

    //STEP 1 : Open the file
    testDataJsonFile, err := os.Open("../testdata/testEnvironmentConfig.json")
    if (nil != err) {
        util.GetLogger().Errorw("Unable to open the file. Error occurred !!", "err", err)
    }
    util.GetLogger().Infow("Opened testEnvironmentConfig.json file successfully !")

    defer testDataJsonFile.Close()

    // STEP 2 : Now get required values depending on environment requested.
    byteValue, _ := ioutil.ReadAll(testDataJsonFile)
    var envConf EnvironmentConf
    json.Unmarshal([]byte(byteValue), &envConf)

    envToTest = envConf.TestEnv

    for i := 0; i < len(envConf.EnvList); i++ {
        if(envToTest == envConf.EnvList[i].EnvironmentName) {
            baseServerUrl = envConf.EnvList[i].BaseServerUrl
            loginUserName = envConf.EnvList[i].LogInUserName
            loginUserPwd = envConf.EnvList[i].LogInUserPwd
        }
    }
    util.GetLogger().Infow("BaseServerUrl: ", "URL:", baseServerUrl)
    return baseServerUrl, loginUserName, loginUserPwd

}

//make the api call to the requested url based on http method requested
func makeApiCall(apiUrl string, method string, body string, setCookie bool)  (*resty.Response, error) {
    var resp *resty.Response
    var err error
    switch method {
        case "GET":
            return getRestyClient(setCookie).R().Get(apiUrl)
        case "POST":
            return getRestyClient(setCookie).R().SetBody(body).Post(apiUrl)
    }
    return resp, err
}

//Log the error and return boolean value indicating whether error occurred or not
func handleError(err error, testName string) bool {
    if(nil != err) {
        util.GetLogger().Errorw("Error occurred while invoking api for test:" +testName, "err", err)
        return true
    }
    return false
}

//support function to return auth token after log in
//TODO : if token is valid, don't call api again, error handling in invoking functions
func getAuthToken() string {
    _, loginUserName, loginUserPwd := getBaseServerDetails()
    jsonString := fmt.Sprintf(`{"username": "%s", "password": "%s"}`,loginUserName,loginUserPwd)
    fmt.Println("jsonString : ", jsonString)
    resp, err := makeApiCall("/orchestrator/api/v1/session", http.MethodPost, jsonString, false)
    if(handleError(err, "getAuthToken")) {
        return ""
    }
    var logInResponse LogInResponse
    json.Unmarshal([]byte(resp.Body()), &logInResponse)
    util.GetLogger().Infow("Getting Auth token : ", "AuthToken:", logInResponse.Result.Token)
    return logInResponse.Result.Token
}