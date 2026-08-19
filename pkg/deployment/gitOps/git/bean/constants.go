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

package bean

const (
	GIT_WORKING_DIR       = "/tmp/gitops/"
	GetRepoUrlStage       = "Get Repo RedirectionUrl"
	CreateRepoStage       = "Create Repo"
	CloneHttpStage        = "Clone Http"
	CreateReadmeStage     = "Create Readme"
	CloneSshStage         = "Clone Ssh"
	GITLAB_PROVIDER       = "GITLAB"
	GITHUB_PROVIDER       = "GITHUB"
	AZURE_DEVOPS_PROVIDER = "AZURE_DEVOPS"
	BITBUCKET_PROVIDER    = "BITBUCKET_CLOUD"
	GITHUB_API_V3         = "api/v3"
	GITHUB_HOST           = "github.com"
	GIT_TLS_DIR           = "/tmp/gitops/tls"
	// BITBUCKET_ACCESS_TOKEN_USERNAME is the fixed username Bitbucket Cloud expects for
	// git-over-HTTPS when authenticating with a repository/project/workspace access token.
	// Ref: https://support.atlassian.com/bitbucket-cloud/docs/using-access-tokens/
	BITBUCKET_ACCESS_TOKEN_USERNAME = "x-token-auth"

	// BITBUCKET_API_TOKEN_USERNAME is the fixed username Bitbucket Cloud expects for
	// git-over-HTTPS when authenticating with an Atlassian API token. (The REST API, by
	// contrast, uses the account email as the Basic-auth username — see NewGitBitbucketClient.)
	// Ref: https://support.atlassian.com/bitbucket-cloud/docs/using-api-tokens/
	BITBUCKET_API_TOKEN_USERNAME = "x-bitbucket-api-token-auth"
)
