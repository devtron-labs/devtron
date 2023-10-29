package bean

import "fmt"

const (
	EntityNull = iota
	EntityTypeCiPipelineId
)

const (
	ImagePathPattern                                              = "%s/%s:%s" // dockerReg/dockerRepo:Tag
	ImageTagUnavailableMessage                                    = "Desired image tag already exists"
	REGEX_PATTERN_FOR_ENSURING_ONLY_ONE_VARIABLE_BETWEEN_BRACKETS = `\{.{2,}\}`
	REGEX_PATTERN_FOR_CHARACTER_OTHER_THEN_X_OR_x                 = `\{[^xX]|{}\}`
	REGEX_PATTERN_FOR_IMAGE_TAG                                   = `^[a-zA-Z0-9]+[a-zA-Z0-9._-]*$`
)

var (
	ErrImagePathInUse = fmt.Errorf(ImageTagUnavailableMessage)
)

const (
	IMAGE_TAG_VARIABLE_NAME_X = "{X}"
	IMAGE_TAG_VARIABLE_NAME_x = "{x}"
)
