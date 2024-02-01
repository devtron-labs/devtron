package lockConfiguration

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"go.uber.org/zap"
	"reflect"
	"time"
)

type LockConfigurationService interface {
	GetLockConfiguration() (*bean.LockConfigResponse, error)
	SaveLockConfiguration(*bean.LockConfigRequest, int32) error
	DeleteActiveLockConfiguration(userId int32) error
	HandleLockConfiguration(currentConfig, savedConfig string, userId int) (*bean.LockValidateErrorResponse, error)
}

type LockConfigurationServiceImpl struct {
	logger                      *zap.SugaredLogger
	lockConfigurationRepository LockConfigurationRepository
	userService                 user.UserService
	mergeUtil                   util.MergeUtil
}

func NewLockConfigurationServiceImpl(logger *zap.SugaredLogger,
	lockConfigurationRepository LockConfigurationRepository,
	userService user.UserService,
	mergeUtil util.MergeUtil) *LockConfigurationServiceImpl {
	return &LockConfigurationServiceImpl{
		logger:                      logger,
		lockConfigurationRepository: lockConfigurationRepository,
		userService:                 userService,
		mergeUtil:                   mergeUtil,
	}
}

func (impl LockConfigurationServiceImpl) GetLockConfiguration() (*bean.LockConfigResponse, error) {
	impl.logger.Infow("Getting active lock configuration")

	lockConfigDto, err := impl.lockConfigurationRepository.GetActiveLockConfig()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("Error in getting active lock config", "err", err)
		return nil, err
	}
	if lockConfigDto == nil {
		return &bean.LockConfigResponse{}, nil
	}
	lockConfig := lockConfigDto.ConvertDBDtoToResponse()
	return lockConfig, nil
}

func (impl LockConfigurationServiceImpl) SaveLockConfiguration(lockConfig *bean.LockConfigRequest, createdBy int32) error {
	// pass tx
	dbConnection := impl.lockConfigurationRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// Delete Active configuration
	err = impl.lockConfigurationRepository.DeleteActiveLockConfigs(int(createdBy))
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in deleting current active lock config", "err", err)
		return err
	}

	newLockConfigDto := lockConfig.ConvertRequestToDBDto()
	newLockConfigDto.AuditLog = sql.NewDefaultAuditLog(createdBy)

	err = impl.lockConfigurationRepository.Create(newLockConfigDto, tx)
	if err != nil {
		impl.logger.Errorw("error while saving lockConfiguration", "error", err)
		return err
	}

	// commit TX
	// TODO log
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("Error in committing tx", "err", err)
		return err
	}
	return err
}

func (impl LockConfigurationServiceImpl) DeleteActiveLockConfiguration(userId int32) error {
	lockConfigDto, err := impl.lockConfigurationRepository.GetActiveLockConfig()
	if err != nil {
		impl.logger.Errorw("Error in getting active lock config", "err", err)
		return err
	}
	dbConnection := impl.lockConfigurationRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}

	// Rollback tx on error.
	defer tx.Rollback()

	lockConfigDto.Active = false
	lockConfigDto.UpdatedOn = time.Now()
	lockConfigDto.UpdatedBy = userId
	err = impl.lockConfigurationRepository.Update(lockConfigDto, tx)
	if err != nil {
		impl.logger.Errorw("Error in updating lock config", "lockConfigId", lockConfigDto.Id, "err", err)
		return err
	}
	// commit TX
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("Error in committing tx", "err", err)
		return err
	}

	return nil
}

func (impl LockConfigurationServiceImpl) HandleLockConfiguration(currentConfig, savedConfig string, userId int) (*bean.LockValidateErrorResponse, error) {

	isSuperAdmin, err := impl.userService.IsSuperAdminForDevtronManaged(userId)

	if err != nil || isSuperAdmin {
		return nil, err
	}

	var savedConfigMap map[string]interface{}
	var currentConfigMap map[string]interface{}

	lockConfig, err := impl.GetLockConfiguration()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(savedConfig), &savedConfigMap)
	if err != nil {
		impl.logger.Errorw("Error in umMarshal data", "err", err, "savedConfig", savedConfig)
		return nil, err
	}
	err = json.Unmarshal([]byte(currentConfig), &currentConfigMap)
	if err != nil {
		impl.logger.Errorw("Error in umMarshal data", "err", err, "currentConfig", currentConfig)
		return nil, err
	}
	// TODO name for disableSaveEligibleChanges, allChanges
	allChanges, disableSaveEligibleChanges := getDiffJson(savedConfigMap, currentConfigMap)
	var isLockConfigError bool
	if lockConfig.ContainAllowedPaths {
		// Will add in v2 of this feature
	} else {
		isLockConfigError = checkLockedChanges(currentConfig, savedConfig, lockConfig.Paths)
	}
	if isLockConfigError {
		//  rename lockedOverride Diff json byte array
		lockedOverride, _ := json.Marshal(allChanges)
		lockConfigErrorResponse := bean.GetLockConfigErrorResponse(string(lockedOverride), "", "", "", disableSaveEligibleChanges)
		return lockConfigErrorResponse, nil
	}
	return nil, nil
}

// TODO add comments
func checkLockedChanges(currentConfig, savedConfig string, lockedConfigJsonPaths []string) bool {
	// TODO RENAME obj odj1 x ys ys1
	obj, err := oj.ParseString(currentConfig)
	if err != nil {
		return false
	}
	obj1, err := oj.ParseString(savedConfig)
	if err != nil {
		return false
	}
	for _, lockedConfigJsonPath := range lockedConfigJsonPaths {
		x, err := jp.ParseString(lockedConfigJsonPath)
		if err != nil {
			return false
		}
		ys := x.Get(obj)
		ys1 := x.Get(obj1)
		if !reflect.DeepEqual(ys, ys1) {
			return true
		}
	}
	return false
}

func checkForLockedArray(savedConfigMap, currentConfigMap []interface{}) []interface{} {
	var lockedArray []interface{}
	var key int
	for key, _ = range savedConfigMap {
		if key >= len(currentConfigMap) {
			lockedArray = append(lockedArray, savedConfigMap[key])
			continue
		}
		if !reflect.DeepEqual(savedConfigMap[key], currentConfigMap[key]) {
			switch reflect.TypeOf(savedConfigMap[key]).Kind() {
			case reflect.Map:
				locked, _ := getDiffJson(savedConfigMap[key].(map[string]interface{}), currentConfigMap[key].(map[string]interface{}))
				if locked != nil && len(locked) != 0 {
					lockedArray = append(lockedArray, locked)
				}
			case reflect.Array, reflect.Slice:
				locked := checkForLockedArray(savedConfigMap[key].([]interface{}), currentConfigMap[key].([]interface{}))
				if locked != nil && len(locked) != 0 {
					lockedArray = append(lockedArray, locked)
				}
			default:
				lockedArray = append(lockedArray, currentConfigMap[key])
			}
		}
	}
	for key1, _ := range currentConfigMap {
		if key1 <= key {
			continue
		}
		lockedArray = append(lockedArray, currentConfigMap[key1])
	}

	return lockedArray
}

func getMinOperationsToChangeArray(word1 []interface{}, word2 []interface{}, i int, j int) (int, []interface{}, []interface{}, []interface{}) {
	var added, modified, deleted []interface{}
	if i < 0 {
		return j + 1, word2[0 : j+1], modified, deleted
	}
	if j < 0 {
		return i + 1, added, modified, word1[0 : i+1]
	}
	if reflect.DeepEqual(word1[i], word2[j]) {
		val, added, modified, deleted := getMinOperationsToChangeArray(word1, word2, i-1, j-1)
		return val, added, modified, deleted
	}
	const MaxUint = ^uint(0)

	ans := int(MaxUint >> 1)
	insert, addedI, modifiedI, deletedI := getMinOperationsToChangeArray(word1, word2, i, j-1) // insert
	if 1+insert < ans {
		ans = 1 + insert
	}
	deletedV, addedD, modifiedD, deletedD := getMinOperationsToChangeArray(word1, word2, i-1, j) //delete
	if 1+deletedV < ans {
		ans = 1 + deletedV
	}
	modifiedV, addedM, modifiedM, deletedM := getMinOperationsToChangeArray(word1, word2, i-1, j-1) //replace
	if 1+modifiedV < ans {
		ans = 1 + modifiedV
	}
	if insert < deletedV {
		if insert < modifiedV {
			added = append(added, word2[j])
			added = append(added, addedI...)
			modified = append(modified, modifiedI...)
			deleted = append(deleted, deletedI...)
		} else {
			val := getModifiedValue(word1, word2, i, j)
			modified = append(modified, val)
			added = append(added, addedM...)
			modified = append(modified, modifiedM...)
			deleted = append(deleted, deletedM...)
		}
	} else {
		if deletedV < modifiedV {
			deleted = append(deleted, word1[i])
			added = append(added, addedD...)
			modified = append(modified, modifiedD...)
			deleted = append(deleted, deletedD...)
		} else {
			val := getModifiedValue(word1, word2, i, j)
			modified = append(modified, val)
			added = append(added, addedM...)
			modified = append(modified, modifiedM...)
			deleted = append(deleted, deletedM...)
		}
	}
	return ans, added, modified, deleted
}

func getModifiedValue(word1 []interface{}, word2 []interface{}, i int, j int) interface{} {
	switch reflect.TypeOf(word1[i]).Kind() {
	case reflect.Map:
		savedConfig := copyMap(word1[i].(map[string]interface{}))
		currentConfig := copyMap(word2[j].(map[string]interface{}))
		locked, _ := getDiffJson(savedConfig, currentConfig)
		return locked
	case reflect.Array, reflect.Slice:
		locked := getArrayDiff(word1[i].([]interface{}), word2[j].([]interface{}))
		return locked
	default:
		return word2[j]
	}
}

func getArrayDiff(word1 []interface{}, word2 []interface{}) []interface{} {
	l1 := len(word1)
	l2 := len(word2)
	_, added, modified, deleted := getMinOperationsToChangeArray(word1, word2, l1-1, l2-1)
	var lockedArray []interface{}
	lockedArray = append(lockedArray, added...)
	lockedArray = append(lockedArray, modified...)
	lockedArray = append(lockedArray, deleted...)
	return lockedArray
}

// TODO ADD comment for return
func getDiffJson(savedConfigMap, currentConfigMap map[string]interface{}) (map[string]interface{}, bool) {
	// Store all the changes
	lockedMap := make(map[string]interface{})
	disableSaveEligibleChanges := false
	for key, _ := range savedConfigMap {

		currMapVal, ok := currentConfigMap[key]
		if !ok || currMapVal == nil {
			if savedConfigMap[key] != nil {
				lockedMap[key] = nil
			}
			continue
		}

		if savedConfigMap[key] == nil {
			lockedMap[key] = currentConfigMap[key]
			continue
		}

		if !reflect.DeepEqual(savedConfigMap[key], currentConfigMap[key]) {
			switch reflect.TypeOf(savedConfigMap[key]).Kind() {
			case reflect.Map:
				locked, isSaveEligibleChangesDisabled := getDiffJson(savedConfigMap[key].(map[string]interface{}), currentConfigMap[key].(map[string]interface{}))
				if locked != nil && len(locked) != 0 {
					lockedMap[key] = locked
				}
				if isSaveEligibleChangesDisabled {
					disableSaveEligibleChanges = true
				}
			case reflect.Array, reflect.Slice:
				locked := getArrayDiff(savedConfigMap[key].([]interface{}), currentConfigMap[key].([]interface{}))
				if locked != nil && len(locked) != 0 {
					lockedMap[key] = locked
				}
				disableSaveEligibleChanges = true
			default:
				lockedMap[key] = currentConfigMap[key]
			}
		} else {
			delete(currentConfigMap, key)
			continue
		}
		switch reflect.TypeOf(currentConfigMap[key]).Kind() {
		case reflect.Map:
			if len(currentConfigMap[key].(map[string]interface{})) == 0 {
				if len(currentConfigMap[key].(map[string]interface{})) == 0 {
					delete(currentConfigMap, key)
				}
			}
		default:
			delete(currentConfigMap, key)
		}
	}
	// Append for the new added keys
	for key, val := range currentConfigMap {
		lockedMap[key] = val
		delete(currentConfigMap, key)
	}
	return lockedMap, disableSaveEligibleChanges
}

func copyMap(map1 map[string]interface{}) map[string]interface{} {
	map2 := make(map[string]interface{}, len(map1))
	for k, v := range map1 {
		map2[k] = v
	}
	return map2
}
