package devtronResource

import (
	"fmt"
	apiBean "github.com/devtron-labs/devtron/api/devtronResource/bean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/devtronResource/helper"
	"github.com/gorilla/mux"
	"net/http"
)

func getKindSubKindVersion(w http.ResponseWriter, r *http.Request) (kind string, subKind string, version string, caughtError bool) {
	vars := mux.Vars(r)
	kindVar := vars[apiBean.PathParamKind]
	versionVar := vars[apiBean.PathParamVersion]
	kind, subKind, statusCode, err := resolveKindSubKindValues(kindVar)
	if err != nil {
		common.WriteJsonResp(w, err, nil, statusCode)
		caughtError = true
	}
	return kind, subKind, versionVar, caughtError
}

func resolveKindSubKindValues(kindVar string) (kind, subKind string, statusCode int, err error) {
	kind, subKind, err = helper.GetKindAndSubKindFrom(kindVar)
	if err != nil {
		err = fmt.Errorf("invalid parameter: kind")
		statusCode = http.StatusBadRequest
	}
	return kind, subKind, statusCode, err
}
