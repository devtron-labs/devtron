package fluxApplication

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/fluxApplication/bean"
	"strconv"
	"strings"
)

func DecodeFluxExternalAppAppId(appId string) (*bean.FluxAppIdentifier, error) {
	component := strings.Split(appId, "|")
	if len(component) != 4 {
		return nil, fmt.Errorf("malformed app id %s", appId)
	}
	clusterId, err := strconv.Atoi(component[0])
	if err != nil {
		return nil, err
	}
	isKustomizeApp, err := strconv.ParseBool(component[3])
	if err != nil {
		return nil, err
	}
	if clusterId <= 0 {
		return nil, fmt.Errorf("target cluster is not provided")
	}
	return &bean.FluxAppIdentifier{
		ClusterId:      clusterId,
		Namespace:      component[1],
		Name:           component[2],
		IsKustomizeApp: isKustomizeApp,
	}, nil
}
