/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package util

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"k8s.io/helm/pkg/tlsutil"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"time"
)

func ReadFromUrlWithRetry(url string) ([]byte, error) {
	var (
		err      error
		response *http.Response
		retries  = 3
	)

	for retries > 0 {
		response, err = http.Get(url)
		if err != nil {
			retries -= 1
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	if response != nil {
		defer response.Body.Close()
		statusCode := response.StatusCode
		if statusCode != http.StatusOK {
			return nil, errors.New(fmt.Sprintf("Error in getting content from url - %s. Status code : %s", url, strconv.Itoa(statusCode)))
		}
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
	return nil, err
}

func GetHost(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err == nil {
		return u.Host, nil
	}
	u, err = url.Parse("//" + urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}
	return u.Host, nil
}

func GetTlsConfig(TLSKey, TLSCert, CACert, folderPath string) (*tls.Config, error) {
	var tlsKeyFileName, tlsCertFileName, caCertFileName string
	var tlsKeyFilePath, tlsCertFilePath, caCertFilePath string

	defer func() {
		err := DeleteFile(tlsKeyFilePath)
		if err != nil {
			fmt.Printf("error in deleting tls file %s", err)
		}
		err = DeleteFile(tlsCertFilePath)
		if err != nil {
			fmt.Printf("error in deleting tls file %s", err)
		}
		err = DeleteFile(caCertFilePath)
		if err != nil {
			fmt.Printf("error in deleting tls file %s", err)
		}
	}()

	var err error
	if len(TLSKey) > 0 && len(TLSCert) > 0 {
		tlsKeyFileName = getTLSKeyFileName()
		tlsKeyFilePath, err = CreateFileWithData(folderPath, tlsKeyFileName, TLSKey)
		if err != nil {
			fmt.Printf("error in creating tls key file %s", err)
			return nil, err
		}
		tlsCertFileName = getCertFileName()
		tlsCertFilePath, err = CreateFileWithData(folderPath, tlsCertFileName, TLSCert)
		if err != nil {
			fmt.Printf("error in creating tls cert file %s", err)
			return nil, err
		}
	}
	if len(caCertFileName) > 0 {
		caCertFileName = getCertFileName()
		caCertFilePath, err = CreateFileWithData(folderPath, caCertFileName, CACert)
		if err != nil {
			fmt.Printf("error in creating caCert file %s", err)
			return nil, err
		}
	}
	tlsConfig, err := tlsutil.NewClientTLS(caCertFilePath, tlsKeyFilePath, tlsCertFilePath)
	if err != nil {
		fmt.Printf("error in creating tls config %s ", err)
	}
	return tlsConfig, nil
}

func GetHTTPClientWithTLSConfig(tlsConfig *tls.Config) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ForceAttemptHTTP2:     true,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		},
	}
}

func getTLSKeyFileName() string {
	randomName := fmt.Sprintf("%v.key", GetRandomName())
	return randomName
}

func getCertFileName() string {
	randomName := fmt.Sprintf("%v.crt", GetRandomName())
	return randomName
}
