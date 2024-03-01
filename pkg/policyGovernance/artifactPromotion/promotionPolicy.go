package artifactPromotion

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"sync"
)

const updateHookMutex string = "update-hook-mutex"
const deleteHookMutex string = "delete-hook-mutex"

type PromotionPolicyCUDService interface {
	UpdatePolicy(userId int32, policyName string, policyBean *bean.PromotionPolicy) error
	CreatePolicy(userId int32, policyBean *bean.PromotionPolicy) error
	DeletePolicy(userId int32, profileName string) error
	AddPreDeleteHook(hook func(tx *pg.Tx, policyId int) error)
	AddPreUpdateHook(hook func(tx *pg.Tx, policyId int) error)
}

type PromotionPolicyServiceImpl struct {
	globalPolicyDataManager         globalPolicy.GlobalPolicyDataManager
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService
	pipelineService                 pipeline.CdPipelineConfigService
	logger                          *zap.SugaredLogger

	// hooks, mutexes can be optional given that the hooks registration happens at the
	// construction of registering services.
	hookMutexes    map[string]*sync.Mutex
	preDeleteHooks []func(tx *pg.Tx, policyId int) error
	preUpdateHooks []func(tx *pg.Tx, policyId int) error
}

func NewPromotionPolicyServiceImpl(globalPolicyDataManager globalPolicy.GlobalPolicyDataManager,
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService,
	pipelineService pipeline.CdPipelineConfigService,
	logger *zap.SugaredLogger,
) *PromotionPolicyServiceImpl {
	preUpdateHooks := make([]func(tx *pg.Tx, policyId int) error, 0)
	preDeleteHooks := make([]func(tx *pg.Tx, policyId int) error, 0)
	hookMutexes := map[string]*sync.Mutex{
		updateHookMutex: &sync.Mutex{},
		deleteHookMutex: &sync.Mutex{},
	}
	return &PromotionPolicyServiceImpl{
		globalPolicyDataManager:         globalPolicyDataManager,
		resourceQualifierMappingService: resourceQualifierMappingService,
		pipelineService:                 pipelineService,
		logger:                          logger,
		preDeleteHooks:                  preDeleteHooks,
		preUpdateHooks:                  preUpdateHooks,
		hookMutexes:                     hookMutexes,
	}
}

func (impl PromotionPolicyServiceImpl) AddPreDeleteHook(hook func(tx *pg.Tx, policyId int) error) {
	impl.hookMutexes[deleteHookMutex].Lock()
	defer impl.hookMutexes[deleteHookMutex].Unlock()
	impl.preDeleteHooks = append(impl.preDeleteHooks, hook)
}

func (impl PromotionPolicyServiceImpl) AddPreUpdateHook(hook func(tx *pg.Tx, policyId int) error) {
	impl.hookMutexes[updateHookMutex].Lock()
	defer impl.hookMutexes[updateHookMutex].Unlock()
	impl.preUpdateHooks = append(impl.preUpdateHooks, hook)
}

func (impl PromotionPolicyServiceImpl) UpdatePolicy(userId int32, policyName string, policyBean *bean.PromotionPolicy) error {

	globalPolicyDataModel, err := policyBean.ConvertToGlobalPolicyDataModel(userId)
	if err != nil {
		impl.logger.Errorw("error in create policy, not able to convert promotion policy object to global policy data model", "policyBean", policyBean, "err", err)
		return err
	}

	policyId, err := impl.globalPolicyDataManager.GetPolicyIdByName(policyName, bean2.GLOBAL_POLICY_TYPE_IMAGE_PROMOTION_POLICY)
	if err != nil {
		impl.logger.Errorw("error in getting the policy by name", "policyName", policyName, "userId", userId, "err", err)
		if errors.Is(err, pg.ErrNoRows) {
			return &util.ApiError{
				HttpStatusCode:  http.StatusNotFound,
				InternalMessage: fmt.Sprintf("policy with name %s not found", policyName),
				UserMessage:     fmt.Sprintf("policy with name %s not found", policyName),
			}
		}
		return err
	}

	tx, err := impl.resourceQualifierMappingService.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting the transaction", "userId", userId, "policyName", policyName, "err", err)
		return err
	}
	defer impl.resourceQualifierMappingService.RollbackTx(tx)
	for _, hook := range impl.preUpdateHooks {
		err = hook(tx, policyId)
		if err != nil {
			impl.logger.Errorw("error in running pre update hook ", "policyName", policyName, "err", err)
			return err
		}
	}
	_, err = impl.globalPolicyDataManager.UpdatePolicyByName(tx, policyName, globalPolicyDataModel)
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

	err = impl.resourceQualifierMappingService.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing the transaction ", "policyName", policyName, "err", err)
		return err
	}

	return nil
}

func (impl PromotionPolicyServiceImpl) CreatePolicy(userId int32, policyBean *bean.PromotionPolicy) error {
	globalPolicyDataModel, err := policyBean.ConvertToGlobalPolicyDataModel(userId)
	if err != nil {
		impl.logger.Errorw("error in create policy, not able to convert promotion policy object to global policy data model", "policyBean", policyBean, "err", err)
		return err
	}

	tx, err := impl.resourceQualifierMappingService.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting the transaction", "userId", userId, "globalPolicyDataModel", globalPolicyDataModel, "err", err)
		return err
	}
	defer impl.resourceQualifierMappingService.RollbackTx(tx)

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

	err = impl.resourceQualifierMappingService.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing the transaction ", "globalPolicyDataModel", globalPolicyDataModel, "err", err)
		return err
	}
	return nil
}

func (impl PromotionPolicyServiceImpl) DeletePolicy(userId int32, policyName string) error {
	tx, err := impl.resourceQualifierMappingService.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting the transaction", "userId", userId, "policyName", policyName, "err", err)
		return err
	}

	policyId, err := impl.globalPolicyDataManager.GetPolicyIdByName(policyName, bean2.GLOBAL_POLICY_TYPE_IMAGE_PROMOTION_POLICY)
	if err != nil {
		impl.logger.Errorw("error in getting the policy by name", "policyName", policyName, "userId", userId, "err", err)
		if errors.Is(err, pg.ErrNoRows) {
			return &util.ApiError{
				HttpStatusCode:  http.StatusNotFound,
				InternalMessage: fmt.Sprintf("policy with name %s not found", policyName),
				UserMessage:     fmt.Sprintf("policy with name %s not found", policyName),
			}
		}
		return err
	}

	for _, hook := range impl.preDeleteHooks {
		err = hook(tx, policyId)
		if err != nil {
			impl.logger.Errorw("error in running pre delete hook ", "policyName", policyName, "err", err)
			return err
		}
	}

	defer impl.resourceQualifierMappingService.RollbackTx(tx)
	err = impl.globalPolicyDataManager.DeletePolicyByName(tx, policyName, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting the promotion policy using name", "policyName", policyName, "userId", userId, "err", err)
		return err
	}
	err = impl.resourceQualifierMappingService.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing the transaction ", "policyName", policyName, "err", err)
		return err
	}
	return nil
}
