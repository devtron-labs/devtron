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

package jira

import (
	b64 "encoding/base64"
	"errors"
	"regexp"
)

func GetEncryptedAuthParams(userName string, userToken string) string {
	authParams := userName + ":" + userToken
	authParamsEnc := b64Encode(authParams)
	return authParamsEnc
}

func b64Encode(token string) string {
	authParamsEnc := b64.StdEncoding.EncodeToString([]byte(token))
	return authParamsEnc
}

func ExtractRegex(regex string, message string) ([]string, error) {
	// For regexp tasks you'll need to `Compile` an optimized `Regexp` struct.
	r := regexp.MustCompile(regex)
	matches := r.FindAllString(message, -1)

	if len(matches) == 0 {
		return nil, errors.New("no matches found in the branch name")
	}

	return matches, nil
}
