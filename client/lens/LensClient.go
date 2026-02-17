/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package lens

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/caarlos0/env"
	"go.uber.org/zap"
)

type LensConfig struct {
	Url     string `env:"LENS_URL" envDefault:"http://lens-milandevtron-service:80" description:"Lens micro-service URL"`
	Timeout int    `env:"LENS_TIMEOUT" envDefault:"0" description:"Lens microservice timeout."` // in seconds
}
type StatusCode int

func (code StatusCode) IsSuccess() bool {
	return code >= 200 && code <= 299
}

type LensClient interface {
	GetAppMetrics(metricRequest *MetricRequest) (resBody []byte, resCode *StatusCode, err error)
	GetBulkAppMetrics(bulkRequest *BulkMetricRequest) (*LensResponse, *StatusCode, error)
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
	RequestBody interface{}
}

type LensResponse struct {
	Code   int             `json:"code,omitempty"`
	Status string          `json:"status,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Errors []*LensApiError `json:"errors,omitempty"`
}
type LensApiError struct {
	HttpStatusCode    int    `json:"-"`
	Code              string `json:"code,omitempty"`
	InternalMessage   string `json:"internalMessage,omitempty"`
	UserMessage       string `json:"userMessage,omitempty"`
	UserDetailMessage string `json:"userDetailMessage,omitempty"`
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
		session.logger.Errorw("error in lens communication ", "err", err)
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

type AppEnvPair struct {
	AppId int `json:"appId"`
	EnvId int `json:"envId"`
}

type BulkMetricRequest struct {
	AppEnvPairs []AppEnvPair `json:"appEnvPairs"`
	From        *time.Time   `json:"from"`
	To          *time.Time   `json:"to"`
}

type Metrics struct {
	AverageCycleTime       float64 `json:"average_cycle_time"`
	AverageLeadTime        float64 `json:"average_lead_time"`
	ChangeFailureRate      float64 `json:"change_failure_rate"`
	AverageRecoveryTime    float64 `json:"average_recovery_time"`
	AverageDeploymentSize  float32 `json:"average_deployment_size"`
	AverageLineAdded       float32 `json:"average_line_added"`
	AverageLineDeleted     float32 `json:"average_line_deleted"`
	LastFailedTime         string  `json:"last_failed_time"`
	RecoveryTimeLastFailed float64 `json:"recovery_time_last_failed"`
}

type AppEnvMetrics struct {
	AppId   int      `json:"appId"`
	EnvId   int      `json:"envId"`
	Metrics *Metrics `json:"metrics"`
	Error   string   `json:"error,omitempty"`
}

type BulkMetricsResponse struct {
	Results []AppEnvMetrics `json:"results"`
}

// DoraMetrics represents the new response structure from Lens API
type DoraMetrics struct {
	AppId                  int     `json:"app_id"`
	EnvId                  int     `json:"env_id"`
	DeploymentFrequency    float64 `json:"deployment_frequency"`       // Deployments per day
	ChangeFailureRate      float64 `json:"change_failure_rate"`        // Percentage
	MeanLeadTimeForChanges float64 `json:"mean_lead_time_for_changes"` // Minutes
	MeanTimeToRecovery     float64 `json:"mean_time_to_recovery"`      // Minutes
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

func (session *LensClientImpl) GetBulkAppMetrics(bulkRequest *BulkMetricRequest) (*LensResponse, *StatusCode, error) {
	u, err := url.Parse("deployment-metrics/bulk")
	if err != nil {
		return nil, nil, err
	}
	req := &ClientRequest{
		Method:      "GET",
		Path:        u.String(),
		RequestBody: bulkRequest,
	}
	session.logger.Infow("lens bulk req", "req", req)
	resBody, resCode, err := session.doRequest(req)
	if err != nil {
		return nil, resCode, err
	}
	if resCode.IsSuccess() {
		apiRes := &LensResponse{}
		err = json.Unmarshal(resBody, apiRes)
		if err != nil {
			return nil, resCode, err
		}
		return apiRes, resCode, nil
	}
	session.logger.Errorw("api err in git sensor response", "res", string(resBody))
	return nil, resCode, fmt.Errorf("res not success, Statuscode: %v", resCode)
}
