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
	"os"
	"strconv"
	"strings"
)

const RELEASE_NOT_EXIST = "release not exist"
const NOT_FOUND = "not found"
const PermissionDenied = "permission denied"

func MoveFileToDestination(filePath, destinationPath string) error {
	err := os.Rename(filePath, destinationPath)
	if err != nil {
		return err
	}
	return nil
}

func CreateFileAtFilePathAndWrite(filePath, fileContent string) (string, error) {
	file, err := os.Create(filePath)
	defer file.Close()
	if err != nil {
		return filePath, err
	}
	_, err = file.Write([]byte(fileContent))
	if err != nil {
		return filePath, err
	}
	return filePath, err
}

func ConvertIntArrayToStringArray(req []int) []string {
	var resp []string
	for _, item := range req {
		resp = append(resp, strconv.Itoa(item))
	}
	return resp
}

func CheckAppReleaseNotExist(err error) bool {
	// RELEASE_NOT_EXIST check for helm App and NOT_FOUND check for argo app
	return strings.Contains(err.Error(), NOT_FOUND) || strings.Contains(err.Error(), RELEASE_NOT_EXIST)
}

func CheckPermissionErrorForArgoCd(err error) bool {
	return strings.Contains(err.Error(), PermissionDenied)
}

func IsExternalChartStoreApp(displayName string) bool {
	return len(displayName) > 0
}
