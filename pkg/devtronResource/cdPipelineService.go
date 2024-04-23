package devtronResource

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/devtronResource/bean"
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

func (impl *DevtronResourceServiceImpl) getMapOfCdPipelineMetadata(pipelineIdsToGetMetadata []int) (map[int]interface{}, error) {
	mapOfCdPipelinesMetadata := make(map[int]interface{})
	var err error
	var pipelineMetadataDtos []*bean.EnvironmentForDependency
	if len(pipelineIdsToGetMetadata) > 0 {
		pipelineMetadataDtos, err = impl.appListingRepository.FetchDependencyMetadataByPipelineIds(pipelineIdsToGetMetadata)
		if err != nil {
			impl.logger.Errorw("error in getting cd pipelines by ids", "err", err, "ids", pipelineIdsToGetMetadata)
			return nil, err
		}
	}
	for _, pipelineMetadata := range pipelineMetadataDtos {
		mapOfCdPipelinesMetadata[pipelineMetadata.PipelineId] = pipelineMetadata
	}
	return mapOfCdPipelinesMetadata, nil
}

func updateCdPipelineMetaDataInDependencyObj(oldObjectId int, metaDataObj *bean2.DependencyMetaDataBean) interface{} {
	return metaDataObj.MapOfCdPipelinesMetadata[oldObjectId]
}
