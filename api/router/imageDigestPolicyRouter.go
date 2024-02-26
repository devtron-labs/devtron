package router

import (
	"github.com/devtron-labs/devtron/api/restHandler/imageDigestPolicy"
	"github.com/gorilla/mux"
)

type ImageDigestPolicyRouter interface {
	initImageDigestPolicyRouter(imageDigestPolicyRouter *mux.Router)
}

type ImageDigestPolicyRouterImpl struct {
	ImageDigestPolicyRestHandler imageDigestPolicy.ImageDigestPolicyRestHandler
}

func NewImageDigestPolicyRouterImpl(
	ImageDigestPolicyRestHandler imageDigestPolicy.ImageDigestPolicyRestHandler,
) *ImageDigestPolicyRouterImpl {
	return &ImageDigestPolicyRouterImpl{
		ImageDigestPolicyRestHandler: ImageDigestPolicyRestHandler,
	}
}

func (router ImageDigestPolicyRouterImpl) initImageDigestPolicyRouter(imageDigestPolicyRouter *mux.Router) {
	imageDigestPolicyRouter.Path("").
		Methods("GET").HandlerFunc(router.ImageDigestPolicyRestHandler.GetAllImageDigestPolicies)
	imageDigestPolicyRouter.Path("").
		Methods("POST").HandlerFunc(router.ImageDigestPolicyRestHandler.SaveOrUpdateImageDigestPolicy)

}
