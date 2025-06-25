package adapter

import "github.com/devtron-labs/devtron/pkg/attributes"

func BuildResponseDTO(request *attributes.UserAttributesDto, mergedValue string) *attributes.UserAttributesDto {
	return &attributes.UserAttributesDto{
		EmailId: request.EmailId,
		Key:     request.Key,
		Value:   mergedValue,
		UserId:  request.UserId,
	}
}
