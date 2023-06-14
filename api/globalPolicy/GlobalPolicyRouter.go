package globalPolicy

import (
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type GlobalPolicyRouter interface {
	InitGlobalPolicyRouter(router *mux.Router)
}

type GlobalPolicyRouterImpl struct {
	logger                  *zap.SugaredLogger
	globalPolicyRestHandler GlobalPolicyRestHandler
}

func NewGlobalPolicyRouterImpl(logger *zap.SugaredLogger,
	globalPolicyRestHandler GlobalPolicyRestHandler) *GlobalPolicyRouterImpl {
	return &GlobalPolicyRouterImpl{
		logger:                  logger,
		globalPolicyRestHandler: globalPolicyRestHandler,
	}
}

func (router *GlobalPolicyRouterImpl) InitGlobalPolicyRouter(policyRouter *mux.Router) {
	policyRouter.Path("/ci-pipeline").
		HandlerFunc(router.globalPolicyRestHandler.GetMandatoryPluginsForACiPipeline).Methods("GET")
	policyRouter.Path("/ci-pipeline/block-state").
		HandlerFunc(router.globalPolicyRestHandler.GetOnlyBlockageStateOfACiPipeline).Methods("GET")
	policyRouter.Path("/all").
		HandlerFunc(router.globalPolicyRestHandler.GetAllGlobalPolicies).Methods("GET")
	policyRouter.Path("/offending-pipeline/wf/tree/list").
		HandlerFunc(router.globalPolicyRestHandler.GetPolicyOffendingPipelinesWfTree).Methods("GET")
	policyRouter.Path("/{id}").
		HandlerFunc(router.globalPolicyRestHandler.GetById).Methods("GET")
	policyRouter.Path("/{id}").
		HandlerFunc(router.globalPolicyRestHandler.DeleteGlobalPolicy).Methods("DELETE")
	policyRouter.Path("").
		HandlerFunc(router.globalPolicyRestHandler.CreateGlobalPolicy).Methods("POST")
	policyRouter.Path("").
		HandlerFunc(router.globalPolicyRestHandler.UpdateGlobalPolicy).Methods("PUT")
}
