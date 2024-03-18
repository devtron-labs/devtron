package artifactPromotion

import (
	"context"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/mocks"
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/globalPolicy/repository"
	"github.com/devtron-labs/devtron/pkg/policyGovernance"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/constants"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	deploymentStatus "github.com/devtron-labs/devtron/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPolicyHooks(t *testing.T) {
	requestService, policyCUDService := getRequestService(t)
	if requestService == nil {
		t.SkipNow()
	}

	t.Run("testing delete policy hook", func(tt *testing.T) {
		//  create policy
		// 	create promotion requests and set above created policy id in the requests
		// 	delete the policy
		//  check all the above created requests are deleted
		policyName := "random-123"
		policy := &bean.PromotionPolicy{
			Name:        policyName,
			Description: "testing delete hook",
			Conditions: []deploymentStatus.ResourceCondition{
				{
					ConditionType: deploymentStatus.PASS,
					Expression:    "true",
				},
			},
			ApprovalMetaData: bean.ApprovalMetaData{
				ApprovalCount: 1,
			},
		}
		ctx, _ := context.WithCancel(context.Background())
		ctx = context.WithValue(ctx, deploymentStatus.UserId, int32(2))
		rctx := deploymentStatus.NewRequestCtx(ctx)
		err := policyCUDService.CreatePolicy(rctx, policy)
		if err != nil {
			t.Fail()
		}

		// create policy
		createdPolicy, err := policyCUDService.globalPolicyDataManager.GetPolicyByName(policyName, bean2.GLOBAL_POLICY_TYPE_IMAGE_PROMOTION_POLICY)
		if err != nil {
			t.Fail()
		}

		requests := []*repository.ArtifactPromotionApprovalRequest{
			{
				PolicyId:                createdPolicy.Id,
				DestinationPipelineId:   1,
				ArtifactId:              1,
				PolicyEvaluationAuditId: 1,
				SourceType:              constants.SOURCE_TYPE_CI.GetSourceType(),
				SourcePipelineId:        1,
				Status:                  constants.AWAITING_APPROVAL,
				AuditLog:                sql.NewDefaultAuditLog(1),
			},

			{
				PolicyId:                createdPolicy.Id,
				DestinationPipelineId:   2,
				ArtifactId:              1,
				PolicyEvaluationAuditId: 1,
				SourceType:              constants.SOURCE_TYPE_CI.GetSourceType(),
				SourcePipelineId:        1,
				Status:                  constants.AWAITING_APPROVAL,
				AuditLog:                sql.NewDefaultAuditLog(1),
			},

			{
				PolicyId:                createdPolicy.Id,
				DestinationPipelineId:   3,
				ArtifactId:              1,
				PolicyEvaluationAuditId: 1,
				SourceType:              constants.SOURCE_TYPE_CI.GetSourceType(),
				SourcePipelineId:        1,
				Status:                  constants.AWAITING_APPROVAL,
				AuditLog:                sql.NewDefaultAuditLog(1),
			},
		}

		tx, err := requestService.transactionManager.StartTx()
		if err != nil {
			t.Fail()
		}
		requestIds := make([]int, 0, len(requests))
		for _, req := range requests {
			createdReq, err := requestService.artifactPromotionApprovalRequestRepository.Create(tx, req)
			if err != nil {
				t.Fail()
			}
			requestIds = append(requestIds, createdReq.Id)
		}
		err = tx.Commit()
		if err != nil {
			t.Fail()
		}
		// delete policies
		err = policyCUDService.DeletePolicy(rctx, policyName)
		if err != nil {
			t.Fail()
		}

		// test if the requests are marked stale after policy deletion
		for _, requestId := range requestIds {
			req, err := requestService.artifactPromotionApprovalRequestRepository.FindById(requestId)
			if err != nil {
				t.Fail()
			}
			assert.Equal(tt, constants.STALE, req.Status)
		}

	})

	t.Run("testing update policy hook", func(tt *testing.T) {

	})

}

func getRequestService(t *testing.T) (*ApprovalRequestServiceImpl, *PromotionPolicyServiceImpl) {

	dbConfig, err := sql.GetConfig()
	if err != nil {
		t.Fail()
	}
	logger, _ := util.NewSugardLogger()
	dbConnection, err := sql.NewDbConnection(dbConfig, logger)
	if err != nil {
		t.SkipNow()
	}

	transactionManager := sql.NewTransactionUtilImpl(dbConnection)
	requestRepo := repository.NewRequestRepositoryImpl(dbConnection)
	celService := resourceFilter.NewCELServiceImpl(logger)
	resourceFilterEvalutionService, _ := resourceFilter.NewResourceFilterEvaluatorImpl(logger, celService)
	auditRepo := resourceFilter.NewFilterEvaluationAuditRepositoryImpl(logger, dbConnection)
	resourceFilterEvalutionAuditService := resourceFilter.NewFilterEvaluationAuditServiceImpl(logger, auditRepo, nil)

	globalPolicyRepo := repository2.NewGlobalPolicyRepositoryImpl(logger, dbConnection)
	globalPolicySearchableKeyRepo := repository2.NewGlobalPolicySearchableFieldRepositoryImpl(logger, dbConnection)
	globalPolicyManager := globalPolicy.NewGlobalPolicyDataManagerImpl(logger, globalPolicyRepo, globalPolicySearchableKeyRepo, nil)
	cdPipelineConfigService := mocks.NewCdPipelineConfigService(t)
	promotionPolicyService := NewPromotionPolicyServiceImpl(globalPolicyManager, cdPipelineConfigService, logger, nil, celService, transactionManager)
	commonPolicyservice := policyGovernance.NewCommonPolicyActionsService(globalPolicyManager, nil, cdPipelineConfigService, nil, nil, logger, transactionManager)
	service := NewApprovalRequestServiceImpl(
		logger,
		nil,
		nil,
		nil,
		nil,
		nil,
		resourceFilterEvalutionService,
		nil,
		nil,
		nil,
		promotionPolicyService,
		commonPolicyservice,
		nil,
		nil,
		resourceFilterEvalutionAuditService,
		transactionManager,
		nil,
		requestRepo,
		nil)

	return service, promotionPolicyService
}
