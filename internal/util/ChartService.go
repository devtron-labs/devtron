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

package util

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/devtron-labs/devtron/util"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/go-pg/pg"
	dirCopy "github.com/otiai10/copy"
	"go.uber.org/zap"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

type ChartWorkingDir string

type ChartTemplateService interface {
	CreateChart(chartMetaData *chart.Metadata, refChartLocation string, templateName string, userId int32) (*ChartValues, *ChartGitAttribute, error)
	GetChartVersion(location string) (string, error)
	CreateChartProxy(chartMetaData *chart.Metadata, refChartLocation string, templateName string, version string, envName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (string, *ChartGitAttribute, error)
	GitPull(clonedDir string, repoUrl string, appStoreName string) error
	GetDir() string
	GetUserEmailIdAndNameForGitOpsCommit(userId int32) (emailId, name string)
	GetGitOpsRepoName(appName string) string
	GetGitOpsRepoNameFromUrl(gitRepoUrl string) string
}
type ChartTemplateServiceImpl struct {
	randSource             rand.Source
	logger                 *zap.SugaredLogger
	chartWorkingDir        ChartWorkingDir
	gitFactory             *GitFactory
	client                 *http.Client
	globalEnvVariables     *util.GlobalEnvVariables
	gitOpsConfigRepository repository.GitOpsConfigRepository
	userRepository         repository2.UserRepository
}

type ChartValues struct {
	Values                  string `json:"values"`            //yaml
	AppOverrides            string `json:"appOverrides"`      //json
	EnvOverrides            string `json:"envOverrides"`      //json
	ReleaseOverrides        string `json:"releaseOverrides"`  //json
	PipelineOverrides       string `json:"pipelineOverrides"` //json
	ImageDescriptorTemplate string `json:"-"`
}

func NewChartTemplateServiceImpl(logger *zap.SugaredLogger,
	chartWorkingDir ChartWorkingDir,
	client *http.Client,
	gitFactory *GitFactory, globalEnvVariables *util.GlobalEnvVariables,
	gitOpsConfigRepository repository.GitOpsConfigRepository,
	userRepository repository2.UserRepository) *ChartTemplateServiceImpl {
	return &ChartTemplateServiceImpl{
		randSource:             rand.NewSource(time.Now().UnixNano()),
		logger:                 logger,
		chartWorkingDir:        chartWorkingDir,
		client:                 client,
		gitFactory:             gitFactory,
		globalEnvVariables:     globalEnvVariables,
		gitOpsConfigRepository: gitOpsConfigRepository,
		userRepository:         userRepository,
	}
}

func (impl ChartTemplateServiceImpl) GetChartVersion(location string) (string, error) {
	if fi, err := os.Stat(location); err != nil {
		return "", err
	} else if !fi.IsDir() {
		return "", fmt.Errorf("%q is not a directory", location)
	}

	chartYaml := filepath.Join(location, "Chart.yaml")
	if _, err := os.Stat(chartYaml); os.IsNotExist(err) {
		return "", fmt.Errorf("Chart.yaml file not present in the directory %q", location)
	}
	//chartYaml = filepath.Join(chartYaml,filepath.Clean(chartYaml))
	chartYamlContent, err := ioutil.ReadFile(filepath.Clean(chartYaml))
	if err != nil {
		return "", fmt.Errorf("cannot read Chart.Yaml in directory %q", location)
	}

	chartContent, err := chartutil.UnmarshalChartfile(chartYamlContent)
	if err != nil {
		return "", fmt.Errorf("cannot read Chart.Yaml in directory %q", location)
	}

	return chartContent.Version, nil
}

func (impl ChartTemplateServiceImpl) CreateChart(chartMetaData *chart.Metadata, refChartLocation string, templateName string, userId int32) (*ChartValues, *ChartGitAttribute, error) {
	chartMetaData.ApiVersion = "v1" // ensure always v1
	dir := impl.GetDir()
	chartDir := filepath.Join(string(impl.chartWorkingDir), dir)
	impl.logger.Debugw("chart dir ", "chart", chartMetaData.Name, "dir", chartDir)
	err := os.MkdirAll(chartDir, os.ModePerm) //hack for concurrency handling
	if err != nil {
		impl.logger.Errorw("err in creating dir", "dir", chartDir, "err", err)
		return nil, nil, err
	}

	defer impl.CleanDir(chartDir)
	err = dirCopy.Copy(refChartLocation, chartDir)

	if err != nil {
		impl.logger.Errorw("error in copying chart for app", "app", chartMetaData.Name, "error", err)
		return nil, nil, err
	}
	archivePath, valuesYaml, err := impl.packageChart(chartDir, chartMetaData)
	if err != nil {
		impl.logger.Errorw("error in creating archive", "err", err)
		return nil, nil, err
	}
	values, err := impl.getValues(chartDir)
	if err != nil {
		impl.logger.Errorw("error in pushing chart", "path", archivePath, "err", err)
		return nil, nil, err
	}
	values.Values = valuesYaml
	gitOpsRepoName := impl.GetGitOpsRepoName(chartMetaData.Name)
	chartGitAttr, err := impl.createAndPushToGit(gitOpsRepoName, templateName, chartMetaData.Version, chartDir, userId)
	if err != nil {
		impl.logger.Errorw("error in pushing chart to git ", "path", archivePath, "err", err)
		return nil, nil, err
	}
	descriptor, err := ioutil.ReadFile(filepath.Clean(filepath.Join(chartDir, ".image_descriptor_template.json")))
	if err != nil {
		impl.logger.Errorw("error in reading descriptor", "path", chartDir, "err", err)
		return nil, nil, err
	}
	values.ImageDescriptorTemplate = string(descriptor)
	return values, chartGitAttr, nil
}

type ChartGitAttribute struct {
	RepoUrl, ChartLocation string
}

func (impl ChartTemplateServiceImpl) createAndPushToGit(gitOpsRepoName, baseTemplateName, version, tmpChartLocation string, userId int32) (chartGitAttribute *ChartGitAttribute, err error) {
	//baseTemplateName  replace whitespace
	space := regexp.MustCompile(`\s+`)
	gitOpsRepoName = space.ReplaceAllString(gitOpsRepoName, "-")

	gitOpsConfigBitbucket, err := impl.gitFactory.gitOpsRepository.GetGitOpsConfigByProvider(BITBUCKET_PROVIDER)
	if err != nil {
		if err == pg.ErrNoRows {
			gitOpsConfigBitbucket.BitBucketWorkspaceId = ""
			gitOpsConfigBitbucket.BitBucketProjectKey = ""
		} else {
			impl.logger.Errorw("error in fetching gitOps bitbucket config", "err", err)
			return nil, err
		}
	}
	//getting user name & emailId for commit author data
	userEmailId, userName := impl.GetUserEmailIdAndNameForGitOpsCommit(userId)
	repoUrl, _, detailedError := impl.gitFactory.Client.CreateRepository(gitOpsRepoName, fmt.Sprintf("helm chart for "+gitOpsRepoName), gitOpsConfigBitbucket.BitBucketWorkspaceId, gitOpsConfigBitbucket.BitBucketProjectKey, userName, userEmailId)

	for _, err := range detailedError.StageErrorMap {
		if err != nil {
			impl.logger.Errorw("error in creating git project", "name", gitOpsRepoName, "err", err)
			return nil, err
		}
	}

	chartDir := fmt.Sprintf("%s-%s", gitOpsRepoName, impl.GetDir())
	clonedDir := impl.gitFactory.gitService.GetCloneDirectory(chartDir)
	if _, err := os.Stat(clonedDir); os.IsNotExist(err) {
		clonedDir, err = impl.gitFactory.gitService.Clone(repoUrl, chartDir)
		if err != nil {
			impl.logger.Errorw("error in cloning repo", "url", repoUrl, "err", err)
			return nil, err
		}
	} else {
		err = impl.GitPull(clonedDir, repoUrl, gitOpsRepoName)
		if err != nil {
			return nil, err
		}
	}

	dir := filepath.Join(clonedDir, baseTemplateName, version)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		impl.logger.Errorw("error in making dir", "err", err)
		return nil, nil
	}
	err = dirCopy.Copy(tmpChartLocation, dir)
	if err != nil {
		impl.logger.Errorw("error copying dir", "err", err)
		return nil, nil
	}
	commit, err := impl.gitFactory.gitService.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
	if err != nil {
		impl.logger.Errorw("error in pushing git", "err", err)
		impl.logger.Warn("re-trying, taking pull and then push again")
		err = impl.GitPull(clonedDir, repoUrl, gitOpsRepoName)
		if err != nil {
			return nil, err
		}
		err = dirCopy.Copy(tmpChartLocation, dir)
		if err != nil {
			impl.logger.Errorw("error copying dir", "err", err)
			return nil, err
		}
		commit, err = impl.gitFactory.gitService.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
		if err != nil {
			impl.logger.Errorw("error in pushing git", "err", err)
			return nil, err
		}
	}
	impl.logger.Debugw("template committed", "url", repoUrl, "commit", commit)

	defer impl.CleanDir(clonedDir)
	return &ChartGitAttribute{RepoUrl: repoUrl, ChartLocation: filepath.Join(baseTemplateName, version)}, nil
}

func (impl ChartTemplateServiceImpl) getValues(directory string) (values *ChartValues, err error) {

	if fi, err := os.Stat(directory); err != nil {
		return nil, err
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", directory)
	}

	files, err := ioutil.ReadDir(directory)
	if err != nil {
		impl.logger.Errorw("failed reading directory", "err", err)
		return nil, fmt.Errorf(" Couldn't read the %q", directory)
	}

	var appOverrideByte, envOverrideByte, releaseOverrideByte, pipelineOverrideByte []byte

	for _, file := range files {
		if !file.IsDir() {
			name := strings.ToLower(file.Name())
			if name == "app-values.yaml" || name == "app-values.yml" {
				appOverrideByte, err = ioutil.ReadFile(filepath.Clean(filepath.Join(directory, file.Name())))
				if err != nil {
					impl.logger.Errorw("failed reading data from file", "err", err)
				}
				appOverrideByte, err = yaml.YAMLToJSON(appOverrideByte)
				if err != nil {
					return nil, err
				}
			}
			if name == "env-values.yaml" || name == "env-values.yml" {
				envOverrideByte, err = ioutil.ReadFile(filepath.Clean(filepath.Join(directory, file.Name())))
				if err != nil {
					impl.logger.Errorw("failed reading data from file", "err", err)
				}
				envOverrideByte, err = yaml.YAMLToJSON(envOverrideByte)
				if err != nil {
					return nil, err
				}
			}
			if name == "release-values.yaml" || name == "release-values.yml" {
				releaseOverrideByte, err = ioutil.ReadFile(filepath.Clean(filepath.Join(directory, file.Name())))
				if err != nil {
					impl.logger.Errorw("failed reading data from file", "err", err)
				}
				releaseOverrideByte, err = yaml.YAMLToJSON(releaseOverrideByte)
				if err != nil {
					return nil, err
				}
			}
			if name == "pipeline-values.yaml" || name == "pipeline-values.yml" {
				pipelineOverrideByte, err = ioutil.ReadFile(filepath.Clean(filepath.Join(directory, file.Name())))
				if err != nil {
					impl.logger.Errorw("failed reading data from file", "err", err)
				}
				pipelineOverrideByte, err = yaml.YAMLToJSON(pipelineOverrideByte)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	val := &ChartValues{
		AppOverrides:      string(appOverrideByte),
		EnvOverrides:      string(envOverrideByte),
		ReleaseOverrides:  string(releaseOverrideByte),
		PipelineOverrides: string(pipelineOverrideByte),
	}
	return val, nil

}

func (impl ChartTemplateServiceImpl) packageChart(directory string, chartMetaData *chart.Metadata) (*string, string, error) {
	valid, err := chartutil.IsChartDir(directory)
	if err != nil {
		impl.logger.Errorw("error in validating base chart", "dir", directory, "err", err)
		return nil, "", err
	}
	if !valid {
		impl.logger.Errorw("invalid chart at ", "dir", directory)
		return nil, "", fmt.Errorf("invalid base chart")
	}

	b, err := yaml.Marshal(chartMetaData)
	if err != nil {
		impl.logger.Errorw("error in marshaling chartMetadata", "err", err)
		return nil, "", err
	}
	err = ioutil.WriteFile(filepath.Join(directory, "Chart.yaml"), b, 0600)
	if err != nil {
		impl.logger.Errorw("err in writing Chart.yaml", "err", err)
		return nil, "", err
	}
	chart, err := chartutil.LoadDir(directory)
	if err != nil {
		impl.logger.Errorw("error in loading chart dir", "err", err, "dir", directory)
		return nil, "", err
	}

	archivePath, err := chartutil.Save(chart, directory)
	if err != nil {
		impl.logger.Errorw("error in saving", "err", err, "dir", directory)
		return nil, "", err
	}
	impl.logger.Debugw("chart archive path", "path", archivePath)
	//chart.Values
	valuesYaml := chart.Values.Raw
	return &archivePath, valuesYaml, nil
}

func (impl ChartTemplateServiceImpl) CleanDir(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		impl.logger.Warnw("error in deleting dir ", "dir", dir)
	}
}

func (impl ChartTemplateServiceImpl) GetDir() string {
	/* #nosec */
	r1 := rand.New(impl.randSource).Int63()
	return strconv.FormatInt(r1, 10)
}

func (impl ChartTemplateServiceImpl) CreateChartProxy(chartMetaData *chart.Metadata, refChartLocation string, templateName string, version string, envName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (string, *ChartGitAttribute, error) {
	chartMetaData.ApiVersion = "v2" // ensure always v2
	dir := impl.GetDir()
	chartDir := filepath.Join(string(impl.chartWorkingDir), dir)
	impl.logger.Debugw("chart dir ", "chart", chartMetaData.Name, "dir", chartDir)
	err := os.MkdirAll(chartDir, os.ModePerm) //hack for concurrency handling
	if err != nil {
		impl.logger.Errorw("err in creating dir", "dir", chartDir, "err", err)
		return "", nil, err
	}
	defer impl.CleanDir(chartDir)
	err = dirCopy.Copy(refChartLocation, chartDir)

	if err != nil {
		impl.logger.Errorw("error in copying chart for app", "app", chartMetaData.Name, "error", err)
		return "", nil, err
	}
	archivePath, valuesYaml, err := impl.packageChart(chartDir, chartMetaData)
	if err != nil {
		impl.logger.Errorw("error in creating archive", "err", err)
		return "", nil, err
	}

	chartGitAttr, err := impl.createAndPushToGitChartProxy(chartMetaData.Name, templateName, version, chartDir, envName, installAppVersionRequest)
	if err != nil {
		impl.logger.Errorw("error in pushing chart to git ", "path", archivePath, "err", err)
		return "", nil, err
	}
	if valuesYaml == "" {
		valuesYaml = "{}"
	} else {
		valuesYamlByte, err := yaml.YAMLToJSON([]byte(valuesYaml))
		if err != nil {
			return "", nil, err
		}
		valuesYaml = string(valuesYamlByte)
	}
	return valuesYaml, chartGitAttr, nil
}

func (impl ChartTemplateServiceImpl) createAndPushToGitChartProxy(appStoreName, baseTemplateName, version, tmpChartLocation string, envName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (chartGitAttribute *ChartGitAttribute, err error) {
	//baseTemplateName  replace whitespace
	space := regexp.MustCompile(`\s+`)
	appStoreName = space.ReplaceAllString(appStoreName, "-")
	if len(installAppVersionRequest.GitOpsRepoName) == 0 {
		gitOpsRepoName := impl.GetGitOpsRepoName(appStoreName)
		installAppVersionRequest.GitOpsRepoName = gitOpsRepoName
	}
	gitOpsConfigBitbucket, err := impl.gitFactory.gitOpsRepository.GetGitOpsConfigByProvider(BITBUCKET_PROVIDER)
	if err != nil {
		if err == pg.ErrNoRows {
			gitOpsConfigBitbucket.BitBucketWorkspaceId = ""
			gitOpsConfigBitbucket.BitBucketProjectKey = ""
		} else {
			impl.logger.Errorw("error in fetching gitOps bitbucket config", "err", err)
			return nil, err
		}
	}
	//getting user name & emailId for commit author data
	userEmailId, userName := impl.GetUserEmailIdAndNameForGitOpsCommit(installAppVersionRequest.UserId)
	repoUrl, _, detailedError := impl.gitFactory.Client.CreateRepository(installAppVersionRequest.GitOpsRepoName, "helm chart for "+installAppVersionRequest.GitOpsRepoName, gitOpsConfigBitbucket.BitBucketWorkspaceId, gitOpsConfigBitbucket.BitBucketProjectKey, userName, userEmailId)
	for _, err := range detailedError.StageErrorMap {
		if err != nil {
			impl.logger.Errorw("error in creating git project", "name", installAppVersionRequest.GitOpsRepoName, "err", err)
			return nil, err
		}
	}
	chartDir := fmt.Sprintf("%s-%s", installAppVersionRequest.AppName, impl.GetDir())
	clonedDir := impl.gitFactory.gitService.GetCloneDirectory(chartDir)
	if _, err := os.Stat(clonedDir); os.IsNotExist(err) {
		clonedDir, err = impl.gitFactory.gitService.Clone(repoUrl, chartDir)
		if err != nil {
			impl.logger.Errorw("error in cloning repo", "url", repoUrl, "err", err)
			return nil, err
		}
	} else {
		err = impl.GitPull(clonedDir, repoUrl, appStoreName)
		if err != nil {
			return nil, err
		}
	}

	acdAppName := fmt.Sprintf("%s-%s", installAppVersionRequest.AppName, envName)
	dir := filepath.Join(clonedDir, acdAppName)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		impl.logger.Errorw("error in making dir", "err", err)
		return nil, err
	}
	err = dirCopy.Copy(tmpChartLocation, dir)
	if err != nil {
		impl.logger.Errorw("error copying dir", "err", err)
		return nil, err
	}
	commit, err := impl.gitFactory.gitService.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
	if err != nil {
		impl.logger.Errorw("error in pushing git", "err", err)
		impl.logger.Warn("re-trying, taking pull and then push again")
		err = impl.GitPull(clonedDir, repoUrl, acdAppName)
		if err != nil {
			return nil, err
		}
		err = dirCopy.Copy(tmpChartLocation, dir)
		if err != nil {
			impl.logger.Errorw("error copying dir", "err", err)
			return nil, err
		}
		commit, err = impl.gitFactory.gitService.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
		if err != nil {
			impl.logger.Errorw("error in pushing git", "err", err)
			return nil, err
		}
	}
	impl.logger.Debugw("template committed", "url", repoUrl, "commit", commit)
	defer impl.CleanDir(clonedDir)
	return &ChartGitAttribute{RepoUrl: repoUrl, ChartLocation: filepath.Join("", acdAppName)}, nil
}

func (impl ChartTemplateServiceImpl) GitPull(clonedDir string, repoUrl string, appStoreName string) error {
	err := impl.gitFactory.gitService.Pull(clonedDir) //TODO check for local repo exists before clone
	if err != nil {
		impl.logger.Errorw("error in pulling git", "clonedDir", clonedDir, "err", err)
		_, err := impl.gitFactory.gitService.Clone(repoUrl, appStoreName)
		if err != nil {
			impl.logger.Errorw("error in cloning repo", "url", repoUrl, "err", err)
			return err
		}
		return nil
	}
	return nil
}

func (impl *ChartTemplateServiceImpl) GetUserEmailIdAndNameForGitOpsCommit(userId int32) (string, string) {
	emailId := "devtron-bot@devtron.ai"
	name := "devtron bot"
	//getting emailId associated with user
	userDetail, _ := impl.userRepository.GetById(userId)
	if userDetail != nil && userDetail.EmailId != "admin" && userDetail.EmailId != "system" && len(userDetail.EmailId) > 0 {
		emailId = userDetail.EmailId
	} else {
		emailIdGitOps, err := impl.gitOpsConfigRepository.GetEmailIdFromActiveGitOpsConfig()
		if err != nil {
			impl.logger.Errorw("error in getting emailId from active gitOps config", "err", err)
		} else if len(emailIdGitOps) > 0 {
			emailId = emailIdGitOps
		}
	}
	//we are getting name from emailId(replacing special characters in <user-name part of email> with space)
	emailComponents := strings.Split(emailId, "@")
	regex, _ := regexp.Compile(`[^\w]`)
	if regex != nil {
		name = regex.ReplaceAllString(emailComponents[0], " ")
	}
	return emailId, name
}

func (impl ChartTemplateServiceImpl) GetGitOpsRepoName(appName string) string {
	var repoName string
	if len(impl.globalEnvVariables.GitOpsRepoPrefix) == 0 {
		repoName = appName
	} else {
		repoName = fmt.Sprintf("%s-%s", impl.globalEnvVariables.GitOpsRepoPrefix, appName)
	}
	return repoName
}

func (impl ChartTemplateServiceImpl) GetGitOpsRepoNameFromUrl(gitRepoUrl string) string {
	gitRepoUrl = gitRepoUrl[strings.LastIndex(gitRepoUrl, "/")+1:]
	gitRepoUrl = strings.ReplaceAll(gitRepoUrl, ".git", "")
	return gitRepoUrl
}
