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
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	dirCopy "github.com/otiai10/copy"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"sigs.k8s.io/yaml"
)

const (
	PIPELINE_DEPLOYMENT_TYPE_ACD               = "argo_cd"
	PIPELINE_DEPLOYMENT_TYPE_HELM              = "helm"
	PIPELINE_DEPLOYMENT_TYPE_MANIFEST_DOWNLOAD = "manifest_download"
	PIPELINE_DEPLOYMENT_TYPE_MANIFEST_PUSH     = "manifest_push"
	CHART_WORKING_DIR_PATH                     = "/tmp/charts/"
)

type ChartCreateRequest struct {
	ChartMetaData       *chart.Metadata
	ChartPath           string
	IncludePackageChart bool
}

type ChartCreateResponse struct {
	BuiltChartPath string
	valuesYaml     string
}

type ChartTemplateService interface {
	FetchValuesFromReferenceChart(chartMetaData *chart.Metadata, refChartLocation string, templateName string, userId int32, pipelineStrategyPath string) (*ChartValues, error)
	GetChartVersion(location string) (string, error)
	BuildChart(ctx context.Context, chartMetaData *chart.Metadata, referenceTemplatePath string) (string, error)
	BuildChartProxyForHelmApps(chartCreateRequest *ChartCreateRequest) (chartCreateResponse *ChartCreateResponse, err error)
	GetDir() string
	CleanDir(dir string)
	GetByteArrayRefChart(chartMetaData *chart.Metadata, referenceTemplatePath string) ([]byte, error)
	LoadChartInBytes(ChartPath string, deleteChart bool, chartName string, chartVersion string) ([]byte, error)
	LoadChartFromDir(dir string) (*chart.Chart, error)
	CreateZipFileForChart(chart *chart.Chart, outputChartPathDir string) ([]byte, error)
	PackageChart(tempReferenceTemplateDir string, chartMetaData *chart.Metadata) (*string, string, error)
}
type ChartTemplateServiceImpl struct {
	randSource rand.Source
	logger     *zap.SugaredLogger
}

type ChartValues struct {
	Values                  string `json:"values"`            //yaml
	AppOverrides            string `json:"appOverrides"`      //json
	EnvOverrides            string `json:"envOverrides"`      //json
	ReleaseOverrides        string `json:"releaseOverrides"`  //json
	PipelineOverrides       string `json:"pipelineOverrides"` //json
	ImageDescriptorTemplate string `json:"-"`
}

func NewChartTemplateServiceImpl(logger *zap.SugaredLogger) *ChartTemplateServiceImpl {
	return &ChartTemplateServiceImpl{
		randSource: rand.NewSource(time.Now().UnixNano()),
		logger:     logger,
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

func (impl ChartTemplateServiceImpl) FetchValuesFromReferenceChart(chartMetaData *chart.Metadata, refChartLocation string, templateName string, userId int32, pipelineStrategyPath string) (*ChartValues, error) {
	chartMetaData.ApiVersion = "v1" // ensure always v1
	dir := impl.GetDir()
	chartDir := filepath.Join(CHART_WORKING_DIR_PATH, dir)
	impl.logger.Debugw("chart dir ", "chart", chartMetaData.Name, "dir", chartDir)
	err := os.MkdirAll(chartDir, os.ModePerm) //hack for concurrency handling
	if err != nil {
		impl.logger.Errorw("err in creating dir", "dir", chartDir, "err", err)
		return nil, err
	}

	defer impl.CleanDir(chartDir)
	err = dirCopy.Copy(refChartLocation, chartDir)

	if err != nil {
		impl.logger.Errorw("error in copying chart for app", "app", chartMetaData.Name, "error", err)
		return nil, err
	}
	archivePath, valuesYaml, err := impl.PackageChart(chartDir, chartMetaData)
	if err != nil {
		impl.logger.Errorw("error in creating archive", "err", err)
		return nil, err
	}
	values, err := impl.getValues(chartDir, pipelineStrategyPath)
	if err != nil {
		impl.logger.Errorw("error in pushing chart", "path", archivePath, "err", err)
		return nil, err
	}
	values.Values = valuesYaml
	descriptor, err := ioutil.ReadFile(filepath.Clean(filepath.Join(chartDir, ".image_descriptor_template.json")))
	if err != nil {
		impl.logger.Errorw("error in reading descriptor", "path", chartDir, "err", err)
		return nil, err
	}
	values.ImageDescriptorTemplate = string(descriptor)
	return values, nil
}

// TODO: convert BuildChart and BuildChartProxyForHelmApps into one function
func (impl ChartTemplateServiceImpl) BuildChart(ctx context.Context, chartMetaData *chart.Metadata, referenceTemplatePath string) (string, error) {
	if chartMetaData.ApiVersion == "" {
		chartMetaData.ApiVersion = "v1" // ensure always v1
	}
	dir := impl.GetDir()
	tempReferenceTemplateDir := filepath.Join(CHART_WORKING_DIR_PATH, dir)
	impl.logger.Debugw("chart dir ", "chart", chartMetaData.Name, "dir", tempReferenceTemplateDir)
	err := os.MkdirAll(tempReferenceTemplateDir, os.ModePerm) //hack for concurrency handling
	if err != nil {
		impl.logger.Errorw("err in creating dir", "dir", tempReferenceTemplateDir, "err", err)
		return "", err
	}
	err = dirCopy.Copy(referenceTemplatePath, tempReferenceTemplateDir)

	if err != nil {
		impl.logger.Errorw("error in copying chart for app", "app", chartMetaData.Name, "error", err)
		return "", err
	}
	_, span := otel.Tracer("orchestrator").Start(ctx, "impl.PackageChart")
	_, _, err = impl.PackageChart(tempReferenceTemplateDir, chartMetaData)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in creating archive", "err", err)
		return "", err
	}
	return tempReferenceTemplateDir, nil
}

func (impl ChartTemplateServiceImpl) BuildChartProxyForHelmApps(chartCreateRequest *ChartCreateRequest) (*ChartCreateResponse, error) {
	chartCreateResponse := &ChartCreateResponse{}
	chartMetaData := chartCreateRequest.ChartMetaData
	chartMetaData.ApiVersion = "v2" // ensure always v2
	dir := impl.GetDir()
	chartDir := filepath.Join(CHART_WORKING_DIR_PATH, dir)
	impl.logger.Debugw("chart dir ", "chart", chartMetaData.Name, "dir", chartDir)
	err := os.MkdirAll(chartDir, os.ModePerm) //hack for concurrency handling
	if err != nil {
		impl.logger.Errorw("err in creating dir", "dir", chartDir, "err", err)
		return chartCreateResponse, err
	}
	err = dirCopy.Copy(chartCreateRequest.ChartPath, chartDir)

	if err != nil {
		impl.logger.Errorw("error in copying chart for app", "app", chartMetaData.Name, "error", err)
		return chartCreateResponse, err
	}
	if chartCreateRequest.IncludePackageChart {
		_, valuesYaml, err := impl.PackageChart(chartDir, chartMetaData)
		if err != nil {
			impl.logger.Errorw("error in creating archive", "err", err)
			return chartCreateResponse, err
		}
		chartCreateResponse.valuesYaml = valuesYaml
	} else {
		b, err := yaml.Marshal(chartMetaData)
		if err != nil {
			impl.logger.Errorw("error in marshaling chartMetadata", "err", err)
			return chartCreateResponse, err
		}
		err = ioutil.WriteFile(filepath.Join(chartDir, "Chart.yaml"), b, 0600)
		if err != nil {
			impl.logger.Errorw("err in writing Chart.yaml", "err", err)
			return chartCreateResponse, err
		}
	}
	chartCreateResponse.BuiltChartPath = chartDir
	return chartCreateResponse, nil
}

func (impl ChartTemplateServiceImpl) getValues(directory, pipelineStrategyPath string) (values *ChartValues, err error) {

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
				} else {
					appOverrideByte, err = yaml.YAMLToJSON(appOverrideByte)
					if err != nil {
						return nil, err
					}
				}
			}
			if name == "env-values.yaml" || name == "env-values.yml" {
				envOverrideByte, err = ioutil.ReadFile(filepath.Clean(filepath.Join(directory, file.Name())))
				if err != nil {
					impl.logger.Errorw("failed reading data from file", "err", err)
				} else {
					envOverrideByte, err = yaml.YAMLToJSON(envOverrideByte)
					if err != nil {
						return nil, err
					}
				}
			}
			if name == "release-values.yaml" || name == "release-values.yml" {
				releaseOverrideByte, err = ioutil.ReadFile(filepath.Clean(filepath.Join(directory, file.Name())))
				if err != nil {
					impl.logger.Errorw("failed reading data from file", "err", err)
				} else {
					releaseOverrideByte, err = yaml.YAMLToJSON(releaseOverrideByte)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}
	pipelineOverrideByte, err = ioutil.ReadFile(filepath.Clean(filepath.Join(directory, pipelineStrategyPath)))
	if err != nil {
		impl.logger.Errorw("failed reading data from file", "err", err)
	} else {
		pipelineOverrideByte, err = yaml.YAMLToJSON(pipelineOverrideByte)
		if err != nil {
			return nil, err
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

func (impl ChartTemplateServiceImpl) PackageChart(tempReferenceTemplateDir string, chartMetaData *chart.Metadata) (*string, string, error) {
	valid, err := chartutil.IsChartDir(tempReferenceTemplateDir)
	if err != nil {
		impl.logger.Errorw("error in validating base chart", "dir", tempReferenceTemplateDir, "err", err)
		return nil, "", err
	}
	if !valid {
		impl.logger.Errorw("invalid chart at ", "dir", tempReferenceTemplateDir)
		return nil, "", fmt.Errorf("invalid base chart")
	}

	b, err := yaml.Marshal(chartMetaData)
	if err != nil {
		impl.logger.Errorw("error in marshaling chartMetadata", "err", err)
		return nil, "", err
	}
	err = ioutil.WriteFile(filepath.Join(tempReferenceTemplateDir, "Chart.yaml"), b, 0600)
	if err != nil {
		impl.logger.Errorw("err in writing Chart.yaml", "err", err)
		return nil, "", err
	}
	chart, err := chartutil.LoadDir(tempReferenceTemplateDir)
	if err != nil {
		impl.logger.Errorw("error in loading chart dir", "err", err, "dir", tempReferenceTemplateDir)
		return nil, "", err
	}

	archivePath, err := chartutil.Save(chart, tempReferenceTemplateDir)
	if err != nil {
		impl.logger.Errorw("error in saving", "err", err, "dir", tempReferenceTemplateDir)
		return nil, "", err
	}
	impl.logger.Debugw("chart archive path", "path", archivePath)
	var valuesYaml string
	if chart.Values != nil {
		valuesYaml = chart.Values.Raw
	} else {
		impl.logger.Warnw("values.yaml not found in helm chart", "dir", tempReferenceTemplateDir)
	}
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

// GetByteArrayRefChart this method will be used for getting byte array from reference chart to store in db
func (impl ChartTemplateServiceImpl) GetByteArrayRefChart(chartMetaData *chart.Metadata, referenceTemplatePath string) ([]byte, error) {
	chartMetaData.ApiVersion = "v1" // ensure always v1
	dir := impl.GetDir()
	tempReferenceTemplateDir := filepath.Join(CHART_WORKING_DIR_PATH, dir)
	impl.logger.Debugw("chart dir ", "chart", chartMetaData.Name, "dir", tempReferenceTemplateDir)
	err := os.MkdirAll(tempReferenceTemplateDir, os.ModePerm) //hack for concurrency handling
	if err != nil {
		impl.logger.Errorw("err in creating dir", "dir", tempReferenceTemplateDir, "err", err)
		return nil, err
	}
	defer impl.CleanDir(tempReferenceTemplateDir)
	err = dirCopy.Copy(referenceTemplatePath, tempReferenceTemplateDir)
	if err != nil {
		impl.logger.Errorw("error in copying chart for app", "app", chartMetaData.Name, "error", err)
		return nil, err
	}
	activePath, _, err := impl.PackageChart(tempReferenceTemplateDir, chartMetaData)
	if err != nil {
		impl.logger.Errorw("error in creating archive", "err", err)
		return nil, err
	}
	file, err := os.Open(*activePath)
	reader, err := gzip.NewReader(file)
	if err != nil {
		impl.logger.Errorw("There is a problem with os.Open", "err", err)
		return nil, err
	}
	// read the complete content of the file h.Name into the bs []byte
	bs, err := ioutil.ReadAll(reader)
	if err != nil {
		impl.logger.Errorw("There is a problem with readAll", "err", err)
		return nil, err
	}
	return bs, nil
}

func (impl ChartTemplateServiceImpl) LoadChartInBytes(ChartPath string, deleteChart bool, chartName string, chartVersion string) ([]byte, error) {
	var chartBytesArr []byte
	//this function is removed in latest helm release and is replaced by Loader in loader package
	chart, err := chartutil.LoadDir(ChartPath)
	if err != nil {
		impl.logger.Errorw("error in loading chart dir", "err", err, "dir")
		return chartBytesArr, err
	}

	if len(chartName) > 0 && len(chartVersion) > 0 {
		chart.Metadata.Name = chartName
		chart.Metadata.Version = chartVersion
	}

	chartBytesArr, err = impl.CreateZipFileForChart(chart, ChartPath)
	if err != nil {
		impl.logger.Errorw("error in saving", "err", err, "dir")
		return chartBytesArr, err
	}

	if deleteChart {
		defer impl.CleanDir(ChartPath)
	}

	return chartBytesArr, err
}

func (impl ChartTemplateServiceImpl) LoadChartFromDir(dir string) (*chart.Chart, error) {
	//this function is removed in latest helm release and is replaced by Loader in loader package
	chart, err := chartutil.LoadDir(dir)
	if err != nil {
		impl.logger.Errorw("error in loading chart dir", "err", err, "dir")
		return chart, err
	}
	return chart, nil
}

func (impl ChartTemplateServiceImpl) CreateZipFileForChart(chart *chart.Chart, outputChartPathDir string) ([]byte, error) {
	var chartBytesArr []byte
	chartZipPath, err := chartutil.Save(chart, outputChartPathDir)
	if err != nil {
		impl.logger.Errorw("error in saving", "err", err, "dir")
		return chartBytesArr, err
	}

	chartBytesArr, err = ioutil.ReadFile(chartZipPath)
	if err != nil {
		impl.logger.Errorw("There is a problem with os.Open", "err", err)
		return nil, err
	}
	return chartBytesArr, nil
}

func IsHelmApp(deploymentAppType string) bool {
	return deploymentAppType == PIPELINE_DEPLOYMENT_TYPE_HELM
}

func IsAcdApp(deploymentAppType string) bool {
	return deploymentAppType == PIPELINE_DEPLOYMENT_TYPE_ACD
}

func IsManifestDownload(deploymentAppType string) bool {
	return deploymentAppType == PIPELINE_DEPLOYMENT_TYPE_MANIFEST_DOWNLOAD
}

func IsManifestPush(deploymentAppType string) bool {
	return deploymentAppType == PIPELINE_DEPLOYMENT_TYPE_MANIFEST_PUSH
}

func IsOCIRegistryChartProvider(ociRegistry dockerRegistryRepository.DockerArtifactStore) bool {
	if ociRegistry.OCIRegistryConfig == nil ||
		len(ociRegistry.OCIRegistryConfig) != 1 ||
		!IsOCIConfigChartProvider(ociRegistry.OCIRegistryConfig[0]) {
		return false
	}
	return true
}

func IsOCIConfigChartProvider(ociRegistryConfig *dockerRegistryRepository.OCIRegistryConfig) bool {
	if ociRegistryConfig.RepositoryType == dockerRegistryRepository.OCI_REGISRTY_REPO_TYPE_CHART &&
		(ociRegistryConfig.RepositoryAction == dockerRegistryRepository.STORAGE_ACTION_TYPE_PULL ||
			ociRegistryConfig.RepositoryAction == dockerRegistryRepository.STORAGE_ACTION_TYPE_PULL_AND_PUSH) &&
		ociRegistryConfig.RepositoryList != "" {
		return true
	}
	return false
}
