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

func BuildDependencyData(id, devtronResourceId, devtronResourceSchemaId int, maxIndex float64, dependencyType bean.DevtronResourceDependencyType, idType bean.IdType) *bean.DevtronResourceDependencyBean {
	maxIndex++
	return &bean.DevtronResourceDependencyBean{
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

func BuildUserSchemaData(id int32, emailId string) *bean.UserSchema {
	return &bean.UserSchema{
		Id:   id,
		Name: emailId,
		Icon: true,
	}
}
func BuildFilterCriteriaDecoder(resource, identifierType, value string) *bean.FilterCriteriaDecoder {
	return &bean.FilterCriteriaDecoder{
		Resource: bean.DevtronResourceKind(resource),
		Type:     bean.FilterCriteriaIdentifier(identifierType),
		Value:    value,
	}
}

func BuildDevtronResourceObjectGetAPIBean() *bean.DevtronResourceObjectGetAPIBean {
	return &bean.DevtronResourceObjectGetAPIBean{
		DevtronResourceObjectDescriptorBean: &bean.DevtronResourceObjectDescriptorBean{},
		DevtronResourceObjectBasicDataBean: &bean.DevtronResourceObjectBasicDataBean{
			Overview: &bean.ResourceOverview{},
		},
	}
}

func BuildDevtronResourceObjectDescriptorBean(id int, kind, subKind bean.DevtronResourceKind,
	version bean.DevtronResourceVersion, userId int32) *bean.DevtronResourceObjectDescriptorBean {
	reqBean := &bean.DevtronResourceObjectDescriptorBean{
		Kind:    kind.ToString(),
		SubKind: subKind.ToString(),
		Version: version.ToString(),
		UserId:  userId,
	}
	SetIdTypeAndResourceIdBasedOnKind(reqBean, id)
	return reqBean
}
