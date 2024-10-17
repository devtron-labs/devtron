package generateManifest

func ConvertPointerDeploymentTemplateResponseToNonPointer(r *DeploymentTemplateResponse) DeploymentTemplateResponse {
	if r != nil {
		return *r
	}
	return DeploymentTemplateResponse{}
}
