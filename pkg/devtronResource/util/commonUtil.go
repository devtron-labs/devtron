package util

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"net/http"
	"strings"
)

func DecodeFilterCriteriaString(criteria string) (*bean.FilterCriteriaDecoder, error) {
	objs := strings.Split(criteria, "|")
	if len(objs) != 3 {
		return nil, &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			Code:            "400",
			InternalMessage: "invalid format filter criteria!",
			UserMessage:     "invalid format filter criteria!",
		}
	}
	criteriaDecoder := adapter.BuildFilterCriteriaDecoder(objs[0], objs[1], objs[2])
	return criteriaDecoder, nil
}
