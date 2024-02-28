package artifactPromotion

import (
	"errors"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type PromotionPolicyCUDService interface {
	UpdatePolicy(userId int32, policyName string, policyBean *bean.PromotionPolicy) error
	CreatePolicy(userId int32, policyBean *bean.PromotionPolicy) error
	DeletePolicy(userId int32, profileName string) error
	ApplyPolicyToIdentifiers(userId int32, applyIdentifiersRequest bean.BulkPromotionPolicyApplyRequest) error
}

type PromotionPolicyServiceImpl struct {
	globalPolicyDataManager         globalPolicy.GlobalPolicyDataManager
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService
	logger                          *zap.SugaredLogger
}

func NewPromotionPolicyServiceImpl(globalPolicyDataManager globalPolicy.GlobalPolicyDataManager,
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService,
	logger *zap.SugaredLogger,
) *PromotionPolicyServiceImpl {
	return &PromotionPolicyServiceImpl{
		globalPolicyDataManager:         globalPolicyDataManager,
		resourceQualifierMappingService: resourceQualifierMappingService,
		logger:                          logger,
	}
}

func (impl PromotionPolicyServiceImpl) UpdatePolicy(userId int32, policyName string, policyBean *bean.PromotionPolicy) error {
	globalPolicyDataModel, err := policyBean.ConvertToGlobalPolicyDataModel(userId)
	if err != nil {
		impl.logger.Errorw("error in create policy, not able to convert promotion policy object to global policy data model", "policyBean", policyBean, "err", err)
		return err
	}

	_, err = impl.globalPolicyDataManager.UpdatePolicyByName(policyName, globalPolicyDataModel)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), bean2.UniqueActiveNameConstraint) {
			err = errors.New("policy name already exists, err: duplicate name")
			statusCode = http.StatusConflict
		}
		return &util.ApiError{
			HttpStatusCode:  statusCode,
			InternalMessage: err.Error(),
			UserMessage:     err.Error(),
		}
	}
	return nil
}

func (impl PromotionPolicyServiceImpl) CreatePolicy(userId int32, policyBean *bean.PromotionPolicy) error {
	globalPolicyDataModel, err := policyBean.ConvertToGlobalPolicyDataModel(userId)
	if err != nil {
		impl.logger.Errorw("error in create policy, not able to convert promotion policy object to global policy data model", "policyBean", policyBean, "err", err)
		return err
	}

	_, err = impl.globalPolicyDataManager.CreatePolicy(globalPolicyDataModel, nil)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), bean2.UniqueActiveNameConstraint) {
			err = errors.New("policy name already exists, err: duplicate name")
			statusCode = http.StatusConflict
		}
		return &util.ApiError{
			HttpStatusCode:  statusCode,
			InternalMessage: err.Error(),
			UserMessage:     err.Error(),
		}
	}
	return nil
}

func (impl PromotionPolicyServiceImpl) DeletePolicy(userId int32, policyName string) error {
	err := impl.globalPolicyDataManager.DeletePolicyByName(policyName, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting the promotion policy using name", "policyName", policyName, "userId", userId, "err", err)
	}
	return err
}

func (impl PromotionPolicyServiceImpl) ApplyPolicyToIdentifiers(userId int32, applyIdentifiersRequest bean.BulkPromotionPolicyApplyRequest) error {
	// updateToProfile, err := impl.globalPolicyDataManager.GetPolicyByName(applyIdentifiersRequest.ApplyToPolicyName)
	// if err != nil {
	// 	statusCode := http.StatusInternalServerError
	// 	if errors.Is(err, pg.ErrNoRows) {
	// 		err = errors.New(fmt.Sprintf("promotion policy with name '%s' does not exist", applyIdentifiersRequest.ApplyToPolicyName))
	// 		statusCode = http.StatusConflict
	// 	}
	// 	return &util.ApiError{
	// 		HttpStatusCode:  statusCode,
	// 		InternalMessage: err.Error(),
	// 		UserMessage:     err.Error(),
	// 	}
	// }
	// if len(applyIdentifiersRequest.ApplicationEnvironments) > 0 {
	//
	// }
	// if applyIdentifiersRequest.AppEnvPolicyListFilter != nil {
	//
	// }
	return nil
}
