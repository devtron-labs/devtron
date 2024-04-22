package adapter

import "github.com/devtron-labs/devtron/pkg/devtronResource/bean"

// SetIdTypeAndResourceIdBasedOnKind - here Resource Id can be
// either bean.DevtronResourceObjectDescriptorBean.(OldObjectId) // DevtronAppId, HelmAppId, ClusterId, CdPipelineId, DevtronJobId
// OR bean.DevtronResourceObjectDescriptorBean.(Id) // ReleaseId, ReleaseTrackId (refers to devtron_resource_object.id)
func SetIdTypeAndResourceIdBasedOnKind(reqBeanDescriptor *bean.DevtronResourceObjectDescriptorBean, id int) {
	if reqBeanDescriptor.Kind == bean.DevtronResourceRelease.ToString() || reqBeanDescriptor.Kind == bean.DevtronResourceReleaseTrack.ToString() {
		// for bean.DevtronResourceReleaseTrack and bean.DevtronResourceRelease
		// there is no OldObjectId, here the Id -> repository.DevtronResourceObject.Id (own id)
		reqBeanDescriptor.IdType = bean.ResourceObjectIdType
	} else {
		reqBeanDescriptor.IdType = bean.OldObjectId
	}
	reqBeanDescriptor.SetResourceIdBasedOnIdType(id)
}

func SetIdTypeForDependencies(reqBean *bean.DevtronResourceObjectBean) {
	//TODO : add common logic for resolving subKind
	for i := range reqBean.Dependencies {
		resourceKind := reqBean.Dependencies[i].DevtronResourceTypeReq.ResourceKind
		if resourceKind == bean.DevtronResourceRelease || resourceKind == bean.DevtronResourceReleaseTrack {
			reqBean.Dependencies[i].IdType = bean.ResourceObjectIdType
		} else {
			reqBean.Dependencies[i].IdType = bean.OldObjectId
		}
	}
	for i := range reqBean.ChildDependencies {
		resourceKind := reqBean.ChildDependencies[i].DevtronResourceTypeReq.ResourceKind
		if resourceKind == bean.DevtronResourceRelease || resourceKind == bean.DevtronResourceReleaseTrack {
			reqBean.ChildDependencies[i].IdType = bean.ResourceObjectIdType
		} else {
			reqBean.ChildDependencies[i].IdType = bean.OldObjectId
		}
		//TODO: dirty logic, improve
		for j := range reqBean.ChildDependencies[i].Dependencies {
			resourceKindNested := reqBean.ChildDependencies[i].Dependencies[j].DevtronResourceTypeReq.ResourceKind
			if resourceKindNested == bean.DevtronResourceRelease || resourceKindNested == bean.DevtronResourceReleaseTrack {
				reqBean.ChildDependencies[i].Dependencies[j].IdType = bean.ResourceObjectIdType
			} else {
				reqBean.ChildDependencies[i].Dependencies[j].IdType = bean.OldObjectId
			}
		}
	}
}

func RemoveRedundantFieldsAndSetDefaultForDependency(dep *bean.DevtronResourceDependencyBean, isChild bool) {
	dep.Metadata = nil //emptying in case UI sends the data back
	//since child dependencies are included separately in the payload and downstream are not declared explicitly setting this as upstream
	if isChild {
		dep.TypeOfDependency = bean.DevtronResourceDependencyTypeChild
	} else if len(dep.TypeOfDependency) == 0 {
		dep.TypeOfDependency = bean.DevtronResourceDependencyTypeUpstream //assuming one level of nesting
	}
}
