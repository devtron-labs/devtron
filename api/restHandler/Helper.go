package restHandler

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/apiToken"
)

func (impl NotificationRestHandlerImpl) createDraftRequest(token string) (*apiToken.DraftApprovalRequest, error) {
	claimBytes, err := impl.userAuthService.GetFieldValuesFromToken(token)
	if err != nil {
		return nil, err
	}
	return CreateDraftApprovalRequest(claimBytes), err
}

func (impl NotificationRestHandlerImpl) createDeploymentApprovalRequest(token string) (*apiToken.DeploymentApprovalRequest, error) {
	claimBytes, err := impl.userAuthService.GetFieldValuesFromToken(token)
	if err != nil {
		return nil, err
	}
	return CreateDeploymentApprovalRequest(claimBytes), err
}

func CreateDraftApprovalRequest(jsonStr []byte) *apiToken.DraftApprovalRequest {
	draftApprovalRequest := &apiToken.DraftApprovalRequest{}
	json.Unmarshal(jsonStr, draftApprovalRequest)
	return draftApprovalRequest
}

func CreateDeploymentApprovalRequest(jsonStr []byte) *apiToken.DeploymentApprovalRequest {
	deploymentApprovalRequest := &apiToken.DeploymentApprovalRequest{}
	json.Unmarshal(jsonStr, deploymentApprovalRequest)
	return deploymentApprovalRequest
}
