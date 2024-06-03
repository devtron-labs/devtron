/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package cluster

import (
	"fmt"
	"strings"
)

func BuildEnvironmentName(clusterName string, namespace string) string {
	// Here we are replacing the (_) with (-) in clusterName as we don't support (_) in environment Name
	clusterName = strings.ReplaceAll(clusterName, "_", "-")
	return fmt.Sprintf("%s--%s", clusterName, namespace)
}
