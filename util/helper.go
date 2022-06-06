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
	"github.com/juju/errors"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

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

type Closer interface {
	Close() error
}

func Close(c Closer, logger *zap.SugaredLogger) {
	if err := c.Close(); err != nil {
		logger.Warnf("failed to close %v: %v", c, err)
	}
}

var chars = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

//Generates random string
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
				if err := os.Mkdir(filepath.Join(chartDir, header.Name), 0755); err != nil {
					return err
				}
			} else {
				break
			}

		case tar.TypeReg:
			outFile, err := os.Create(filepath.Join(chartDir, header.Name))
			if err != nil {
				return err
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
