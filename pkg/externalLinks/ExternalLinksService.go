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

package externalLinks

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	"go.uber.org/zap"
	"time"
)

type ExternalLinksService interface {
	Create(requests []*ExternalLinkoutRequest, userId int32) (*ExternalLinksCreateUpdateResponse, error)
	GetAllActiveTools() ([]ExternalLinksMonitoringToolsRequest, error)
	FetchAllActiveLinks(clusterIds int) ([]*ExternalLinkoutRequest, error)
	Update(request *ExternalLinkoutRequest) (*ExternalLinksCreateUpdateResponse, error)
	DeleteLink(id int, userId int32) (*ExternalLinksCreateUpdateResponse, error)
}
type ExternalLinksServiceImpl struct {
	logger                          *zap.SugaredLogger
	externalLinksToolsRepository    ExternalLinksToolsRepository
	externalLinksClustersRepository ExternalLinksClustersRepository
	externalLinksRepository         ExternalLinksRepository
	userAuthService                 user.UserAuthService
}
type ExternalLinksMonitoringToolsRequest struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}
type ExternalLinkoutRequest struct {
	Id               int    `json:"id"`
	Name             string `json:"name"`
	Url              string `json:"url"`
	Active           bool   `json:"active"`
	MonitoringToolId int    `json:"monitoringToolId"`
	ClusterIds       []int  `json:"clusterIds"`
	UserId           int32  `json:"-"`
}

type ExternalLinksCreateUpdateResponse struct {
	Success bool `json:"success"`
}

func NewExternalLinksServiceImpl(logger *zap.SugaredLogger, externalLinksToolsRepository ExternalLinksToolsRepository,
	externalLinksClustersRepository ExternalLinksClustersRepository, externalLinksRepository ExternalLinksRepository, userAuthService user.UserAuthService) *ExternalLinksServiceImpl {
	return &ExternalLinksServiceImpl{
		logger:                          logger,
		externalLinksToolsRepository:    externalLinksToolsRepository,
		externalLinksClustersRepository: externalLinksClustersRepository,
		externalLinksRepository:         externalLinksRepository,
		userAuthService:                 userAuthService,
	}
}

func (impl ExternalLinksServiceImpl) Create(requests []*ExternalLinkoutRequest, userId int32) (*ExternalLinksCreateUpdateResponse, error) {
	impl.logger.Debugw("external linkout create request", "req", requests)
	for _, request := range requests {
		t := &ExternalLinks{
			Name:                          request.Name,
			Active:                        true,
			ExternalLinksMonitoringToolId: request.MonitoringToolId,
			Url:                           request.Url,
			AuditLog:                      sql.AuditLog{CreatedOn: time.Now(), CreatedBy: userId, UpdatedOn: time.Now(), UpdatedBy: userId},
		}
		err := impl.externalLinksRepository.Save(t)
		if err != nil {
			impl.logger.Errorw("error in saving link", "data", t, "err", err)
			err = &util.ApiError{
				InternalMessage: "external link failed to create in db",
				UserMessage:     "external link failed to create in db",
			}
			return nil, err
		}

		for _, clusterId := range request.ClusterIds {
			externalLinksMapping := &ExternalLinksClusters{
				ExternalLinksId: t.Id,
				ClusterId:       clusterId,
				Active:          true,
				AuditLog:        sql.AuditLog{CreatedOn: time.Now(), CreatedBy: userId, UpdatedOn: time.Now(), UpdatedBy: userId},
			}
			err := impl.externalLinksClustersRepository.Save(externalLinksMapping)
			if err != nil {
				impl.logger.Errorw("error in saving cluster id's", "data", t, "err", err)
				err = &util.ApiError{
					InternalMessage: "cluster id failed to create in db",
					UserMessage:     "cluster id failed to create in db",
				}
				return nil, err
			}
		}
	}
	externalLinksCreateUpdateResponse := &ExternalLinksCreateUpdateResponse{
		Success: true,
	}
	return externalLinksCreateUpdateResponse, nil
}

func (impl ExternalLinksServiceImpl) GetAllActiveTools() ([]ExternalLinksMonitoringToolsRequest, error) {
	impl.logger.Debug("fetch all links from db")
	tools, err := impl.externalLinksToolsRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetch all tools", "err", err)
		return nil, err
	}
	var toolRequests []ExternalLinksMonitoringToolsRequest
	for _, tool := range tools {
		providerRes := ExternalLinksMonitoringToolsRequest{
			Id:   tool.Id,
			Name: tool.Name,
			Icon: tool.Icon,
		}
		toolRequests = append(toolRequests, providerRes)
	}
	return toolRequests, err
}

func (impl ExternalLinksServiceImpl) FetchAllActiveLinks(clusterId int) ([]*ExternalLinkoutRequest, error) {
	impl.logger.Debug("fetch all links from db")
	var err error
	var mappedExternalLinksIds []int
	filterByCluster := make(map[int]int)
	externalLinksMap := make(map[int]int)
	allActiveExternalLinks, err := impl.externalLinksClustersRepository.FindAllActive()
	for _, link := range allActiveExternalLinks {
		externalLinksMap[link.ExternalLinksId] = link.ExternalLinksId
	}
	for _, externalLinksId := range externalLinksMap {
		mappedExternalLinksIds = append(mappedExternalLinksIds, externalLinksId)
	}

	var externalLinkResponse []*ExternalLinkoutRequest
	response := make(map[int]*ExternalLinkoutRequest)
	for _, link := range allActiveExternalLinks {

		//requested all links
		if clusterId > 0 {
			if link.ClusterId == clusterId {
				filterByCluster[link.ExternalLinksId] = link.ExternalLinksId
			}
		}
		if _, ok := response[link.ExternalLinksId]; !ok {
			response[link.ExternalLinksId] = &ExternalLinkoutRequest{
				Id:               link.ExternalLinksId,
				Name:             link.ExternalLinks.Name,
				Url:              link.ExternalLinks.Url,
				Active:           link.ExternalLinks.Active,
				MonitoringToolId: link.ExternalLinks.ExternalLinksMonitoringToolId,
			}
		}
		response[link.ExternalLinksId].ClusterIds = append(response[link.ExternalLinksId].ClusterIds, link.ClusterId)
	}

	for k, v := range response {
		if _, ok := filterByCluster[k]; ok {
			externalLinkResponse = append(externalLinkResponse, v)
		} else if clusterId == 0 {
			externalLinkResponse = append(externalLinkResponse, v)
		}
	}

	//now add all the links which are not mapped to any clusters
	additionalExternalLinks, err := impl.externalLinksRepository.FindAllNonMapped(mappedExternalLinksIds)
	if err != nil {
		impl.logger.Errorw("error in fetch all links", "err", err)
		return nil, err
	}
	for _, link := range additionalExternalLinks {
		providerRes := &ExternalLinkoutRequest{
			Id:               link.Id,
			Name:             link.Name,
			Url:              link.Url,
			Active:           link.Active,
			MonitoringToolId: link.ExternalLinksMonitoringToolId,
			ClusterIds:       []int{},
		}
		externalLinkResponse = append(externalLinkResponse, providerRes)
	}

	if externalLinkResponse == nil {
		externalLinkResponse = make([]*ExternalLinkoutRequest, 0)
	}
	return externalLinkResponse, err
}
func (impl ExternalLinksServiceImpl) Update(request *ExternalLinkoutRequest) (*ExternalLinksCreateUpdateResponse, error) {
	impl.logger.Debugw("link update request", "req", request)
	externalLinks, err0 := impl.externalLinksRepository.FindOne(request.Id)
	if err0 != nil {
		impl.logger.Errorw("No matching entry found for update.", "id", request.Id)
		return nil, err0
	}
	externalLinks.Name = request.Name
	externalLinks.Url = request.Url
	externalLinks.Active = true
	externalLinks.ExternalLinksMonitoringToolId = request.MonitoringToolId
	externalLinks.UpdatedBy = int32(request.UserId)
	externalLinks.UpdatedOn = time.Now()
	err := impl.externalLinksRepository.Update(&externalLinks)
	if err != nil {
		impl.logger.Errorw("error in updating link", "data", externalLinks, "err", err)
		return nil, err
	}

	allExternalLinksMapping, err := impl.externalLinksClustersRepository.FindAllByExternalLinkId(request.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching link", "data", externalLinks, "err", err)
		return nil, err
	}
	for _, model := range allExternalLinksMapping {
		model.Active = false
		model.UpdatedBy = int32(request.UserId)
		model.UpdatedOn = time.Now()
		err := impl.externalLinksClustersRepository.Update(model)
		if err != nil {
			impl.logger.Errorw("error in updating clusters to false", "data", model, "err", err)
			return nil, err
		}
	}
	for _, requestedClusterId := range request.ClusterIds {
		externalLinkClusterId := 0
		var externalLinkCluster *ExternalLinksClusters
		for _, model := range allExternalLinksMapping {
			if requestedClusterId == model.ClusterId {
				externalLinkClusterId = model.Id
				externalLinkCluster = model
				break
			}
		}
		if externalLinkClusterId > 0 && externalLinkCluster != nil {
			externalLinkCluster.Active = true
			externalLinkCluster.UpdatedOn = time.Now()
			externalLinkCluster.UpdatedBy = request.UserId
			err = impl.externalLinksClustersRepository.Update(externalLinkCluster)
		} else {
			externalLinkCluster := &ExternalLinksClusters{
				ExternalLinksId: request.Id,
				ClusterId:       requestedClusterId,
				Active:          true,
				AuditLog:        sql.AuditLog{CreatedOn: time.Now(), CreatedBy: request.UserId, UpdatedOn: time.Now(), UpdatedBy: request.UserId},
			}
			err = impl.externalLinksClustersRepository.Save(externalLinkCluster)
		}
		if err != nil {
			impl.logger.Errorw("error in saving cluster id's", "data", externalLinkCluster, "err", err)
			err = &util.ApiError{
				InternalMessage: "cluster id failed to create in db",
				UserMessage:     "cluster id failed to create in db",
			}
			return nil, err
		}
	}
	externalLinksCreateUpdateResponse := &ExternalLinksCreateUpdateResponse{
		Success: true,
	}
	return externalLinksCreateUpdateResponse, nil
}
func (impl ExternalLinksServiceImpl) DeleteLink(id int, userId int32) (*ExternalLinksCreateUpdateResponse, error) {
	impl.logger.Debugw("link delete request", "req", id)
	externalLinksMapping, err := impl.externalLinksClustersRepository.FindAllActiveByExternalLinkId(id)
	if err != nil {
		return nil, err
	}
	for _, externalLink := range externalLinksMapping {
		externalLink.Active = false
		externalLink.UpdatedOn = time.Now()
		externalLink.UpdatedBy = userId
		err := impl.externalLinksClustersRepository.Update(externalLink)
		if err != nil {
			impl.logger.Errorw("error in deleting clusters to false", "data", externalLink, "err", err)
			return nil, err
		}
	}

	externalLinks, err := impl.externalLinksRepository.FindOne(id)
	if err != nil {
		return nil, err
	}
	externalLinks.Active = false
	externalLinks.UpdatedOn = time.Now()
	externalLinks.UpdatedBy = userId
	err = impl.externalLinksRepository.Update(&externalLinks)
	if err != nil {
		impl.logger.Errorw("error in deleting link", "data", externalLinks, "err", err)
		return nil, err
	}

	externalLinksCreateUpdateResponse := &ExternalLinksCreateUpdateResponse{
		Success: true,
	}
	return externalLinksCreateUpdateResponse, nil
}
