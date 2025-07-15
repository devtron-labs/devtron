package adapter

import (
	"github.com/devtron-labs/devtron/pkg/attributes/bean"
)

func BuildResponseDTO(request *bean.UserAttributesDto, mergedValue string) *bean.UserAttributesDto {
	return &bean.UserAttributesDto{
		EmailId: request.EmailId,
		Key:     request.Key,
		Value:   mergedValue,
		UserId:  request.UserId,
	}
}
