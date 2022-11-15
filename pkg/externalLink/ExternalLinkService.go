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

package externalLink

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strconv"
	"time"
)

const (
	ADMIN_ROLE         string = "admin"
	SUPER_ADMIN_ROLE   string = "superAdmin"
	CLUSTER_LEVEL_LINK string = "clusterLevel"
	APP_LEVEL_LINK     string = "appLevel"
)

type AppIdentifier int

const (
	CLUSTER               AppIdentifier = 0
	DEVTRON_APP           AppIdentifier = 1
	DEVTRON_INSTALLED_APP AppIdentifier = 2
	EXTERNAL_HELM_APP     AppIdentifier = 3
)

var TypeMappings = map[string]AppIdentifier{
	"cluster":               CLUSTER,
	"devtron-app":           DEVTRON_APP,
	"devtron-installed-app": DEVTRON_INSTALLED_APP,
	"external-helm-app":     EXTERNAL_HELM_APP,
}

type ExternalLinkService interface {
	Create(requests []*ExternalLinkDto, userId int32, userRole string) (*ExternalLinkApiResponse, error)
	GetAllActiveTools() ([]ExternalLinkMonitoringToolDto, error)
	FetchAllActiveLinksByLinkIdentifier(linkIdentifier *LinkIdentifier, clusterId int, userRole string, userId int) ([]*ExternalLinkDto, error)
	Update(request *ExternalLinkDto, userRole string) (*ExternalLinkApiResponse, error)
	DeleteLink(id int, userId int32, userRole string) (*ExternalLinkApiResponse, error)
}
type ExternalLinkServiceImpl struct {
	logger                                  *zap.SugaredLogger
	externalLinkMonitoringToolRepository    ExternalLinkMonitoringToolRepository
	externalLinkIdentifierMappingRepository ExternalLinkIdentifierMappingRepository
	externalLinkRepository                  ExternalLinkRepository
}
type ExternalLinkMonitoringToolDto struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	Category int    `json:"category"`
}
type LinkIdentifier struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
	EnvId      int    `json:"envId"`
	AppId      int    `json:"appId"`
	ClusterId  int    `json:"clusterId"`
}
type ExternalLinkDto struct {
	Id               int              `json:"id"`
	Name             string           `json:"name"`
	Url              string           `json:"url"`
	Active           bool             `json:"active"`
	MonitoringToolId int              `json:"monitoringToolId"`
	Type             string           `json:"type"`
	Identifiers      []LinkIdentifier `json:"identifiers"`
	IsEditable       bool             `json:"isEditable"`
	Description      string           `json:"description"`
	UpdatedOn        time.Time        `json:"updatedOn"`
	UserId           int32            `json:"-"`
}

type ExternalLinkApiResponse struct {
	Success bool `json:"success"`
}

func getType(identifier AppIdentifier) string {
	for key, val := range TypeMappings {
		if val == identifier {
			return key
		}
	}
	return ""
}
func NewExternalLinkServiceImpl(logger *zap.SugaredLogger, externalLinksToolsRepository ExternalLinkMonitoringToolRepository,
	externalLinkIdentifierMappingRepository ExternalLinkIdentifierMappingRepository, externalLinksRepository ExternalLinkRepository) *ExternalLinkServiceImpl {
	return &ExternalLinkServiceImpl{
		logger:                                  logger,
		externalLinkMonitoringToolRepository:    externalLinksToolsRepository,
		externalLinkIdentifierMappingRepository: externalLinkIdentifierMappingRepository,
		externalLinkRepository:                  externalLinksRepository,
	}
}

func (impl ExternalLinkServiceImpl) Create(requests []*ExternalLinkDto, userId int32, userRole string) (*ExternalLinkApiResponse, error) {
	impl.logger.Debugw("external links create request", "req", requests)
	dbConnection := impl.externalLinkRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in establishing connection", "err", err)
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	externalLinksCreateUpdateResponse := &ExternalLinkApiResponse{
		Success: false,
	}
	for _, request := range requests {
		//if user is admin make isEditable true ,else if user is sup_adm get it from request
		if userRole == ADMIN_ROLE {
			request.IsEditable = true
		}
		linkType := request.Type
		if linkType == CLUSTER_LEVEL_LINK {
			request.Name = request.Name + "|" + "C"
		} else {
			request.Name = request.Name + "|" + "A"
		}
		//data storing in external links table in db
		externalLink := &ExternalLink{
			Name:                         request.Name,
			Active:                       true,
			ExternalLinkMonitoringToolId: request.MonitoringToolId,
			Url:                          request.Url,
			IsEditable:                   request.IsEditable,
			Description:                  request.Description,
			AuditLog:                     sql.AuditLog{CreatedOn: time.Now(), CreatedBy: userId, UpdatedOn: time.Now(), UpdatedBy: userId},
		}
		err := impl.externalLinkRepository.Save(externalLink, tx)
		if err != nil {
			impl.logger.Errorw("error in saving link", "data", externalLink, "err", err)
			err = &util.ApiError{
				InternalMessage: "external link failed to create in db",
				UserMessage:     "external link failed to create in db",
			}
			return externalLinksCreateUpdateResponse, err
		}
		//for all identifiers, check if it is clusterLevel/appLevel
		//if appLevel, get type and identifier else get clusterId
		//save it in external_link_type_mapping table
		for _, linkIdentifier := range request.Identifiers {
			if linkType == CLUSTER_LEVEL_LINK {
				linkIdentifier.Type = getType(CLUSTER)
				linkIdentifier.Identifier = ""
			} else if linkType == APP_LEVEL_LINK {
				if linkIdentifier.Type == getType(DEVTRON_APP) || linkIdentifier.Type == getType(DEVTRON_INSTALLED_APP) {
					appId, err := strconv.Atoi(linkIdentifier.Identifier)
					if err != nil {
						impl.logger.Errorw("error while parsing appId", "appId", appId, "err", err)
						return externalLinksCreateUpdateResponse, err
					}
					linkIdentifier.AppId = appId
				}
				linkIdentifier.ClusterId = 0
			} else {
				return nil, fmt.Errorf("link is neither app level or cluster level")
			}
			externalLinkIdentifierMapping := &ExternalLinkIdentifierMapping{
				ExternalLinkId: externalLink.Id,
				Type:           TypeMappings[linkIdentifier.Type],
				Identifier:     linkIdentifier.Identifier,
				AppId:          linkIdentifier.AppId,
				ClusterId:      linkIdentifier.ClusterId,
				Active:         true,
				AuditLog:       sql.AuditLog{CreatedOn: time.Now(), CreatedBy: userId, UpdatedOn: time.Now(), UpdatedBy: userId},
			}
			err := impl.externalLinkIdentifierMappingRepository.Save(externalLinkIdentifierMapping, tx)
			if err != nil {
				impl.logger.Errorw("error in saving external-link-identifier-mappings", "data", externalLinkIdentifierMapping, "err", err)
				err = &util.ApiError{
					InternalMessage: "external-link-identifier-mapping failed to create in db",
					UserMessage:     "external-link-identifier-mapping failed to create in db",
				}
				return externalLinksCreateUpdateResponse, err
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	externalLinksCreateUpdateResponse.Success = true
	return externalLinksCreateUpdateResponse, nil
}

func (impl ExternalLinkServiceImpl) GetAllActiveTools() ([]ExternalLinkMonitoringToolDto, error) {
	tools, err := impl.externalLinkMonitoringToolRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetch all tools", "err", err)
		return nil, err
	}
	var response []ExternalLinkMonitoringToolDto
	for _, tool := range tools {
		morningTool := ExternalLinkMonitoringToolDto{
			Id:       tool.Id,
			Name:     tool.Name,
			Icon:     tool.Icon,
			Category: tool.Category,
		}
		response = append(response, morningTool)
	}
	return response, err
}

func (impl ExternalLinkServiceImpl) processResult(records []ExternalLinkExternalMappingJoinResponse) ([]*ExternalLinkDto, error) {
	var externalLinkResponse = make([]*ExternalLinkDto, 0)
	responseMap := make(map[int]*ExternalLinkDto)
	for _, record := range records {
		_, ok := responseMap[record.Id]
		if !ok {
			externalLinkDto := &ExternalLinkDto{
				Id:               record.Id,
				Name:             record.Name,
				Url:              record.Url,
				Active:           true,
				MonitoringToolId: record.ExternalLinkMonitoringToolId,
				IsEditable:       record.IsEditable,
				Description:      record.Description,
				Identifiers:      []LinkIdentifier{},
				UpdatedOn:        record.UpdatedOn,
			}
			responseMap[externalLinkDto.Id] = externalLinkDto
		}

		if record.Type == 0 && record.Identifier == "" && record.AppId == 0 && record.ClusterId == 0 {
			responseMap[record.Id].Type = CLUSTER_LEVEL_LINK
		} else if record.Active {
			identifier := LinkIdentifier{
				Type:       getType(record.Type),
				Identifier: record.Identifier,
				AppId:      record.AppId,
				ClusterId:  record.ClusterId,
			}
			responseMap[record.Id].Type = identifier.Type
			responseMap[record.Id].Identifiers = append(responseMap[record.Id].Identifiers, identifier)
		}

	}
	for _, res := range responseMap {
		externalLinkResponse = append(externalLinkResponse, res)
	}
	return externalLinkResponse, nil
}
func (impl ExternalLinkServiceImpl) FetchAllActiveLinksByLinkIdentifier(linkIdentifier *LinkIdentifier, clusterId int, userRole string, userId int) ([]*ExternalLinkDto, error) {
	//linkIdentifier and clusterId nil and 0 respectively to fetch links at global level(get all active links)
	//linkIdentifier and clusterId passed to get all active links for a particular app(linkIdentifier.ClusterId will be 0)
	var err error
	if linkIdentifier == nil {
		if userRole != SUPER_ADMIN_ROLE {
			impl.logger.Debugw("user is not super_admin", "userId", userId, "userRole", userRole)
			return nil, fmt.Errorf("user role is not super_admin")
		}

		records, err := impl.externalLinkIdentifierMappingRepository.FindAllActiveByJoin()
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error while fetching external links from external_links_identifier mappings table", "err", err)
			return nil, err
		}
		return impl.processResult(records)
	}

	if linkIdentifier.Type == getType(DEVTRON_APP) || linkIdentifier.Type == getType(DEVTRON_INSTALLED_APP) {
		appId, err := strconv.Atoi(linkIdentifier.Identifier)
		if err != nil {
			impl.logger.Errorw("error while parsing appId", "appId", appId, "err", err)
			return nil, err
		}
		linkIdentifier.AppId = appId
	}
	linkIdentifier.ClusterId = 0
	records, err := impl.externalLinkIdentifierMappingRepository.FindAllActiveByLinkIdentifier(linkIdentifier, clusterId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching external links from external_links_identifier mappings table", "err", err)
		return nil, err
	}
	return impl.processResult(records)
}

func (impl ExternalLinkServiceImpl) Update(request *ExternalLinkDto, userRole string) (*ExternalLinkApiResponse, error) {
	impl.logger.Debugw("link update request", "req", request)
	dbConnection := impl.externalLinkRepository.GetConnection()
	tx, err := dbConnection.Begin()
	externalLinksCreateUpdateResponse := &ExternalLinkApiResponse{
		Success: false,
	}
	if err != nil {
		impl.logger.Errorw("error in establishing connection", "err", err)
		return externalLinksCreateUpdateResponse, err
	}

	// Rollback tx on error.
	defer tx.Rollback()
	externalLink, err0 := impl.externalLinkRepository.FindOne(request.Id)
	if err0 != nil {
		impl.logger.Errorw("No matching entry found for update.", "id", request.Id)
		msg := "no row found for external link	"
		err = &util.ApiError{InternalMessage: msg, UserMessage: msg}
		return externalLinksCreateUpdateResponse, err0
	}
	if userRole == ADMIN_ROLE && !externalLink.IsEditable {
		impl.logger.Infow("app admin not allowed to update or delete the external link", "external-link-id", externalLink.Id, "user-id", request.UserId)
		return externalLinksCreateUpdateResponse, fmt.Errorf("user not allowed to perform update or delete")
	}
	externalLink.Name = request.Name
	externalLink.Url = request.Url
	externalLink.ExternalLinkMonitoringToolId = request.MonitoringToolId
	externalLink.UpdatedBy = int32(request.UserId)
	externalLink.UpdatedOn = time.Now()
	externalLink.IsEditable = request.IsEditable
	externalLink.Description = request.Description
	err = impl.externalLinkRepository.Update(&externalLink, tx)
	if err != nil {
		impl.logger.Errorw("error in updating link", "data", externalLink, "err", err)
		return externalLinksCreateUpdateResponse, err
	}

	allExternalLinksMapping, err := impl.externalLinkIdentifierMappingRepository.FindAllActiveByExternalLinkId(request.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching link", "data", externalLink, "err", err)
		return externalLinksCreateUpdateResponse, err
	}

	//make all the existing mappings of this external link inactive
	for _, model := range allExternalLinksMapping {
		if model.Active == true {
			model.Active = false
			model.UpdatedBy = int32(request.UserId)
			model.UpdatedOn = time.Now()
			err := impl.externalLinkIdentifierMappingRepository.Update(model, tx)
			if err != nil {
				impl.logger.Errorw("error in updating external_link_identifier mappings to false", "data", model, "err", err)
				return externalLinksCreateUpdateResponse, err
			}
		}
	}

	//store all the external_link_identifier_mappings fetched from db in a set
	requestedLinkIdentifiersMap := make(map[LinkIdentifier]*ExternalLinkIdentifierMapping)
	for _, model := range allExternalLinksMapping {
		linkIdentifier := LinkIdentifier{
			Type:       getType(model.Type),
			AppId:      model.AppId,
			EnvId:      model.EnvId,
			Identifier: model.Identifier,
			ClusterId:  model.ClusterId,
		}
		requestedLinkIdentifiersMap[linkIdentifier] = model
	}
	//update if request identifier present in set else save a new record
	for _, identifier := range request.Identifiers {
		if request.Type == APP_LEVEL_LINK {
			if identifier.Type == getType(DEVTRON_APP) || identifier.Type == getType(DEVTRON_INSTALLED_APP) {
				appId, err := strconv.Atoi(identifier.Identifier)
				if err != nil {
					impl.logger.Errorw("error while parsing appId", "appId", appId, "err", err)
					return externalLinksCreateUpdateResponse, err
				}
				identifier.AppId = appId
			}
			identifier.ClusterId = 0
		} else {
			identifier.Identifier = ""
			identifier.Type = getType(CLUSTER)
		}
		if model, ok := requestedLinkIdentifiersMap[identifier]; ok {
			model.Active = true
			model.UpdatedOn = time.Now()
			model.UpdatedBy = request.UserId
			err = impl.externalLinkIdentifierMappingRepository.Update(model, tx)
		} else {
			externalLinkIdentifier := &ExternalLinkIdentifierMapping{
				ExternalLinkId: request.Id,
				Type:           TypeMappings[identifier.Type],
				AppId:          identifier.AppId,
				EnvId:          identifier.EnvId,
				Identifier:     identifier.Identifier,
				ClusterId:      identifier.ClusterId,
				Active:         true,
				AuditLog:       sql.AuditLog{CreatedOn: time.Now(), CreatedBy: request.UserId, UpdatedOn: time.Now(), UpdatedBy: request.UserId},
			}
			err = impl.externalLinkIdentifierMappingRepository.Save(externalLinkIdentifier, tx)
		}
		if err != nil {
			impl.logger.Errorw("error in saving external_link_identifier mapping", "identifier", identifier, "err", err)
			err = &util.ApiError{
				InternalMessage: "external_link_identifier mapping failed to create in db",
				UserMessage:     "external_link_identifier mapping failed to create in db",
			}
			return externalLinksCreateUpdateResponse, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return externalLinksCreateUpdateResponse, err
	}
	externalLinksCreateUpdateResponse.Success = true
	return externalLinksCreateUpdateResponse, nil
}

func (impl ExternalLinkServiceImpl) DeleteLink(id int, userId int32, userRole string) (*ExternalLinkApiResponse, error) {
	impl.logger.Debugw("external link delete request", "external_link_id", id)
	dbConnection := impl.externalLinkRepository.GetConnection()
	tx, err := dbConnection.Begin()
	externalLinksCreateUpdateResponse := &ExternalLinkApiResponse{
		Success: false,
	}
	if err != nil {
		impl.logger.Errorw("error in establishing connection", "err", err)
		return externalLinksCreateUpdateResponse, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	// mark the link inactive if user has edit access
	externalLink, err := impl.externalLinkRepository.FindOne(id)
	if err != nil {
		return externalLinksCreateUpdateResponse, err
	}
	if userRole == ADMIN_ROLE && !externalLink.IsEditable {
		impl.logger.Infow("app admin not allowed to update or delete the external link", "external-link-id", externalLink.Id, "user-id", userId)
		return externalLinksCreateUpdateResponse, fmt.Errorf("user not allowed to perform update or delete")
	}
	externalLink.Active = false
	externalLink.UpdatedOn = time.Now()
	externalLink.UpdatedBy = userId
	err = impl.externalLinkRepository.Update(&externalLink, tx)
	if err != nil {
		impl.logger.Errorw("error in update external link", "data", externalLink, "err", err)
		return externalLinksCreateUpdateResponse, err
	}

	externalLinkIdentifierMappings, err := impl.externalLinkIdentifierMappingRepository.FindAllActiveByExternalLinkId(id)
	if err != nil {
		return externalLinksCreateUpdateResponse, err
	}
	//mark all the mappings inactive
	for _, externalLinkMapping := range externalLinkIdentifierMappings {
		externalLinkMapping.Active = false
		externalLinkMapping.UpdatedOn = time.Now()
		externalLinkMapping.UpdatedBy = userId
		err := impl.externalLinkIdentifierMappingRepository.Update(externalLinkMapping, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting external_link_identifier mappings to false", "data", externalLink, "err", err)
			return externalLinksCreateUpdateResponse, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return externalLinksCreateUpdateResponse, err
	}
	externalLinksCreateUpdateResponse.Success = true
	return externalLinksCreateUpdateResponse, nil
}
