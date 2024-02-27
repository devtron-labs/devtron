package artifactPromotionApprovalRequest

import "github.com/gorilla/mux"

type PromotionApprovalRouter interface {
	InitPromotionApprovalRouter(promotionApprovalRequest *mux.Router)
}

type PromotionApprovalRouterImpl struct {
	promotionApprovalRequestRestHandler  PromotionApprovalRequestRestHandler
	promotionApprovalMaterialRestHandler PromotionApprovalMaterialRestHandler
}

func NewPromotionApprovalRequestRouterImpl(promotionApprovalRequestRestHandler PromotionApprovalRequestRestHandler,
	promotionApprovalMaterialRestHandler PromotionApprovalMaterialRestHandler,
) *PromotionApprovalRouterImpl {
	return &PromotionApprovalRouterImpl{
		promotionApprovalRequestRestHandler:  promotionApprovalRequestRestHandler,
		promotionApprovalMaterialRestHandler: promotionApprovalMaterialRestHandler,
	}
}

func (router *PromotionApprovalRouterImpl) InitPromotionApprovalRouter(promotionApprovalRouter *mux.Router) {
	promotionApprovalRouter.Path("").HandlerFunc(router.promotionApprovalRequestRestHandler.HandleArtifactPromotionRequest).
		Methods("POST")
	promotionApprovalRouter.Path("").HandlerFunc(router.promotionApprovalRequestRestHandler.GetByPromotionRequestId).Queries("promotionRequestId", "{promotionRequestId}").
		Methods("GET")
	promotionApprovalRouter.Path("/env/approval-metadata").HandlerFunc(router.promotionApprovalRequestRestHandler.FetchAwaitingApprovalEnvListForArtifact).
		Methods("GET")
	promotionApprovalRouter.Path("/material").HandlerFunc(router.promotionApprovalMaterialRestHandler.GetArtifactsForPromotion).
		Methods("GET")

}
