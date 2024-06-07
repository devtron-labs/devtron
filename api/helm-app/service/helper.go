/*
 * Copyright (c) 2024. Devtron Inc.
 */

package service

import (
	"fmt"
	"strconv"
	"strings"
)

func DecodeExternalAppAppId(appId string) (*AppIdentifier, error) {
	component := strings.Split(appId, "|")
	if len(component) != 3 {
		return nil, fmt.Errorf("malformed app id %s", appId)
	}
	clusterId, err := strconv.Atoi(component[0])
	if err != nil {
		return nil, err
	}
	if clusterId <= 0 {
		return nil, fmt.Errorf("target cluster is not provided")
	}
	return &AppIdentifier{
		ClusterId:   clusterId,
		Namespace:   component[1],
		ReleaseName: component[2],
	}, nil
}
