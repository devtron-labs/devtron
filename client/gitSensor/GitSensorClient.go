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

package gitSensor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"time"
)

//-----------
type GitSensorResponse struct {
	Code   int                  `json:"code,omitempty"`
	Status string               `json:"status,omitempty"`
	Result json.RawMessage      `json:"result,omitempty"`
	Errors []*GitSensorApiError `json:"errors,omitempty"`
}
type GitSensorApiError struct {
	HttpStatusCode    int    `json:"-"`
	Code              string `json:"code,omitempty"`
	InternalMessage   string `json:"internalMessage,omitempty"`
	UserMessage       string `json:"userMessage,omitempty"`
	UserDetailMessage string `json:"userDetailMessage,omitempty"`
}

//---------------
type FetchScmChangesRequest struct {
	PipelineMaterialId int    `json:"pipelineMaterialId"`
	From               string `json:"from"`
	To                 string `json:"to"`
}
type HeadRequest struct {
	MaterialIds []int `json:"materialIds"`
}

type SourceType string

type CiPipelineMaterial struct {
	Id            int
	GitMaterialId int
	Type          SourceType
	Value         string
	Active        bool
	GitCommit     GitCommit
}

type GitMaterial struct {
	Id               int
	GitProviderId    int
	Url              string
	Name             string
	CheckoutLocation string
	CheckoutStatus   bool
	CheckoutMsgAny   string
	Deleted          bool
	FetchSubmodules  bool
}
type GitProvider struct {
	Id            int
	Name          string
	Url           string
	UserName      string
	Password      string
	SshPrivateKey string
	AccessToken   string
	Active        bool
	AuthMode      repository.AuthMode
}

type GitCommit struct {
	Commit      string //git hash
	Author      string
	Date        time.Time
	Message     string
	Changes     []string
	WebhookData *WebhookData
}

type WebhookData struct {
	Id              int               `json:"id"`
	EventActionType string            `json:"eventActionType"`
	Data            map[string]string `json:"data"`
}

type MaterialChangeResp struct {
	Commits        []*GitCommit `json:"commits"`
	LastFetchTime  time.Time    `json:"lastFetchTime"`
	IsRepoError    bool         `json:"isRepoError"`
	RepoErrorMsg   string       `json:"repoErrorMsg"`
	IsBranchError  bool         `json:"isBranchError"`
	BranchErrorMsg string       `json:"branchErrorMsg"`
}

type CommitMetadataRequest struct {
	PipelineMaterialId int    `json:"pipelineMaterialId"`
	GitHash            string `json:"gitHash"`
	GitTag             string `json:"gitTag"`
	BranchName         string `json:"branchName"`
}

type RefreshGitMaterialRequest struct {
	GitMaterialId int `json:"gitMaterialId"`
}

type RefreshGitMaterialResponse struct {
	Message       string    `json:"message"`
	ErrorMsg      string    `json:"errorMsg"`
	LastFetchTime time.Time `json:"lastFetchTime"`
}

type WebhookDataRequest struct {
	Id int `json:"id"`
}

type WebhookEventConfigRequest struct {
	GitHostId int `json:"gitHostId"`
	EventId   int `json:"eventId"`
}

type WebhookEventConfig struct {
	Id            int       `json:"id"`
	GitHostId     int       `json:"gitHostId"`
	Name          string    `json:"name"`
	EventTypesCsv string    `json:"eventTypesCsv"`
	ActionType    string    `json:"actionType"`
	IsActive      bool      `json:"isActive"`
	CreatedOn     time.Time `json:"createdOn"`
	UpdatedOn     time.Time `json:"updatedOn"`

	Selectors []*WebhookEventSelectors `json:"selectors"`
}

type WebhookEventSelectors struct {
	Id               int       `json:"id"`
	EventId          int       `json:"eventId"`
	Name             string    `json:"name"`
	Selector         string    `json:"selector"`
	ToShow           bool      `json:"toShow"`
	ToShowInCiFilter bool      `json:"toShowInCiFilter"`
	FixValue         string    `json:"fixValue"`
	PossibleValues   string    `json:"possibleValues"`
	IsActive         bool      `json:"isActive"`
	CreatedOn        time.Time `json:"createdOn"`
	UpdatedOn        time.Time `json:"updatedOn"`
}

type WebhookPayloadDataRequest struct {
	CiPipelineMaterialId int    `json:"ciPipelineMaterialId"`
	Limit                int    `json:"limit"`
	Offset               int    `json:"offset"`
	EventTimeSortOrder   string `json:"eventTimeSortOrder"`
}

type WebhookPayloadDataResponse struct {
	Filters       map[string]string                     `json:"filters"`
	RepositoryUrl string                                `json:"repositoryUrl"`
	Payloads      []*WebhookPayloadDataPayloadsResponse `json:"payloads"`
}

type WebhookPayloadDataPayloadsResponse struct {
	ParsedDataId        int       `json:"parsedDataId"`
	EventTime           time.Time `json:"eventTime"`
	MatchedFiltersCount int       `json:"matchedFiltersCount"`
	FailedFiltersCount  int       `json:"failedFiltersCount"`
	MatchedFilters      bool      `json:"matchedFilters"`
}

type WebhookPayloadFilterDataRequest struct {
	CiPipelineMaterialId int `json:"ciPipelineMaterialId"`
	ParsedDataId         int `json:"parsedDataId"`
}

type WebhookPayloadFilterDataResponse struct {
	PayloadId     int                                         `json:"payloadId"`
	PayloadJson   string                                      `json:"payloadJson"`
	SelectorsData []*WebhookPayloadFilterDataSelectorResponse `json:"selectorsData"`
}

type WebhookPayloadFilterDataSelectorResponse struct {
	SelectorName      string `json:"selectorName"`
	SelectorCondition string `json:"selectorCondition"`
	SelectorValue     string `json:"selectorValue"`
	Match             bool   `json:"match"`
}

type GitSensorClient interface {
	GetHeadForPipelineMaterials(req *HeadRequest) (material []*CiPipelineMaterial, err error)
	FetchChanges(changeRequest *FetchScmChangesRequest) (materialChangeResp *MaterialChangeResp, err error)
	GetCommitMetadata(commitMetadataRequest *CommitMetadataRequest) (*GitCommit, error)

	SaveGitProvider(provider *GitProvider) (providerRes *GitProvider, err error)
	AddRepo(material []*GitMaterial) (materialRes []*GitMaterial, err error)
	UpdateRepo(material *GitMaterial) (materialRes *GitMaterial, err error)
	SavePipelineMaterial(material []*CiPipelineMaterial) (materialRes []*CiPipelineMaterial, err error)
	RefreshGitMaterial(req *RefreshGitMaterialRequest) (refreshRes *RefreshGitMaterialResponse, err error)

	GetWebhookData(req *WebhookDataRequest) (*WebhookData, error)

	GetAllWebhookEventConfigForHost(req *WebhookEventConfigRequest) (webhookEvents []*WebhookEventConfig, err error)
	GetWebhookEventConfig(req *WebhookEventConfigRequest) (webhookEvent *WebhookEventConfig, err error)
	GetWebhookPayloadDataForPipelineMaterialId(req *WebhookPayloadDataRequest) (response *WebhookPayloadDataResponse, err error)
	GetWebhookPayloadFilterDataForPipelineMaterialId(req *WebhookPayloadFilterDataRequest) (response *WebhookPayloadFilterDataResponse, err error)
}

//----------------------impl
type GitSensorConfig struct {
	Url     string `env:"GIT_SENSOR_URL" envDefault:"http://localhost:9999"`
	Timeout int    `env:"GIT_SENSOR_TIMEOUT" envDefault:"0"` // in seconds
}

type GitSensorClientImpl struct {
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

func GetGitSensorConfig() (*GitSensorConfig, error) {
	cfg := &GitSensorConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func (session *GitSensorClientImpl) doRequest(clientRequest *ClientRequest) (resBody []byte, resCode *StatusCode, err error) {
	if clientRequest.ResponseBody == nil {
		return nil, nil, fmt.Errorf("response body cant be nil")
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
		apiRes := &GitSensorResponse{}
		err = json.Unmarshal(resBody, apiRes)
		if apiStatus := StatusCode(apiRes.Code); apiStatus.IsSuccess() {
			err = json.Unmarshal(apiRes.Result, clientRequest.ResponseBody)
			return resBody, &apiStatus, err
		} else {
			session.logger.Errorw("api err", "res", apiRes.Errors)
			return resBody, &apiStatus, fmt.Errorf("err in api res")
		}
	} else {
		session.logger.Errorw("api err", "res", string(resBody))
		return resBody, &status, fmt.Errorf("res not success, code: %d ", status)
	}
	return resBody, &status, err
}

func NewGitSensorSession(config *GitSensorConfig, logger *zap.SugaredLogger) (session *GitSensorClientImpl, err error) {
	baseUrl, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: time.Duration(config.Timeout) * time.Second}
	return &GitSensorClientImpl{httpClient: client, logger: logger, baseUrl: baseUrl}, nil
}

func (session GitSensorClientImpl) GetHeadForPipelineMaterials(req *HeadRequest) (material []*CiPipelineMaterial, err error) {
	request := &ClientRequest{ResponseBody: &material, Method: "POST", RequestBody: req, Path: "git-head"}
	_, _, err = session.doRequest(request)
	return material, err
}

func (session GitSensorClientImpl) FetchChanges(changeRequest *FetchScmChangesRequest) (materialChangeResp *MaterialChangeResp, err error) {
	materialChangeResp = new(MaterialChangeResp)
	request := &ClientRequest{ResponseBody: materialChangeResp, Method: "POST", RequestBody: changeRequest, Path: "git-changes"}
	_, _, err = session.doRequest(request)
	return materialChangeResp, err
}

func (session GitSensorClientImpl) SaveGitProvider(provider *GitProvider) (providerRes *GitProvider, err error) {
	providerRes = new(GitProvider)
	request := &ClientRequest{ResponseBody: providerRes, Method: "POST", RequestBody: provider, Path: "git-provider"}
	_, _, err = session.doRequest(request)
	return providerRes, err
}

func (session GitSensorClientImpl) AddRepo(material []*GitMaterial) (materialRes []*GitMaterial, err error) {
	request := &ClientRequest{ResponseBody: &materialRes, Method: "POST", RequestBody: material, Path: "git-repo"}
	_, _, err = session.doRequest(request)
	return materialRes, err
}

func (session GitSensorClientImpl) UpdateRepo(material *GitMaterial) (materialRes *GitMaterial, err error) {
	request := &ClientRequest{ResponseBody: &materialRes, Method: "PUT", RequestBody: material, Path: "git-repo"}
	_, _, err = session.doRequest(request)
	return materialRes, err
}

func (session GitSensorClientImpl) SavePipelineMaterial(material []*CiPipelineMaterial) (materialRes []*CiPipelineMaterial, err error) {
	request := &ClientRequest{ResponseBody: &materialRes, Method: "POST", RequestBody: &material, Path: "git-pipeline-material"}
	_, _, err = session.doRequest(request)
	return materialRes, err
}

func (session GitSensorClientImpl) GetCommitMetadata(commitMetadataRequest *CommitMetadataRequest) (*GitCommit, error) {
	commit := new(GitCommit)
	request := &ClientRequest{ResponseBody: commit, Method: "POST", RequestBody: commitMetadataRequest, Path: "commit-metadata"}
	_, _, err := session.doRequest(request)
	return commit, err
}

func (session GitSensorClientImpl) RefreshGitMaterial(req *RefreshGitMaterialRequest) (refreshRes *RefreshGitMaterialResponse, err error) {
	refreshRes = new(RefreshGitMaterialResponse)
	request := &ClientRequest{ResponseBody: refreshRes, Method: "POST", RequestBody: req, Path: "git-repo/refresh"}
	_, _, err = session.doRequest(request)
	return refreshRes, err
}

func (session GitSensorClientImpl) GetWebhookData(req *WebhookDataRequest) (*WebhookData, error) {
	webhookData := new(WebhookData)
	request := &ClientRequest{ResponseBody: webhookData, Method: "GET", RequestBody: req, Path: "webhook/data"}
	_, _, err := session.doRequest(request)
	return webhookData, err
}

func (session GitSensorClientImpl) GetAllWebhookEventConfigForHost(req *WebhookEventConfigRequest) (webhookEvents []*WebhookEventConfig, err error) {
	request := &ClientRequest{ResponseBody: &webhookEvents, Method: "GET", RequestBody: req, Path: "/webhook/host/events"}
	_, _, err = session.doRequest(request)
	return webhookEvents, err
}

func (session GitSensorClientImpl) GetWebhookEventConfig(req *WebhookEventConfigRequest) (webhookEvent *WebhookEventConfig, err error) {
	request := &ClientRequest{ResponseBody: &webhookEvent, Method: "GET", RequestBody: req, Path: "/webhook/host/event"}
	_, _, err = session.doRequest(request)
	return webhookEvent, err
}

func (session GitSensorClientImpl) GetWebhookPayloadDataForPipelineMaterialId(req *WebhookPayloadDataRequest) (response *WebhookPayloadDataResponse, err error) {
	request := &ClientRequest{ResponseBody: &response, Method: "GET", RequestBody: req, Path: "/webhook/ci-pipeline-material/payload-data"}
	_, _, err = session.doRequest(request)
	return response, err
}

func (session GitSensorClientImpl) GetWebhookPayloadFilterDataForPipelineMaterialId(req *WebhookPayloadFilterDataRequest) (response *WebhookPayloadFilterDataResponse, err error) {
	request := &ClientRequest{ResponseBody: &response, Method: "GET", RequestBody: req, Path: "/webhook/ci-pipeline-material/payload-filter-data"}
	_, _, err = session.doRequest(request)
	return response, err
}
