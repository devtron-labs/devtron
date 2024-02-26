package artifactPromotionApprovalRequest

import "github.com/gorilla/mux"

type PromotionApprovalRouter interface {
	InitPromotionApprovalRouter(promotionApprovalRequest *mux.Router)
}

type PromotionApprovalRouterImpl struct {
	promotionApprovalRequestRestHandler PromotionApprovalRequestRestHandler
}

func NewPromotionApprovalRequestRouterImpl(promotionApprovalRequestRestHandler PromotionApprovalRequestRestHandler) *PromotionApprovalRouterImpl {
	return &PromotionApprovalRouterImpl{promotionApprovalRequestRestHandler: promotionApprovalRequestRestHandler}
}

func (router *PromotionApprovalRouterImpl) InitPromotionApprovalRouter(promotionApprovalRouter *mux.Router) {
	promotionApprovalRouter.Path("").HandlerFunc(router.promotionApprovalRequestRestHandler.HandleArtifactPromotionRequest).
		Methods("POST")
	promotionApprovalRouter.Path("").HandlerFunc(router.promotionApprovalRequestRestHandler.GetByPromotionRequestId).Queries("promotionRequestId", "{promotionRequestId}").
		Methods("GET")
	promotionApprovalRouter.Path("/env/metadata").HandlerFunc(router.promotionApprovalRequestRestHandler.FetchAwaitingApprovalEnvListForArtifact).
		Methods("GET")

}
