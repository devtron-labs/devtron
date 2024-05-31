/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
