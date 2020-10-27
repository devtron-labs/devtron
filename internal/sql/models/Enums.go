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

package models

type ChartStatus int

const (
	CHARTSTATUS_NEW                    ChartStatus = 1
	CHARTSTATUS_DEPLOYMENT_IN_PROGRESS ChartStatus = 2
	CHARTSTATUS_SUCCESS                ChartStatus = 3
	CHARTSTATUS_ERROR                  ChartStatus = 4
	CHARTSTATUS_ROLLBACK               ChartStatus = 5
	CHARTSTATUS_UNKNOWN                ChartStatus = 6
)

func (s ChartStatus) String() string {
	return [...]string{"CHARTSTATUS_NEW", "CHARTSTATUS_DEPLOYMENT_IN_PROGRESS", "CHARTSTATUS_SUCCESS", "CHARTSTATUS_ERROR", "CHARTSTATUS_ROLLBACK", "CHARTSTATUS_UNKNOWN"}[s]
}

type DeploymentType int

const (
	DEPLOYMENTTYPE_UNKNOWN DeploymentType = iota
	DEPLOYMENTTYPE_DEPLOY
	DEPLOYMENTTYPE_ROLLBACK
	DEPLOYMENTTYPE_STOP
	DEPLOYMENTTYPE_START
)

func (d DeploymentType) String() string {
	return [...]string{"DEPLOYMENTTYPE_UNKNOWN", "DEPLOYMENTTYPE_DEPLOY", "DEPLOYMENTTYPE_ROLLBACK", "DEPLOYMENTTYPE_STOP", "DEPLOYMENTTYPE_START"}[d]
}
