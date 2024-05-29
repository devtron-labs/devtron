/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package util

var (
	GitCommit            = ""
	BuildTime            = ""
	ServerMode           = "FULL"
	SERVER_MODE_FULL     = "FULL"
	SERVER_MODE_HYPERION = "EA_ONLY"
)

type ServerVersion struct {
	GitCommit  string `json:"gitCommit"`
	BuildTime  string `json:"buildTime"`
	ServerMode string `json:"serverMode"`
}

func GetDevtronVersion() *ServerVersion {
	return &ServerVersion{BuildTime: BuildTime, GitCommit: GitCommit, ServerMode: ServerMode}
}

func IsBaseStack() bool {
	return GetDevtronVersion().ServerMode == SERVER_MODE_HYPERION
}

func IsFullStack() bool {
	return GetDevtronVersion().ServerMode == SERVER_MODE_FULL
}

func IsHelmApp(appOfferingMode string) bool {
	return appOfferingMode == SERVER_MODE_HYPERION
}
