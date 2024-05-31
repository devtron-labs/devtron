/*
 * Copyright (c) 2024. Devtron Inc.
 */

package artifactPromotionApprovalRequest

import "github.com/gorilla/mux"

type Router interface {
	InitPromotionApprovalRouter(promotionApprovalRequest *mux.Router)
}

type RouterImpl struct {
	promotionApprovalRequestRestHandler  RestHandler
	promotionApprovalMaterialRestHandler MaterialRestHandler
}

func NewRouterImpl(promotionApprovalRequestRestHandler RestHandler,
	promotionApprovalMaterialRestHandler MaterialRestHandler,
) *RouterImpl {
	return &RouterImpl{
		promotionApprovalRequestRestHandler:  promotionApprovalRequestRestHandler,
		promotionApprovalMaterialRestHandler: promotionApprovalMaterialRestHandler,
	}
}

func (router *RouterImpl) InitPromotionApprovalRouter(promotionApprovalRouter *mux.Router) {
	promotionApprovalRouter.Path("").HandlerFunc(router.promotionApprovalRequestRestHandler.HandleArtifactPromotionRequest).
		Methods("POST")
	promotionApprovalRouter.Path("/env/approval-metadata").HandlerFunc(router.promotionApprovalRequestRestHandler.FetchAwaitingApprovalEnvListForArtifact).
		Methods("GET")
	promotionApprovalRouter.Path("/material").HandlerFunc(router.promotionApprovalMaterialRestHandler.GetArtifactsForPromotion).
		Methods("GET")
	promotionApprovalRouter.Path("/env/list").HandlerFunc(router.promotionApprovalRequestRestHandler.FetchWorkflowPromoteNodeList).
		Methods("GET")

}
