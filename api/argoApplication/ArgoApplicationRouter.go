package argoApplication

import "github.com/gorilla/mux"

type ArgoApplicationRouter interface {
	InitArgoApplicationRouter(argoApplicationRouter *mux.Router)
}

type ArgoApplicationRouterImpl struct {
	argoApplicationRestHandler ArgoApplicationRestHandler
}

func NewArgoApplicationRouterImpl(argoApplicationRestHandler ArgoApplicationRestHandler) *ArgoApplicationRouterImpl {
	return &ArgoApplicationRouterImpl{
		argoApplicationRestHandler: argoApplicationRestHandler,
	}
}

func (impl *ArgoApplicationRouterImpl) InitArgoApplicationRouter(argoApplicationRouter *mux.Router) {
	argoApplicationRouter.Path("").
		Methods("GET").
		HandlerFunc(impl.argoApplicationRestHandler.ListApplications)

	argoApplicationRouter.Path("/detail").
		Methods("GET").
		HandlerFunc(impl.argoApplicationRestHandler.GetApplicationDetail)
}
