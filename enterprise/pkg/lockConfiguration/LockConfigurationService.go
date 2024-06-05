/*
 * Copyright (c) 2024. Devtron Inc.
 */

package lockConfiguration

import (
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"go.uber.org/zap"
	"reflect"
	"sort"
	"time"
)

type LockConfigurationService interface {
	GetLockConfiguration() (*bean.LockConfigResponse, error)
	SaveLockConfiguration(*bean.LockConfigRequest, int32) error
	DeleteActiveLockConfiguration(userId int32) error
	HandleLockConfiguration(currentConfig, savedConfig string, userId int) (*bean.LockValidateErrorResponse, error)
}

type LockConfigurationServiceConfig struct {
	ArrayDiffMemoization bool `env:"ARRAY_DIFF_MEMOIZATION" envDefault:"false"`
}

type LockConfigurationServiceImpl struct {
	logger                         *zap.SugaredLogger
	lockConfigurationRepository    LockConfigurationRepository
	userService                    user.UserService
	mergeUtil                      util.MergeUtil
	lockConfigurationServiceConfig *LockConfigurationServiceConfig
}

func NewLockConfigurationServiceImpl(logger *zap.SugaredLogger,
	lockConfigurationRepository LockConfigurationRepository,
	userService user.UserService,
	mergeUtil util.MergeUtil) *LockConfigurationServiceImpl {
	config := &LockConfigurationServiceConfig{}
	err := env.Parse(config)
	if err != nil {
		logger.Warnw("error in initialising UserTerminalSessionConfig but continuing with rest of initialisation", err)
	}
	logger.Infow("env var ", "ARRAY_DIFF_MEMOIZATION", config.ArrayDiffMemoization)
	return &LockConfigurationServiceImpl{
		logger:                         logger,
		lockConfigurationRepository:    lockConfigurationRepository,
		userService:                    userService,
		mergeUtil:                      mergeUtil,
		lockConfigurationServiceConfig: config,
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
	lockConfig, err := impl.GetLockConfiguration()
	if err != nil {
		impl.logger.Errorw("error in getting active lock configuration", "err", err)
		return nil, err
	}
	if lockConfig.Id == 0 {
		return nil, nil
	}
	isSuperAdmin, err := impl.userService.IsSuperAdminForDevtronManaged(userId)

	if err != nil || isSuperAdmin {
		return nil, err
	}

	var savedConfigMap map[string]interface{}
	var currentConfigMap map[string]interface{}

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
	allChanges, deletedMap, addedMap, modifiedMap, containChangesInArray, deletedPaths := impl.getDiffJson(savedConfigMap, currentConfigMap, "")
	var isLockConfigError bool
	if lockConfig.ContainAllowedPaths {
		// Will add in v2 of this feature
	} else {
		isLockConfigError = checkLockedChanges(currentConfig, savedConfig, lockConfig.Paths)
	}
	if isLockConfigError {
		//  rename lockedOverride Diff json byte array
		lockedOverride, _ := json.Marshal(allChanges)
		deletedOverride, _ := json.Marshal(deletedMap)
		addedOverride, _ := json.Marshal(addedMap)
		modifiedOverride, _ := json.Marshal(modifiedMap)
		lockConfigErrorResponse := bean.GetLockConfigErrorResponse(string(lockedOverride), string(modifiedOverride), string(addedOverride), string(deletedOverride), containChangesInArray, deletedPaths)
		return lockConfigErrorResponse, nil
	}
	return nil, nil
}

// Here we are checking whether the values at lockedPath in currentConfig & savedConfig are same or not
func checkLockedChanges(currentConfig, savedConfig string, lockedConfigJsonPaths []string) bool {
	currentConfigParsed, err := oj.ParseString(currentConfig)
	if err != nil {
		return false
	}
	savedConfigParsed, err := oj.ParseString(savedConfig)
	if err != nil {
		return false
	}
	for _, lockedConfigJsonPath := range lockedConfigJsonPaths {
		parsedLockedConfigJsonPath, err := jp.ParseString(lockedConfigJsonPath)
		if err != nil {
			return false
		}
		currentConfigValue := parsedLockedConfigJsonPath.Get(currentConfigParsed)
		savedConfigValue := parsedLockedConfigJsonPath.Get(savedConfigParsed)
		// Sort slices before comparison
		sort.Slice(currentConfigValue, func(i, j int) bool {
			return fmt.Sprintf("%v", currentConfigValue[i]) < fmt.Sprintf("%v", currentConfigValue[j])
		})
		sort.Slice(savedConfigValue, func(i, j int) bool {
			return fmt.Sprintf("%v", savedConfigValue[i]) < fmt.Sprintf("%v", savedConfigValue[j])
		})
		if !reflect.DeepEqual(currentConfigValue, savedConfigValue) {
			return true
		}
	}
	return false
}

func (impl LockConfigurationServiceImpl) getMinOperationsToChangeArrayWithMemoization(word1 []interface{}, word2 []interface{}, i int, j int,
	memoizedArrayI *[][][][]interface{}, memoizedIntArrayI *[][]int,
	memoizedArrayM *[][][][]interface{}, memoizedIntArrayM *[][]int,
	memoizedArrayD *[][][][]interface{}, memoizedIntArrayD *[][]int) (int, []interface{}, []interface{}, []interface{}) {
	var added, modified, deleted []interface{}
	if i < 0 {
		return j + 1, word2[0 : j+1], modified, deleted
	}
	if j < 0 {
		return i + 1, added, modified, word1[0 : i+1]
	}
	if reflect.DeepEqual(word1[i], word2[j]) {
		val, added, modified, deleted := impl.getMinOperationsToChangeArrayWithMemoization(word1, word2, i-1, j-1, memoizedArrayI, memoizedIntArrayI, memoizedArrayM, memoizedIntArrayM, memoizedArrayD, memoizedIntArrayD)
		return val, added, modified, deleted
	}
	const MaxUint = ^uint(0)

	ans := int(MaxUint >> 1)

	// insert
	insert := (*memoizedIntArrayI)[i][j]
	addedI := (*memoizedArrayI)[i][j][0]
	modifiedI := (*memoizedArrayI)[i][j][1]
	deletedI := (*memoizedArrayI)[i][j][2]
	if insert == -1 {

		insert, addedI, modifiedI, deletedI = impl.getMinOperationsToChangeArrayWithMemoization(word1, word2, i, j-1,
			memoizedArrayI, memoizedIntArrayI,
			memoizedArrayM, memoizedIntArrayM,
			memoizedArrayD, memoizedIntArrayD) // insert

		//insert = tmpInsert
		(*memoizedIntArrayI)[i][j] = insert
		(*memoizedArrayI)[i][j][0] = addedI
		(*memoizedArrayI)[i][j][1] = modifiedI
		(*memoizedArrayI)[i][j][2] = deletedI

	}
	if 1+insert < ans {
		ans = 1 + insert
	}

	//delete
	deletedV := (*memoizedIntArrayD)[i][j]
	addedD := (*memoizedArrayD)[i][j][0]
	modifiedD := (*memoizedArrayD)[i][j][1]
	deletedD := (*memoizedArrayD)[i][j][2]
	if deletedV == -1 {
		deletedV, addedD, modifiedD, deletedD = impl.getMinOperationsToChangeArrayWithMemoization(word1, word2, i-1, j,
			memoizedArrayI, memoizedIntArrayI,
			memoizedArrayM, memoizedIntArrayM,
			memoizedArrayD, memoizedIntArrayD) //delete

		//deletedV = tmpDeletedV
		(*memoizedIntArrayD)[i][j] = deletedV
		(*memoizedArrayD)[i][j][0] = addedD
		(*memoizedArrayD)[i][j][1] = modifiedD
		(*memoizedArrayD)[i][j][2] = deletedD
	}
	if 1+deletedV < ans {
		ans = 1 + deletedV
	}

	//replace
	modifiedV := (*memoizedIntArrayM)[i][j]
	addedM := (*memoizedArrayM)[i][j][0]
	modifiedM := (*memoizedArrayM)[i][j][1]
	deletedM := (*memoizedArrayM)[i][j][2]
	if modifiedV == -1 {

		modifiedV, addedM, modifiedM, deletedM = impl.getMinOperationsToChangeArrayWithMemoization(word1, word2, i-1, j-1,
			memoizedArrayI, memoizedIntArrayI,
			memoizedArrayM, memoizedIntArrayM,
			memoizedArrayD, memoizedIntArrayD) //replace

		//modifiedV = tmpModifiedV
		(*memoizedIntArrayM)[i][j] = modifiedV
		(*memoizedArrayM)[i][j][0] = addedM
		(*memoizedArrayM)[i][j][1] = modifiedM
		(*memoizedArrayM)[i][j][2] = deletedM

	}

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
			val := impl.getModifiedValue(word1, word2, i, j)
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
			val := impl.getModifiedValue(word1, word2, i, j)
			modified = append(modified, val)
			added = append(added, addedM...)
			modified = append(modified, modifiedM...)
			deleted = append(deleted, deletedM...)
		}
	}
	return ans, added, modified, deleted
}

func (impl LockConfigurationServiceImpl) getMinOperationsToChangeArray(word1 []interface{}, word2 []interface{}, i int, j int) (int, []interface{}, []interface{}, []interface{}) {
	var added, modified, deleted []interface{}
	if i < 0 {
		return j + 1, word2[0 : j+1], modified, deleted
	}
	if j < 0 {
		return i + 1, added, modified, word1[0 : i+1]
	}
	if reflect.DeepEqual(word1[i], word2[j]) {
		val, added, modified, deleted := impl.getMinOperationsToChangeArray(word1, word2, i-1, j-1)
		return val, added, modified, deleted
	}
	const MaxUint = ^uint(0)

	ans := int(MaxUint >> 1)
	insert, addedI, modifiedI, deletedI := impl.getMinOperationsToChangeArray(word1, word2, i, j-1) // insert
	if 1+insert < ans {
		ans = 1 + insert
	}
	deletedV, addedD, modifiedD, deletedD := impl.getMinOperationsToChangeArray(word1, word2, i-1, j) //delete
	if 1+deletedV < ans {
		ans = 1 + deletedV
	}
	modifiedV, addedM, modifiedM, deletedM := impl.getMinOperationsToChangeArray(word1, word2, i-1, j-1) //replace
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
			val := impl.getModifiedValue(word1, word2, i, j)
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
			val := impl.getModifiedValue(word1, word2, i, j)
			modified = append(modified, val)
			added = append(added, addedM...)
			modified = append(modified, modifiedM...)
			deleted = append(deleted, deletedM...)
		}
	}
	return ans, added, modified, deleted
}

func (impl LockConfigurationServiceImpl) getModifiedValue(word1 []interface{}, word2 []interface{}, i int, j int) interface{} {
	switch reflect.TypeOf(word1[i]).Kind() {
	case reflect.Map:
		savedConfig := copyMap(word1[i].(map[string]interface{}))
		currentConfig := copyMap(word2[j].(map[string]interface{}))
		locked, _, _, _, _, _ := impl.getDiffJson(savedConfig, currentConfig, "/")
		return locked
	case reflect.Array, reflect.Slice:
		locked, _, _, _ := impl.getArrayDiff(word1[i].([]interface{}), word2[j].([]interface{}))
		return locked
	default:
		return word2[j]
	}
}

func (impl LockConfigurationServiceImpl) getArrayDiff(word1 []interface{}, word2 []interface{}) ([]interface{}, []interface{}, []interface{}, []interface{}) {
	if impl.lockConfigurationServiceConfig.ArrayDiffMemoization {
		return impl.getArrayDiffWithMemoization(word1, word2)
	} else {
		return impl.getArrayDiffWithoutMemoization(word1, word2)
	}
}

func (impl LockConfigurationServiceImpl) getArrayDiffWithMemoization(word1 []interface{}, word2 []interface{}) ([]interface{}, []interface{}, []interface{}, []interface{}) {
	l1 := len(word1)
	l2 := len(word2)
	memoizedArrayI := initializeAndGetArray(l1, l2)
	memoizedIntArrayI := initializeAndGetIntArray(l1, l2)
	memoizedArrayM := initializeAndGetArray(l1, l2)
	memoizedIntArrayM := initializeAndGetIntArray(l1, l2)
	memoizedArrayD := initializeAndGetArray(l1, l2)
	memoizedIntArrayD := initializeAndGetIntArray(l1, l2)
	_, added, modified, deleted := impl.getMinOperationsToChangeArrayWithMemoization(word1, word2, l1-1, l2-1, &memoizedArrayI, &memoizedIntArrayI,
		&memoizedArrayM, &memoizedIntArrayM,
		&memoizedArrayD, &memoizedIntArrayD)
	var lockedArray []interface{}
	lockedArray = append(lockedArray, added...)
	lockedArray = append(lockedArray, modified...)
	lockedArray = append(lockedArray, deleted...)
	return lockedArray, added, modified, deleted
}

func (impl LockConfigurationServiceImpl) getArrayDiffWithoutMemoization(word1 []interface{}, word2 []interface{}) ([]interface{}, []interface{}, []interface{}, []interface{}) {
	l1 := len(word1)
	l2 := len(word2)
	_, added, modified, deleted := impl.getMinOperationsToChangeArray(word1, word2, l1-1, l2-1)
	var lockedArray []interface{}
	lockedArray = append(lockedArray, added...)
	lockedArray = append(lockedArray, modified...)
	lockedArray = append(lockedArray, deleted...)
	return lockedArray, added, modified, deleted
}

func copyMap(map1 map[string]interface{}) map[string]interface{} {
	map2 := make(map[string]interface{}, len(map1))
	for k, v := range map1 {
		map2[k] = v
	}
	return map2
}

// Here we are returning 4 maps
// The first one contains all the changes
// The second contains the deleted values
// The third contains the added values
// The fourth contains the modified values
func (impl LockConfigurationServiceImpl) getDiffJson(savedConfigMap, currentConfigMap map[string]interface{}, path string) (map[string]interface{}, map[string]interface{}, map[string]interface{}, map[string]interface{}, bool, []string) {
	// Store all the changes
	lockedMap := make(map[string]interface{})
	deletedMap := make(map[string]interface{})
	addedMap := make(map[string]interface{})
	modifiedMap := make(map[string]interface{})
	var allDeletedPaths []string
	disableSaveEligibleChanges := false
	for key, _ := range savedConfigMap {

		currMapVal, ok := currentConfigMap[key]
		if !ok || currMapVal == nil {
			if savedConfigMap[key] != nil {
				lockedMap[key] = nil
				deletedMap[key] = savedConfigMap[key]
				allDeletedPaths = append(allDeletedPaths, path+"/"+key)

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
				if currentConfigMap[key] == nil {
					lockedMap[key] = currentConfigMap[key]
					modifiedMap[key] = currentConfigMap[key]
					continue
				}
				locked, deleted, added, modified, isSaveEligibleChangesDisabled, deletedPaths := impl.getDiffJson(savedConfigMap[key].(map[string]interface{}), currentConfigMap[key].(map[string]interface{}), path+"/"+key)
				assignMapValueToMap(lockedMap, locked, key)
				assignMapValueToMap(deletedMap, deleted, key)
				assignMapValueToMap(addedMap, added, key)
				assignMapValueToMap(modifiedMap, modified, key)
				allDeletedPaths = append(allDeletedPaths, deletedPaths...)
				if isSaveEligibleChangesDisabled {
					disableSaveEligibleChanges = true
				}
			case reflect.Array, reflect.Slice:
				locked, added, modified, deleted := impl.getArrayDiff(savedConfigMap[key].([]interface{}), currentConfigMap[key].([]interface{}))
				assignArrayValueToMap(lockedMap, locked, key)
				assignArrayValueToMap(deletedMap, deleted, key)
				if len(deleted) != 0 {
					allDeletedPaths = append(allDeletedPaths, path+"/"+key)
				}
				assignArrayValueToMap(addedMap, added, key)
				assignArrayValueToMap(modifiedMap, modified, key)
				disableSaveEligibleChanges = true
			default:
				lockedMap[key] = currentConfigMap[key]
				modifiedMap[key] = currentConfigMap[key]
			}
		} else {
			delete(currentConfigMap, key)
			continue
		}
		switch reflect.TypeOf(currentConfigMap[key]).Kind() {
		case reflect.Map:
			if len(currentConfigMap[key].(map[string]interface{})) == 0 {
				delete(currentConfigMap, key)
			}
		default:
			delete(currentConfigMap, key)
		}
	}
	// Append for the new added keys
	for key, val := range currentConfigMap {
		lockedMap[key] = val
		addedMap[key] = val
		delete(currentConfigMap, key)
	}
	return lockedMap, deletedMap, addedMap, modifiedMap, disableSaveEligibleChanges, allDeletedPaths
}

func appendMapValueToArray(array []interface{}, val map[string]interface{}) []interface{} {
	if len(val) != 0 {
		array = append(array, val)
	}
	return array
}

func appendArrayValueToArray(array []interface{}, val []interface{}) []interface{} {
	if len(val) != 0 {
		array = append(array, val)
	}
	return array
}

func assignArrayValueToMap(mp map[string]interface{}, val []interface{}, key string) {
	if len(val) != 0 {
		mp[key] = val
	}
}

func assignMapValueToMap(mp map[string]interface{}, val map[string]interface{}, key string) {
	if len(val) != 0 {
		mp[key] = val
	}
}

func initializeAndGetArray(l1 int, l2 int) [][][][]interface{} {
	memoizedArray := make([][][][]interface{}, l1)

	for i := 0; i < l1; i++ {
		memoizedArray[i] = make([][][]interface{}, l2)
		for j := 0; j < l2; j++ {
			memoizedArray[i][j] = make([][]interface{}, 3)
			for k := 0; k < 3; k++ {
				memoizedArray[i][j][k] = []interface{}{}
			}
		}
	}

	return memoizedArray
}

func initializeAndGetIntArray(l1 int, l2 int) [][]int {
	memoizedArray := make([][]int, l1)

	for i := 0; i < l1; i++ {
		memoizedArray[i] = make([]int, l2)
		for k := 0; k < l2; k++ {
			memoizedArray[i][k] = -1
		}
	}

	return memoizedArray
}
