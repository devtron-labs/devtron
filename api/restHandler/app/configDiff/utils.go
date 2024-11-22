package configDiff

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/configDiff/bean"
)

var validConfigCategories = map[string]bool{"secret": true, "cm": true, "dt": true, "ps": true}
var ErrInvalidConfigCategory = errors.New("invalid config category provided")
var ErrInvalidComparisonItems = errors.New("invalid comparison items, only 2 items are supported for comparison")

func validateComparisonRequest(configCategory string, comparisonRequestDto bean.ComparisonRequestDto) error {
	if ok := validConfigCategories[configCategory]; !ok {
		return ErrInvalidConfigCategory
	}
	if len(comparisonRequestDto.ComparisonItems) > 2 {
		return ErrInvalidComparisonItems
	}
	return nil
}
