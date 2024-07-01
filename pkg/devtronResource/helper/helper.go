package helper

import (
	"github.com/devtron-labs/devtron/internal/util"
	"net/http"
	"strings"
)

func GetKindAndSubKindFrom(resourceKindVar string) (kind, subKind string, err error) {
	kindSplits := strings.Split(resourceKindVar, "/")
	if len(kindSplits) == 1 {
		kind = kindSplits[0]
	} else if len(kindSplits) == 2 {
		kind = kindSplits[0]
		subKind = kindSplits[1]
	} else {
		return kind, subKind, &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			Code:            "400",
			InternalMessage: "invalid kind!",
			UserMessage:     "invalid kind!",
		}
	}
	return kind, subKind, nil
}
