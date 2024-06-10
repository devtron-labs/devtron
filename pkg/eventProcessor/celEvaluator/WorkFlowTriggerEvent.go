package celEvaluator

import (
	"errors"
	"github.com/devtron-labs/devtron/cel"
	repository "github.com/devtron-labs/devtron/internal/sql/repository/imageTagging"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/attributes"
	attributesBean "github.com/devtron-labs/devtron/pkg/attributes/bean"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strings"
)

type TriggerEventEvaluator interface {
	IsPriorityDeployment(valuesOverrideResponse *app.ValuesOverrideResponse) (isPriorityEvent bool, err error)
}

type TriggerEventEvaluatorImpl struct {
	logger              *zap.SugaredLogger
	imageTagRepository  repository.ImageTaggingRepository // TODO: fix import cycle issue for pipeline.ImageTaggingService
	teamService         team.TeamService
	attributesService   attributes.AttributesService
	celEvaluatorService cel.EvaluatorService
}

func NewTriggerEventEvaluatorImpl(logger *zap.SugaredLogger,
	imageTagRepository repository.ImageTaggingRepository,
	teamService team.TeamService,
	attributesService attributes.AttributesService,
	celEvaluatorService cel.EvaluatorService) (*TriggerEventEvaluatorImpl, error) {
	impl := &TriggerEventEvaluatorImpl{
		logger:              logger,
		imageTagRepository:  imageTagRepository,
		teamService:         teamService,
		attributesService:   attributesService,
		celEvaluatorService: celEvaluatorService,
	}
	return impl, nil
}

func (impl *TriggerEventEvaluatorImpl) IsPriorityDeployment(valuesOverrideResponse *app.ValuesOverrideResponse) (isPriorityEvent bool, err error) {
	expression, err := impl.getPriorityDeploymentExpression()
	if err != nil {
		impl.logger.Errorw("error while getting priority deployment CEL expression", "err", err)
		return isPriorityEvent, err
	}
	params, err := impl.getParamsForPriorityDeployment(valuesOverrideResponse)
	if err != nil {
		impl.logger.Errorw("error while getting priority deployment CEL expression metadata", "err", err)
		return isPriorityEvent, err
	}
	evalReq := cel.Request{
		Expression: expression,
		ExpressionMetadata: cel.ExpressionMetadata{
			Params: params,
		},
	}
	return impl.celEvaluatorService.EvaluateCELRequest(evalReq)
}

func (impl *TriggerEventEvaluatorImpl) getPriorityDeploymentExpression() (string, error) {
	attribute, err := impl.attributesService.GetByKey(attributesBean.PRIORITY_DEPLOYMENT_CONDITION)
	if err != nil {
		impl.logger.Errorw("error while getting attribute by key", "key", attributesBean.PRIORITY_DEPLOYMENT_CONDITION, "error", err)
		return "", err
	}
	return attribute.Value, nil
}

func (impl *TriggerEventEvaluatorImpl) getParamsForPriorityDeployment(valuesOverrideResponse *app.ValuesOverrideResponse) ([]cel.ExpressionParam, error) {
	imageReleaseTags, err := impl.imageTagRepository.GetTagsByArtifactId(valuesOverrideResponse.Artifact.Id)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching image tags using artifactId", "err", err, "artifactId", valuesOverrideResponse.Artifact.Id)
		return nil, err
	}
	imageLabels := make([]string, 0, len(imageReleaseTags))
	for _, imageTag := range imageReleaseTags {
		imageLabels = append(imageLabels, imageTag.TagName)
	}
	project, err := impl.teamService.FetchOne(valuesOverrideResponse.Pipeline.App.TeamId)
	if err != nil {
		impl.logger.Errorw("error while getting project", "projectId", valuesOverrideResponse.Pipeline.App.TeamId, "err", err)
		return nil, err
	}
	lastColonIndex := strings.LastIndex(valuesOverrideResponse.Artifact.Image, ":")
	var containerRepository, containerImageTag string
	if len(valuesOverrideResponse.Artifact.Image)-1 > lastColonIndex && lastColonIndex != 1 {
		containerRepository = valuesOverrideResponse.Artifact.Image[:lastColonIndex]
		containerImageTag = valuesOverrideResponse.Artifact.Image[lastColonIndex+1:]
	}
	containerImage := valuesOverrideResponse.Artifact.Image
	params := []cel.ExpressionParam{
		{
			ParamName: cel.AppName,
			Value:     valuesOverrideResponse.Pipeline.App.AppName,
			Type:      cel.ParamTypeString,
		},
		{
			ParamName: cel.ProjectName,
			Value:     project.Name,
			Type:      cel.ParamTypeString,
		},
		{
			ParamName: cel.EnvName,
			Value:     valuesOverrideResponse.EnvOverride.Environment.Name,
			Type:      cel.ParamTypeString,
		},
		{
			ParamName: cel.CdPipelineName,
			Value:     valuesOverrideResponse.Pipeline.Name,
			Type:      cel.ParamTypeString,
		},
		{
			ParamName: cel.CdPipelineTriggerType,
			Value:     valuesOverrideResponse.Pipeline.TriggerType.ToString(),
			Type:      cel.ParamTypeString,
		},
		{
			ParamName: cel.IsProdEnv,
			Value:     valuesOverrideResponse.EnvOverride.Environment.Default,
			Type:      cel.ParamTypeBool,
		},
		{
			ParamName: cel.ClusterName,
			Value:     valuesOverrideResponse.EnvOverride.Environment.Cluster.ClusterName,
			Type:      cel.ParamTypeString,
		},
		{
			ParamName: cel.ChartRefId,
			Value:     valuesOverrideResponse.EnvOverride.Chart.ChartRefId,
			Type:      cel.ParamTypeInteger,
		},
		{
			ParamName: cel.ContainerRepo,
			Value:     containerRepository,
			Type:      cel.ParamTypeString,
		},
		{
			ParamName: cel.ContainerImage,
			Value:     containerImage,
			Type:      cel.ParamTypeString,
		},
		{
			ParamName: cel.ContainerImageTag,
			Value:     containerImageTag,
			Type:      cel.ParamTypeString,
		},
		{
			ParamName: cel.ImageLabels,
			Value:     imageLabels,
			Type:      cel.ParamTypeList,
		},
	}
	return params, nil
}
