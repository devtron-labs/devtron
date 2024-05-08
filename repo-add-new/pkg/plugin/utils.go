package plugin

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
)

func getStageType(stageTypeReq string) (int, error) {
	var stageType int
	switch stageTypeReq {
	case repository.CI_STAGE_TYPE:
		stageType = repository.CI
	case repository.CD_STAGE_TYPE:
		stageType = repository.CD
	case repository.CI_CD_STAGE_TYPE:
		stageType = repository.CI_CD
	default:
		return 0, errors.New("stage type not recognised, please add valid stage type in query parameter")
	}
	return stageType, nil
}
