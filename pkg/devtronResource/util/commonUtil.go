package util

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
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
