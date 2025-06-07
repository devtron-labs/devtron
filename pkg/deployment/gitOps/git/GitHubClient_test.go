/*
 * Copyright (c) 2024. Devtron Inc.
 */

package git

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/api/bean"
	apiBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/commandManager"
	validationBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/validation/bean"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/rand"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestGitHubProvider(t *testing.T) {
	t.Run("GitHubClient Tests", func(t *testing.T) {
		t.Run("CreateRepository", newGithubTestClient(t).CreateRepositoryByClient)
		t.Run("CommitValues", newGithubTestClient(t).CommitValuesByClient)
	})
}

// gitHubTestConfig holds the configuration for GitHub tests
type gitHubTestConfig struct {
	GitHubUsername string `env:"GITHUB_USERNAME"`
	GitHubToken    string `env:"GITHUB_TOKEN"`
	GitHubOrgName  string `env:"GITHUB_ORG_NAME"`
}

func newGitHubTestConfig() (*gitHubTestConfig, error) {
	cfg := &gitHubTestConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func (github *gitHubTestConfig) getProvider() string {
	return GITHUB_PROVIDER
}

func (github *gitHubTestConfig) getHost() string {
	return "https://github.com/"
}

func (github *gitHubTestConfig) getOrg() string {
	return github.GitHubOrgName
}

func (github *gitHubTestConfig) getUsername() string {
	return github.GitHubUsername
}

func (github *gitHubTestConfig) getToken() string {
	return github.GitHubToken
}

func (github *gitHubTestConfig) getBasicAuth() *commandManager.BasicAuth {
	return &commandManager.BasicAuth{
		Username: github.getUsername(),
		Password: github.getToken(),
	}
}

func getGitHubTestHttpClient(t *testing.T) (GitHubClient, *GitOpsHelper, *gitHubTestConfig) {
	logger, err := util.NewSugardLogger()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	githubCfg, err := newGitHubTestConfig()
	if err != nil {
		t.Fatalf("failed to load GitHub test config: %v", err)
	}
	gitService, err := NewGitOpsHelperImpl(
		githubCfg.getBasicAuth(), logger,
		&bean.TLSConfig{}, false)
	if err != nil {
		t.Fatalf("failed to create GitOpsHelperImpl: %v", err)
	}
	githubClient, err := NewGithubClient(
		githubCfg.getHost(), githubCfg.getToken(), githubCfg.getOrg(),
		logger, gitService, &tls.Config{})
	if err != nil {
		t.Fatalf("failed to create GitHub client: %v", err)
	}
	return githubClient, gitService, githubCfg
}

type githubTestClient struct {
	gitHubClient GitHubClient
	gitOpsHelper *GitOpsHelper
	config       *gitHubTestConfig
}

func newGithubTestClient(t *testing.T) *githubTestClient {
	gitHubClient, gitOpsHelper, config := getGitHubTestHttpClient(t)
	return &githubTestClient{
		gitHubClient: gitHubClient,
		gitOpsHelper: gitOpsHelper,
		config:       config,
	}
}

// GitHubClient Tests -----------------------------
func (impl *githubTestClient) CreateRepositoryByClient(t *testing.T) {
	// t.SkipNow()
	tests := []struct {
		name               string
		getGitOpsConfigDto func(*gitHubTestConfig) *apiBean.GitOpsConfigDto
		wantIsNew          bool
		wantIsEmpty        bool
		wantErrCount       int
	}{
		{
			name: "valid case - create new repository",
			getGitOpsConfigDto: func(validGithubCfg *gitHubTestConfig) *apiBean.GitOpsConfigDto {
				return &apiBean.GitOpsConfigDto{
					GitRepoName: "test-repo-" + strings.ToLower(rand.SafeEncodeString(rand.String(5))),
					Description: "Test repository for GitHub client",
					Provider:    validGithubCfg.getProvider(),
					Username:    validGithubCfg.getUsername(),
					Token:       validGithubCfg.getToken(),
					GitHubOrgId: validGithubCfg.getOrg(),
					Host:        validGithubCfg.getHost(),
					Active:      true,
					UserEmailId: fmt.Sprintf("%s@devtron.ai", validGithubCfg.getUsername()),
					UserId:      1,
				}
			},
			wantIsNew:    true,
			wantIsEmpty:  false,
			wantErrCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitOpsConfigDTO := tt.getGitOpsConfigDto(impl.config)
			// Create a new repository
			_, gotIsNew, gotIsEmpty, detailedErrorGitOpsConfigActions := impl.gitHubClient.CreateRepository(context.Background(), gitOpsConfigDTO)
			// Clean up by deleting the repository
			defer func() {
				if slices.Contains(detailedErrorGitOpsConfigActions.SuccessfulStages, validationBean.CreateRepoStage) {
					err := impl.gitHubClient.DeleteRepository(gitOpsConfigDTO)
					assert.NoError(t, err, "DeleteRepository() error = %v", err)
				}
			}()
			errorCount := len(detailedErrorGitOpsConfigActions.StageErrorMap)
			assert.Equalf(t, tt.wantErrCount, errorCount, "CreateRepository() error count = %v, want 0", errorCount)
			if tt.wantErrCount == 0 {
				assert.Equalf(t, tt.wantIsNew, gotIsNew, "CreateRepository() gotIsNew = %v, want %v", gotIsNew, tt.wantIsNew)
				assert.Equalf(t, tt.wantIsEmpty, gotIsEmpty, "CreateRepository() gotIsEmpty = %v, want %v", gotIsEmpty, tt.wantIsEmpty)
			}
			// Create the same repository again to check if it is not created again
			_, gotIsNew, gotIsEmpty, detailedErrorGitOpsConfigActions = impl.gitHubClient.CreateRepository(context.Background(), gitOpsConfigDTO)
			errorCount = len(detailedErrorGitOpsConfigActions.StageErrorMap)
			assert.Equalf(t, tt.wantErrCount, errorCount, "CreateRepository() error count = %v, want 0", errorCount)
			if tt.wantErrCount == 0 {
				assert.Equalf(t, false, gotIsNew, "CreateRepository() gotIsNew = %v, want %v", gotIsNew, false)
				// If the repository just created, GitHub API calculates the size hourly,
				// So it may not be empty immediately,
				// Hence we skip the assertion for gotIsEmpty in the second call.
				// assert.Equalf(t, false, gotIsEmpty, "CreateRepository() gotIsEmpty = %v, want %v", gotIsEmpty, false)
			}
		})
	}
}

func (impl *githubTestClient) CommitValuesByClient(t *testing.T) {
	// t.SkipNow()
	gitOpsConfigDTO := &apiBean.GitOpsConfigDto{
		GitRepoName: "test-repo-" + strings.ToLower(rand.SafeEncodeString(rand.String(5))),
		Description: "Test repository for GitHub client",
		Provider:    impl.config.getProvider(),
		Username:    impl.config.getUsername(),
		Token:       impl.config.getToken(),
		GitHubOrgId: impl.config.getOrg(),
		Host:        impl.config.getHost(),
		Active:      true,
		UserEmailId: fmt.Sprintf("%s@devtron.ai", impl.config.getUsername()),
		UserId:      1,
	}
	repoUrl, isNew, isEmpty, detailedErrorGitOpsConfigActions := impl.gitHubClient.CreateRepository(context.Background(), gitOpsConfigDTO)
	// Clean up by deleting the repository
	defer func() {
		if slices.Contains(detailedErrorGitOpsConfigActions.SuccessfulStages, validationBean.CreateRepoStage) {
			err := impl.gitHubClient.DeleteRepository(gitOpsConfigDTO)
			assert.NoError(t, err, "DeleteRepository() error = %v", err)
		}
	}()
	// Assert that the repository was created successfully
	assert.Equal(t, true, isNew)
	// Assert that the repository is not empty as we commit a README file
	assert.Equal(t, false, isEmpty)
	// Assert that the repository URL is not empty
	assert.NotEmpty(t, repoUrl)
	// Assert that there are no errors in the detailed error map
	assert.Empty(t, detailedErrorGitOpsConfigActions.StageErrorMap)

	// Test case for committing values to the repository
	tests := []struct {
		name        string
		chartConfig func(gitOpsConfigDTO *apiBean.GitOpsConfigDto, repoUrl string) *ChartConfig
		assertError assert.ErrorAssertionFunc
	}{
		{
			name: "valid commit",
			chartConfig: func(gitOpsConfigDTO *apiBean.GitOpsConfigDto, repoUrl string) *ChartConfig {
				return &ChartConfig{
					FileName: "README.md",
					FileContent: `## Description:
This is a test chart for testing GitHub client commit functionality.
Signature: @devtron-labs`,
					ReleaseMessage:   "test commit",
					ChartRepoName:    gitOpsConfigDTO.GitRepoName,
					TargetRevision:   "master",
					UseDefaultBranch: true,
					UserName:         gitOpsConfigDTO.Username,
					UserEmailId:      gitOpsConfigDTO.UserEmailId,
				}
			},
			assertError: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chartConfig := tt.chartConfig(gitOpsConfigDTO, repoUrl)
			commitHash, commitTime, err := impl.gitHubClient.CommitValues(context.Background(), chartConfig, gitOpsConfigDTO, false)
			if !tt.assertError(t, err, "CommitValues() error = %v", err) {
				return
			}
			assert.NotEmpty(t, commitHash)
			assert.NotEqual(t, time.Time{}, commitTime)
			// Verify the commit by checking the file content

		})
	}
}

// ------------------------------------------------
