package lockConfiguration

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
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
	DeleteActiveLockConfiguration(userId int32, tx *pg.Tx) error
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
	err = impl.DeleteActiveLockConfiguration(createdBy, tx)
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

func (impl LockConfigurationServiceImpl) DeleteActiveLockConfiguration(userId int32, tx *pg.Tx) error {
	lockConfigDto, err := impl.lockConfigurationRepository.GetActiveLockConfig()
	if err != nil {
		impl.logger.Errorw("Error in getting active lock config", "err", err)
		return err
	}
	dbConnection := impl.lockConfigurationRepository.GetConnection()
	isNewTx := false
	if tx == nil {
		tx, err = dbConnection.Begin()
		if err != nil {
			return err
		}
		isNewTx = true
	}

	// Rollback tx on error.
	if isNewTx {
		defer tx.Rollback()
	}

	lockConfigDto.Active = false
	lockConfigDto.UpdatedOn = time.Now()
	lockConfigDto.UpdatedBy = userId
	err = impl.lockConfigurationRepository.Update(lockConfigDto, tx)
	if err != nil {
		impl.logger.Errorw("Error in updating lock config", "lockConfigId", lockConfigDto.Id, "err", err)
		return err
	}
	// commit TX
	if isNewTx {
		err = tx.Commit()
		if err != nil {
			impl.logger.Errorw("Error in committing tx", "err", err)
			return err
		}
	}
	return nil
}

func (impl LockConfigurationServiceImpl) HandleLockConfiguration(currentConfig, savedConfig string, userId int) (*bean.LockValidateErrorResponse, error) {

	isSuperAdmin, err := impl.userService.IsSuperAdmin(userId)

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
	var lockedMap []interface{}
	var key int
	for key, _ = range savedConfigMap {
		if key >= len(currentConfigMap) {
			break
		}
		if !reflect.DeepEqual(savedConfigMap[key], currentConfigMap[key]) {
			switch reflect.TypeOf(savedConfigMap[key]).Kind() {
			case reflect.Map:
				locked, _ := getDiffJson(savedConfigMap[key].(map[string]interface{}), currentConfigMap[key].(map[string]interface{}))
				if locked != nil && len(locked) != 0 {
					lockedMap = append(lockedMap, locked)
				}
			case reflect.Array, reflect.Slice:
				locked := checkForLockedArray(savedConfigMap[key].([]interface{}), currentConfigMap[key].([]interface{}))
				if locked != nil && len(locked) != 0 {
					lockedMap = append(lockedMap, locked)
				}
			default:
				lockedMap = append(lockedMap, currentConfigMap[key])

			}
		}
	}
	for key1, _ := range currentConfigMap {
		if key1 <= key {
			continue
		}
		lockedMap = append(lockedMap, currentConfigMap[key1])
	}
	return lockedMap
}

// TODO ADD comment for return
func getDiffJson(savedConfigMap, currentConfigMap map[string]interface{}) (map[string]interface{}, bool) {
	// Store all the changes
	lockedMap := make(map[string]interface{})
	disableSaveEligibleChanges := false
	for key, _ := range savedConfigMap {
		// check for the deleted keys
		if _, ok := currentConfigMap[key]; !ok {
			lockedMap[key] = nil
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
				locked := checkForLockedArray(savedConfigMap[key].([]interface{}), currentConfigMap[key].([]interface{}))
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
			if currentConfigMap[key] == nil || len(currentConfigMap[key].(map[string]interface{})) == 0 {
				delete(currentConfigMap, key)
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
