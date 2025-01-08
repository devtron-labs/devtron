package appList

import (
	"github.com/devtron-labs/devtron/api/bean/AppView"
	"net/http"
)

func (handler AppListingRestHandlerImpl) FetchAppPolicyConsequences(w http.ResponseWriter, r *http.Request) {
}

func (handler AppListingRestHandlerImpl) FetchAutocompleteJobCiPipelines(w http.ResponseWriter, r *http.Request) {
}

func (handler AppListingRestHandlerImpl) GetAllAppEnvsFromResourceNames(w http.ResponseWriter, r *http.Request) {
}

func (handler AppListingRestHandlerImpl) updateApprovalConfigDataInAppDetailResp(appDetail AppView.AppDetailContainer, appId, envId int) error {
	return nil
}
