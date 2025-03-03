package helper

import (
	apiBean "github.com/devtron-labs/devtron/api/userResource/bean"
	"github.com/devtron-labs/devtron/internal/util"
	bean5 "github.com/devtron-labs/devtron/pkg/userResource/bean"
	"net/http"
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
