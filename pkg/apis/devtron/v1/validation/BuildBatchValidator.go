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

package validation

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/apis/devtron/v1"
	"github.com/devtron-labs/devtron/util"
	"strings"
)

var validateBuildFunc = []func(build *v1.Build) error{validateBuildVersion, validateBuildClone}
var validBuildVersions = []string{"app/v1"}

func ValidateBuild(build *v1.Build) error {
	if len(build.GetOperation()) == 0 {
		return fmt.Errorf(v1.OperationUndefinedError, "build")
	}
	if len(build.ApiVersion) == 0 || !util.ContainsString(validBuildVersions, build.ApiVersion) {
		return fmt.Errorf(v1.UnsupportedVersion, build.ApiVersion, "build")
	}
	errs := make([]string, 0)
	for _, f := range validateBuildFunc {
		err := f(build)
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	return nil
}

func validateBuildVersion(build *v1.Build) error {
	if len(build.ApiVersion) == 0 || !util.ContainsString(validDeploymentVersions, build.ApiVersion) {
		return fmt.Errorf(v1.UnsupportedVersion, build.ApiVersion, "build")
	}
	return nil
}

func validateBuildClone(build *v1.Build) error {
	if build.GetOperation() != v1.Clone {
		return nil
	}

	return fmt.Errorf(v1.OperationUnimplementedError, "clone", "build")
}
