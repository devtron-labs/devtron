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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	u "net/url"
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

// DoHttpPOSTRequest only handles post request, todo: can be made generic for all types
func DoHttpPOSTRequest(url string, queryParams map[string]string, payload interface{}) (bool, error) {
	client := http.Client{}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		klog.Errorln("error while marshaling event request ", "err", err)
		return false, err
	}

	if len(queryParams) > 0 {
		params := u.Values{}
		for key, val := range queryParams {
			params.Set(key, val)
		}
		url = fmt.Sprintf("%s?%s", url, params.Encode())
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		klog.Errorln("error while writing event", "err", err)
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		klog.Errorln("error in sending notification rest request ", "err", err)
		return false, err
	}
	klog.Infof("notification response %s", resp.Status)
	defer resp.Body.Close()
	return true, err
}
