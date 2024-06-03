/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package validation

import (
	"fmt"
	"strings"

	v1 "github.com/devtron-labs/devtron/pkg/apis/devtron/v1"
	"github.com/devtron-labs/devtron/util"
)

var validateBuildFunc = []func(build *v1.Build) error{validateBuildVersion, validateBuildClone}
var validBuildVersions = []string{"app/v1"}

func ValidateBuild(build *v1.Build) error {
	if len(build.GetOperation()) == 0 {
		return fmt.Errorf(v1.OperationUndefinedError, "build")
	}
	if build.ApiVersion == "" || !util.ContainsString(validBuildVersions, build.ApiVersion) {
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
	if build.ApiVersion == "" || !util.ContainsString(validDeploymentVersions, build.ApiVersion) {
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
