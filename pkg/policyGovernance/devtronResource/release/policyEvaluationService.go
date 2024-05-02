package release

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/util"
	bean3 "github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/read"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/devtronResource/release/bean"
	"go.uber.org/zap"
	"net/http"
)

type PolicyEvaluationService interface {
	EvaluateReleaseStatusChangeAndGetAutoAction(stateToReq,
		stateFromReq *bean.ReleaseStatusDefinitionState) (isValid bool, autoAction *bean.ReleaseStatusDefinitionState, err error)
}

type PolicyEvaluationServiceImpl struct {
	logger                  *zap.SugaredLogger
	globalPolicyReadService read.ReadService
}

func NewPolicyEvaluationServiceImpl(logger *zap.SugaredLogger,
	globalPolicyReadService read.ReadService) *PolicyEvaluationServiceImpl {
	return &PolicyEvaluationServiceImpl{
		logger:                  logger,
		globalPolicyReadService: globalPolicyReadService,
	}
}

func (impl *PolicyEvaluationServiceImpl) EvaluateReleaseStatusChangeAndGetAutoAction(stateToReq,
	stateFromReq *bean.ReleaseStatusDefinitionState) (isValid bool, autoAction *bean.ReleaseStatusDefinitionState, err error) {
	policyDetail, err := impl.getReleaseStatusPolicy()
	if err != nil {
		impl.logger.Errorw("error getting release status policy", "err", err)
		return false, nil, err
	}
	//checking all policy definitions and matching change done
	for i := range policyDetail.Definitions {
		definition := policyDetail.Definitions[i]
		//matching the stateTo
		if matchDefinitionState(definition.StateTo, stateToReq) {
			//matched, moving to match from states
			for _, possibleFromState := range definition.PossibleFromStates {
				if matchDefinitionState(possibleFromState, stateFromReq) {
					isValid = true
					autoAction = definition.AutoAction
					return isValid, autoAction, nil
				}
			}
		}
	}
	return false, nil, nil
}

func (impl *PolicyEvaluationServiceImpl) getReleaseStatusPolicy() (*bean.ReleaseStatusPolicy, error) {
	//getting global policy related to release
	policies, err := impl.globalPolicyReadService.GetAllGlobalPoliciesByTypeAndVersion(bean2.GLOBAL_POLICY_TYPE_RELEASE_STATUS, bean2.GLOBAL_POLICY_VERSION_V1)
	if err != nil {
		impl.logger.Errorw("error in getting release status policy", "err", err)
		return nil, err
	}
	if len(policies) != 1 {
		impl.logger.Errorw("invalid release status policy found", "policies", policies)
		return nil, &util.ApiError{
			HttpStatusCode:    http.StatusConflict,
			InternalMessage:   "invalid release status policy found",
			UserDetailMessage: "invalid release status policy found",
		}
	}
	policyJson := policies[0].JsonData
	//getting release status policy in detail
	policyDetail := &bean.ReleaseStatusPolicy{}
	err = json.Unmarshal([]byte(policyJson), policyDetail)
	if err != nil {
		impl.logger.Errorw("error in unmarshalling release status policy", "err", err, "policyJson", policyJson)
		return nil, err
	}
	return policyDetail, nil
}

func matchDefinitionState(stateInPolicy, stateInRequest *bean.ReleaseStatusDefinitionState) bool {
	//setting default for draft and readyToRelease/Hold
	if len(stateInRequest.ReleaseRolloutStatus) == 0 {
		stateInRequest.ReleaseRolloutStatus = bean3.NotDeployedReleaseRolloutStatus
	}
	if stateInRequest.ConfigStatus == bean3.ReadyForReleaseStatus || stateInRequest.ConfigStatus == bean3.HoldReleaseStatus &&
		len(stateInRequest.DependencyArtifactStatus) == 0 {
		stateInRequest.DependencyArtifactStatus = bean3.AllSelectedDependencyArtifactStatus
	}
	return stateInPolicy.ConfigStatus == stateInRequest.ConfigStatus &&
		stateInPolicy.ReleaseRolloutStatus == stateInRequest.ReleaseRolloutStatus &&
		stateInPolicy.DependencyArtifactStatus == stateInRequest.DependencyArtifactStatus &&
		(stateInPolicy.LockStatus == nil || *stateInPolicy.LockStatus == *stateInRequest.LockStatus)
}
