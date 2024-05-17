package artifactPromotion

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/enterprise/pkg/expressionEvaluators"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

const policyNotFoundErr = "policy with name %s not found"
const policyAlreadyExistsErr = "policy name already exists, err: duplicate name"
const policyupdationErr = "error in updating policy"

type PolicyCUDService interface {
	UpdatePolicy(ctx *util2.RequestCtx, policyName string, policyBean *bean.PromotionPolicy) error
	CreatePolicy(ctx *util2.RequestCtx, policyBean *bean.PromotionPolicy) error
	DeletePolicy(ctx *util2.RequestCtx, profileName string) error
}

type PolicyEventNotifier interface {
	AddDeleteEventObserver(hook func(tx *pg.Tx, policyId int) error)
	AddUpdateEventObserver(hook func(tx *pg.Tx, policy *bean.PromotionPolicy) error)
}

type PromotionPolicyServiceImpl struct {
	globalPolicyDataManager         globalPolicy.GlobalPolicyDataManager
	pipelineService                 pipeline.CdPipelineConfigService
	logger                          *zap.SugaredLogger
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService
	celEvaluatorService             expressionEvaluators.CELEvaluatorService
	transactionManager              sql.TransactionWrapper

	// hooks
	preDeleteHooks []func(tx *pg.Tx, policyId int) error
	preUpdateHooks []func(tx *pg.Tx, policy *bean.PromotionPolicy) error
}

func NewPromotionPolicyServiceImpl(globalPolicyDataManager globalPolicy.GlobalPolicyDataManager,
	pipelineService pipeline.CdPipelineConfigService,
	logger *zap.SugaredLogger,
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService,
	celEvaluatorService expressionEvaluators.CELEvaluatorService,
	transactionManager sql.TransactionWrapper,
) *PromotionPolicyServiceImpl {
	preDeleteHooks := make([]func(tx *pg.Tx, policyId int) error, 0)
	preUpdateHooks := make([]func(tx *pg.Tx, policy *bean.PromotionPolicy) error, 0)

	return &PromotionPolicyServiceImpl{
		globalPolicyDataManager:         globalPolicyDataManager,
		pipelineService:                 pipelineService,
		logger:                          logger,
		resourceQualifierMappingService: resourceQualifierMappingService,
		celEvaluatorService:             celEvaluatorService,
		transactionManager:              transactionManager,
		preDeleteHooks:                  preDeleteHooks,
		preUpdateHooks:                  preUpdateHooks,
	}
}

func (impl *PromotionPolicyServiceImpl) AddDeleteEventObserver(hook func(tx *pg.Tx, policyId int) error) {
	impl.preDeleteHooks = append(impl.preDeleteHooks, hook)
}

func (impl *PromotionPolicyServiceImpl) AddUpdateEventObserver(hook func(tx *pg.Tx, policy *bean.PromotionPolicy) error) {
	impl.preUpdateHooks = append(impl.preUpdateHooks, hook)
}

func (impl *PromotionPolicyServiceImpl) UpdatePolicy(ctx *util2.RequestCtx, policyName string, policyBean *bean.PromotionPolicy) error {
	validateResp, inValid := impl.celEvaluatorService.ValidateCELRequest(expressionEvaluators.ValidateRequestResponse{Conditions: policyBean.Conditions})
	if inValid {
		err := errors.New("invalid filter conditions : " + fmt.Sprint(validateResp))
		return util.NewApiError().WithHttpStatusCode(http.StatusUnprocessableEntity).WithUserMessage("invalid conditions statements : " + err.Error())
	}
	globalPolicyDataModel, err := policyBean.ConvertToGlobalPolicyDataModel(ctx.GetUserId())
	if err != nil {
		impl.logger.Errorw("error in create policy, not able to convert promotion policy object to global policy data model", "policyBean", policyBean, "err", err)
		return err
	}

	policyId, err := impl.globalPolicyDataManager.GetPolicyIdByName(policyName, bean2.GLOBAL_POLICY_TYPE_IMAGE_PROMOTION_POLICY)
	if err != nil {
		impl.logger.Errorw("error in getting the policy by name", "policyName", policyName, "userId", ctx.GetUserId(), "err", err)
		if errors.Is(err, pg.ErrNoRows) {
			errMsg := fmt.Sprintf(policyNotFoundErr, policyName)
			return util.NewApiError().WithHttpStatusCode(http.StatusNotFound).WithUserMessage(errMsg).WithInternalMessage(errMsg)
		}
		return err
	}
	policyBean.Id = policyId

	// todo: create a transaction manager
	tx, err := impl.transactionManager.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting the transaction", "userId", ctx.GetUserId(), "policyName", policyName, "err", err)
		return err
	}
	defer impl.transactionManager.RollbackTx(tx)
	_, err = impl.globalPolicyDataManager.UpdatePolicyByName(tx, policyName, globalPolicyDataModel)
	if err != nil {
		errResp := util.NewApiError().WithHttpStatusCode(http.StatusInternalServerError).WithInternalMessage(err.Error()).WithUserMessage(policyupdationErr)
		if strings.Contains(err.Error(), bean2.UniqueActiveNameConstraint) {
			errResp = errResp.WithHttpStatusCode(http.StatusConflict).WithUserMessage(policyAlreadyExistsErr)
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

func (impl *PromotionPolicyServiceImpl) CreatePolicy(ctx *util2.RequestCtx, policyBean *bean.PromotionPolicy) error {
	validateResp, valid := impl.celEvaluatorService.ValidateCELRequest(expressionEvaluators.ValidateRequestResponse{Conditions: policyBean.Conditions})
	if valid {
		err := errors.New("invalid filter conditions : " + fmt.Sprint(validateResp))
		return util.NewApiError().WithHttpStatusCode(http.StatusUnprocessableEntity).WithUserMessage("invalid conditions statements : " + err.Error())
	}
	globalPolicyDataModel, err := policyBean.ConvertToGlobalPolicyDataModel(ctx.GetUserId())
	if err != nil {
		impl.logger.Errorw("error in create policy, not able to convert promotion policy object to global policy data model", "policyBean", policyBean, "err", err)
		return err
	}

	_, err = impl.globalPolicyDataManager.CreatePolicy(globalPolicyDataModel, nil)
	if err != nil {
		errResp := util.NewApiError().WithHttpStatusCode(http.StatusInternalServerError).WithInternalMessage(err.Error()).WithUserMessage(policyupdationErr)
		if strings.Contains(err.Error(), bean2.UniqueActiveNameConstraint) {
			errResp = errResp.WithHttpStatusCode(http.StatusConflict).WithUserMessage(policyAlreadyExistsErr)
		}
		return errResp
	}
	return nil
}

func (impl *PromotionPolicyServiceImpl) DeletePolicy(ctx *util2.RequestCtx, policyName string) error {
	tx, err := impl.transactionManager.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting the transaction", "userId", ctx.GetUserId(), "policyName", policyName, "err", err)
		return err
	}

	policyId, err := impl.globalPolicyDataManager.GetPolicyIdByName(policyName, bean2.GLOBAL_POLICY_TYPE_IMAGE_PROMOTION_POLICY)
	if err != nil {
		impl.logger.Errorw("error in getting the policy by name", "policyName", policyName, "userId", ctx.GetUserId(), "err", err)
		if errors.Is(err, pg.ErrNoRows) {
			errMsg := fmt.Sprintf(policyNotFoundErr, policyName)
			return util.NewApiError().WithHttpStatusCode(http.StatusNotFound).WithUserMessage(errMsg).WithInternalMessage(errMsg)
		}
		return err
	}

	defer impl.transactionManager.RollbackTx(tx)
	err = impl.globalPolicyDataManager.DeletePolicyByName(tx, policyName, bean2.GLOBAL_POLICY_TYPE_IMAGE_PROMOTION_POLICY, ctx.GetUserId())
	if err != nil {
		impl.logger.Errorw("error in deleting the promotion policy using name", "policyName", policyName, "userId", ctx.GetUserId(), "err", err)
		return err
	}

	err = impl.resourceQualifierMappingService.DeleteAllQualifierMappingsByResourceTypeAndId(resourceQualifiers.ImagePromotionPolicy, policyId, sql.NewDefaultAuditLog(ctx.GetUserId()), tx)
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
