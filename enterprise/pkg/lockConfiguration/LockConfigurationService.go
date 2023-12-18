package lockConfiguration

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/go-pg/pg"
	"github.com/mdaverde/jsonpath"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"go.uber.org/zap"
	"reflect"
	"strings"
	"time"
)

type LockConfigurationService interface {
	GetLockConfiguration() (*bean.LockConfigResponse, error)
	SaveLockConfiguration(*bean.LockConfigRequest, int32) error
	DeleteActiveLockConfiguration(userId int32, tx *pg.Tx) error
	HandleLockConfiguration(currentConfig, savedConfig string, userId int) (bool, string, string, string, string, error)
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
	err = impl.DeleteActiveLockConfiguration(createdBy, tx)
	if err != nil && err != pg.ErrNoRows {
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
		return err
	}
	return err
}

func (impl LockConfigurationServiceImpl) DeleteActiveLockConfiguration(userId int32, tx *pg.Tx) error {
	lockConfigDto, err := impl.lockConfigurationRepository.GetActiveLockConfig()
	if err != nil {
		return err
	}
	dbConnection := impl.lockConfigurationRepository.GetConnection()
	if tx == nil {
		tx, err = dbConnection.Begin()
		if err != nil {
			return err
		}
	}

	// Rollback tx on error.
	defer tx.Rollback()

	lockConfigDto.Active = false
	lockConfigDto.UpdatedOn = time.Now()
	lockConfigDto.UpdatedBy = userId
	err = impl.lockConfigurationRepository.Update(lockConfigDto, tx)
	if err != nil {
		return err
	}
	// commit TX
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (impl LockConfigurationServiceImpl) HandleLockConfiguration(currentConfig, savedConfig string, userId int) (bool, string, string, string, string, error) {

	isSuperAdmin, err := impl.userService.IsSuperAdmin(userId)

	if err != nil || isSuperAdmin {
		return false, "", "", "", "", err
	}

	var mp map[string]interface{}
	var mp2 map[string]interface{}

	lockConfig, err := impl.GetLockConfiguration()
	if err != nil {
		return false, "", "", "", "", err
	}

	json.Unmarshal([]byte(savedConfig), &mp)
	json.Unmarshal([]byte(currentConfig), &mp2)
	changes, deleted := getChanges(mp, mp2)
	// allChanges := getAllChanges(mp, mp2)
	deletedByte, _ := json.Marshal(deleted)
	addedByte, _ := json.Marshal(mp2)
	var isLockConfigError bool
	if lockConfig.Allowed {
		isLockConfigError = checkAllowedChanges(changes, lockConfig.Config)
	} else {
		isLockConfigError = checkLockedChanges(currentConfig, savedConfig, lockConfig.Config)
	}
	if isLockConfigError {
		modifiedByte, _ := json.Marshal(changes)
		changesByte, err := impl.mergeUtil.JsonPatch(modifiedByte, deletedByte)
		if err != nil {
			return false, "", "", "", "", err
		}
		changesByte, err = impl.mergeUtil.JsonPatch(changesByte, addedByte)
		if err != nil {
			return false, "", "", "", "", err
		}

		return true, string(changesByte), string(modifiedByte), string(deletedByte), string(addedByte), nil
	}
	return false, "", "", "", "", nil
}

func checkAllowedChanges(diffJson map[string]interface{}, configs []string) bool {
	diffJson = setAllJsonValue(diffJson, "%devtron%")
	for _, config := range configs {
		path := config
		if strings.Contains(path, "$.") {
			path = strings.Split(path, "$.")[1]
		}
		err := jsonpath.Set(&diffJson, path, "%devtron2%")
		if err != nil {

		}
	}
	diffJsonByte, _ := json.Marshal(diffJson)
	diffJsonStr := string(diffJsonByte)
	return strings.Contains(diffJsonStr, "%devtron%")
}

func checkLockedChanges(currentConfig, savedConfig string, configs []string) bool {
	obj, err := oj.ParseString(currentConfig)
	if err != nil {
		return false
	}
	obj1, err := oj.ParseString(savedConfig)
	if err != nil {
		return false
	}
	for _, config := range configs {
		x, err := jp.ParseString(config)
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

//func getLockedAndAllowedArray(mp1, mp2 []interface{}, currentPath string, lockedPath []string) ([]interface{}, []interface{}) {
//	var lockedMap, allowedMap []interface{}
//	for key, _ := range mp1 {
//		if !reflect.DeepEqual(mp1[key], mp2[key]) {
//			if slices.Contains(lockedPath, currentPath+strconv.Itoa(key)) {
//				lockedMap[key] = mp2[key]
//				continue
//			}
//			switch reflect.TypeOf(mp1[key]).Kind() {
//			case reflect.Map:
//				locked, allowed := getLockedAndAllowed(mp1[key].(map[string]interface{}), mp2[key].(map[string]interface{}), currentPath+strconv.Itoa(key)+"/", lockedPath)
//				if locked != nil && len(locked) != 0 {
//					lockedMap = append(lockedMap, locked)
//				}
//				if allowed != nil && len(allowed) != 0 {
//					allowedMap = append(allowedMap, allowed)
//				}
//			case reflect.Array:
//				locked, allowed := getLockedAndAllowedArray(mp1[key].([]interface{}), mp2[key].([]interface{}), currentPath+strconv.Itoa(key)+"/", lockedPath)
//				if locked != nil && len(locked) != 0 {
//					lockedMap = append(lockedMap, locked)
//				}
//				if allowed != nil && len(allowed) != 0 {
//					allowedMap = append(allowedMap, allowed)
//				}
//			default:
//				allowedMap = append(allowedMap, mp2[key])
//			}
//		} else {
//			allowedMap = append(allowedMap, mp2[key])
//		}
//
//	}
//	return lockedMap, allowedMap
//
//}
//
//func getLockedAndAllowed(mp1, mp2 map[string]interface{}, currentPath string, lockedPath []string) (map[string]interface{}, map[string]interface{}) {
//	lockedMap := make(map[string]interface{})
//	allowedMap := make(map[string]interface{})
//	for key, _ := range mp1 {
//		if _, ok := mp2[key]; !ok {
//
//		}
//		if !reflect.DeepEqual(mp1[key], mp2[key]) {
//			if slices.Contains(lockedPath, currentPath+key) {
//				lockedMap[key] = mp2[key]
//				continue
//			}
//			switch reflect.TypeOf(mp1[key]).Kind() {
//			case reflect.Map:
//				locked, allowed := getLockedAndAllowed(mp1[key].(map[string]interface{}), mp2[key].(map[string]interface{}), currentPath+key+"/", lockedPath)
//				if locked != nil && len(locked) != 0 {
//					lockedMap[key] = locked
//				}
//				if allowed != nil && len(allowed) != 0 {
//					allowedMap[key] = allowed
//				}
//			case reflect.Array, reflect.Slice:
//				locked, allowed := getLockedAndAllowedArray(mp1[key].([]interface{}), mp2[key].([]interface{}), currentPath+key+"/", lockedPath)
//				if locked != nil && len(locked) != 0 {
//					lockedMap[key] = locked
//				}
//				if allowed != nil && len(allowed) != 0 {
//					allowedMap[key] = allowed
//				}
//			default:
//				allowedMap[key] = mp2[key]
//			}
//		}
//
//	}
//	return lockedMap, allowedMap
//
//}

func checkForLockedArray(mp1, mp2 []interface{}) []interface{} {
	var lockedMap []interface{}
	for key, _ := range mp1 {
		if key >= len(mp2) {
			break
		}
		if !reflect.DeepEqual(mp1[key], mp2[key]) {
			switch reflect.TypeOf(mp1[key]).Kind() {
			case reflect.Map:
				locked := getAllChanges(mp1[key].(map[string]interface{}), mp2[key].(map[string]interface{}))
				if locked != nil && len(locked) != 0 {
					lockedMap = append(lockedMap, locked)
				}
			case reflect.Array, reflect.Slice:
				locked := checkForLockedArray(mp1[key].([]interface{}), mp2[key].([]interface{}))
				if locked != nil && len(locked) != 0 {
					lockedMap = append(lockedMap, locked)
				}
			default:
				lockedMap = append(lockedMap, mp2[key])

			}
		}
	}
	return lockedMap
}

func getChanges(mp1, mp2 map[string]interface{}) (map[string]interface{}, map[string]interface{}) {
	lockedMap := make(map[string]interface{})
	deletedMap := make(map[string]interface{})
	for key, _ := range mp1 {
		if _, ok := mp2[key]; !ok {
			deletedMap[key] = mp1[key]
			continue
		}
		if !reflect.DeepEqual(mp1[key], mp2[key]) {
			switch reflect.TypeOf(mp1[key]).Kind() {
			case reflect.Map:
				locked, deleted := getChanges(mp1[key].(map[string]interface{}), mp2[key].(map[string]interface{}))
				if locked != nil && len(locked) != 0 {
					lockedMap[key] = locked
				}
				if deleted != nil && len(deleted) != 0 {
					deletedMap[key] = deleted
				}
			default:
				lockedMap[key] = mp2[key]
			}
		}
		delete(mp2, key)
	}
	return lockedMap, deletedMap
}

func getAllChanges(mp1, mp2 map[string]interface{}) map[string]interface{} {
	lockedMap := make(map[string]interface{})
	for key, _ := range mp1 {
		if _, ok := mp2[key]; !ok {
			continue
		}
		if !reflect.DeepEqual(mp1[key], mp2[key]) {
			switch reflect.TypeOf(mp1[key]).Kind() {
			case reflect.Map:
				locked := getAllChanges(mp1[key].(map[string]interface{}), mp2[key].(map[string]interface{}))
				if locked != nil && len(locked) != 0 {
					lockedMap[key] = locked
				}
			case reflect.Array, reflect.Slice:
				locked := checkForLockedArray(mp1[key].([]interface{}), mp2[key].([]interface{}))
				if locked != nil && len(locked) != 0 {
					lockedMap[key] = locked
				}
			default:
				lockedMap[key] = mp2[key]
			}
		}

	}
	return lockedMap
}

func setAllJsonValue(mp map[string]interface{}, val string) map[string]interface{} {
	for key, _ := range mp {
		switch reflect.TypeOf(mp[key]).Kind() {
		case reflect.Map:
			childVal := setAllJsonValue(mp[key].(map[string]interface{}), val)
			mp[key] = childVal
		case reflect.Array, reflect.Slice:
			childVal := setArrayValue(mp[key].([]interface{}), val)
			mp[key] = childVal
		default:
			mp[key] = val
		}
	}
	return mp
}

func setArrayValue(mp []interface{}, val string) []interface{} {
	for key, _ := range mp {
		switch reflect.TypeOf(mp[key]).Kind() {
		case reflect.Map:
			childVal := setAllJsonValue(mp[key].(map[string]interface{}), val)
			mp[key] = childVal
		case reflect.Array, reflect.Slice:
			childVal := setArrayValue(mp[key].([]interface{}), val)
			mp[key] = childVal
		default:
			mp[key] = val
		}
	}
	return mp
}
