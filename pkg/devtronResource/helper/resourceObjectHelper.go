package helper

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
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
