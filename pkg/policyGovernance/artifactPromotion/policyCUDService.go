package artifactPromotion

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

const policyNotFoundErrMsg = "policy with name %s not found"

type PolicyCUDService interface {
	UpdatePolicy(userId int32, policyName string, policyBean *bean.PromotionPolicy) error
	CreatePolicy(userId int32, policyBean *bean.PromotionPolicy) error
	DeletePolicy(userId int32, profileName string) error
	AddDeleteHook(hook func(tx *pg.Tx, policyId int) error)
	AddUpdateHook(hook func(tx *pg.Tx, policy *bean.PromotionPolicy) error)
}

type PromotionPolicyServiceImpl struct {
	globalPolicyDataManager globalPolicy.GlobalPolicyDataManager
	pipelineService         pipeline.CdPipelineConfigService
	logger                  *zap.SugaredLogger
	transactionManager      sql.TransactionWrapper

	// hooks
	preDeleteHooks []func(tx *pg.Tx, policyId int) error
	preUpdateHooks []func(tx *pg.Tx, policy *bean.PromotionPolicy) error
}

func NewPromotionPolicyServiceImpl(globalPolicyDataManager globalPolicy.GlobalPolicyDataManager,
	pipelineService pipeline.CdPipelineConfigService,
	logger *zap.SugaredLogger, transactionManager sql.TransactionWrapper,
) *PromotionPolicyServiceImpl {
	preDeleteHooks := make([]func(tx *pg.Tx, policyId int) error, 0)
	preUpdateHooks := make([]func(tx *pg.Tx, policy *bean.PromotionPolicy) error, 0)

	return &PromotionPolicyServiceImpl{
		globalPolicyDataManager: globalPolicyDataManager,
		pipelineService:         pipelineService,
		logger:                  logger,
		transactionManager:      transactionManager,
		preDeleteHooks:          preDeleteHooks,
		preUpdateHooks:          preUpdateHooks,
	}
}

func (impl *PromotionPolicyServiceImpl) AddDeleteHook(hook func(tx *pg.Tx, policyId int) error) {
	impl.preDeleteHooks = append(impl.preDeleteHooks, hook)
}

func (impl *PromotionPolicyServiceImpl) AddUpdateHook(hook func(tx *pg.Tx, policy *bean.PromotionPolicy) error) {
	impl.preUpdateHooks = append(impl.preUpdateHooks, hook)
}

func (impl *PromotionPolicyServiceImpl) UpdatePolicy(userId int32, policyName string, policyBean *bean.PromotionPolicy) error {

	globalPolicyDataModel, err := policyBean.ConvertToGlobalPolicyDataModel(userId)
	if err != nil {
		impl.logger.Errorw("error in create policy, not able to convert promotion policy object to global policy data model", "policyBean", policyBean, "err", err)
		return err
	}

	policyId, err := impl.globalPolicyDataManager.GetPolicyIdByName(policyName, bean2.GLOBAL_POLICY_TYPE_IMAGE_PROMOTION_POLICY)
	if err != nil {
		impl.logger.Errorw("error in getting the policy by name", "policyName", policyName, "userId", userId, "err", err)
		if errors.Is(err, pg.ErrNoRows) {
			errMsg := fmt.Sprintf(policyNotFoundErrMsg, policyName)
			return util.NewApiError().WithHttpStatusCode(http.StatusNotFound).WithUserMessage(errMsg).WithInternalMessage(errMsg)
		}
		return err
	}
	policyBean.Id = policyId

	// todo: create a transaction manager
	tx, err := impl.transactionManager.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting the transaction", "userId", userId, "policyName", policyName, "err", err)
		return err
	}
	defer impl.transactionManager.RollbackTx(tx)
	_, err = impl.globalPolicyDataManager.UpdatePolicyByName(tx, policyName, globalPolicyDataModel)
	if err != nil {
		errResp := util.NewApiError().WithHttpStatusCode(http.StatusInternalServerError).WithInternalMessage(err.Error()).WithUserMessage("error in updating policy")
		if strings.Contains(err.Error(), bean2.UniqueActiveNameConstraint) {
			errResp = errResp.WithHttpStatusCode(http.StatusConflict).WithUserMessage("policy name already exists, err: duplicate name")
		}
		return errResp
	}

	for _, hook := range impl.preUpdateHooks {
		err = hook(tx, policyBean)
		if err != nil {
			impl.logger.Errorw("error in running pre update hook ", "policyName", policyName, "err", err)
			return err
		}
	}

	err = impl.transactionManager.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing the transaction ", "policyName", policyName, "err", err)
		return err
	}

	return nil
}

func (impl *PromotionPolicyServiceImpl) CreatePolicy(userId int32, policyBean *bean.PromotionPolicy) error {
	globalPolicyDataModel, err := policyBean.ConvertToGlobalPolicyDataModel(userId)
	if err != nil {
		impl.logger.Errorw("error in create policy, not able to convert promotion policy object to global policy data model", "policyBean", policyBean, "err", err)
		return err
	}

	_, err = impl.globalPolicyDataManager.CreatePolicy(globalPolicyDataModel, nil)
	if err != nil {
		errResp := util.NewApiError().WithHttpStatusCode(http.StatusInternalServerError).WithInternalMessage(err.Error()).WithUserMessage("error in updating policy")
		if strings.Contains(err.Error(), bean2.UniqueActiveNameConstraint) {
			errResp = errResp.WithHttpStatusCode(http.StatusConflict).WithUserMessage("policy name already exists, err: duplicate name")
		}
		return errResp
	}
	return nil
}

func (impl *PromotionPolicyServiceImpl) DeletePolicy(userId int32, policyName string) error {
	tx, err := impl.transactionManager.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting the transaction", "userId", userId, "policyName", policyName, "err", err)
		return err
	}

	policyId, err := impl.globalPolicyDataManager.GetPolicyIdByName(policyName, bean2.GLOBAL_POLICY_TYPE_IMAGE_PROMOTION_POLICY)
	if err != nil {
		impl.logger.Errorw("error in getting the policy by name", "policyName", policyName, "userId", userId, "err", err)
		if errors.Is(err, pg.ErrNoRows) {
			errMsg := fmt.Sprintf(policyNotFoundErrMsg, policyName)
			return util.NewApiError().WithHttpStatusCode(http.StatusNotFound).WithUserMessage(errMsg).WithInternalMessage(errMsg)
		}
		return err
	}

	defer impl.transactionManager.RollbackTx(tx)
	err = impl.globalPolicyDataManager.DeletePolicyByName(tx, policyName, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting the promotion policy using name", "policyName", policyName, "userId", userId, "err", err)
		return err
	}
	for _, hook := range impl.preDeleteHooks {
		err = hook(tx, policyId)
		if err != nil {
			impl.logger.Errorw("error in running pre delete hook ", "policyName", policyName, "err", err)
			return err
		}
	}
	err = impl.transactionManager.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing the transaction ", "policyName", policyName, "err", err)
		return err
	}
	return nil
}
