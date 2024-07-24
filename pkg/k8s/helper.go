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

package k8s

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"strings"
)

func IsResourceNotFoundErr(err error) bool {
	return k8sErrors.IsNotFound(err)
}

func IsBadRequestErr(err error) bool {
	return k8sErrors.IsBadRequest(err)
}

func IsServerTimeoutErr(err error) bool {
	return k8sErrors.IsServerTimeout(err)
}

func GetClientErrorMessage(err error) string {
	if status, ok := err.(k8sErrors.APIStatus); ok || errors.As(err, &status) {
		return status.Status().Message
	}
	return err.Error()
}

// StripPrereleaseFromK8sVersion takes in k8sVersion and stripe pre-release from semver version and return sanitized k8sVersion
// or error if invalid version provided, e.g. if k8sVersion = "1.25.16-eks-b9c9ed7", then it returns "1.25.16".
func StripPrereleaseFromK8sVersion(k8sVersion string) string {
	version, err := semver.NewVersion(k8sVersion)
	if err != nil {
		fmt.Printf("error in stripping pre-release from k8sServerVersion due to invalid k8sServerVersion:= %s, err:= %v", k8sVersion, err)
		return k8sVersion
	}
	if len(version.Prerelease()) > 0 {
		stringToReplace := "-" + version.Prerelease()
		k8sVersion = strings.Replace(k8sVersion, stringToReplace, "", 1)
	}
	return k8sVersion
}
