package artifactPromotion

import (
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"go.uber.org/zap"
)

type PromotionPolicyService interface {
	GetByAppAndEnvId(appId, envId int) (*bean.PromotionPolicy, error)
	GetByAppNameAndEnvName(appName string, envNames []string) (map[string]*bean.PromotionPolicy, error)
	GetById(id int) (*bean.PromotionPolicy, error)
	GetByIds(ids []int) ([]*bean.PromotionPolicy, error)
}

type PromotionPolicyServiceImpl struct {
	globalPolicyService             globalPolicy.GlobalPolicyService
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService
	logger                          *zap.SugaredLogger
}

func NewPromotionPolicyServiceImpl(globalPolicyService globalPolicy.GlobalPolicyService,
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService,
	logger *zap.SugaredLogger,
) *PromotionPolicyServiceImpl {
	return &PromotionPolicyServiceImpl{
		globalPolicyService:             globalPolicyService,
		resourceQualifierMappingService: resourceQualifierMappingService,
		logger:                          logger,
	}
}

func (impl PromotionPolicyServiceImpl) GetByAppAndEnvId(appId, envId int) (*bean.PromotionPolicy, error) {

	//scope := &resourceQualifiers.Scope{AppId: appId, EnvId: envId}
	//
	//qualifierMapping, err := impl.resourceQualifierMappingService.GetResourceMappingsForScopes(
	//	resourceQualifiers.ImagePromotionPolicy,
	//	resourceQualifiers.ApplicationEnvironmentSelector,
	//	[]*resourceQualifiers.Scope{scope},
	//)
	//if err != nil {
	//	impl.logger.Errorw("error in fetching resource qualifier mapping by scope", "resource", resourceQualifiers.ImagePromotionPolicy, "scope", scope, "err", err)
	//	return nil, err
	//}
	//
	//policyId := qualifierMapping[0].ResourceId
	//
	////TODO; get from new service
	//promotionPolicyDao, err := impl.globalPolicyService.GetById(policyId)
	//if err!=nil{
	//	impl.logger.Errorw("error in fetching policy by id", "policyId", policyId)
	//	return nil, err
	//}
	//
	return &bean.PromotionPolicy{}, nil
}


func (impl PromotionPolicyServiceImpl) GetByAppNameAndEnvName(appName string, envNames []string) (map[string]*bean.PromotionPolicy, error) {

	// scope := &resourceQualifiers.Scope{AppId: appId, EnvId: envId}
	//
	// qualifierMapping, err := impl.resourceQualifierMappingService.GetResourceMappingsForScopes(
	//	resourceQualifiers.ImagePromotionPolicy,
	//	resourceQualifiers.ApplicationEnvironmentSelector,
	//	[]*resourceQualifiers.Scope{scope},
	// )
	// if err != nil {
	//	impl.logger.Errorw("error in fetching resource qualifier mapping by scope", "resource", resourceQualifiers.ImagePromotionPolicy, "scope", scope, "err", err)
	//	return nil, err
	// }
	//
	// policyId := qualifierMapping[0].ResourceId
	//
	// //TODO; get from new service
	// promotionPolicyDao, err := impl.globalPolicyService.GetById(policyId)
	// if err!=nil{
	//	impl.logger.Errorw("error in fetching policy by id", "policyId", policyId)
	//	return nil, err
	// }
	//
	return nil, nil
}

func (impl PromotionPolicyServiceImpl) GetById(id int) (*bean.PromotionPolicy, error) {
	return nil, nil
}

func (impl PromotionPolicyServiceImpl) GetByIds(ids []int) ([]*bean.PromotionPolicy, error) {
	return nil, nil
}
