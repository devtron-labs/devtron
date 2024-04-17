package adapter

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
)

func GetResourceObjectRequirementRequest(reqBean *bean.DevtronResourceObjectBean, objectDataPath string, skipJsonSchemaValidation bool) *bean.ResourceObjectRequirementRequest {
	return &bean.ResourceObjectRequirementRequest{
		ReqBean:                  reqBean,
		ObjectDataPath:           objectDataPath,
		SkipJsonSchemaValidation: skipJsonSchemaValidation,
	}
}

func BuildConfigStatusSchemaData(status *bean.ConfigStatus) *bean.ReleaseConfigSchema {
	return &bean.ReleaseConfigSchema{
		Status:  status.Status.ToString(),
		Comment: status.Comment,
		Lock:    status.IsLocked,
	}
}

func BuildDependencyData(name string, id, devtronResourceId, devtronResourceSchemaId int,
	maxIndex float64, dependencyType bean.DevtronResourceDependencyType, idType bean.IdType) *bean.DevtronResourceDependencyBean {
	maxIndex++
	return &bean.DevtronResourceDependencyBean{
		Name:                    name,
		OldObjectId:             id,
		IdType:                  idType,
		Index:                   int(maxIndex),
		TypeOfDependency:        dependencyType,
		DevtronResourceId:       devtronResourceId,
		DevtronResourceSchemaId: devtronResourceSchemaId,
		Dependencies:            make([]*bean.DevtronResourceDependencyBean, 0),
	}
}

func GetSuccessPassResponse() *bean.SuccessResponse {
	return &bean.SuccessResponse{
		Success: true,
	}
}

func GetSuccessFailResponse() *bean.SuccessResponse {
	return &bean.SuccessResponse{
		Success: false,
	}
}

func BuildUserSchemaData(id int32, emailId string) *bean.UserSchema {
	return &bean.UserSchema{
		Id:   id,
		Name: emailId,
		Icon: true,
	}
}
