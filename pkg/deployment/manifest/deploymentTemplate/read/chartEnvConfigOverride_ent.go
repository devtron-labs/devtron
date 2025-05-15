package read

import "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/bean"

type EnvConfigOverrideServiceEnt interface {
}

func (impl EnvConfigOverrideReadServiceImpl) getOverrideDataWithUpdatedPatchDataUnResolved(overrideDTO *bean.EnvConfigOverride, appId int) (*bean.EnvConfigOverride, error) {
	return overrideDTO, nil
}
