package devtronResource

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"go.uber.org/zap"
)

type DevtronResourceSearchableKeyService interface {
	GetAllSearchableKeyNameIdMap() map[bean.DevtronResourceSearchableKeyName]int
	GetAllSearchableKeyIdNameMap() map[int]bean.DevtronResourceSearchableKeyName
}

type DevtronResourceSearchableKeyServiceImpl struct {
	logger                                 *zap.SugaredLogger
	devtronResourceSearchableKeyRepository repository.DevtronResourceSearchableKeyRepository
	searchableKeyNameIdMap                 map[bean.DevtronResourceSearchableKeyName]int
	searchableKeyIdNameMap                 map[int]bean.DevtronResourceSearchableKeyName
}

func NewDevtronResourceSearchableKeyServiceImpl(logger *zap.SugaredLogger,
	devtronResourceSearchableKeyRepository repository.DevtronResourceSearchableKeyRepository) (*DevtronResourceSearchableKeyServiceImpl, error) {
	impl := &DevtronResourceSearchableKeyServiceImpl{
		logger:                                 logger,
		devtronResourceSearchableKeyRepository: devtronResourceSearchableKeyRepository,
	}
	searchableKeyNameIdMap, searchableKeyIdNameMap, err := impl.getAllSearchableKeyNameIdAndIdNameMaps()
	if err != nil {
		impl.logger.Errorw("error, GetAllSearchableKeyNameIdAndIdNameMaps", "err", err)
		return nil, err
	}
	impl.searchableKeyNameIdMap = searchableKeyNameIdMap
	impl.searchableKeyIdNameMap = searchableKeyIdNameMap
	return impl, nil
}

func (impl *DevtronResourceSearchableKeyServiceImpl) getAllSearchableKeyNameIdAndIdNameMaps() (map[bean.DevtronResourceSearchableKeyName]int,
	map[int]bean.DevtronResourceSearchableKeyName, error) {
	//getting searchable keys from db
	searchableKeys, err := impl.devtronResourceSearchableKeyRepository.GetAll()
	if err != nil {
		impl.logger.Errorw("error in getting all attributes from db", "err", err)
		return nil, nil, err
	}
	searchableKeyNameIdMap := make(map[bean.DevtronResourceSearchableKeyName]int)
	searchableKeyIdNameMap := make(map[int]bean.DevtronResourceSearchableKeyName)
	for _, searchableKey := range searchableKeys {
		searchableKeyNameIdMap[searchableKey.Name] = searchableKey.Id
		searchableKeyIdNameMap[searchableKey.Id] = searchableKey.Name
	}
	return searchableKeyNameIdMap, searchableKeyIdNameMap, nil
}

func (impl *DevtronResourceSearchableKeyServiceImpl) GetAllSearchableKeyNameIdMap() map[bean.DevtronResourceSearchableKeyName]int {
	return impl.searchableKeyNameIdMap
}

func (impl *DevtronResourceSearchableKeyServiceImpl) GetAllSearchableKeyIdNameMap() map[int]bean.DevtronResourceSearchableKeyName {
	return impl.searchableKeyIdNameMap
}
