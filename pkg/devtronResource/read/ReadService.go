package read

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type ReadService interface {
	CheckIfExistingDevtronObject(id, devtronResourceSchemaId int, idType bean.IdType, identifier string) (exists bool, err error)
	CheckIfDevtronObjectExistsByIdAndIdType(id, devtronResourceSchemaId int, idType bean.IdType) (exists bool, err error)
	GetTaskRunSourceInfoForReleases(releaseIds []int) (map[int]*bean.TaskRunSource, error)
}
type ReadServiceImpl struct {
	logger                          *zap.SugaredLogger
	devtronResourceObjectRepository repository.DevtronResourceObjectRepository
}

func NewReadServiceImpl(logger *zap.SugaredLogger,
	devtronResourceObjectRepository repository.DevtronResourceObjectRepository) *ReadServiceImpl {
	return &ReadServiceImpl{
		logger:                          logger,
		devtronResourceObjectRepository: devtronResourceObjectRepository,
	}
}

// CheckIfExistingDevtronObject : this method check if it is existing object in the db.
func (impl *ReadServiceImpl) CheckIfExistingDevtronObject(id, devtronResourceSchemaId int, idType bean.IdType, identifier string) (exists bool, err error) {
	if id > 0 {
		exists, err = impl.CheckIfDevtronObjectExistsByIdAndIdType(id, devtronResourceSchemaId, idType)
		if err != nil {
			impl.logger.Errorw("error in checking object exists by id", "err", err, "id", id, "idType", idType, "devtronResourceSchemaId", devtronResourceSchemaId)
			return exists, err
		}
	} else if len(identifier) > 0 {
		exists, err = impl.devtronResourceObjectRepository.CheckIfExistByIdentifier(identifier, devtronResourceSchemaId)
		if err != nil {
			impl.logger.Errorw("error in checking object exists by identifier", "err", err, "identifier", identifier, "devtronResourceSchemaId", devtronResourceSchemaId)
			return exists, err
		}
	}
	return exists, err
}

// CheckIfDevtronObjectExistsByIdAndIdType : this method check if it is existing object in the db by resource id or oldObjectId.
func (impl *ReadServiceImpl) CheckIfDevtronObjectExistsByIdAndIdType(id, devtronResourceSchemaId int, idType bean.IdType) (exists bool, err error) {
	if idType == bean.ResourceObjectIdType {
		exists, err = impl.devtronResourceObjectRepository.CheckIfExistById(id, devtronResourceSchemaId)
		if err != nil {
			return exists, err
		}
	} else if idType == bean.OldObjectId {
		exists, err = impl.devtronResourceObjectRepository.CheckIfExistById(id, devtronResourceSchemaId)
		if err != nil {
			return exists, err
		}
	} else {
		return exists, fmt.Errorf(bean.IdTypeNotSupportedError)
	}
	return exists, err
}

func (impl *ReadServiceImpl) GetTaskRunSourceInfoForReleases(releaseIds []int) (map[int]*bean.TaskRunSource, error) {
	//getting release objects
	releases, err := impl.devtronResourceObjectRepository.FindByIds(releaseIds)
	if err != nil {
		impl.logger.Errorw("error in getting release objects", "err", err, "releaseIds", releaseIds)
		return nil, err
	}
	releaseTrackIdReleaseIdMap := make(map[int][]int)
	releaseTrackIds := make([]int, 0)
	resp := make(map[int]*bean.TaskRunSource, len(releases))
	for _, release := range releases {
		resp[release.Id] = &bean.TaskRunSource{
			Kind:           bean.DevtronResourceRelease,
			Version:        bean.DevtronResourceVersionAlpha1,
			Id:             release.Id,
			Identifier:     release.Identifier,
			ReleaseVersion: gjson.Get(release.ObjectData, bean.ReleaseResourceObjectReleaseVersionPath).String(),
			Name:           gjson.Get(release.ObjectData, bean.ResourceObjectNamePath).String(),
		}
		//getting parent dependency
		parentDep := gjson.Get(release.ObjectData, `dependencies.#(typeOfDependency=="parent")#`)
		releaseTrackDepStr := parentDep.Array()[0].String()
		releaseTrackId := int(gjson.Get(releaseTrackDepStr, bean.IdKey).Int())
		releaseTrackIds = append(releaseTrackIds, releaseTrackId)
		releaseTrackIdReleaseIdMap[releaseTrackId] = append(releaseTrackIdReleaseIdMap[releaseTrackId], release.Id)
	}
	//getting release tracks
	releaseTracks, err := impl.devtronResourceObjectRepository.FindByIds(releaseTrackIds)
	if err != nil {
		impl.logger.Errorw("error in getting releaseTrack objects", "err", err, "releaseTrackIds", releaseTrackIds)
		return nil, err
	}
	for _, releaseTrack := range releaseTracks {
		releaseTrackName := gjson.Get(releaseTrack.ObjectData, bean.ResourceObjectNamePath).String()
		fmt.Println(releaseTrackIdReleaseIdMap[releaseTrack.Id])
		for _, releaseId := range releaseTrackIdReleaseIdMap[releaseTrack.Id] {
			resp[releaseId].ReleaseTrackName = releaseTrackName
		}
	}
	return resp, nil
}
