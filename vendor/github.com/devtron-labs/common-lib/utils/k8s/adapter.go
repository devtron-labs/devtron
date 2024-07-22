package k8s

func NewK8sRequestBean() *K8sRequestBean {
	return &K8sRequestBean{}
}

func (req *K8sRequestBean) WithResourceIdentifier(resourceIdentifier *ResourceIdentifier) *K8sRequestBean {
	if resourceIdentifier == nil {
		resourceIdentifier = &ResourceIdentifier{}
	}
	req.ResourceIdentifier = *resourceIdentifier
	return req
}

func NewResourceIdentifier() *ResourceIdentifier {
	return &ResourceIdentifier{}
}

func (req *ResourceIdentifier) WithName(name string) *ResourceIdentifier {
	req.Name = name
	return req
}

func (req *ResourceIdentifier) WithNameSpace(namespace string) *ResourceIdentifier {
	req.Namespace = namespace
	return req
}

func (req *ResourceIdentifier) WithGroup(group string) *ResourceIdentifier {
	req.GroupVersionKind.Group = group
	return req
}

func (req *ResourceIdentifier) WithVersion(version string) *ResourceIdentifier {
	req.GroupVersionKind.Version = version
	return req
}

func (req *ResourceIdentifier) WithKind(kind string) *ResourceIdentifier {
	req.GroupVersionKind.Kind = kind
	return req
}
