/*
 * Copyright (c) 2024. Devtron Inc.
 */

package k8s

import (
	"fmt"
	"github.com/Masterminds/semver"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func IsResourceNotFoundErr(err error) bool {
	if errStatus, ok := err.(*k8sErrors.StatusError); ok && errStatus.Status().Reason == metav1.StatusReasonNotFound {
		return true
	}
	return false
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
