package generateManifest

func ConvertPointerDeploymentTemplateResponseToNonPointer(r *DeploymentTemplateResponse) DeploymentTemplateResponse {
	if r != nil {
		return DeploymentTemplateResponse{
			Data:                r.Data,
			ResolvedData:        r.ResolvedData,
			VariableSnapshot:    r.VariableSnapshot,
			TemplateVersion:     r.TemplateVersion,
			IsAppMetricsEnabled: r.IsAppMetricsEnabled,
		}
	}
	return DeploymentTemplateResponse{}
}
