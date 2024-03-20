package bean

type PromotionRequestWfMetadata struct {
	cdPipelineIds     []int
	authCdPipelineIds []int
}

func (p *PromotionRequestWfMetadata) SetCdPipelineIds(cdPipelineIds []int) {
	p.cdPipelineIds = cdPipelineIds
}

func (p *PromotionRequestWfMetadata) SetAuthCdPipelineIds(authCdPipelineIds []int) {
	p.authCdPipelineIds = authCdPipelineIds
}

func (p PromotionRequestWfMetadata) GetCdPipelineIds() []int {
	return p.cdPipelineIds
}

func (p PromotionRequestWfMetadata) GetAuthCdPipelineIds() []int {
	return p.authCdPipelineIds
}
