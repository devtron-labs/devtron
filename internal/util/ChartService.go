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
	"github.com/ghodss/yaml"
	dirCopy "github.com/otiai10/copy"
	"go.uber.org/zap"
	"io/ioutil"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

type ChartWorkingDir string

type ChartTemplateService interface {
	CreateChart(chartMetaData *chart.Metadata, refChartLocation string, templateName string) (*ChartValues, *ChartGitAttribute, error)
	GetChartVersion(location string) (string, error)
	CreateChartProxy(chartMetaData *chart.Metadata, refChartLocation string, templateName string, version string, envName string, appName string) (string, *ChartGitAttribute, error)
	GitPull(clonedDir string, repoUrl string, appStoreName string) error
}
type ChartTemplateServiceImpl struct {
	randSource      rand.Source
	logger          *zap.SugaredLogger
	chartWorkingDir ChartWorkingDir
	gitFactory      *GitFactory
	client          *http.Client
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
	gitFactory *GitFactory) *ChartTemplateServiceImpl {
	return &ChartTemplateServiceImpl{
		randSource:      rand.NewSource(time.Now().UnixNano()),
		logger:          logger,
		chartWorkingDir: chartWorkingDir,
		client:          client,
		gitFactory:      gitFactory,
	}
}

func (ChartTemplateServiceImpl) GetChartVersion(location string) (string, error) {
	if fi, err := os.Stat(location); err != nil {
		return "", err
	} else if !fi.IsDir() {
		return "", fmt.Errorf("%q is not a directory", location)
	}

	chartYaml := filepath.Join(location, "Chart.yaml")
	if _, err := os.Stat(chartYaml); os.IsNotExist(err) {
		return "", fmt.Errorf("no Chart.yaml exists in directory %q", location)
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

func (impl ChartTemplateServiceImpl) CreateChart(chartMetaData *chart.Metadata, refChartLocation string, templateName string) (*ChartValues, *ChartGitAttribute, error) {
	chartMetaData.ApiVersion = "v1" // ensure always v1
	dir := impl.getDir()
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
	chartGitAttr, err := impl.createAndPushToGit(chartMetaData.Name, templateName, chartMetaData.Version, chartDir)
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

func (impl ChartTemplateServiceImpl) createAndPushToGit(appName, baseTemplateName, version, tmpChartLocation string) (chartGitAttribute *ChartGitAttribute, err error) {
	//baseTemplateName  replace whitespace
	space := regexp.MustCompile(`\s+`)
	appName = space.ReplaceAllString(appName, "-")
	repoUrl, _, err := impl.gitFactory.Client.CreateRepository(appName, "helm chart for "+appName)
	if err != nil {
		impl.logger.Errorw("error in creating git project", "name", appName, "err", err)
		return nil, err
	}

	chartDir := fmt.Sprintf("%s-%s", appName, impl.getDir())
	clonedDir := impl.gitFactory.gitService.GetCloneDirectory(chartDir)
	if _, err := os.Stat(clonedDir); os.IsNotExist(err) {
		clonedDir, err = impl.gitFactory.gitService.Clone(repoUrl, chartDir)
		if err != nil {
			impl.logger.Errorw("error in cloning repo", "url", repoUrl, "err", err)
			return nil, err
		}
	} else {
		err = impl.GitPull(clonedDir, repoUrl, appName)
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
	commit, err := impl.gitFactory.gitService.CommitAndPushAllChanges(clonedDir, "first commit")
	if err != nil {
		impl.logger.Errorw("error in pushing git", "err", err)
		impl.logger.Warn("re-trying, taking pull and then push again")
		err = impl.GitPull(clonedDir, repoUrl, appName)
		if err != nil {
			return nil, err
		}
		err = dirCopy.Copy(tmpChartLocation, dir)
		if err != nil {
			impl.logger.Errorw("error copying dir", "err", err)
			return nil, err
		}
		commit, err = impl.gitFactory.gitService.CommitAndPushAllChanges(clonedDir, "first commit")
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
	appOverrideByte, err := ioutil.ReadFile(filepath.Clean(filepath.Join(directory, "app-values.yaml")))
	if err != nil {
		return nil, err
	}
	appOverrideByte, err = yaml.YAMLToJSON(appOverrideByte)
	if err != nil {
		return nil, err
	}
	envOverrideByte, err := ioutil.ReadFile(filepath.Clean(filepath.Join(directory, "env-values.yaml")))
	if err != nil {
		return nil, err
	}
	envOverrideByte, err = yaml.YAMLToJSON(envOverrideByte)
	if err != nil {
		return nil, err
	}
	releaseOverrideByte, err := ioutil.ReadFile(filepath.Clean(filepath.Join(directory, "release-values.yaml")))
	if err != nil {
		return nil, err
	}
	releaseOverrideByte, err = yaml.YAMLToJSON(releaseOverrideByte)
	if err != nil {
		return nil, err
	}

	pipelineOverrideByte, err := ioutil.ReadFile(filepath.Clean(filepath.Join(directory, "pipeline-values.yaml")))
	if err != nil {
		return nil, err
	}
	pipelineOverrideByte, err = yaml.YAMLToJSON(pipelineOverrideByte)
	if err != nil {
		return nil, err
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

func (impl ChartTemplateServiceImpl) getDir() string {
	/* #nosec */
	r1 := rand.New(impl.randSource).Int63()
	return strconv.FormatInt(r1, 10)
}

func (impl ChartTemplateServiceImpl) CreateChartProxy(chartMetaData *chart.Metadata, refChartLocation string, templateName string, version string, envName string, appName string) (string, *ChartGitAttribute, error) {
	chartMetaData.ApiVersion = "v1" // ensure always v1
	dir := impl.getDir()
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

	chartGitAttr, err := impl.createAndPushToGitChartProxy(chartMetaData.Name, templateName, version, chartDir, envName, appName)
	if err != nil {
		impl.logger.Errorw("error in pushing chart to git ", "path", archivePath, "err", err)
		return "", nil, err
	}
	if len(valuesYaml) == 0 {
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

func (impl ChartTemplateServiceImpl) createAndPushToGitChartProxy(appStoreName, baseTemplateName, version, tmpChartLocation string, envName string, appName string) (chartGitAttribute *ChartGitAttribute, err error) {
	//baseTemplateName  replace whitespace
	space := regexp.MustCompile(`\s+`)
	appStoreName = space.ReplaceAllString(appStoreName, "-")
	repoUrl, _, err := impl.gitFactory.Client.CreateRepository(appStoreName, "helm chart for "+appStoreName)
	if err != nil {
		impl.logger.Errorw("error in creating git project", "name", appStoreName, "err", err)
		return nil, err
	}

	chartDir := fmt.Sprintf("%s-%s", appName, impl.getDir())
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

	acdAppName := fmt.Sprintf("%s-%s", appName, envName)
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
	commit, err := impl.gitFactory.gitService.CommitAndPushAllChanges(clonedDir, "first commit")
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
		commit, err = impl.gitFactory.gitService.CommitAndPushAllChanges(clonedDir, "first commit")
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
