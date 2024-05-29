package service

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/helm-app/service/bean"
	"strconv"
	"strings"
)

func DecodeExternalAppAppId(appId string) (*bean.AppIdentifier, error) {
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
	return &bean.AppIdentifier{
		ClusterId:   clusterId,
		Namespace:   component[1],
		ReleaseName: component[2],
	}, nil
}
