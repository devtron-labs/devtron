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
		reqBeanDescriptor.Id = id
		reqBeanDescriptor.OldObjectId = 0 // reqBean.Id and reqBean.OldObjectId both can not be used at a time

	} else {
		reqBeanDescriptor.IdType = bean.OldObjectId
		reqBeanDescriptor.OldObjectId = id // from FE, we are taking the id of the resource (devtronApp, helmApp, cluster, job) from their respective tables
		reqBeanDescriptor.Id = 0           // reqBean.Id and reqBean.OldObjectId both can not be used at a time

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
