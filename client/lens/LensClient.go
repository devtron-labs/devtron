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

package lens

import (
	"bytes"
	"encoding/json"
	"github.com/caarlos0/env"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type LensConfig struct {
	Url     string `env:"LENS_URL" envDefault:"http://lens-milandevtron-service:80"`
	Timeout int    `env:"LENS_TIMEOUT" envDefault:"0"` // in seconds
}
type StatusCode int

func (code StatusCode) IsSuccess() bool {
	return code >= 200 && code <= 299
}

type LensClient interface {
	GetAppMetrics(metricRequest *MetricRequest) (resBody []byte, resCode *StatusCode, err error)
}
type LensClientImpl struct {
	httpClient *http.Client
	logger     *zap.SugaredLogger
	baseUrl    *url.URL
}

func GetLensConfig() (*LensConfig, error) {
	cfg := &LensConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func NewLensClientImpl(config *LensConfig, logger *zap.SugaredLogger) (*LensClientImpl, error) {
	baseUrl, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: time.Duration(config.Timeout) * time.Second}
	return &LensClientImpl{httpClient: client, logger: logger, baseUrl: baseUrl}, nil
}

type ClientRequest struct {
	Method      string
	Path        string
	RequestBody *MetricRequest
}

func (session *LensClientImpl) doRequest(clientRequest *ClientRequest) (resBody []byte, resCode *StatusCode, err error) {
	rel, err := session.baseUrl.Parse(clientRequest.Path)
	if err != nil {
		return nil, nil, err
	}
	var body io.Reader
	if clientRequest.RequestBody != nil {
		if req, err := json.Marshal(clientRequest.RequestBody); err != nil {
			return nil, nil, err
		} else {
			session.logger.Infow("argo req with body", "body", string(req))
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
	return resBody, &status, err
}

type MetricRequest struct {
	AppId int    `json:"app_id"`
	EnvId int    `json:"env_id"`
	From  string `json:"from"`
	To    string `json:"to"`
}

func (session *LensClientImpl) GetAppMetrics(metricRequest *MetricRequest) (resBody []byte, resCode *StatusCode, err error) {
	params := url.Values{}
	params.Add("app_id", strconv.Itoa(metricRequest.AppId))
	params.Add("env_id", strconv.Itoa(metricRequest.EnvId))
	params.Add("from", metricRequest.From)
	params.Add("to", metricRequest.To)
	u, err := url.Parse("deployment-metrics")
	if err != nil {
		return nil, nil, err
	}
	u.RawQuery = params.Encode()
	req := &ClientRequest{
		Method: "GET",
		Path:   u.String(),
	}
	session.logger.Infow("lens req", "req", req)
	resBody, resCode, err = session.doRequest(req)
	return resBody, resCode, err
}
