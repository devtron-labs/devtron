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

package providers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/devtron-labs/common-lib/cloud-provider-identifier/bean"
	"go.uber.org/zap"
)

type digitalOceanMetadataResponse struct {
	DropletID int `json:"droplet_id"`
}

type IdentifyDigitalOcean struct {
	Logger *zap.SugaredLogger
}

func (impl *IdentifyDigitalOcean) Identify() (string, error) {
	data, err := os.ReadFile(bean.DigitalOceanSysFile)
	if err != nil {
		impl.Logger.Errorw("error while reading file", "error", err)
		return bean.Unknown, err
	}
	if strings.Contains(string(data), bean.DigitalOceanIdentifierString) {
		return bean.DigitalOcean, nil
	}
	return bean.Unknown, nil
}

func (impl *IdentifyDigitalOcean) IdentifyViaMetadataServer(detected chan<- string) {
	r := digitalOceanMetadataResponse{}
	req, err := http.NewRequest("GET", bean.DigitalOceanMetadataServer, nil)
	if err != nil {
		impl.Logger.Errorw("error while creating new request", "error", err)
		detected <- bean.Unknown
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		impl.Logger.Errorw("error while requesting", "error", err, "request", req)
		detected <- bean.Unknown
		return
	}
	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			impl.Logger.Errorw("error while reading response body", "error", err, "respBody", resp.Body)
			detected <- bean.Unknown
			return
		}
		err = json.Unmarshal(body, &r)
		if err != nil {
			impl.Logger.Errorw("error while unmarshaling json", "error", err, "body", body)
			detected <- bean.Unknown
			return
		}
		if r.DropletID > 0 {
			detected <- bean.DigitalOcean
			return
		}
	} else {
		detected <- bean.Unknown
		return
	}
}
