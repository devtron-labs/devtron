/*
 * Copyright (c) 2024. Devtron Inc.
 */

package resourceQualifiers

func BuildScope(appId, envId, clusterId, projectId int, isProdEnv bool) Scope {
	return Scope{
		AppId:     appId,
		EnvId:     envId,
		ClusterId: clusterId,
		ProjectId: projectId,
		IsProdEnv: isProdEnv}

}
