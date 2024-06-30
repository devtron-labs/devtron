/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type PProfRouter interface {
	initPProfRouter(router *mux.Router)
}

type PProfRouterImpl struct {
	logger           *zap.SugaredLogger
	pprofRestHandler restHandler.PProfRestHandler
}

func NewPProfRouter(logger *zap.SugaredLogger,
	pprofRestHandler restHandler.PProfRestHandler) *PProfRouterImpl {
	return &PProfRouterImpl{
		logger:           logger,
		pprofRestHandler: pprofRestHandler,
	}
}

func (ppr PProfRouterImpl) initPProfRouter(router *mux.Router) {
	router.HandleFunc("/", ppr.pprofRestHandler.Index)
	router.HandleFunc("/cmdline", ppr.pprofRestHandler.Cmdline)
	router.HandleFunc("/profile", ppr.pprofRestHandler.Profile)
	router.HandleFunc("/symbol", ppr.pprofRestHandler.Symbol)
	router.HandleFunc("/trace", ppr.pprofRestHandler.Trace)
	router.HandleFunc("/goroutine", ppr.pprofRestHandler.Goroutine)
	router.HandleFunc("/threadcreate", ppr.pprofRestHandler.Threadcreate)
	router.HandleFunc("/heap", ppr.pprofRestHandler.Heap)
	router.HandleFunc("/block", ppr.pprofRestHandler.Block)
	router.HandleFunc("/mutex", ppr.pprofRestHandler.Mutex)
	router.HandleFunc("/allocs", ppr.pprofRestHandler.Allocs)
}
