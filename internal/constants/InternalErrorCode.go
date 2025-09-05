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

package constants

import "fmt"

// Error Codes Sequence

//	Cluster -		1000-1999
//	Environment -	2000-2999
//	Global Config -	3000-3999
//	Application -	4000-4999
//	Pipeline -		5000-5999
//	User -			6000-6999
//	Other -			7000-7999

type ErrorCode struct {
	Code           string
	userErrMessage string
}

func (code ErrorCode) UserMessage(params ...interface{}) string {
	return fmt.Sprintf(code.userErrMessage, params)
}

const (
	// Cluster Errors Start ------------------
	// Sequence 1000-1999

	ClusterCreateDBFailed      string = "1001"
	ClusterCreateACDFailed     string = "1002"
	ClusterDBRollbackFailed    string = "1003"
	ClusterUpdateDBFailed      string = "1004"
	ClusterUpdateACDFailed     string = "1005"
	ClusterCreateBadRequestACD string = "1006"
	ClusterUpdateBadRequestACD string = "1007"

	// Cluster Errors End --------------------
)

const (
	// Environment Errors Start --------------
	// Sequence 2000-2999

	EnvironmentCreateDBFailed          string = "2001"
	EnvironmentUpdateDBFailed          string = "2002"
	EnvironmentUpdateEnvOverrideFailed string = "2003"

	// Environment Errors End ----------------
)

const (
	// Global Config Errors Start ------------
	// Sequence 3000-3999

	// Docker Registry Errors

	DockerRegCreateFailedInDb   string = "3001"
	DockerRegCreateFailedInGocd string = "3002"
	DockerRegUpdateFailedInDb   string = "3003"
	DockerRegUpdateFailedInGocd string = "3004"

	// Git Provider Errors

	GitProviderCreateFailedAlreadyExists string = "3005"
	GitProviderCreateFailedInDb          string = "3006"
	GitProviderUpdateProviderNotExists   string = "3007"
	GitProviderUpdateFailedInDb          string = "3008"
	DockerRegDeleteFailedInDb            string = "3009"
	DockerRegDeleteFailedInGocd          string = "3010"
	GitProviderUpdateFailedInSync        string = "3011"
	GitProviderUpdateRequestIsInvalid    string = "3012"

	// ---------------------------------------

	// For Global Config conflicts use 3900 series

	GitOpsConfigValidationConflict string = "3900"

	// Global Config Errors End --------------
)

const (
// Application Errors Start --------------
// Sequence 4000-4999

// Application Errors End ----------------
)

const (
	// Pipeline Errors Start -----------------
	// Sequence 5000-5999

	ChartCreatedAlreadyExists string = "5001"
	ChartNameAlreadyReserved  string = "5002"

	// Pipeline Config GitOps Config Errors
	// Sequence 5100-5199
	GitOpsNotConfigured                 string = "5100"
	GitOpsOrganisationMismatch          string = "5101"
	GitOpsURLAlreadyInUse               string = "5102"
	InvalidDeploymentAppTypeForPipeline string = "5103"

	// Pipeline Errors End -------------------
)

const (
	// User Errors Start ---------------------
	// Sequence 6000-6999

	UserCreateDBFailed        string = "6001"
	UserCreatePolicyFailed    string = "6002"
	UserUpdateDBFailed        string = "6003"
	UserUpdatePolicyFailed    string = "6004"
	UserNoTokenProvided       string = "6005"
	UserNotFoundForToken      string = "6006"
	UserCreateFetchRoleFailed string = "6007"
	UserUpdateFetchRoleFailed string = "6008"

	// User Errors End -----------------------
)

const (
	// Other Errors Start --------------------
	// Sequence 7000-7999

	AppDetailResourceTreeNotFound string = "7000"
	HelmReleaseNotFound           string = "7001"

	CasbinPolicyNotCreated string = "8000"

	GitHostCreateFailedAlreadyExists string = "9001"
	GitHostCreateFailedInDb          string = "9002"

	// Other Errors End ----------------------
)

const (
	// Feasibility Errors Start --------------
	// Sequence 10000-10999

	VulnerabilityFound                   string = "10001"
	ApprovalNodeFail                     string = "10002"
	FilteringConditionFail               string = "10003"
	DeploymentWindowFail                 string = "10004"
	PreCDDoesNotExists                   string = "10005"
	PostCDDoesNotExists                  string = "10006"
	ArtifactNotAvailable                 string = "10007"
	DeploymentWindowByPassed             string = "10008"
	MandatoryPluginNotAdded              string = "10009"
	MandatoryTagNotAdded                 string = "10010"
	SecurityScanFail                     string = "10011"
	ApprovalConfigDependentActionFailure string = "10012"

	// Feasibility Errors End ----------------
)

const (
	// Generic API Errors Start --------------
	// Sequence 11000-11999

	InvalidPathParameter  string = "11001"
	InvalidRequestBody    string = "11002"
	InvalidQueryParameter string = "11003"
	ValidationFailed      string = "11004"
	MissingRequiredField  string = "11005"
	ResourceNotFound      string = "11006"
	DuplicateResource     string = "11007"

	// Generic API Errors End ----------------
)

const (
	// Not Processed Internal Errors Start ---
	// Sequence 11000-11999

	NotProcessed string = "11001"
	NotExecuted  string = "11002"

	// Not Processed Internal Errors End -----
)

const (
	HttpStatusUnprocessableEntity = "422"
)

const (
	HttpClientSideTimeout = 499
)

var AppAlreadyExists = &ErrorCode{"4001", "application %s already exists"}
var AppDoesNotExist = &ErrorCode{"4004", "application %s does not exist"}

const (
	ErrorDeletingPipelineForDeletedArgoAppMsg = "error in deleting devtron pipeline for deleted argocd app"
	ArgoAppDeletedErrMsg                      = "argocd app deleted"
	UnableToFetchResourceTreeErrMsg           = "unable to fetch resource tree"
	UnableToFetchResourceTreeForAcdErrMsg     = "app detail fetched, failed to get resource tree from acd"
	CannotGetAppWithRefreshErrMsg             = "cannot get application with refresh"
	NoDataFoundErrMsg                         = "no data found"
)
