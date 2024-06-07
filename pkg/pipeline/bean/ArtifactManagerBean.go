/*
 * Copyright (c) 2024. Devtron Inc.
 */

package bean

type PromotionRequestWfMetadata struct {
	cdPipelineIds       []int
	authCdPipelineIds   []int
	hasProdEnv          bool
	pipelineIdToEnvName map[int]string
}

func (p *PromotionRequestWfMetadata) WithCdPipelineIds(cdPipelineIds []int) *PromotionRequestWfMetadata {
	p.cdPipelineIds = cdPipelineIds
	return p
}

func (p *PromotionRequestWfMetadata) WithAuthCdPipelineIds(authCdPipelineIds []int) *PromotionRequestWfMetadata {
	p.authCdPipelineIds = authCdPipelineIds
	return p
}

func (p *PromotionRequestWfMetadata) WithPipelineIdToEnvNameMap(mapping map[int]string) *PromotionRequestWfMetadata {
	p.pipelineIdToEnvName = mapping
	return p
}

func (p *PromotionRequestWfMetadata) WithHasProdEnv(hasProdEnv bool) *PromotionRequestWfMetadata {
	p.hasProdEnv = hasProdEnv
	return p
}

func (p *PromotionRequestWfMetadata) GetCdPipelineIds() []int {
	return p.cdPipelineIds
}

func (p *PromotionRequestWfMetadata) GetPipelineIdToEnvNameMap() map[int]string {
	return p.pipelineIdToEnvName
}

func (p *PromotionRequestWfMetadata) GetAuthCdPipelineIds() []int {
	return p.authCdPipelineIds
}

func (p *PromotionRequestWfMetadata) GetHasProdEnv() bool {
	return p.hasProdEnv
}
