/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pipeline

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"regexp"
	"strconv"
	"strings"
)

type CustomTagService interface {
	CreateOrUpdateCustomTag(tag *bean.CustomTag) error
	GetCustomTagByEntityKeyAndValue(entityKey int, entityValue string) (*repository.CustomTag, error)
	GetActiveCustomTagByEntityKeyAndValue(entityKey int, entityValue string) (*repository.CustomTag, error)
	GenerateImagePath(entityKey int, entityValue string, dockerRegistryURL string, dockerRepo string) (*repository.ImagePathReservation, error)
	DeleteCustomTagIfExists(tag bean.CustomTag) error
	DeactivateImagePathReservation(id int) error
	GetCustomTag(entityKey int, entityValue string) (*repository.CustomTag, string, error)
	ReserveImagePath(imagePath string, customTagId int) (*repository.ImagePathReservation, error)
	DeactivateImagePathReservationByImagePath(imagePaths []string) error
	DeactivateImagePathReservationByImageIds(imagePathReservationIds []int) error
	DisableCustomTagIfExist(tag bean.CustomTag) error
}

type CustomTagServiceImpl struct {
	Logger              *zap.SugaredLogger
	customTagRepository repository.ImageTagRepository
}

func NewCustomTagService(logger *zap.SugaredLogger, customTagRepo repository.ImageTagRepository) *CustomTagServiceImpl {
	return &CustomTagServiceImpl{
		Logger:              logger,
		customTagRepository: customTagRepo,
	}
}

func (impl *CustomTagServiceImpl) DeactivateImagePathReservation(id int) error {
	return impl.customTagRepository.DeactivateImagePathReservation(id)
}

func (impl *CustomTagServiceImpl) CreateOrUpdateCustomTag(tag *bean.CustomTag) error {

	if len(tag.TagPattern) == 0 && tag.Enabled {
		return fmt.Errorf("tag pattern cannot be empty")
	}
	if tag.Enabled {
		if err := validateTagPattern(tag.TagPattern); err != nil {
			return err
		}
	}
	var customTagData repository.CustomTag
	customTagData = repository.CustomTag{
		EntityKey:            tag.EntityKey,
		EntityValue:          tag.EntityValue,
		TagPattern:           strings.ReplaceAll(tag.TagPattern, bean2.IMAGE_TAG_VARIABLE_NAME_X, bean2.IMAGE_TAG_VARIABLE_NAME_x),
		AutoIncreasingNumber: tag.AutoIncreasingNumber,
		Metadata:             tag.Metadata,
		Active:               true,
		Enabled:              tag.Enabled,
	}
	oldTagObject, err := impl.customTagRepository.FetchCustomTagData(customTagData.EntityKey, customTagData.EntityValue)
	if err != nil && err != pg.ErrNoRows {
		return err
	}
	if oldTagObject.Id == 0 {
		return impl.customTagRepository.CreateImageTag(&customTagData)
	} else {
		customTagData.Id = oldTagObject.Id
		customTagData.Active = true
		return impl.customTagRepository.UpdateImageTag(&customTagData)
	}
}

func (impl *CustomTagServiceImpl) DeleteCustomTagIfExists(tag bean.CustomTag) error {
	return impl.customTagRepository.DeleteByEntityKeyAndValue(tag.EntityKey, tag.EntityValue)
}

func (impl *CustomTagServiceImpl) GetCustomTagByEntityKeyAndValue(entityKey int, entityValue string) (*repository.CustomTag, error) {
	return impl.customTagRepository.FetchCustomTagData(entityKey, entityValue)
}

func (impl *CustomTagServiceImpl) GetActiveCustomTagByEntityKeyAndValue(entityKey int, entityValue string) (*repository.CustomTag, error) {
	return impl.customTagRepository.FetchActiveCustomTagData(entityKey, entityValue)
}

func (impl *CustomTagServiceImpl) GenerateImagePath(entityKey int, entityValue string, dockerRegistryURL string, dockerRepo string) (*repository.ImagePathReservation, error) {
	connection := impl.customTagRepository.GetConnection()
	tx, err := connection.Begin()
	if err != nil {
		return nil, nil
	}
	defer tx.Rollback()
	customTagData, err := impl.customTagRepository.IncrementAndFetchByEntityKeyAndValue(tx, entityKey, entityValue)
	if err != nil {
		return nil, err
	}
	tag, err := validateAndConstructTag(customTagData)
	if err != nil {
		return nil, err
	}
	imagePath := fmt.Sprintf(bean2.ImagePathPattern, dockerRegistryURL, dockerRepo, tag)
	imagePathReservations, err := impl.customTagRepository.FindByImagePath(tx, imagePath)
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	if len(imagePathReservations) > 0 {
		return nil, bean2.ErrImagePathInUse
	}
	imagePathReservation := &repository.ImagePathReservation{
		ImagePath:   imagePath,
		CustomTagId: customTagData.Id,
	}
	err = impl.customTagRepository.InsertImagePath(tx, imagePathReservation)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return imagePathReservation, nil
}

func validateAndConstructTag(customTagData *repository.CustomTag) (string, error) {
	err := validateTagPattern(customTagData.TagPattern)
	if err != nil {
		return "", err
	}
	if customTagData.AutoIncreasingNumber < 0 {
		return "", fmt.Errorf("counter {x} can not be negative")
	}
	dockerImageTag := strings.ReplaceAll(customTagData.TagPattern, bean2.IMAGE_TAG_VARIABLE_NAME_x, strconv.Itoa(customTagData.AutoIncreasingNumber-1)) //-1 because number is already incremented, current value will be used next time
	if !isValidDockerImageTag(dockerImageTag) {
		return dockerImageTag, fmt.Errorf("invalid docker tag")
	}
	return dockerImageTag, nil
}

func validateTagPattern(customTagPattern string) error {
	if len(customTagPattern) == 0 {
		return fmt.Errorf("tag length can not be zero")
	}

	variableCount := 0
	variableCount = variableCount + strings.Count(customTagPattern, bean2.IMAGE_TAG_VARIABLE_NAME_x)
	variableCount = variableCount + strings.Count(customTagPattern, bean2.IMAGE_TAG_VARIABLE_NAME_X)

	if variableCount == 0 {
		// there can be case when there is only one {x} or {x}
		return fmt.Errorf("variable with format {x} or {X} not found")
	} else if variableCount > 1 {
		return fmt.Errorf("only one variable with format {x} or {X} allowed")
	}

	// replacing variable with 1 (dummy value) and checking if resulting string is valid tag
	tagWithDummyValue := strings.ReplaceAll(customTagPattern, bean2.IMAGE_TAG_VARIABLE_NAME_x, "1")
	tagWithDummyValue = strings.ReplaceAll(tagWithDummyValue, bean2.IMAGE_TAG_VARIABLE_NAME_X, "1")

	if !isValidDockerImageTag(tagWithDummyValue) {
		return fmt.Errorf("not a valid image tag")
	}

	return nil
}

func isValidDockerImageTag(tag string) bool {
	// Define the regular expression for a valid Docker image tag
	re := regexp.MustCompile(bean2.REGEX_PATTERN_FOR_IMAGE_TAG)
	return re.MatchString(tag)
}

func (impl *CustomTagServiceImpl) GetCustomTag(entityKey int, entityValue string) (*repository.CustomTag, string, error) {
	connection := impl.customTagRepository.GetConnection()
	tx, err := connection.Begin()
	customTagData, err := impl.customTagRepository.IncrementAndFetchByEntityKeyAndValue(tx, entityKey, entityValue)
	if err != nil {
		return nil, "", err
	}
	err = tx.Commit()
	if err != nil {
		impl.Logger.Errorw("Error in fetching custom tag", "err", err)
		return customTagData, "", err
	}
	var dockerTag string
	if customTagData != nil && len(customTagData.TagPattern) == 0 {
		return customTagData, dockerTag, nil
	}
	dockerTag, err = validateAndConstructTag(customTagData)
	if err != nil {
		return nil, "", err
	}
	return customTagData, dockerTag, nil

}

func (impl *CustomTagServiceImpl) ReserveImagePath(imagePath string, customTagId int) (*repository.ImagePathReservation, error) {
	connection := impl.customTagRepository.GetConnection()
	tx, err := connection.Begin()
	if err != nil {
		return nil, err
	}
	imagePathReservations, err := impl.customTagRepository.FindByImagePath(tx, imagePath)
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	if len(imagePathReservations) > 0 {
		return nil, bean2.ErrImagePathInUse
	}
	imagePathReservation := &repository.ImagePathReservation{
		ImagePath:   imagePath,
		CustomTagId: customTagId,
	}
	err = impl.customTagRepository.InsertImagePath(tx, imagePathReservation)
	if err != nil {
		return imagePathReservation, err
	}
	err = tx.Commit()
	if err != nil {
		impl.Logger.Errorw("Error in fetching custom tag", "err", err)
		return imagePathReservation, err
	}
	return imagePathReservation, err
}

func (impl *CustomTagServiceImpl) DeactivateImagePathReservationByImagePath(imagePaths []string) error {
	connection := impl.customTagRepository.GetConnection()
	tx, err := connection.Begin()
	if err != nil {
		return nil
	}
	err = impl.customTagRepository.DeactivateImagePathReservationByImagePaths(tx, imagePaths)
	if err != nil {
		impl.Logger.Errorw("error in marking image path unreserved")
		return err
	}
	err = tx.Commit()
	if err != nil {
		impl.Logger.Errorw("Error in fetching custom tag", "err", err)
		return err
	}
	return nil
}

func (impl *CustomTagServiceImpl) DeactivateImagePathReservationByImageIds(imagePathReservationIds []int) error {
	connection := impl.customTagRepository.GetConnection()
	tx, err := connection.Begin()
	if err != nil {
		return nil
	}
	err = impl.customTagRepository.DeactivateImagePathReservationByImagePathReservationIds(tx, imagePathReservationIds)
	if err != nil {
		impl.Logger.Errorw("error in marking image path unreserved")
		return err
	}
	err = tx.Commit()
	if err != nil {
		impl.Logger.Errorw("Error in fetching custom tag", "err", err)
		return err
	}
	return nil
}

func (impl *CustomTagServiceImpl) DisableCustomTagIfExist(tag bean.CustomTag) error {
	return impl.customTagRepository.DisableCustomTag(tag.EntityKey, tag.EntityValue)
}
