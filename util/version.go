/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package util

var (
	GitCommit            = ""
	BuildTime            = ""
	ServerMode           = ""
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
