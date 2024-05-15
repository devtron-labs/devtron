package helper

import (
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"net/http"
	"strings"
	"time"
)

func GetOverviewTags(objectData string) map[string]string {
	tagsResults := gjson.Get(objectData, bean.ResourceObjectTagsPath).Map()
	tags := make(map[string]string, len(tagsResults))
	for key, value := range tagsResults {
		tags[key] = value.String()
	}
	return tags
}

func GetCreatedOnTime(objectData string) (createdOn time.Time, err error) {
	createdOnStr := gjson.Get(objectData, bean.ResourceObjectCreatedOnPath).String()
	if len(createdOnStr) != 0 {
		createdOn, err = time.Parse(time.RFC3339, createdOnStr)
		if err != nil {
			return createdOn, err
		}
	}
	return createdOn, nil
}

func PatchResourceObjectDataAtAPath(objectData string, path string, value interface{}) (string, error) {
	return sjson.Set(objectData, path, value)
}

func UpdateKindAndSubKindParentConfig(parentConfig *bean.ResourceIdentifier) error {
	kind, subKind, err := GetKindAndSubKindFrom(parentConfig.ResourceKind.ToString())
	if err != nil {
		return err
	}
	parentConfig.ResourceKind = bean.DevtronResourceKind(kind)
	parentConfig.ResourceSubKind = bean.DevtronResourceKind(subKind)
	return nil
}

func DecodeFilterCriteriaString(criteria string) (*bean.FilterCriteriaDecoder, error) {
	objs := strings.Split(criteria, "|")
	if len(objs) != 3 {
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidFilterCriteria, bean.InvalidFilterCriteria)
	}
	criteriaDecoder := adapter.BuildFilterCriteriaDecoder(objs[0], objs[1], objs[2])
	return criteriaDecoder, nil
}

func DecodeSearchKeyString(searchKey string) (*bean.SearchCriteriaDecoder, error) {
	objs := strings.Split(searchKey, "|")
	if len(objs) != 2 {
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidSearchKey, bean.InvalidSearchKey)
	}
	searchDecoder := adapter.BuildSearchCriteriaDecoder(objs[0], objs[1])
	return searchDecoder, nil
}

func DecodeFiltersForDeployAndRolloutStatus(filters []string) ([]int, []string, []int, []string, map[bean2.WorkflowType][]string, []string, error) {
	filterLen := len(filters)
	appIdsFilters := make([]int, 0, filterLen)
	appIdentifierFilters := make([]string, 0, filterLen)
	envIdsFilters := make([]int, 0, filterLen)
	envIdentifierFilters := make([]string, 0, filterLen)
	deploymentStatus := make(map[bean2.WorkflowType][]string, filterLen)
	rolloutStatus := make([]string, 0, filterLen)

	for _, filter := range filters {
		objs := strings.Split(filter, "|")
		if len(objs) == 0 {
			return nil, nil, nil, nil, nil, nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidFilterCriteria, bean.InvalidFilterCriteria)
		}
		switch objs[0] {
		case bean.DevtronApplicationFilter.ToString():
			{
				if len(objs) != 3 {
					return nil, nil, nil, nil, nil, nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", fmt.Sprintf("%s:%s", bean.InvalidFilterCriteria, bean.DevtronApplicationFilter), fmt.Sprintf("%s:%s", bean.InvalidFilterCriteria, bean.DevtronApplicationFilter))
				}
				ids, identifiers, err := getIdsAndIdentifierBasedOnType(objs)
				if err != nil {
					return appIdsFilters, appIdentifierFilters, envIdsFilters, envIdentifierFilters, deploymentStatus, rolloutStatus, err
				}
				appIdsFilters = append(appIdsFilters, ids...)
				appIdentifierFilters = append(appIdentifierFilters, identifiers...)
			}
		case bean.EnvironmentFilter.ToString():
			{
				if len(objs) != 3 {
					return nil, nil, nil, nil, nil, nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", fmt.Sprintf("%s:%s", bean.InvalidFilterCriteria, bean.EnvironmentFilter), fmt.Sprintf("%s:%s", bean.InvalidFilterCriteria, bean.EnvironmentFilter))
				}
				ids, identifiers, err := getIdsAndIdentifierBasedOnType(objs)
				if err != nil {
					return appIdsFilters, appIdentifierFilters, envIdsFilters, envIdentifierFilters, deploymentStatus, rolloutStatus, err
				}
				envIdsFilters = append(envIdsFilters, ids...)
				envIdentifierFilters = append(envIdentifierFilters, identifiers...)
			}

		case bean.DeploymentStatusFilter.ToString():
			{
				if len(objs) != 3 {
					return nil, nil, nil, nil, nil, nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", fmt.Sprintf("%s:%s", bean.InvalidFilterCriteria, bean.DeploymentStatusFilter), fmt.Sprintf("%s:%s", bean.InvalidFilterCriteria, bean.DeploymentStatusFilter))
				}
				statuses := strings.Split(objs[2], ",")
				deploymentStatus[bean2.WorkflowType(objs[1])] = append(deploymentStatus[bean2.WorkflowType(objs[1])], statuses...)
			}
		case bean.ReleaseDeploymentStatusFilter.ToString():
			{
				if len(objs) != 2 {
					return nil, nil, nil, nil, nil, nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", fmt.Sprintf("%s:%s", bean.InvalidFilterCriteria, bean.ReleaseDeploymentStatusFilter), fmt.Sprintf("%s:%s", bean.InvalidFilterCriteria, bean.ReleaseDeploymentStatusFilter))
				}
				statuses := strings.Split(objs[1], ",")
				rolloutStatus = append(rolloutStatus, statuses...)
			}
		default:
			return appIdsFilters, appIdentifierFilters, envIdsFilters, envIdentifierFilters, deploymentStatus, rolloutStatus, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidFilterCriteria, bean.InvalidFilterCriteria)
		}
	}
	return appIdsFilters, appIdentifierFilters, envIdsFilters, envIdentifierFilters, deploymentStatus, rolloutStatus, nil
}

func getIdsAndIdentifierBasedOnType(objs []string) ([]int, []string, error) {
	ids := make([]int, 0)
	identifiers := make([]string, 0)
	if objs[1] == bean.Identifier.ToString() {
		identifiers = strings.Split(objs[2], ",")
		return ids, identifiers, nil
	} else if objs[1] == bean.Id.ToString() {
		ids, err := util2.ConvertStringSliceToIntSlice(objs[2])
		if err != nil {
			return ids, identifiers, util.GetApiErrorAdapter(http.StatusBadRequest, "400", fmt.Sprintf("%s:%s", bean.InvalidFilterCriteria, err.Error()), fmt.Sprintf("%s:%s", bean.InvalidFilterCriteria, err.Error()))
		}
		return ids, identifiers, nil
	}
	return ids, identifiers, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidFilterCriteria, bean.InvalidFilterCriteria)

}
