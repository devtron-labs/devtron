package adapter

import "github.com/devtron-labs/devtron/pkg/devtronResource/bean"

func removeInternalOnlyFieldsFromDescriptorObjectBean(reqBean *bean.DevtronResourceObjectDescriptorBean) {
	if len(reqBean.Identifier) != 0 {
		reqBean.Identifier = ""
	}
	if len(reqBean.IdType) != 0 {
		reqBean.IdType = ""
	}
}

func RemoveInternalOnlyFieldsFromGetResourceObjectBean(reqBean *bean.DevtronResourceObjectGetAPIBean) *bean.DevtronResourceObjectGetAPIBean {
	if reqBean.DevtronResourceObjectDescriptorBean != nil {
		removeInternalOnlyFieldsFromDescriptorObjectBean(reqBean.DevtronResourceObjectDescriptorBean)
	}
	return reqBean
}

func removeInternalOnlyFieldsFromDependencyObjectBean(dependencies []*bean.DevtronResourceDependencyBean) {
	for i := range dependencies {
		if len(dependencies[i].IdType) != 0 {
			dependencies[i].IdType = ""
		}
		if dependencies[i].Dependencies != nil || len(dependencies[i].Dependencies) != 0 {
			removeInternalOnlyFieldsFromDependencyObjectBean(dependencies[i].Dependencies)
		}
	}
}

func RemoveInternalOnlyFieldsFromResourceObjectBean(reqBean *bean.DevtronResourceObjectBean) *bean.DevtronResourceObjectBean {
	if reqBean.DevtronResourceObjectDescriptorBean != nil {
		removeInternalOnlyFieldsFromDescriptorObjectBean(reqBean.DevtronResourceObjectDescriptorBean)
	}
	if reqBean.Dependencies != nil || len(reqBean.Dependencies) != 0 {
		removeInternalOnlyFieldsFromDependencyObjectBean(reqBean.Dependencies)
	}
	if reqBean.ChildDependencies != nil || len(reqBean.ChildDependencies) != 0 {
		removeInternalOnlyFieldsFromDependencyObjectBean(reqBean.ChildDependencies)
	}
	return reqBean
}
