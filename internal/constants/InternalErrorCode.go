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

package constants

import "fmt"

/**
 	Cluster - 			1000-1999
	Environment - 		2000-2999
	Global Config - 	3000-3999
	Pipeline Config -	4000-4999
	Pipeline - 			5000-5999
	User - 				6000-6999
	Other -				7000-7999
*/

type ErrorCode struct {
	Code           string
	userErrMessage string
}

func (code ErrorCode) UserMessage(params ...interface{}) string {
	return fmt.Sprintf(code.userErrMessage, params)
}

const (
	//Cluster Errors
	ClusterCreateDBFailed      string = "1001"
	ClusterCreateACDFailed     string = "1002"
	ClusterDBRollbackFailed    string = "1003"
	ClusterUpdateDBFailed      string = "1004"
	ClusterUpdateACDFailed     string = "1005"
	ClusterCreateBadRequestACD string = "1006"
	ClusterUpdateBadRequestACD string = "1007"
	//Environment Errors
	EnvironmentCreateDBFailed          string = "2001"
	EnvironmentUpdateDBFailed          string = "2002"
	EnvironmentUpdateEnvOverrideFailed string = "2003"
	//Global Config Errors
	DockerRegCreateFailedInDb            string = "3001"
	DockerRegCreateFailedInGocd          string = "3002"
	DockerRegUpdateFailedInDb            string = "3003"
	DockerRegUpdateFailedInGocd          string = "3004"
	GitProviderCreateFailedAlreadyExists string = "3005"
	GitProviderCreateFailedInDb          string = "3006"
	GitProviderUpdateProviderNotExists   string = "3007"
	GitProviderUpdateFailedInDb          string = "3008"
	DockerRegDeleteFailedInDb            string = "3009"
	DockerRegDeleteFailedInGocd          string = "3010"
	GitProviderUpdateFailedInSync        string = "3011"
	ChartCreatedAlreadyExists            string = "5001"
	UserCreateDBFailed                   string = "6001"
	UserCreatePolicyFailed               string = "6002"
	UserUpdateDBFailed                   string = "6003"
	UserUpdatePolicyFailed               string = "6004"
	UserNoTokenProvided                  string = "6005"
	UserNotFoundForToken                 string = "6006"
	UserCreateFetchRoleFailed            string = "6007"
	UserUpdateFetchRoleFailed            string = "6008"

	AppDetailResourceTreeNotFound string = "7000"

	CasbinPolicyNotCreated string = "8000"

	GitHostCreateFailedAlreadyExists string = "9001"
	GitHostCreateFailedInDb          string = "9002"
)

var AppAlreadyExists = &ErrorCode{"4001", "application %s already exists"}
