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

package telemetry

import (
	"encoding/base64"
	"encoding/json"
	"github.com/caarlos0/env"
	"github.com/patrickmn/go-cache"
	"github.com/posthog/posthog-go"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"time"
)

type PosthogClient struct {
	Client posthog.Client
	cache  *cache.Cache
}

type PosthogConfig struct {
	PosthogApiKey           string `env:"POSTHOG_API_KEY" envDefault:""`
	PosthogEndpoint         string `env:"POSTHOG_ENDPOINT" envDefault:"https://app.posthog.com"`
	SummaryInterval         int    `env:"SUMMARY_INTERVAL" envDefault:"24"`
	HeartbeatInterval       int    `env:"HEARTBEAT_INTERVAL" envDefault:"3"`
	CacheExpiry             int    `env:"CACHE_EXPIRY" envDefault:"120"`
	TelemetryApiKeyEndpoint string `env:"TELEMETRY_API_KEY_ENDPOINT" envDefault:"aHR0cHM6Ly90ZWxlbWV0cnkuZGV2dHJvbi5haS9kZXZ0cm9uL3RlbGVtZXRyeS9hcGlrZXk="`
}

func GetPosthogConfig() (*PosthogConfig, error) {
	cfg := &PosthogConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, err
}

func NewPosthogClient(logger *zap.SugaredLogger) (*PosthogClient, error) {
	cfg := &PosthogConfig{}
	err := env.Parse(cfg)
	if err != nil {
		logger.Errorw("exception caught while parsing posthog config", "err", err)
		//return &PosthogClient{}, err
	}
	if len(cfg.PosthogApiKey) == 0 {
		apiKey, err := getPosthogApiKey(cfg.TelemetryApiKeyEndpoint)
		if err != nil {
			logger.Errorw("exception caught while getting api key", "err", err)
			//return &PosthogClient{}, err
		}
		cfg.PosthogApiKey = apiKey
	}
	client, _ := posthog.NewWithConfig(cfg.PosthogApiKey, posthog.Config{Endpoint: cfg.PosthogEndpoint})
	//defer client.Close()
	d := time.Duration(cfg.CacheExpiry)
	c := cache.New(d*time.Minute, 240*time.Minute)
	pgClient := &PosthogClient{
		Client: client,
		cache:  c,
	}
	return pgClient, nil
}

func getPosthogApiKey(encodedPosthogApiKeyUrl string) (string, error) {
	dncodedPosthogApiKeyUrl, err := base64.StdEncoding.DecodeString(encodedPosthogApiKeyUrl)
	if err != nil {
		return "", err
	}
	apiKeyUrl := string(dncodedPosthogApiKeyUrl)
	req, err := http.NewRequest(http.MethodGet, apiKeyUrl, nil)
	if err != nil {
		return "", err
	}
	//var client *http.Client
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		var apiRes map[string]interface{}
		err = json.Unmarshal(resBody, &apiRes)
		if err != nil {
			return "", err
		}
		encodedApiKey := apiRes["result"].(string)
		apiKey, err := base64.StdEncoding.DecodeString(encodedApiKey)
		if err != nil {
			return "", err
		}
		return string(apiKey), err
	}
	return "", err
}
