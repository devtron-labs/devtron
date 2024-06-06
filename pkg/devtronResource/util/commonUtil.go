package util

import (
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	bean3 "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	util2 "github.com/devtron-labs/devtron/util"
	"golang.org/x/exp/slices"
	"net/http"
	"strconv"
	"strings"
)

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

func DecodeTaskRunSourceIdentifier(identifier string) (*bean.TaskRunIdentifier, error) {
	splits := strings.Split(identifier, "|")
	if len(splits) < 4 {
		return nil, fmt.Errorf("error in getting task run source identifier")
	}
	id, err := strconv.Atoi(splits[0])
	if err != nil {
		return nil, fmt.Errorf("error in getting task run source identifier")
	}
	dtResourceId, err := strconv.Atoi(splits[2])
	if err != nil {
		return nil, fmt.Errorf("error in getting task run source identifier")
	}
	dtResourceSchemaId, err := strconv.Atoi(splits[3])
	if err != nil {
		return nil, fmt.Errorf("error in getting task run source identifier")
	}
	return &bean.TaskRunIdentifier{
		Id:                 id,
		IdType:             bean.IdType(splits[1]),
		DtResourceId:       dtResourceId,
		DtResourceSchemaId: dtResourceSchemaId,
	}, nil
}

func DecodeFiltersForDeployAndRolloutStatus(filters []string) ([]int, []string, []int, []string, map[bean2.WorkflowType][]string, []string, error) {
	filterLen := len(filters)
	appIdsFilters := make([]int, 0, filterLen)
	appIdentifierFilters := make([]string, 0, filterLen)
	envIdsFilters := make([]int, 0, filterLen)
	envIdentifierFilters := make([]string, 0, filterLen)
	stageWiseDeploymentStatus := make(map[bean2.WorkflowType][]string, filterLen)
	releaseDeploymentStatus := make([]string, 0, filterLen)

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
					return appIdsFilters, appIdentifierFilters, envIdsFilters, envIdentifierFilters, stageWiseDeploymentStatus, releaseDeploymentStatus, err
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
					return appIdsFilters, appIdentifierFilters, envIdsFilters, envIdentifierFilters, stageWiseDeploymentStatus, releaseDeploymentStatus, err
				}
				envIdsFilters = append(envIdsFilters, ids...)
				envIdentifierFilters = append(envIdentifierFilters, identifiers...)
			}

		case bean.StageWiseDeploymentStatusFilter.ToString():
			{
				if len(objs) != 3 {
					return nil, nil, nil, nil, nil, nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", fmt.Sprintf("%s:%s", bean.InvalidFilterCriteria, bean.StageWiseDeploymentStatusFilter), fmt.Sprintf("%s:%s", bean.InvalidFilterCriteria, bean.StageWiseDeploymentStatusFilter))
				}
				statuses := strings.Split(objs[2], ",")
				// doing this for others and fall back cases, others signifies missing and unknown and unable to fetch, and timedOut
				if slices.Contains(statuses, bean.Others) {
					statuses = append(statuses, bean.Missing, bean.Unknown, pipelineConfig.WorkflowUnableToFetchState, pipelineConfig.WorkflowTimedOut)
				}
				if slices.Contains(statuses, pipelineConfig.WorkflowInProgress) {
					statuses = append(statuses, pipelineConfig.WorkflowStarting, bean.RunningStatus, pipelineConfig.WorkflowInitiated)
				}
				if slices.Contains(statuses, pipelineConfig.WorkflowFailed) {
					statuses = append(statuses, pipelineConfig.WorkflowAborted, pipelineConfig.WorkflowTimedOut, bean3.Degraded, bean.Error, executors.WorkflowCancel)
				}
				if slices.Contains(statuses, pipelineConfig.WorkflowSucceeded) {
					statuses = append(statuses, bean3.Healthy)
				}
				stageWiseDeploymentStatus[bean2.WorkflowType(objs[1])] = append(stageWiseDeploymentStatus[bean2.WorkflowType(objs[1])], statuses...)
			}
		case bean.ReleaseDeploymentRolloutStatusFilter.ToString():
			{
				if len(objs) != 2 {
					return nil, nil, nil, nil, nil, nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", fmt.Sprintf("%s:%s", bean.InvalidFilterCriteria, bean.ReleaseDeploymentRolloutStatusFilter), fmt.Sprintf("%s:%s", bean.InvalidFilterCriteria, bean.ReleaseDeploymentRolloutStatusFilter))
				}
				statuses := strings.Split(objs[1], ",")
				releaseDeploymentStatus = append(releaseDeploymentStatus, statuses...)
			}
		default:
			return appIdsFilters, appIdentifierFilters, envIdsFilters, envIdentifierFilters, stageWiseDeploymentStatus, releaseDeploymentStatus, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidFilterCriteria, bean.InvalidFilterCriteria)
		}
	}
	return appIdsFilters, appIdentifierFilters, envIdsFilters, envIdentifierFilters, stageWiseDeploymentStatus, releaseDeploymentStatus, nil
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
