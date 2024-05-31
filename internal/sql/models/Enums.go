/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
	DEPLOYMENTTYPE_PRE
	DEPLOYMENTTYPE_POST
)

func (d DeploymentType) String() string {
	return [...]string{"DEPLOYMENTTYPE_UNKNOWN", "DEPLOYMENTTYPE_DEPLOY", "DEPLOYMENTTYPE_ROLLBACK", "DEPLOYMENTTYPE_STOP", "DEPLOYMENTTYPE_START"}[d]
}

type ChartsViewEditorType string

const (
	EDITOR_TYPE_BASIC    ChartsViewEditorType = "BASIC"
	EDITOR_TYPE_ADVANCED ChartsViewEditorType = "ADVANCED"
	//default value
	EDITOR_TYPE_UNDEFINED ChartsViewEditorType = "UNDEFINED"
)
