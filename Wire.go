//go:build wireinject
// +build wireinject

/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	cloudProviderIdentifier "github.com/devtron-labs/common-lib/cloud-provider-identifier"
	pubsub1 "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/commonWireset"
	"github.com/devtron-labs/devtron/internals/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {

	wire.Build(
		// ----- wireset start
		sql.PgSqlWireSet,
		AuthWireSet,
		util.NewSugardLogger,
		util.NewHttpClient,
		NewApp,
		util.IntValidator,
		pubsub1.NewPubSubClientServiceImpl,
		cloudProviderIdentifier.NewProviderIdentifierServiceImpl,
		wire.Bind(new(cloudProviderIdentifier.ProviderIdentifierService), new(*cloudProviderIdentifier.ProviderIdentifierServiceImpl)),
		commonWireset.CommonWireSet,
	)
	return &App{}, nil
}
