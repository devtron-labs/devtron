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

package util

import (
	"encoding/base64"
	"encoding/json"
)

type SecretTransformMode int

const (
	EncodeSecret SecretTransformMode = 1
	DecodeSecret SecretTransformMode = 2
)

func GetDecodedAndEncodedData(data json.RawMessage, transformer SecretTransformMode) ([]byte, error) {
	dataMap := make(map[string]string)
	err := json.Unmarshal(data, &dataMap)
	if err != nil {
		return nil, err
	}
	var transformedData []byte
	for key, value := range dataMap {
		switch transformer {
		case EncodeSecret:
			transformedData = []byte(base64.StdEncoding.EncodeToString([]byte(value)))
		case DecodeSecret:
			transformedData, err = base64.StdEncoding.DecodeString(value)
			if err != nil {
				return nil, err
			}
		}
		dataMap[key] = string(transformedData)
	}
	marshal, err := json.Marshal(dataMap)
	if err != nil {
		return nil, err
	}
	return marshal, nil
}
