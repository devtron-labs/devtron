package pkg

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"regexp"
	"strconv"
	"strings"
)

const (
	EntityNull = iota
	EntityTypeCiPipelineId
)

const (
	imagePathPattern                                              = "%s/%s:%s" // dockerReg/dockerRepo:Tag
	ImageTagUnavailableMessage                                    = "Desired image tag already exists"
	REGEX_PATTERN_FOR_ENSURING_ONLY_ONE_VARIABLE_BETWEEN_BRACKETS = `\{.{2,}\}`
	REGEX_PATTERN_FOR_CHARACTER_OTHER_THEN_X_OR_x                 = `\{[^xX]|{}\}`
	REGEX_PATTERN_FOR_IMAGE_TAG                                   = `^[a-zA-Z0-9][a-zA-Z0-9._-]{0,126}[a-zA-Z0-9]$`
)

var (
	ErrImagePathInUse = fmt.Errorf(ImageTagUnavailableMessage)
)

type CustomTagService interface {
	CreateOrUpdateCustomTag(tag *bean.CustomTag) error
	GetCustomTagByEntityKeyAndValue(entityKey int, entityValue string) (*repository.CustomTag, error)
	GetActiveCustomTagByEntityKeyAndValue(entityKey int, entityValue string) (*repository.CustomTag, error)
	GenerateImagePath(entityKey int, entityValue string, dockerRegistryURL string, dockerRepo string) (*repository.ImagePathReservation, error)
	DeleteCustomTagIfExists(tag bean.CustomTag) error
	DeactivateImagePathReservation(id int) error
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
	if err := validateTagPattern(tag.TagPattern); err != nil {
		return err
	}
	customTagData := repository.CustomTag{
		EntityKey:            tag.EntityKey,
		EntityValue:          tag.EntityValue,
		TagPattern:           strings.ReplaceAll(tag.TagPattern, "{X}", "{x}"),
		AutoIncreasingNumber: tag.AutoIncreasingNumber,
		Metadata:             tag.Metadata,
		Active:               true,
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
	imagePath := fmt.Sprintf(imagePathPattern, dockerRegistryURL, dockerRepo, tag)
	imagePathReservations, err := impl.customTagRepository.FindByImagePath(tx, imagePath)
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	if len(imagePathReservations) > 0 {
		return nil, ErrImagePathInUse
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
	dockerImageTag := strings.ReplaceAll(customTagData.TagPattern, "{x}", strconv.Itoa(customTagData.AutoIncreasingNumber-1)) //-1 because number is already incremented, current value will be used next time
	if isValidDockerImageTag(dockerImageTag) {
		return dockerImageTag, fmt.Errorf("invalid docker tag")
	}
	return dockerImageTag, nil
}

func validateTagPattern(customTagPattern string) error {
	if len(customTagPattern) == 0 {
		return fmt.Errorf("tag length can not be zero")
	}

	// for patterns like v1.0.{x} we will calculate count with . in {x} i.e .{x}
	variableCount := 0
	variableCount = variableCount + strings.Count(customTagPattern, ".{x}")
	variableCount = variableCount + strings.Count(customTagPattern, ".{X}")

	if variableCount == 0 {
		// there can be case when there is only one {x} or {x}
		IsOnlyVariableTag := 0
		IsOnlyVariableTag = IsOnlyVariableTag + strings.Count(customTagPattern, "{x}")
		IsOnlyVariableTag = IsOnlyVariableTag + strings.Count(customTagPattern, "{X}")
		if IsOnlyVariableTag == 0 {
			return fmt.Errorf("variable with format {x} or {X} not found")
		} else if IsOnlyVariableTag > 1 {
			return fmt.Errorf("only one variable with format {x} or {X} allowed")
		}
	} else if variableCount > 1 {
		return fmt.Errorf("only one variable with format {x} or {X} allowed")
	}

	// replacing variable with 1 (dummy value) and checking if resulting string is valid tag
	tagWithDummyValue := strings.ReplaceAll(customTagPattern, ".{x}", "1")
	tagWithDummyValue = strings.ReplaceAll(customTagPattern, ".{X}", "1")
	tagWithDummyValue = strings.ReplaceAll(customTagPattern, "{x}", "1")
	tagWithDummyValue = strings.ReplaceAll(customTagPattern, "{X}", "1")

	if !isValidDockerImageTag(tagWithDummyValue) {
		return fmt.Errorf("not a valid image tag")
	}

	return nil
}

func isValidDockerImageTag(tag string) bool {
	// Define the regular expression for a valid Docker image tag
	re := regexp.MustCompile(REGEX_PATTERN_FOR_IMAGE_TAG)
	return re.MatchString(tag)
}

func validateTag(imageTag string) error {
	if len(imageTag) == 0 || len(imageTag) > 128 {
		return fmt.Errorf("image tag should be of len 1-128 only, imageTag: %s", imageTag)
	}
	allowedSymbols := ".abcdefghijklmnopqrstuvwxyz_ABCDEFGHIJKLMNOPQRSTUVWXYZ-0987654321"
	allowedCharSet := make(map[int32]struct{})
	for _, c := range allowedSymbols {
		allowedCharSet[c] = struct{}{}
	}
	firstChar := imageTag[0:1]
	if firstChar == "." || firstChar == "-" {
		return fmt.Errorf("image tag can not start with a period or a hyphen, imageTag: %s", imageTag)
	}
	return nil
}
