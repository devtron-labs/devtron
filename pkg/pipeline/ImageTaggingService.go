package pipeline

import (
	repository "github.com/devtron-labs/devtron/internal/sql/repository/imageTagging"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/go-pg/pg"
)

type ImageTaggingResponseDTO struct {
	ImageReleaseTags []repository.ImageTag   `json:"imageReleaseTags"`
	AppReleaseTags   []repository.ImageTag   `json:"appReleaseTags"`
	ImageComment     repository.ImageComment `json:"imageComments"`
	ProdEnvExists    bool                    `json:"prodEnvExists"`
}

type ImageTaggingRequestDTO struct {
	CreateTags     []repository.ImageTag
	SoftDeleteTags []repository.ImageTag
	ImageComment   []repository.ImageComment
	HardDeleteTags []repository.ImageTag
}

type ImageTaggingService interface {
	GetTagsData(ciPipelineId, appId, artifactId int) (*ImageTaggingResponseDTO, error)
	CreateUpdateImageTagging(ciPipelineId, appId, artifactId int, iamgeTaggingRequest *ImageTaggingRequestDTO) (*ImageTaggingResponseDTO, error)
	GetEnvFromParentAndLinkedWorkflow(ciPipelineId int) (bool, error)
}

type ImageTaggingServiceImpl struct {
	imageTaggingRepo      repository.ImageTaggingRepository
	ciPipelineRepository  pipelineConfig.CiPipelineRepository
	environmentRepository repository2.EnvironmentRepository
}

func NewImageTaggingServiceImpl(imageTaggingRepo repository.ImageTaggingRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	environmentRepository repository2.EnvironmentRepository) *ImageTaggingServiceImpl {
	return &ImageTaggingServiceImpl{
		imageTaggingRepo:      imageTaggingRepo,
		ciPipelineRepository:  ciPipelineRepository,
		environmentRepository: environmentRepository,
	}
}

func (impl ImageTaggingServiceImpl) GetTagsData(ciPipelineId, appId, artifactId int) (*ImageTaggingResponseDTO, error) {
	resp := &ImageTaggingResponseDTO{}
	imageComment, err := impl.imageTaggingRepo.GetImageComment(artifactId)
	if err != nil && err != pg.ErrNoRows {
		//log error
		return resp, err
	}
	appReleaseTags, err := impl.imageTaggingRepo.GetTagsByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		//log error
		return resp, err
	}
	imageReleaseTags, err := impl.imageTaggingRepo.GetTagsByArtifactId(artifactId)
	if err != nil && err != pg.ErrNoRows {
		//log error
		return resp, err
	}
	prodEnvExists, err := impl.GetEnvFromParentAndLinkedWorkflow(ciPipelineId)
	if err != nil {
		//log error
		return resp, err
	}
	resp.AppReleaseTags = appReleaseTags
	resp.ImageReleaseTags = imageReleaseTags
	resp.ImageComment = imageComment
	resp.ProdEnvExists = prodEnvExists
	return resp, err
}

func (impl ImageTaggingServiceImpl) CreateUpdateImageTagging(ciPipelineId, appId, artifactId int, iamgeTaggingRequest *ImageTaggingRequestDTO) (*ImageTaggingResponseDTO, error) {
	return impl.GetTagsData(ciPipelineId, appId, artifactId)
}
func (impl ImageTaggingServiceImpl) GetEnvFromParentAndLinkedWorkflow(ciPipelineId int) (bool, error) {
	prodEnvExists := false
	pipelines, err := impl.ciPipelineRepository.FindByParentCiPipelineId(ciPipelineId)
	if err != nil {
		//add log
		return prodEnvExists, err
	}

	//get all the pipeline ids liked with the requested ciPipelineId
	pipelineIds := make([]int, len(pipelines)+1)
	pipelineIds[0] = ciPipelineId
	for i := 0; i < len(pipelines); i++ {
		pipelineIds[i+1] = pipelines[i].Id
	}

	envs, err := impl.environmentRepository.FindEnvLinkedWithCiPipelines(pipelineIds)
	if err != nil {
		//add log
		return prodEnvExists, err
	}

	for _, env := range envs {
		//env id prod ,return true
		if env.Default {
			prodEnvExists = true
			break
		}
	}

	return prodEnvExists, nil

}
