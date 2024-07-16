package argoApplication

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	"strconv"
	"strings"
)

func DecodeExternalArgoAppId(appId string) (*bean.ArgoAppIdentifier, error) {
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
	return &bean.ArgoAppIdentifier{
		ClusterId: clusterId,
		Namespace: component[1],
		AppName:   component[2],
	}, nil
}
