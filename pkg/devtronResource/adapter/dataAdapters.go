package adapter

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
)

func GetRequirementReqForCreateRequest(reqBean *bean.DtResourceObjectCreateReqBean, objectDataPath string, skipJsonSchemaValidation bool) *bean.ResourceObjectRequirementRequest {
	return &bean.ResourceObjectRequirementRequest{
		ReqBean: &bean.DtResourceObjectInternalBean{
			DevtronResourceObjectDescriptorBean: reqBean.DevtronResourceObjectDescriptorBean,
			Overview:                            reqBean.Overview,
			ParentConfig:                        reqBean.ParentConfig,
		},
		ObjectDataPath:           objectDataPath,
		SkipJsonSchemaValidation: skipJsonSchemaValidation,
	}
}

func GetRequirementRequestForCatalogRequest(reqBean *bean.DtResourceObjectCatalogReqBean, skipJsonSchemaValidation bool) *bean.ResourceObjectRequirementRequest {
	return &bean.ResourceObjectRequirementRequest{
		ReqBean: &bean.DtResourceObjectInternalBean{
			DevtronResourceObjectDescriptorBean: reqBean.DevtronResourceObjectDescriptorBean,
			ObjectData:                          reqBean.ObjectData,
		},
		ObjectDataPath:           bean.ResourceObjectMetadataPath,
		SkipJsonSchemaValidation: skipJsonSchemaValidation,
	}
}

func GetRequirementRequestForDependenciesRequest(reqBean *bean.DtResourceObjectDependenciesReqBean, skipJsonSchemaValidation bool) (*bean.ResourceObjectRequirementRequest, error) {
	marshaledDependencies, err := json.Marshal(reqBean.Dependencies)
	if err != nil {
		return nil, err
	}
	return &bean.ResourceObjectRequirementRequest{
		ReqBean: &bean.DtResourceObjectInternalBean{
			DevtronResourceObjectDescriptorBean: reqBean.DevtronResourceObjectDescriptorBean,
			Dependencies:                        reqBean.Dependencies,
			ObjectData:                          string(marshaledDependencies),
		},
		ObjectDataPath:           bean.ResourceObjectDependenciesPath,
		SkipJsonSchemaValidation: skipJsonSchemaValidation,
	}, nil
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
