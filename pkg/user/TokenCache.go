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

package user

import (
	"context"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"time"
)

type TokenCache struct {
	cache           *cache.Cache
	logger          *zap.SugaredLogger
	aCDAuthConfig   *ACDAuthConfig
	userAuthService UserAuthService
}

func NewTokenCache(logger *zap.SugaredLogger, aCDAuthConfig *ACDAuthConfig, userAuthService UserAuthService) *TokenCache {
	tokenCache := &TokenCache{
		cache:           cache.New(cache.NoExpiration, 5*time.Minute),
		logger:          logger,
		aCDAuthConfig:   aCDAuthConfig,
		userAuthService: userAuthService,
	}
	return tokenCache
}
func (impl *TokenCache) BuildACDSynchContext() (acdContext context.Context, err error) {
	token, found := impl.cache.Get("token")
	if !found {
		token, err := impl.userAuthService.HandleLogin(impl.aCDAuthConfig.ACDUsername, impl.aCDAuthConfig.ACDPassword)
		if err != nil {
			impl.logger.Errorw("error while acd login", "err", err)
			return nil, err
		}
		impl.cache.Set("token", token, cache.NoExpiration)
	}
	token, _ = impl.cache.Get("token")
	ctx := context.Background()
	ctx = context.WithValue(ctx, "token", token)
	return ctx, nil
}

type ACDAuthConfig struct {
	ACDUsername           string `env:"ACD_USERNAME" `
	ACDPassword           string `env:"ACD_PASSWORD" `
	ACDConfigMapName      string `env:"ACD_CM" `
	ACDConfigMapNamespace string `env:"ACD_NAMESPACE" `
}

func GetACDAuthConfig() (*ACDAuthConfig, error) {
	cfg := &ACDAuthConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	if len(cfg.ACDPassword) == 0 {
		return nil, fmt.Errorf("ACD_PASSWORD is not present in environment")
	}
	if len(cfg.ACDUsername) == 0 {
		return nil, fmt.Errorf("ACD_USERNAME is not present in environment")
	}
	return cfg, err
}
