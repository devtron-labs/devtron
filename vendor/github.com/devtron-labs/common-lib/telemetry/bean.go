/*
 * Copyright (c) 2024. Devtron Inc.
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

package telemetry

var (
	PosthogApiKey        string = ""
	PosthogEndpoint      string = "https://app.posthog.com"
	SummaryCronExpr      string = "0 0 * * *" // Run once a day, midnight
	HeartbeatCronExpr    string = "0 0/6 * * *"
	CacheExpiry          int    = 1440
	PosthogEncodedApiKey string = ""
	IsOptOut             bool   = false
)

const (
	TelemetryApiKeyEndpoint   string = "aHR0cHM6Ly90ZWxlbWV0cnkuZGV2dHJvbi5haS9kZXZ0cm9uL3RlbGVtZXRyeS9wb3N0aG9nSW5mbw=="
	TelemetryOptOutApiBaseUrl string = "aHR0cHM6Ly90ZWxlbWV0cnkuZGV2dHJvbi5haS9kZXZ0cm9uL3RlbGVtZXRyeS9vcHQtb3V0"
	ResponseApiKey            string = "PosthogApiKey"
	ResponseUrlKey            string = "PosthogEndpoint"
)
