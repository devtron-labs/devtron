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
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/juju/errors"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

type CDMetrics struct {
	AppName         string
	DeploymentType  string
	Status          string
	EnvironmentName string
	Time            float64
}

type CIMetrics struct {
	CacheDownDuration  float64   `json:"cacheDownDuration"`
	PreCiDuration      float64   `json:"preCiDuration"`
	BuildDuration      float64   `json:"buildDuration"`
	PostCiDuration     float64   `json:"postCiDuration"`
	CacheUpDuration    float64   `json:"cacheUpDuration"`
	TotalDuration      float64   `json:"totalDuration"`
	CacheDownStartTime time.Time `json:"cacheDownStartTime"`
	PreCiStartTime     time.Time `json:"preCiStart"`
	BuildStartTime     time.Time `json:"buildStartTime"`
	PostCiStartTime    time.Time `json:"postCiStartTime"`
	CacheUpStartTime   time.Time `json:"cacheUpStartTime"`
	TotalStartTime     time.Time `json:"totalStartTime"`
}

func ContainsString(list []string, element string) bool {
	if len(list) == 0 {
		return false
	}
	for _, l := range list {
		if l == element {
			return true
		}
	}
	return false
}

func AppendErrorString(errs []string, err error) []string {
	if err != nil {
		errs = append(errs, err.Error())
	}
	return errs
}

func GetErrorOrNil(errs []string) error {
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	return nil
}

func ExtractChartVersion(chartVersion string) (int, int, error) {
	if chartVersion == "" {
		return 0, 0, nil
	}
	chartVersions := strings.Split(chartVersion, ".")
	chartMajorVersion, err := strconv.Atoi(chartVersions[0])
	if err != nil {
		return 0, 0, err
	}
	chartMinorVersion, err := strconv.Atoi(chartVersions[1])
	if err != nil {
		return 0, 0, err
	}
	return chartMajorVersion, chartMinorVersion, nil
}

func ExtractEcrImage(registryId, region, repoName, tag string) string {
	return fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:%s", registryId, region, repoName, tag)
}

type Closer interface {
	Close() error
}

func Close(c Closer, logger *zap.SugaredLogger) {
	if c == nil {
		return
	}
	if err := c.Close(); err != nil {
		logger.Warnf("failed to close %v: %v", c, err)
	}
}

var chars = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// Generates random string
func Generate(size int) string {
	rand.Seed(time.Now().UnixNano())
	var b strings.Builder
	for i := 0; i < size; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()
	return str
}

func HttpRequest(url string) (map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	//var client *http.Client
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		var apiRes map[string]interface{}
		err = json.Unmarshal(resBody, &apiRes)
		if err != nil {
			return nil, err
		}
		return apiRes, err
	}
	return nil, err
}

func CheckForMissingFiles(chartLocation string) error {
	imageDescriptorPath := filepath.Clean(filepath.Join(chartLocation, ".image_descriptor_template.json"))
	if _, err := os.Stat(imageDescriptorPath); os.IsNotExist(err) {
		return errors.New(".image_descriptor_template.json file not present in the directory")
	}
	return nil
}

func ExtractTarGz(gzipStream io.Reader, chartDir string) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)
	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(filepath.Join(chartDir, header.Name)); os.IsNotExist(err) {
				if err := os.MkdirAll(filepath.Join(chartDir, header.Name), 0755); err != nil {
					return err
				}
			} else {
				break
			}

		case tar.TypeReg:
			outFile, err := os.Create(filepath.Join(chartDir, header.Name))
			if err != nil {
				dirName := filepath.Dir(header.Name)
				if _, err1 := os.Stat(filepath.Join(chartDir, dirName)); os.IsNotExist(err1) {
					if err1 = os.MkdirAll(filepath.Join(chartDir, dirName), 0755); err1 != nil {
						return err1
					}
					outFile, err = os.Create(filepath.Join(chartDir, header.Name))
					if err != nil {
						return err
					}
				}
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			outFile.Close()

		default:
			return err

		}

	}
	return nil
}

func BuildDevtronBomUrl(bomUrl string, version string) string {
	return fmt.Sprintf(bomUrl, version)
}

// InterfaceToMapAdapter it will convert any golang struct into map
func InterfaceToMapAdapter(resp interface{}) map[string]interface{} {
	var dat map[string]interface{}
	b, err := json.Marshal(resp)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return dat
	}
	if err := json.Unmarshal(b, &dat); err != nil {
		fmt.Printf("Error: %s", err)
		return dat
	}
	return dat
}

func TriggerCDMetrics(wfr CDMetrics, exposeCDMetrics bool) {
	if exposeCDMetrics && (wfr.Status == WorkflowFailed || wfr.Status == WorkflowSucceeded) {
		middleware.CdDuration.WithLabelValues(wfr.AppName, wfr.Status, wfr.EnvironmentName, wfr.DeploymentType).Observe(wfr.Time)
	}
}

func TriggerCIMetrics(Metrics CIMetrics, exposeCIMetrics bool, PipelineName string, AppName string) {
	if exposeCIMetrics {
		middleware.CacheDownloadDuration.WithLabelValues(PipelineName, AppName).Observe(Metrics.CacheDownDuration)
		middleware.CiDuration.WithLabelValues(PipelineName, AppName).Observe(Metrics.TotalDuration)
		if Metrics.CacheUpDuration != 0 {
			middleware.CacheUploadDuration.WithLabelValues(PipelineName, AppName).Observe(Metrics.CacheUpDuration)
		}
		if Metrics.PostCiDuration != 0 {
			middleware.PostCiDuration.WithLabelValues(PipelineName, AppName).Observe(Metrics.PostCiDuration)
		}
		if Metrics.PreCiDuration != 0 {
			middleware.PreCiDuration.WithLabelValues(PipelineName, AppName).Observe(Metrics.PreCiDuration)
		}
		middleware.BuildDuration.WithLabelValues(PipelineName, AppName).Observe(Metrics.BuildDuration)
	}
}

func TriggerGitOpsMetrics(operation string, method string, startTime time.Time, err error) {
	status := "Success"
	if err != nil {
		status = "Failed"
	}
	middleware.GitOpsDuration.WithLabelValues(operation, method, status).Observe(time.Since(startTime).Seconds())
}

func InterfaceToString(resp interface{}) string {
	var dat string
	b, err := json.Marshal(resp)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return dat
	}
	if err := json.Unmarshal(b, &dat); err != nil {
		fmt.Printf("Error: %s", err)
		return dat
	}
	return dat
}

func InterfaceToFloat(resp interface{}) float64 {
	var dat float64
	b, err := json.Marshal(resp)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return dat
	}
	if err := json.Unmarshal(b, &dat); err != nil {
		fmt.Printf("Error: %s", err)
		return dat
	}
	return dat
}

type HpaResourceRequest struct {
	ResourceName    string
	ReqReplicaCount float64
	ReqMaxReplicas  float64
	ReqMinReplicas  float64
	IsEnable        bool
	Group           string
	Version         string
	Kind            string
}

func ConvertStringSliceToMap(inputs []string) map[string]bool {
	m := make(map[string]bool, len(inputs))
	for _, input := range inputs {
		m[input] = true
	}
	return m
}

func MatchRegexExpression(exp string, text string) (bool, error) {
	rExp, err := regexp.Compile(exp)
	if err != nil {
		return false, err
	}
	matched := rExp.Match([]byte(text))
	return matched, nil
}

func GetLatestImageAccToImagePushedAt(imageDetails []types.ImageDetail) types.ImageDetail {
	sort.Slice(imageDetails, func(i, j int) bool {
		return imageDetails[i].ImagePushedAt.After(*imageDetails[j].ImagePushedAt)
	})
	return imageDetails[0]
}

func GetReverseSortedImageDetails(imageDetails []types.ImageDetail) []types.ImageDetail {
	sort.Slice(imageDetails, func(i, j int) bool {
		return imageDetails[i].ImagePushedAt.Before(*imageDetails[j].ImagePushedAt)
	})
	return imageDetails
}
