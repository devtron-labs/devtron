package devtronResource

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
)

func (impl *DevtronResourceServiceImpl) getCdPipelineIdentifierByPipelineId(pipelineId int) (string, error) {
	pipeline, err := impl.pipelineRepository.FindById(pipelineId)
	if err != nil {
		impl.logger.Errorw("error in finding pipeline by pipeline id", "err", err, "pipelineId", pipelineId)
		return "", err
	}
	return fmt.Sprintf("%s-%s", pipeline.App.AppName, pipeline.Name), nil
}

func (impl *DevtronResourceServiceImpl) buildIdentifierForCdPipelineResourceObj(object *repository.DevtronResourceObject) (string, error) {
	return impl.getCdPipelineIdentifierByPipelineId(object.OldObjectId)
}
