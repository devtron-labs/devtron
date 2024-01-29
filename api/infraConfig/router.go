package infraConfig

import "github.com/gorilla/mux"

type InfraConfigRouter interface {
	InitInfraConfigRouter(configRouter *mux.Router)
}

type InfraConfigRouterImpl struct {
	infraConfigRestHandler InfraConfigRestHandler
}

func NewInfraProfileRouterImpl(infraConfigRestHandler InfraConfigRestHandler) *InfraConfigRouterImpl {
	return &InfraConfigRouterImpl{
		infraConfigRestHandler: infraConfigRestHandler,
	}
}

func (impl *InfraConfigRouterImpl) InitInfraConfigRouter(configRouter *mux.Router) {
	configRouter.Path("/profile/{name}").
		HandlerFunc(impl.infraConfigRestHandler.GetProfile).
		Methods("GET")

	configRouter.Path("/profile/{name}").
		HandlerFunc(impl.infraConfigRestHandler.UpdateInfraProfile).
		Methods("PUT")

	configRouter.Path("/profile/{name}").
		HandlerFunc(impl.infraConfigRestHandler.DeleteProfile).
		Methods("DELETE")

	configRouter.Path("/profile").
		HandlerFunc(impl.infraConfigRestHandler.CreateProfile).
		Methods("POST")

	configRouter.Path("/list/profile").
		// Queries("search", "{profileNameLike}").
		HandlerFunc(impl.infraConfigRestHandler.GetProfileList).
		Methods("GET")

	configRouter.Path("/list/identifier/{identifierType}").
		// Queries("search", "{identifierNameLike}",
		// 	"sort", "{sortOrder}",
		// 	"profileName", "{profileName}",
		// 	"size", "{size}",
		// 	"offset", "{offset}").
		HandlerFunc(impl.infraConfigRestHandler.GetIdentifierList).
		Methods("GET")

	configRouter.Path("/identifier/{identifierType}/apply").
		HandlerFunc(impl.infraConfigRestHandler.ApplyProfileToIdentifiers).
		Methods("POST")

	// todo: @gireesh create a lite weight api autocomplete api for profile name and id data
}
