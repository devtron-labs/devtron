package appList

import (
	"context"
	"github.com/devtron-labs/devtron/api/bean/AppView"
	userBean "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"net/http"
)

func (handler AppListingRestHandlerImpl) FetchAppPolicyConsequences(w http.ResponseWriter, r *http.Request) {
}

func (handler AppListingRestHandlerImpl) FetchAutocompleteJobCiPipelines(w http.ResponseWriter, r *http.Request) {
}

func (handler AppListingRestHandlerImpl) GetAllAppEnvsFromResourceNames(w http.ResponseWriter, r *http.Request) {
}

func (handler AppListingRestHandlerImpl) updateApprovalConfigDataInAppDetailResp(ctx context.Context, appDetail AppView.AppDetailContainer, appId, envId int, userMetadata *userBean.UserMetadata) (AppView.AppDetailContainer, error) {
	return appDetail, nil
}
