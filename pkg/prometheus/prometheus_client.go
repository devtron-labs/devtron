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

package prometheus

import (
	"fmt"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
)

var prometheusAPI v1.API
var prometheusAPIMap map[string]v1.API

func Context(prometheusUrl string) (v1.API, error) {

	if prometheusAPI != nil {
		return prometheusAPI, nil
	}

	client, err := api.NewClient(api.Config{Address: prometheusUrl})
	if err != nil {
		fmt.Println("Error creating client")
		fmt.Println(err)
		return nil, err
	}
	prometheusAPI = v1.NewAPI(client)

	return prometheusAPI, nil
}

func ContextByEnv(env string, prometheusUrl string) (v1.API, error) {
	if prometheusAPIMap == nil {
		prometheusAPIMap = make(map[string]v1.API)
	}
	if _, ok := prometheusAPIMap[env]; ok {
		prometheusAPI = prometheusAPIMap[env]
	} else {
		client, err := api.NewClient(api.Config{Address: prometheusUrl})
		if err != nil {
			fmt.Println("Error creating client")
			fmt.Println(err)
			return nil, err
		}
		prometheusAPI = v1.NewAPI(client)
		prometheusAPIMap[env] = prometheusAPI
	}
	return prometheusAPI, nil
}
