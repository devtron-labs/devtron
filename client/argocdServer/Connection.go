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

package argocdServer

import (
	"fmt"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"log"

	"github.com/argoproj/argo-cd/v2/util/settings"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

func init() {
	grpc_prometheus.EnableClientHandlingTimeHistogram()
}

type ArgoCdConnection interface {
	GetConnection(token string) *grpc.ClientConn
}
type ArgoCdConnectionImpl struct {
	logger           *zap.SugaredLogger
	settingsManager  *settings.SettingsManager
	moduleRepository moduleRepo.ModuleRepository
}

func NewArgoCdConnectionImpl(Logger *zap.SugaredLogger, settingsManager *settings.SettingsManager,
	moduleRepository moduleRepo.ModuleRepository) (*ArgoCdConnectionImpl, error) {
	argoUserServiceImpl := &ArgoCdConnectionImpl{
		logger:           Logger,
		settingsManager:  settingsManager,
		moduleRepository: moduleRepository,
	}
	return argoUserServiceImpl, nil
}

func (impl *ArgoCdConnectionImpl) GetConnection(token string) *grpc.ClientConn {
	conf, err := GetConfig()
	if err != nil {
		impl.logger.Errorw("error on get acd connection", "err", err)
		log.Fatal(err)
	}
	module, err := impl.moduleRepository.FindOne("acd")
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error on get acd connection", "err", err)
		log.Fatal(err)
	}
	if module == nil || module.Status != "installed" {
		impl.logger.Errorw("error on get acd connection", "err", err)
		log.Fatal(err)
	}
	settings, err := impl.settingsManager.GetSettings()
	if err != nil {
		impl.logger.Errorw("error on get acd connection", "err", err)
		log.Fatal(err)
	}
	var option []grpc.DialOption
	option = append(option, grpc.WithTransportCredentials(GetTLS(settings.Certificate)))
	if len(token) > 0 {
		option = append(option, grpc.WithPerRPCCredentials(TokenAuth{token: token}))
	}
	option = append(option, grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor), grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor))
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", conf.Host, conf.Port), option...)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}
