/*
 * Copyright (c) 2024. Devtron Inc.
 */

package release

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/read"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/devtronResource/release/bean"
	util2 "github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"net/http"
)

type PolicyEvaluationService interface {
	EvaluateReleaseActionRequest(operationTypeFromReq bean.PolicyReleaseOperationType,
		operationPathFromReq []bean.PolicyReleaseOperationPath, stateFromReq *bean.ReleaseStatusDefinitionState) (isValid bool, err error)

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

func (impl *PolicyEvaluationServiceImpl) EvaluateReleaseActionRequest(operationTypeFromReq bean.PolicyReleaseOperationType,
	operationPathFromReq []bean.PolicyReleaseOperationPath, stateFromReq *bean.ReleaseStatusDefinitionState) (isValid bool, err error) {
	policyDetail, err := impl.getReleaseActionCheckPolicy()
	if err != nil {
		impl.logger.Errorw("error getting release status policy", "err", err)
		return false, err
	}
	//checking all policy definitions and matching change done
	for i := range policyDetail.Definitions {
		definition := policyDetail.Definitions[i]
		if operationTypeFromReq == definition.OperationType {
			toCheckStates := len(definition.OperationPaths) == 0 //if no operation path then check, else iterate and see if any operation path match
			for _, operationPath := range definition.OperationPaths {
				if util2.ContainsStringAlias(operationPathFromReq, operationPath) {
					toCheckStates = true
					break
				}
			}
			if toCheckStates {
				for _, possibleFromState := range definition.PossibleFromStates {
					if matchDefinitionState(possibleFromState, stateFromReq) {
						isValid = true
						return isValid, nil
					}
				}
			}
		}
	}
	return false, nil
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

func (impl *PolicyEvaluationServiceImpl) getReleaseActionCheckPolicy() (*bean.ReleaseActionCheckPolicy, error) {
	//getting global policy related to release action check
	policyJson, err := impl.getReleaseRelatedPolicyJsonByType(bean2.GLOBAL_POLICY_TYPE_RELEASE_ACTION_CHECK)
	if err != nil {
		impl.logger.Errorw("error, getReleaseRelatedPolicyJsonByType", "err", err, "policyType", bean2.GLOBAL_POLICY_TYPE_RELEASE_ACTION_CHECK)
		return nil, err
	}
	//getting release status policy in detail
	policyDetail := &bean.ReleaseActionCheckPolicy{}
	err = json.Unmarshal([]byte(policyJson), policyDetail)
	if err != nil {
		impl.logger.Errorw("error in unmarshalling release status policy", "err", err, "policyJson", policyJson)
		return nil, err
	}
	return policyDetail, nil
}

func (impl *PolicyEvaluationServiceImpl) getReleaseStatusPolicy() (*bean.ReleaseStatusPolicy, error) {
	//getting global policy related to release status
	policyJson, err := impl.getReleaseRelatedPolicyJsonByType(bean2.GLOBAL_POLICY_TYPE_RELEASE_STATUS)
	if err != nil {
		impl.logger.Errorw("error, getReleaseRelatedPolicyJsonByType", "err", err, "policyType", bean2.GLOBAL_POLICY_TYPE_RELEASE_STATUS)
		return nil, err
	}
	//getting release status policy in detail
	policyDetail := &bean.ReleaseStatusPolicy{}
	err = json.Unmarshal([]byte(policyJson), policyDetail)
	if err != nil {
		impl.logger.Errorw("error in unmarshalling release status policy", "err", err, "policyJson", policyJson)
		return nil, err
	}
	return policyDetail, nil
}

func (impl *PolicyEvaluationServiceImpl) getReleaseRelatedPolicyJsonByType(policyType bean2.GlobalPolicyType) (policyJson string, err error) {
	//getting global policy related to release
	policies, err := impl.globalPolicyReadService.GetAllGlobalPoliciesByTypeAndVersion(policyType, bean2.GLOBAL_POLICY_VERSION_V1)
	if err != nil {
		impl.logger.Errorw("error in getting release status policy", "err", err)
		return policyJson, err
	}
	if len(policies) != 1 {
		impl.logger.Errorw("invalid release status policy found", "policies", policies)
		return policyJson, &util.ApiError{
			HttpStatusCode:    http.StatusConflict,
			InternalMessage:   "invalid release status policy found",
			UserDetailMessage: "invalid release status policy found",
		}
	}
	policyJson = policies[0].JsonData
	return policyJson, nil
}

func matchDefinitionState(stateInPolicy, stateInRequest *bean.ReleaseStatusDefinitionState) bool {
	configStatusCheck := stateInPolicy.ConfigStatus == bean.PolicyReleaseConfigStatusAny ||
		stateInPolicy.ConfigStatus == stateInRequest.ConfigStatus
	rolloutStatusCheck := stateInPolicy.ReleaseRolloutStatus == bean.PolicyReleaseRolloutStatusAny ||
		stateInPolicy.ReleaseRolloutStatus == stateInRequest.ReleaseRolloutStatus
	depArtifactStatusCheck := stateInPolicy.DependencyArtifactStatus == bean.PolicyDependencyArtifactStatusAny ||
		stateInPolicy.DependencyArtifactStatus == stateInRequest.DependencyArtifactStatus
	lockStatusCheck := stateInPolicy.LockStatus == bean.PolicyLockStatusAny ||
		stateInPolicy.LockStatus == stateInRequest.LockStatus
	return configStatusCheck && rolloutStatusCheck && depArtifactStatusCheck && lockStatusCheck
}
