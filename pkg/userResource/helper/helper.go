package helper

import (
	"fmt"
	"net/http"
	"strings"

	apiBean "github.com/devtron-labs/devtron/api/userResource/bean"
	"github.com/devtron-labs/devtron/internal/util"
	bean5 "github.com/devtron-labs/devtron/pkg/userResource/bean"
)

func ValidateResourceOptionReqBean(reqBean *apiBean.ResourceOptionsReqDto) error {
	if reqBean == nil {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean5.InvalidPayloadMessage, bean5.InvalidPayloadMessage)
	}
	if len(reqBean.EntityAccessType.Entity) == 0 {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean5.InvalidEntityMessage, bean5.InvalidEntityMessage)
	}
	return nil
}

func GetValidatedClusterIds(reqBean *apiBean.ResourceOptionsReqDto) ([]int, error) {
	invalid := util.GetApiErrorAdapter(http.StatusBadRequest, "400",
		bean5.InvalidClusterIdMessage, bean5.InvalidClusterIdMessage)
	if reqBean == nil {
		return nil, invalid
	}
	seen := make(map[int]bool, len(reqBean.ClusterIds)+1)
	clusterIds := make([]int, 0, len(reqBean.ClusterIds)+1)
	appendIfValid := func(clusterId int) {
		if clusterId > 0 && !seen[clusterId] {
			seen[clusterId] = true
			clusterIds = append(clusterIds, clusterId)
		}
	}
	for _, clusterId := range reqBean.ClusterIds {
		appendIfValid(clusterId)
	}
	if reqBean.ResourceRequestBean != nil {
		appendIfValid(reqBean.ClusterId)
	}
	if len(clusterIds) == 0 {
		return nil, invalid
	}
	return clusterIds, nil
}

func FilterExternalGitOpsAppsByEnvIdentifier(apps []*bean5.ExternalGitOpsAppDto, envIdentifiers []string) []*bean5.ExternalGitOpsAppDto {
	if len(envIdentifiers) == 0 {
		return apps
	}
	selected := make(map[string]bool, len(envIdentifiers))
	for _, identifier := range envIdentifiers {
		selected[strings.ToLower(identifier)] = true
	}
	filtered := make([]*bean5.ExternalGitOpsAppDto, 0, len(apps))
	for _, app := range apps {
		if app == nil {
			continue
		}
		identifier := fmt.Sprintf("%s__%s", app.ClusterName, app.Namespace)
		if selected[strings.ToLower(identifier)] {
			filtered = append(filtered, app)
		}
	}
	return filtered
}
