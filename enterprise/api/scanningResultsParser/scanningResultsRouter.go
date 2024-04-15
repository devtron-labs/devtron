package scanningResultsParser

import "github.com/gorilla/mux"

type ScanningResultRouter interface {
	InitScanningResultRouter(configRouter *mux.Router)
}

type ScanningResultRouterImpl struct {
	ScanningResultRestHandler ScanningResultRestHandler
}

func NewScanningResultRouterImpl(ScanningResultRestHandler ScanningResultRestHandler) *ScanningResultRouterImpl {
	return &ScanningResultRouterImpl{ScanningResultRestHandler: ScanningResultRestHandler}
}

func (router *ScanningResultRouterImpl) InitScanningResultRouter(configRouter *mux.Router) {
	configRouter.Path("/scanResults").HandlerFunc(router.ScanningResultRestHandler.ScanResults).Methods("GET")
}
