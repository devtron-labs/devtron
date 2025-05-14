/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package beans

import (
	"github.com/devtron-labs/devtron/client/events/bean"
	util "github.com/devtron-labs/devtron/util/event"
)

const (
	SLACK_URL   = "https://hooks.slack.com/"
	WEBHOOK_URL = "https://"
)

type WebhookVariable string

const (
	// these fields will be configurable in future
	DevtronContainerImageTag  WebhookVariable = "{{devtronContainerImageTag}}"
	DevtronContainerImageRepo WebhookVariable = "{{devtronContainerImageRepo}}"
	DevtronAppName            WebhookVariable = "{{devtronAppName}}"
	DevtronAppId              WebhookVariable = "{{devtronAppId}}"
	DevtronEnvName            WebhookVariable = "{{devtronEnvName}}"
	DevtronEnvId              WebhookVariable = "{{devtronEnvId}}"
	DevtronCiPipelineId       WebhookVariable = "{{devtronCiPipelineId}}"
	DevtronCdPipelineId       WebhookVariable = "{{devtronCdPipelineId}}"
	DevtronTriggeredByEmail   WebhookVariable = "{{devtronTriggeredByEmail}}"
	DevtronBuildGitCommitHash WebhookVariable = "{{devtronBuildGitCommitHash}}"
	DevtronPipelineType       WebhookVariable = "{{devtronPipelineType}}"
	EventType                 WebhookVariable = "{{eventType}}"
)

const (
	AllNonProdEnvsName = "All non-prod environments"
	AllProdEnvsName    = "All prod environments"
)

type NotificationConfigRequest struct {
	Id int `json:"id"`

	TeamId    []*int `json:"teamId"`
	AppId     []*int `json:"appId"`
	EnvId     []*int `json:"envId"`
	ClusterId []*int `json:"clusterId"`

	PipelineId   *int              `json:"pipelineId"`
	PipelineType util.PipelineType `json:"pipelineType" validate:"required"`
	EventTypeIds []int             `json:"eventTypeIds" validate:"required"`
	Providers    []*bean.Provider  `json:"providers"`
}

func (notificationSettingsRequest *NotificationConfigRequest) GenerateSettingCombinationsV1() []*LocalRequest {

	var tempRequest []*LocalRequest
	if len(notificationSettingsRequest.TeamId) == 0 && len(notificationSettingsRequest.EnvId) == 0 && len(notificationSettingsRequest.AppId) > 0 {
		for _, item := range notificationSettingsRequest.AppId {
			tempRequest = append(tempRequest, &LocalRequest{AppId: item})
		}
	} else if len(notificationSettingsRequest.TeamId) == 0 && len(notificationSettingsRequest.EnvId) > 0 && len(notificationSettingsRequest.AppId) == 0 {
		for _, item := range notificationSettingsRequest.EnvId {
			tempRequest = append(tempRequest, &LocalRequest{EnvId: item})
		}
	} else if len(notificationSettingsRequest.TeamId) > 0 && len(notificationSettingsRequest.EnvId) == 0 && len(notificationSettingsRequest.AppId) == 0 {
		for _, item := range notificationSettingsRequest.TeamId {
			tempRequest = append(tempRequest, &LocalRequest{TeamId: item})
		}
	} else if len(notificationSettingsRequest.TeamId) == 0 && len(notificationSettingsRequest.EnvId) > 0 && len(notificationSettingsRequest.AppId) > 0 {
		for _, itemE := range notificationSettingsRequest.EnvId {
			for _, itemA := range notificationSettingsRequest.AppId {
				tempRequest = append(tempRequest, &LocalRequest{EnvId: itemE, AppId: itemA})
			}
		}
	} else if len(notificationSettingsRequest.TeamId) > 0 && len(notificationSettingsRequest.EnvId) > 0 && len(notificationSettingsRequest.AppId) == 0 {
		for _, itemT := range notificationSettingsRequest.TeamId {
			for _, itemE := range notificationSettingsRequest.EnvId {
				tempRequest = append(tempRequest, &LocalRequest{TeamId: itemT, EnvId: itemE})
			}
		}
	} else if len(notificationSettingsRequest.TeamId) > 0 && len(notificationSettingsRequest.EnvId) == 0 && len(notificationSettingsRequest.AppId) > 0 {
		for _, itemT := range notificationSettingsRequest.TeamId {
			for _, itemA := range notificationSettingsRequest.AppId {
				tempRequest = append(tempRequest, &LocalRequest{TeamId: itemT, AppId: itemA})
			}
		}
	} else if len(notificationSettingsRequest.TeamId) > 0 && len(notificationSettingsRequest.EnvId) > 0 && len(notificationSettingsRequest.AppId) > 0 {
		for _, itemT := range notificationSettingsRequest.TeamId {
			for _, itemE := range notificationSettingsRequest.EnvId {
				for _, itemA := range notificationSettingsRequest.AppId {
					tempRequest = append(tempRequest, &LocalRequest{TeamId: itemT, EnvId: itemE, AppId: itemA})
				}
			}
		}
	} else {
		tempRequest = append(tempRequest, &LocalRequest{PipelineId: notificationSettingsRequest.PipelineId})
	}

	return tempRequest
}

// GetIdsByTypeIndex
// if new criteria fields are added, add a case to this function
func (notificationConfigRequest *NotificationConfigRequest) GetIdsByTypeIndex(index selectorIndex) []*int {
	switch index {
	case teams:
		return notificationConfigRequest.TeamId
	case apps:
		return notificationConfigRequest.AppId
	case envs:
		return notificationConfigRequest.EnvId
	case clusters:
		return notificationConfigRequest.ClusterId
	}
	return nil
}

// if new criteria fields are added, create a index for it here
type selectorIndex int

const teams, apps, envs, clusters selectorIndex = 0, 1, 2, 3

// GenerateSettingCombinations
// if new criteria is added , add a similar condition for that criteria index below
func (notificationConfigRequest *NotificationConfigRequest) GenerateSettingCombinations() []*LocalRequest {
	selectorIndices := make([]selectorIndex, 0)
	if len(notificationConfigRequest.TeamId) > 0 {
		selectorIndices = append(selectorIndices, teams)
	}

	if len(notificationConfigRequest.AppId) > 0 {
		selectorIndices = append(selectorIndices, apps)
	}

	if len(notificationConfigRequest.EnvId) > 0 {
		selectorIndices = append(selectorIndices, envs)
	}

	if len(notificationConfigRequest.ClusterId) > 0 {
		selectorIndices = append(selectorIndices, clusters)
	}

	result := make([]*LocalRequest, 0)
	if len(selectorIndices) == 0 {
		return append(result, &LocalRequest{PipelineId: notificationConfigRequest.PipelineId})
	}

	selectorsMappedWithIdx := make([][]*int, 0)
	for _, selectorIdx := range selectorIndices {
		selectorsMappedWithIdx = append(selectorsMappedWithIdx, notificationConfigRequest.GetIdsByTypeIndex(selectorIdx))
	}
	generateCombinationSettings(selectorsMappedWithIdx, LocalRequest{}, 0, &result, selectorIndices)
	return result
}

// example:
// teams: [1,2]
// apps: [3,4]
// envs: [5,6,7]
//
//	we should generate the following combinations:
//	(1,3,5), (1,3,6), (1,3,7),
//	(1,4,5), (1,4,6), (1,4,7),
//	(2,3,5), (2,3,6), (2,3,7),
//	(2,4,5), (2,4,6), (2,4,7)
//
// these can be generated by just running  3 nested for loops. but whenever we use a new criteria lets say cluster,
// we should add another for loop. this is also not a big challange, but the challenge is when we have random criterias ids, i.e  when we have
// only (teamIds,appIds and clusterIds) or (teamIds,envIds and clusterIds) etc...
// this requires us to select the combination of nested loops based on the available criteria.
// this essentially requires 2^n combinations(n -> criteria selectors) of nested loop logic. to avoid this we are using the recursive approach.
// generateCombinationSettings:
// Function to generate all combinations of arrays corresponding to the selectorIndices
func generateCombinationSettings(arrays [][]*int, current LocalRequest, index int, result *[]*LocalRequest, selectorIndices []selectorIndex) {
	// Base case: when we reach the end of the arrays, we append the combination to the result
	if index == len(arrays) {
		// Create a copy of the current result and append it
		comb := current
		*result = append(*result, &comb)
		return
	}

	// Iterate over the current array and generate combinations
	for _, value := range arrays[index] {
		valCopy := *value               // Take a copy for pointer safety
		switch selectorIndices[index] { // Set the correct index in the struct based on set bits
		case teams:
			current.TeamId = &valCopy
		case apps:
			current.AppId = &valCopy
		case envs:
			current.EnvId = &valCopy
		case clusters:
			current.ClusterId = &valCopy
		}

		generateCombinationSettings(arrays, current, index+1, result, selectorIndices)
	}
}

type NSConfig struct {
	TeamId       []*int            `json:"teamId"`
	AppId        []*int            `json:"appId"`
	EnvId        []*int            `json:"envId"`
	PipelineId   *int              `json:"pipelineId"`
	ClusterId    []*int            `json:"clusterId"`
	PipelineType util.PipelineType `json:"pipelineType" validate:"required"`
	EventTypeIds []int             `json:"eventTypeIds" validate:"required"`
	Providers    []*bean.Provider  `json:"providers" validate:"required"`
}

type NotificationSettingRequest struct {
	Id     int  `json:"id"`
	TeamId int  `json:"teamId"`
	AppId  *int `json:"appId"`
	EnvId  *int `json:"envId"`
	//Pipelines    []int             `json:"pipelineIds"`
	PipelineType util.PipelineType `json:"pipelineType" validate:"required"`
	EventTypeIds []int             `json:"eventTypeIds" validate:"required"`
	Providers    []bean.Provider   `json:"providers" validate:"required"`
}

type Providers struct {
	Providers []bean.Provider `json:"providers"`
}

type NSDeleteRequest struct {
	Id []*int `json:"id"`
}

type NotificationRequest struct {
	UpdateType                util.UpdateType              `json:"updateType,omitempty"`
	Providers                 []*bean.Provider             `json:"providers"`
	NotificationConfigRequest []*NotificationConfigRequest `json:"notificationConfigRequest" validate:"required"`
	// will be deprecated in future
	SesConfigId int `json:"sesConfigId"`
}

type NotificationUpdateRequest struct {
	UpdateType                util.UpdateType              `json:"updateType,omitempty"`
	NotificationConfigRequest []*NotificationConfigRequest `json:"notificationConfigRequest" validate:"required"`
}

type NSViewResponse struct {
	Total                        int                             `json:"total"`
	NotificationSettingsResponse []*NotificationSettingsResponse `json:"settings"`
}

type NotificationSettingsResponse struct {
	Id               int                `json:"id"`
	ConfigName       string             `json:"configName"`
	TeamResponse     []*TeamResponse    `json:"team"`
	AppResponse      []*AppResponse     `json:"app"`
	EnvResponse      []*EnvResponse     `json:"environment"`
	ClusterResponse  []*ClusterResponse `json:"cluster"`
	PipelineResponse *PipelineResponse  `json:"pipeline"`
	PipelineType     string             `json:"pipelineType"`
	ProvidersConfig  []*ProvidersConfig `json:"providerConfigs"`
	EventTypes       []int              `json:"eventTypes"`
}

type SearchFilterResponse struct {
	TeamResponse     []*TeamResponse    `json:"team"`
	AppResponse      []*AppResponse     `json:"app"`
	EnvResponse      []*EnvResponse     `json:"environment"`
	ClusterResponse  []*ClusterResponse `json:"cluster"`
	PipelineResponse *PipelineResponse  `json:"pipeline"`
	PipelineType     string             `json:"pipelineType"`
}

type TeamResponse struct {
	Id   *int   `json:"id"`
	Name string `json:"name"`
}

type AppResponse struct {
	Id   *int   `json:"id"`
	Name string `json:"name"`
}

type EnvResponse struct {
	Id   *int   `json:"id"`
	Name string `json:"name"`
}

type ClusterResponse struct {
	Id   *int   `json:"id"`
	Name string `json:"name"`
}

type PipelineResponse struct {
	Id              *int     `json:"id"`
	Name            string   `json:"name"`
	EnvironmentName string   `json:"environmentName,omitempty"`
	AppName         string   `json:"appName,omitempty"`
	Branches        []string `json:"branches,omitempty"`
	ClusterName     string   `json:"clusterName"`
}

type ProvidersConfig struct {
	Id         int    `json:"id"`
	Dest       string `json:"dest"`
	ConfigName string `json:"name"`
	Recipient  string `json:"recipient"`
}

type LocalRequest struct {
	Id         int  `json:"id"`
	TeamId     *int `json:"teamId"`
	AppId      *int `json:"appId"`
	EnvId      *int `json:"envId"`
	PipelineId *int `json:"pipelineId"`
	ClusterId  *int `json:"clusterId"`
}

type NotificationChannelAutoResponse struct {
	ConfigName string `json:"configName"`
	Id         int    `json:"id"`
	TeamId     int    `json:"-"`
}

type NotificationRecipientListingResponse struct {
	Dest      util.Channel `json:"dest"`
	ConfigId  int          `json:"configId"`
	Recipient string       `json:"recipient"`
}

//SES

type SESChannelConfig struct {
	Channel       util.Channel    `json:"channel" validate:"required"`
	SESConfigDtos []*SESConfigDto `json:"configs"`
}

type SESConfigDto struct {
	OwnerId      int32  `json:"userId" validate:"number"`
	TeamId       int    `json:"teamId" validate:"number"`
	Region       string `json:"region" validate:"required"`
	AccessKey    string `json:"accessKey" validate:"required"`
	SecretKey    string `json:"secretKey" validate:"required"`
	FromEmail    string `json:"fromEmail" validate:"email,required"`
	ToEmail      string `json:"toEmail"`
	SessionToken string `json:"sessionToken"`
	ConfigName   string `json:"configName" validate:"required"`
	Description  string `json:"description"`
	Id           int    `json:"id" validate:"number"`
	Default      bool   `json:"default,notnull"`
}

//Slack

type SlackChannelConfig struct {
	Channel         util.Channel     `json:"channel" validate:"required"`
	SlackConfigDtos []SlackConfigDto `json:"configs"`
}

type SlackConfigDto struct {
	OwnerId     int32  `json:"userId" validate:"number"`
	TeamId      int    `json:"teamId" validate:"required"`
	WebhookUrl  string `json:"webhookUrl" validate:"required"`
	ConfigName  string `json:"configName" validate:"required"`
	Description string `json:"description"`
	Id          int    `json:"id" validate:"number"`
}

//SMTP

type SMTPChannelConfig struct {
	Channel        util.Channel     `json:"channel" validate:"required"`
	SMTPConfigDtos []*SMTPConfigDto `json:"configs"`
}

type SMTPConfigDto struct {
	Id           int    `json:"id"`
	Port         string `json:"port"`
	Host         string `json:"host"`
	AuthType     string `json:"authType"`
	AuthUser     string `json:"authUser"`
	AuthPassword string `json:"authPassword"`
	FromEmail    string `json:"fromEmail"`
	ConfigName   string `json:"configName"`
	Description  string `json:"description"`
	OwnerId      int32  `json:"ownerId"`
	Default      bool   `json:"default"`
	Deleted      bool   `json:"deleted"`
}

//webhook

type WebhookChannelConfig struct {
	Channel           util.Channel        `json:"channel" validate:"required"`
	WebhookConfigDtos *[]WebhookConfigDto `json:"configs"`
}

type WebhookConfigDto struct {
	OwnerId     int32                  `json:"userId" validate:"number"`
	WebhookUrl  string                 `json:"webhookUrl" validate:"required"`
	ConfigName  string                 `json:"configName" validate:"required"`
	Header      map[string]interface{} `json:"header"`
	Payload     string                 `json:"payload"`
	Description string                 `json:"description"`
	Id          int                    `json:"id" validate:"number"`
}

type Config struct {
	AppId        int               `json:"appId"`
	EnvId        int               `json:"envId"`
	Pipelines    []int             `json:"pipelineIds"`
	PipelineType util.PipelineType `json:"pipelineType" validate:"required"`
	EventTypeIds []int             `json:"eventTypeIds" validate:"required"`
	Providers    []bean.Provider   `json:"providers" validate:"required"`
}
