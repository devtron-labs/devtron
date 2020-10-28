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

package ArgoUtil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
)

type ArgoConfig struct {
	Url                string `env:"ACD_URL" envDefault:""`
	UserName           string `env:"ACD_USER" `
	Password           string `env:"ACD_PASSWORD" `
	Timeout            int    `env:"ACD_TIMEOUT" envDefault:"0"`        // in seconds
	InsecureSkipVerify bool   `env:"ACD_SKIP_VERIFY" envDefault:"true"` //ignore ssl verification
}

type ArgoSession struct {
	httpClient *http.Client
	logger     *zap.SugaredLogger
	baseUrl    *url.URL
}
type StatusCode int

func (code StatusCode) IsSuccess() bool {
	return code >= 200 && code <= 299
}

type ClientRequest struct {
	Method       string
	Path         string
	RequestBody  interface{}
	ResponseBody interface{}
}

func GetArgoConfig() (*ArgoConfig, error) {
	cfg := &ArgoConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func (session *ArgoSession) DoRequest(clientRequest *ClientRequest) (resBody []byte, resCode *StatusCode, err error) {
	if clientRequest.ResponseBody == nil {
		return nil, nil, fmt.Errorf("responce body cant be nil")
	}
	if reflect.ValueOf(clientRequest.ResponseBody).Kind() != reflect.Ptr {
		return nil, nil, fmt.Errorf("responsebody non pointer")
	}
	rel, err := session.baseUrl.Parse(clientRequest.Path)
	if err != nil {
		return nil, nil, err
	}
	var body io.Reader
	if clientRequest.RequestBody != nil {
		if req, err := json.Marshal(clientRequest.RequestBody); err != nil {
			return nil, nil, err
		} else {
			session.logger.Debugw("argo req with body", "body", string(req))
			body = bytes.NewBuffer(req)
		}

	}
	httpReq, err := http.NewRequest(clientRequest.Method, rel.String(), body)
	if err != nil {
		return nil, nil, err
	}
	httpRes, err := session.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, err
	}
	defer httpRes.Body.Close()
	resBody, err = ioutil.ReadAll(httpRes.Body)
	if err != nil {
		session.logger.Errorw("error in argocd communication ", "err", err)
		return nil, nil, err
	}
	status := StatusCode(httpRes.StatusCode)
	if status.IsSuccess() {
		err = json.Unmarshal(resBody, clientRequest.ResponseBody)
	} else {
		session.logger.Errorw("api err", "res", string(resBody))
		return resBody, &status, fmt.Errorf("res not success, code: %d ", status)
	}
	return resBody, &status, err
}

func NewArgoSession(config *ArgoConfig, logger *zap.SugaredLogger) (session *ArgoSession, err error) {
	/*location := "/api/v1/session"
	baseUrl, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}
	rel, err := baseUrl.Parse(location)
	param := map[string]string{}
	param["username"] = "admin"
	param["password"] = "argocd-server-6cd5bcffd4-j6kcx"
	paramJson, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("POST", rel.String(), bytes.NewBuffer(paramJson))
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.InsecureSkipVerify},
	}
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Transport: transCfg, Jar: cookieJar, Timeout: time.Duration(config.Timeout)}
	res, err := client.Do(req)
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, err
	}*/
	return &ArgoSession{}, nil
}
