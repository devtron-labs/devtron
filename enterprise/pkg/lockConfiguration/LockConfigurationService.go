package lockConfiguration

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
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
	HandleLockConfiguration(currentConfig, savedConfig string) (bool, string, error)
}

type LockConfigurationServiceImpl struct {
	logger                      *zap.SugaredLogger
	lockConfigurationRepository LockConfigurationRepository
}

func NewLockConfigurationServiceImpl(logger *zap.SugaredLogger,
	lockConfigurationRepository LockConfigurationRepository) *LockConfigurationServiceImpl {
	return &LockConfigurationServiceImpl{
		logger:                      logger,
		lockConfigurationRepository: lockConfigurationRepository,
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
	lockConfigDto, err := impl.lockConfigurationRepository.GetActiveLockConfig()
	if err != nil && err != pg.ErrNoRows {
		return err
	}
	dbConnection := impl.lockConfigurationRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	if lockConfigDto != nil {
		lockConfigDto.Active = false
		lockConfigDto.UpdatedOn = time.Now()
		lockConfigDto.UpdatedBy = createdBy
		err := impl.lockConfigurationRepository.Update(lockConfigDto, tx)
		if err != nil {
			return err
		}
	}

	newLockConfigDto := lockConfig.ConvertRequestToDBDto()
	newLockConfigDto.AuditLog = sql.NewDefaultAuditLog(createdBy)

	err = impl.lockConfigurationRepository.Create(newLockConfigDto, tx)
	if err != nil {
		impl.logger.Errorw("error while saving global tags", "error", err)
		return err
	}

	// commit TX
	err = tx.Commit()
	if err != nil {
		return err
	}
	return err
}

func (impl LockConfigurationServiceImpl) HandleLockConfiguration(currentConfig, savedConfig string) (bool, string, error) {
	emptyJson := `{
    }`
	patch, err := jsonpatch.CreatePatch([]byte(savedConfig), []byte(currentConfig))
	if err != nil {
		fmt.Printf("Error creating JSON patch:%v", err)
		return false, "", err
	}
	patch1, err := jsonpatch.CreatePatch([]byte(currentConfig), []byte(emptyJson))
	if err != nil {
		fmt.Printf("Error creating JSON patch:%v", err)
		return false, "", err
	}
	paths := make(map[string]bool)
	for _, path := range patch {
		res := strings.Split(path.Path, "/")
		fmt.Println(res[1])
		paths["/"+res[1]] = true
	}
	for index, path := range patch1 {
		if paths[path.Path] {
			fmt.Println(path)
			patch1 = append(patch1[:index], patch1[index+1:]...)
		}
	}
	marsh, _ := json.Marshal(patch1)
	patche, err := jsonpatch1.DecodePatch(marsh)
	if err != nil {
		panic(err)
	}
	modified, err := patche.Apply([]byte(currentConfig))
	if err != nil {
		panic(err)
	}
	obj, err := oj.ParseString(string(modified))
	if err != nil {
		panic(err)
	}
	lockConfig, err := impl.GetLockConfiguration()
	isLockConfigError := false
	for _, config := range lockConfig.Config {
		x, err := jp.ParseString(config)
		if err != nil {
			panic(err)
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
