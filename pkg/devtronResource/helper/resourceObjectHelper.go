/*
 * Copyright (c) 2024. Devtron Inc.
 */

package helper

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
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
