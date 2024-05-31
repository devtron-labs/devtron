/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package util

import (
	"context"
	"time"

	"github.com/devtron-labs/devtron/pkg/auth/user"

	"github.com/caarlos0/env"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

type TokenCache struct {
	cache           *cache.Cache
	logger          *zap.SugaredLogger
	aCDAuthConfig   *ACDAuthConfig
	userAuthService user.UserAuthService
}

func NewTokenCache(logger *zap.SugaredLogger, aCDAuthConfig *ACDAuthConfig, userAuthService user.UserAuthService) *TokenCache {
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
	impl.logger.Debugw("building acd context", "found", found)
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
	ACDUsername                      string `env:"ACD_USERNAME" envDefault:"admin"`
	ACDPassword                      string `env:"ACD_PASSWORD" `
	ACDConfigMapName                 string `env:"ACD_CM" envDefault:"argocd-cm"`
	ACDConfigMapNamespace            string `env:"ACD_NAMESPACE" envDefault:"devtroncd"`
	GitOpsSecretName                 string `env:"GITOPS_SECRET_NAME" envDefault:"devtron-gitops-secret"`
	ResourceListForReplicas          string `env:"RESOURCE_LIST_FOR_REPLICAS" envDefault:"Deployment,Rollout,StatefulSet,ReplicaSet"`
	ResourceListForReplicasBatchSize int    `env:"RESOURCE_LIST_FOR_REPLICAS_BATCH_SIZE" envDefault:"5"`
}

func GetACDAuthConfig() (*ACDAuthConfig, error) {
	cfg := &ACDAuthConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, err
}
