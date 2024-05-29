/*
 * Copyright (c) 2024. Devtron Inc.
 */

package read

import (
	"github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ReadService interface {
	GetAllGlobalPoliciesByTypeAndVersion(policyOf bean.GlobalPolicyType, policyVersion bean.GlobalPolicyVersion) ([]*bean.GlobalPolicyBaseModel, error)
}

type ReadServiceImpl struct {
	logger                 *zap.SugaredLogger
	globalPolicyRepository repository.GlobalPolicyRepository
}

func NewReadServiceImpl(logger *zap.SugaredLogger,
	globalPolicyRepository repository.GlobalPolicyRepository) *ReadServiceImpl {
	return &ReadServiceImpl{
		logger:                 logger,
		globalPolicyRepository: globalPolicyRepository,
	}
}

func (impl *ReadServiceImpl) GetAllGlobalPoliciesByTypeAndVersion(policyOf bean.GlobalPolicyType,
	policyVersion bean.GlobalPolicyVersion) ([]*bean.GlobalPolicyBaseModel, error) {
	// getting all global policy entries
	globalPolicies, err := impl.globalPolicyRepository.GetAllByPolicyOfAndVersion(policyOf, policyVersion)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting all policies", "err", err, "policyOf", policyOf, "policyVersion", policyVersion)
		return nil, err
	}
	globalPolicyDtos := make([]*bean.GlobalPolicyBaseModel, 0, len(globalPolicies))
	for _, globalPolicy := range globalPolicies {
		globalPolicyDtos = append(globalPolicyDtos, globalPolicy.GetGlobalPolicyBaseModel())
	}
	return globalPolicyDtos, nil
}
