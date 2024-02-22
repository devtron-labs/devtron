package artifactPromotionApprovalRequest

import "github.com/gorilla/mux"

type PromotionApprovalRequestRouter interface {
	InitPromotionApprovalRequestRouter(promotionApprovalRequest *mux.Router)
}

type PromotionApprovalRequestRouterImpl struct {
	promotionApprovalRequestRestHandler PromotionApprovalRequestRestHandler
}

func NewPromotionApprovalRequestRouterImpl(promotionApprovalRequestRestHandler PromotionApprovalRequestRestHandler) *PromotionApprovalRequestRouterImpl {
	return &PromotionApprovalRequestRouterImpl{promotionApprovalRequestRestHandler: promotionApprovalRequestRestHandler}
}

func (router *PromotionApprovalRequestRouterImpl) InitPromotionApprovalRequestRouter(promotionApprovalRequest *mux.Router) {
	promotionApprovalRequest.Path("/request").HandlerFunc(router.promotionApprovalRequestRestHandler.HandleArtifactPromotionRequest).
		Methods("POST")
	promotionApprovalRequest.Path("/request").HandlerFunc(router.promotionApprovalRequestRestHandler.GetByPromotionRequestId).Queries("promotionRequestId", "{promotionRequestId}").
		Methods("GET")
	//promotionApprovalRequest.Path("/material").
}
