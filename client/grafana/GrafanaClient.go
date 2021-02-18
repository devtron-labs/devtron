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

package grafana

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type GrafanaClientConfig struct {
	GrafanaUsername string `env:"GRAFANA_USERNAME" envDefault:"admin"`
	GrafanaPassword string `env:"GRAFANA_PASSWORD" envDefault:"prom-operator"`
	GrafanaOrgId    int    `env:"GRAFANA_ORG_ID" envDefault:"2"`
	DestinationURL  string `env:"GRAFANA_URL" envDefault:""`
}

const PromDatasource = "/api/datasources"
const AddPromDatasource = "/api/datasources"
const DeletePromDatasource = "/api/datasources/%d"
const UpdatePromDatasource = "/api/datasources/%d"
const GetPromDatasource = "/api/datasources/%d"

func GetGrafanaClientConfig() (*GrafanaClientConfig, error) {
	cfg := &GrafanaClientConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, errors.New("could not get grafana service url")
	}
	return cfg, err
}

type GrafanaClient interface {
	GetAllDatasource() ([]*GetPrometheusDatasourceResponse, error)
	CreateDatasource(createDatasourceRequest CreateDatasourceRequest) (*DatasourceResponse, error)
	GetDatasource(datasourceId int) (*GetPrometheusDatasourceResponse, error)
	UpdateDatasource(updateDatasourceRequest UpdateDatasourceRequest, datasourceId int) (*DatasourceResponse, error)
}

type StatusCode int

func (code StatusCode) IsSuccess() bool {
	return code >= 200 && code <= 299
}

type DatasourceResponse struct {
	Id      int    `json:"id"`
	Message string `json:"message"`
	Name    string `json:"name"`
}

type UpdateDatasourceRequest struct {
	Id                int              `json:"id"`
	OrgId             int              `json:"orgId"`
	Name              string           `json:"name"`
	Type              string           `json:"type"`
	TypeLogoUrl       string           `json:"typeLogoUrl"`
	Access            string           `json:"access"`
	Url               string           `json:"url"`
	Password          string           `json:"password"`
	User              string           `json:"user"`
	Database          string           `json:"database"`
	BasicAuth         bool             `json:"basicAuth"`
	BasicAuthUser     string           `json:"basicAuthUser"`
	BasicAuthPassword string           `json:"basicAuthPassword"`
	WithCredentials   bool             `json:"withCredentials"`
	IsDefault         bool             `json:"isDefault"`
	JsonData          JsonData         `json:"jsonData"`
	SecureJsonFields  SecureJsonFields `json:"secureJsonFields"`
	Version           *int             `json:"version"`
	ReadOnly          bool             `json:"readOnly"`
	BasicAuthPayload  `json:",inline"`
}

type SecureJsonFields struct {
}

type JsonData struct {
	HttpMethod    string   `json:"httpMethod,omitempty"`
	KeepCookies   []string `json:"keepCookies,omitempty"`
	AuthType      string   `json:"authType,omitempty"`
	DefaultRegion string   `json:"defaultRegion,omitempty"`
	TlsAuth       bool     `json:"tlsAuth"`
}

type GetPrometheusDatasourceResponse struct {
	Id                int              `json:"id"`
	OrgId             int              `json:"orgId"`
	Name              string           `json:"name"`
	Type              string           `json:"type"`
	TypeLogoUrl       string           `json:"typeLogoUrl"`
	Access            string           `json:"access"`
	Url               string           `json:"url"`
	Password          string           `json:"password"`
	User              string           `json:"user"`
	Database          string           `json:"database"`
	BasicAuth         bool             `json:"basicAuth"`
	BasicAuthUser     string           `json:"basicAuthUser"`
	BasicAuthPassword string           `json:"basicAuthPassword"`
	WithCredentials   bool             `json:"withCredentials"`
	IsDefault         bool             `json:"isDefault"`
	JsonData          JsonData         `json:"jsonData"`
	SecureJsonFields  SecureJsonFields `json:"secureJsonFields"`
	Version           *int             `json:"version"`
	ReadOnly          bool             `json:"readOnly"`
}

type CreateDatasourceRequest struct {
	Name             string `json:"name"`
	Type             string `json:"type"`
	Url              string `json:"url"`
	Access           string `json:"access"`
	BasicAuth        bool   `json:"basicAuth"`
	BasicAuthPayload `json:",inline"`
}

type BasicAuthPayload struct {
	BasicAuthUser     string          `json:"basicAuthUser,omitempty"`
	BasicAuthPassword string          `json:"basicAuthPassword,omitempty"`
	SecureJsonData    *SecureJsonData `json:"secureJsonData,omitempty"`
	JsonData          *JsonData       `json:"jsonData,omitempty"`
}

type SecureJsonData struct {
	BasicAuthPassword string `json:"basicAuthPassword,omitempty"`
	TlsClientCert     string `json:"tlsClientCert,omitempty"`
	TlsClientKey      string `json:"tlsClientKey,omitempty"`
}

type GrafanaClientImpl struct {
	logger            *zap.SugaredLogger
	client            *http.Client
	config            *GrafanaClientConfig
	attributesService attributes.AttributesService
}

func NewGrafanaClientImpl(logger *zap.SugaredLogger, client *http.Client, config *GrafanaClientConfig, attributesService attributes.AttributesService) *GrafanaClientImpl {
	return &GrafanaClientImpl{logger: logger, client: client, config: config, attributesService: attributesService}
}

func (impl *GrafanaClientImpl) GetAllDatasource() ([]*GetPrometheusDatasourceResponse, error) {
	if len(impl.config.DestinationURL) == 0 {
		hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
		if err != nil {
			return nil, err
		}
		if hostUrl != nil {
			impl.config.DestinationURL = strings.ReplaceAll(hostUrl.Value, "//", "//%s:%s")
		}
	}
	url := impl.config.DestinationURL + PromDatasource
	url = fmt.Sprintf(url, impl.config.GrafanaUsername, impl.config.GrafanaPassword)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		impl.logger.Errorw("error while fetching data source", "err", err)
		return nil, err
	}
	req.Header.Set("X-Grafana-Org-Id", strconv.Itoa(impl.config.GrafanaOrgId))
	resp, err := impl.client.Do(req)
	if err != nil {
		impl.logger.Errorw("err", err)
		return nil, err
	}
	status := StatusCode(resp.StatusCode)
	resBody, err := ioutil.ReadAll(resp.Body)
	var apiRes []*GetPrometheusDatasourceResponse
	if status.IsSuccess() {
		if err != nil {
			impl.logger.Errorw("error in grafana communication ", "err", err)
			return nil, err
		}
		err = json.Unmarshal(resBody, &apiRes)
		if err != nil {
			impl.logger.Errorw("error in grafana resp unmarshalling ", "err", err)
			return nil, err
		}
	} else {
		impl.logger.Errorw("api err", "res", string(resBody))
		return nil, fmt.Errorf("res not success, code: %d ,response body: %s", status, string(resBody))
	}

	impl.logger.Debugw("grafana resp", "body", apiRes)
	return apiRes, nil
}

func (impl *GrafanaClientImpl) GetDatasource(datasourceId int) (*GetPrometheusDatasourceResponse, error) {
	if len(impl.config.DestinationURL) == 0 {
		hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
		if err != nil {
			return nil, err
		}
		if hostUrl != nil {
			impl.config.DestinationURL = strings.ReplaceAll(hostUrl.Value, "//", "//%s:%s")
		}
	}
	url := impl.config.DestinationURL + GetPromDatasource
	url = fmt.Sprintf(url, impl.config.GrafanaUsername, impl.config.GrafanaPassword, datasourceId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		impl.logger.Errorw("error while fetching data source", "err", err)
		return nil, err
	}
	req.Header.Set("X-Grafana-Org-Id", strconv.Itoa(impl.config.GrafanaOrgId))
	resp, err := impl.client.Do(req)
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return nil, err
	}
	status := StatusCode(resp.StatusCode)
	resBody, err := ioutil.ReadAll(resp.Body)
	apiRes := &GetPrometheusDatasourceResponse{}
	if status.IsSuccess() {
		if err != nil {
			impl.logger.Errorw("error in grafana communication ", "err", err)
			return nil, err
		}
		err = json.Unmarshal(resBody, apiRes)
		if err != nil {
			impl.logger.Errorw("error in grafana resp unmarshal ", "err", err)
			return nil, err
		}
	} else {
		impl.logger.Errorw("api err", "res", string(resBody))
		return nil, fmt.Errorf("res not success, code: %d ,response body: %s", status, string(resBody))
	}
	impl.logger.Debugw("grafana resp", "body", apiRes)
	return apiRes, nil
}

func (impl *GrafanaClientImpl) UpdateDatasource(updateDatasourceRequest UpdateDatasourceRequest, datasourceId int) (*DatasourceResponse, error) {
	if len(impl.config.DestinationURL) == 0 {
		hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
		if err != nil {
			return nil, err
		}
		if hostUrl != nil {
			impl.config.DestinationURL = strings.ReplaceAll(hostUrl.Value, "//", "//%s:%s")
		}
	}
	updateDatasourceRequest.OrgId = impl.config.GrafanaOrgId
	body, err := json.Marshal(updateDatasourceRequest)
	if err != nil {
		impl.logger.Errorw("error while marshaling request ", "err", err)
		return nil, err
	}
	var reqBody = []byte(body)
	url := impl.config.DestinationURL + UpdatePromDatasource
	url = fmt.Sprintf(url, impl.config.GrafanaUsername, impl.config.GrafanaPassword, datasourceId)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(reqBody))
	if err != nil {
		impl.logger.Errorw("error while updating data source", "err", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Grafana-Org-Id", strconv.Itoa(impl.config.GrafanaOrgId))
	resp, err := impl.client.Do(req)
	if err != nil {
		impl.logger.Errorw("err", err)
		return nil, err
	}
	status := StatusCode(resp.StatusCode)
	resBody, err := ioutil.ReadAll(resp.Body)
	apiRes := &DatasourceResponse{}
	if status.IsSuccess() {
		if err != nil {
			impl.logger.Errorw("error in grafana communication ", "err", err)
			return nil, err
		}
		err = json.Unmarshal(resBody, apiRes)
		if err != nil {
			impl.logger.Errorw("error in grafana resp unmarshalling ", "err", err)
			return nil, err
		}
	} else {
		impl.logger.Errorw("api err", "res", string(resBody))
		return nil, fmt.Errorf("res not success, code: %d ,response body: %s", status, string(resBody))
	}

	impl.logger.Debugw("grafana resp", "body", apiRes)
	return apiRes, nil
}

func (impl *GrafanaClientImpl) deleteDatasource(updateDatasourceRequest CreateDatasourceRequest, datasourceId int) (*DatasourceResponse, error) {
	if len(impl.config.DestinationURL) == 0 {
		hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
		if err != nil {
			return nil, err
		}
		if hostUrl != nil {
			impl.config.DestinationURL = strings.ReplaceAll(hostUrl.Value, "//", "//%s:%s")
		}
	}
	body, err := json.Marshal(updateDatasourceRequest)
	if err != nil {
		impl.logger.Errorw("error while marshaling request ", "err", err)
		return nil, err
	}
	var reqBody = []byte(body)
	url := impl.config.DestinationURL + DeletePromDatasource
	url = fmt.Sprintf(url, impl.config.GrafanaUsername, impl.config.GrafanaPassword, datasourceId)
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(reqBody))
	if err != nil {
		impl.logger.Errorw("error while updating data source", "err", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Grafana-Org-Id", strconv.Itoa(impl.config.GrafanaOrgId))
	resp, err := impl.client.Do(req)
	if err != nil {
		impl.logger.Errorw("err", err)
		return nil, err
	}
	status := StatusCode(resp.StatusCode)
	resBody, err := ioutil.ReadAll(resp.Body)
	apiRes := &DatasourceResponse{}
	if status.IsSuccess() {
		if err != nil {
			impl.logger.Errorw("error in grafana communication ", "err", err)
			return nil, err
		}
		err = json.Unmarshal(resBody, apiRes)
		if err != nil {
			impl.logger.Errorw("error in grafana resp unmarshalling ", "err", err)
			return nil, err
		}
	} else {
		impl.logger.Errorw("api err", "res", string(resBody))
		return nil, fmt.Errorf("res not success, code: %d ,response body: %s", status, string(resBody))
	}

	impl.logger.Debugw("grafana resp", "body", apiRes)
	return apiRes, nil
}

func (impl *GrafanaClientImpl) CreateDatasource(createDatasourceRequest CreateDatasourceRequest) (*DatasourceResponse, error) {
	if len(impl.config.DestinationURL) == 0 {
		hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
		if err != nil {
			return nil, err
		}
		if hostUrl != nil {
			impl.config.DestinationURL = strings.ReplaceAll(hostUrl.Value, "//", "//%s:%s")
		}
	}

	body, err := json.Marshal(createDatasourceRequest)
	if err != nil {
		impl.logger.Errorw("error while marshaling request ", "err", err)
		return nil, err
	}
	var reqBody = []byte(body)
	url := impl.config.DestinationURL + AddPromDatasource
	url = fmt.Sprintf(url, impl.config.GrafanaUsername, impl.config.GrafanaPassword)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	req.Header.Set("X-Grafana-Org-Id", strconv.Itoa(impl.config.GrafanaOrgId))
	if err != nil {
		impl.logger.Errorw("error while adding datasource", "err", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := impl.client.Do(req)
	if err != nil {
		impl.logger.Errorw("err", err)
		return nil, err
	}
	status := StatusCode(resp.StatusCode)
	resBody, err := ioutil.ReadAll(resp.Body)

	apiRes := &DatasourceResponse{}
	if status.IsSuccess() {
		if err != nil {
			impl.logger.Errorw("error in grafana communication ", "err", err)
			return nil, err
		}
		err = json.Unmarshal(resBody, apiRes)
		if err != nil {
			impl.logger.Errorw("error in grafana resp unmarshalling ", "err", err)
			return nil, err
		}
	} else {
		impl.logger.Errorw("api err", "status", status, "res", string(resBody))
		err = &util.ApiError{Code: strconv.Itoa(int(status)), UserMessage: "Data source with same name already exists", InternalMessage: string(resBody)}
		return nil, err
	}

	impl.logger.Debugw("grafana resp", "body", apiRes)
	return apiRes, nil
}
