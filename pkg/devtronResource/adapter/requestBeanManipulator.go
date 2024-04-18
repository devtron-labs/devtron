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
