package lockConfiguration

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	jsonpatch1 "github.com/evanphx/json-patch"
	"github.com/go-pg/pg"
	"github.com/mattbaird/jsonpatch"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"go.uber.org/zap"
	"strings"
	"time"
)

type LockConfigurationService interface {
	GetLockConfiguration() (*bean.LockConfigResponse, error)
	SaveLockConfiguration(*bean.LockConfigRequest, int32) error
	DeleteActiveLockConfiguration(userId int32, tx *pg.Tx) error
	HandleLockConfiguration(currentConfig, savedConfig string, userId int) (bool, string, error)
}

type LockConfigurationServiceImpl struct {
	logger                      *zap.SugaredLogger
	lockConfigurationRepository LockConfigurationRepository
	userService                 user.UserService
}

func NewLockConfigurationServiceImpl(logger *zap.SugaredLogger,
	lockConfigurationRepository LockConfigurationRepository,
	userService user.UserService) *LockConfigurationServiceImpl {
	return &LockConfigurationServiceImpl{
		logger:                      logger,
		lockConfigurationRepository: lockConfigurationRepository,
		userService:                 userService,
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

//TODO refactoring

func (impl LockConfigurationServiceImpl) HandleLockConfiguration(currentConfig, savedConfig string, userId int) (bool, string, error) {

	isSuperAdmin, err := impl.userService.IsSuperAdmin(userId)
	if err != nil || isSuperAdmin {
		return false, "", err
	}

	emptyJson := `{
    }`
	patch, err := jsonpatch.CreatePatch([]byte(savedConfig), []byte(currentConfig))
	if err != nil {
		fmt.Printf("Error creating JSON patch:%v", err)
		return false, "", err
	}
	patch1, err := jsonpatch.CreatePatch([]byte(currentConfig), []byte(emptyJson))
	if err != nil {
		fmt.Printf("Error creating JSON patch: %v", err)
		return false, "", err
	}
	paths := make(map[string]bool)
	for _, path := range patch {
		// ADD
		res := strings.Split(path.Path, "/")
		paths["/"+res[1]] = true
	}
	for index, path := range patch1 {
		if paths[path.Path] {
			patch1 = append(patch1[:index], patch1[index+1:]...)
		}
	}
	marsh, _ := json.Marshal(patch1)
	patche, err := jsonpatch1.DecodePatch(marsh)
	if err != nil {
		return false, "", err
	}
	modified, err := patche.Apply([]byte(currentConfig))
	if err != nil {
		return false, "", err
	}
	obj, err := oj.ParseString(string(modified))
	if err != nil {
		return false, "", err
	}
	lockConfig, err := impl.GetLockConfiguration()
	isLockConfigError := false
	for _, config := range lockConfig.Config {
		x, err := jp.ParseString(config)
		if err != nil {
			return false, "", err
		}
		ys := x.Get(obj)
		if len(ys) != 0 {
			isLockConfigError = true
		}
	}
	if isLockConfigError {
		return true, string(modified), nil
	}

	return false, "", nil
}
