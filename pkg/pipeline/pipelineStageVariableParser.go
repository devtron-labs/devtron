package pipeline

import (
	"errors"
	"fmt"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/plugin"
	"github.com/go-pg/pg"
	errors1 "github.com/juju/errors"
	"go.uber.org/zap"
	"strings"
)

type copyContainerImagePluginInputVariable = string
type RefPluginName = string

const (
	COPY_CONTAINER_IMAGE RefPluginName = "Copy container image"
	EMPTY_STRING                       = " "
)

const (
	DESTINATION_INFO                copyContainerImagePluginInputVariable = "DESTINATION_INFO"
	SOURCE_REGISTRY_CREDENTIALS_KEY                                       = "SOURCE_REGISTRY_CREDENTIAL"
)

type PluginInputVariableParser interface {
	HandleCopyContainerImagePluginInputVariables(inputVariables []*bean.VariableObject, dockerImageTag string, pluginTriggerImage string, sourceImageDockerRegistry string) (registryDestinationImageMap map[string][]string, registryCredentials map[string]plugin.RegistryCredentials, err error)
}

type PluginInputVariableParserImpl struct {
	logger               *zap.SugaredLogger
	dockerRegistryConfig DockerRegistryConfig
	customTagService     CustomTagService
}

func NewPluginInputVariableParserImpl(
	logger *zap.SugaredLogger,
	dockerRegistryConfig DockerRegistryConfig,
	customTagService CustomTagService,
) *PluginInputVariableParserImpl {
	return &PluginInputVariableParserImpl{
		logger:               logger,
		dockerRegistryConfig: dockerRegistryConfig,
		customTagService:     customTagService,
	}
}

func (impl *PluginInputVariableParserImpl) HandleCopyContainerImagePluginInputVariables(inputVariables []*bean.VariableObject,
	dockerImageTag string,
	pluginTriggerImage string,
	sourceImageDockerRegistry string) (registryDestinationImageMap map[string][]string, registryCredentials map[string]plugin.RegistryCredentials, err error) {

	var DestinationInfo string
	for _, ipVariable := range inputVariables {
		if ipVariable.Name == DESTINATION_INFO {
			DestinationInfo = ipVariable.Value
		}
	}

	if len(pluginTriggerImage) == 0 {
		return nil, nil, errors.New("no image provided during trigger time")
	}

	if len(DestinationInfo) == 0 {
		return nil, nil, errors.New("destination info now")
	}

	if len(dockerImageTag) == 0 {
		// case when custom tag is not configured - source image tag will be taken as docker image tag
		pluginTriggerImageSplit := strings.Split(pluginTriggerImage, ":")
		dockerImageTag = pluginTriggerImageSplit[len(pluginTriggerImageSplit)-1]
	}

	registryRepoMapping := impl.getRegistryRepoMapping(DestinationInfo)
	registryCredentials, err = impl.getRegistryDetails(registryRepoMapping, sourceImageDockerRegistry)
	if err != nil {
		return nil, nil, err
	}
	registryDestinationImageMap = impl.getRegistryDestinationImageMapping(registryRepoMapping, dockerImageTag, registryCredentials)

	err = impl.createEcrRepoIfRequired(registryCredentials, registryRepoMapping)
	if err != nil {
		impl.logger.Errorw("error in creating ecr repo", "err", err)
		return registryDestinationImageMap, registryCredentials, err
	}

	return registryDestinationImageMap, registryCredentials, nil
}

func (impl *PluginInputVariableParserImpl) getRegistryRepoMapping(destinationInfo string) map[string][]string {
	/*
		creating map with registry as key and list of repositories in that registry where we need to copy image
			destinationInfo format (each registry detail is separated by new line) :
				<registryName1> | <comma separated repoNames>
				<registryName2> | <comma separated repoNames>
	*/
	destinationRegistryRepositoryMap := make(map[string][]string)
	destinationRegistryRepoDetails := strings.Split(destinationInfo, "\n")
	for _, detail := range destinationRegistryRepoDetails {
		registryRepoSplit := strings.Split(detail, "|")
		registryName := strings.Trim(registryRepoSplit[0], EMPTY_STRING)
		repositoryValuesSplit := strings.Split(registryRepoSplit[1], ",")
		var repositories []string
		for _, repositoryName := range repositoryValuesSplit {
			repositoryName = strings.Trim(repositoryName, EMPTY_STRING)
			repositories = append(repositories, repositoryName)
		}
		destinationRegistryRepositoryMap[registryName] = repositories
	}
	return destinationRegistryRepositoryMap
}

func (impl *PluginInputVariableParserImpl) getRegistryDetails(destinationRegistryRepositoryMap map[string][]string, sourceRegistry string) (map[string]plugin.RegistryCredentials, error) {
	registryCredentialsMap := make(map[string]plugin.RegistryCredentials)
	//saving source registry credentials
	sourceRegistry = strings.Trim(sourceRegistry, " ")
	sourceRegistryCredentials, err := impl.getPluginRegistryCredentialsByRegistryName(sourceRegistry)
	if err != nil {
		return nil, err
	}
	registryCredentialsMap[SOURCE_REGISTRY_CREDENTIALS_KEY] = *sourceRegistryCredentials

	// saving destination registry credentials; destinationRegistryRepositoryMap -> map[registryName]= [<repo1>, <repo2>]
	for registry, _ := range destinationRegistryRepositoryMap {
		destinationRegistryCredential, err := impl.getPluginRegistryCredentialsByRegistryName(registry)
		if err != nil {
			return nil, err
		}
		registryCredentialsMap[registry] = *destinationRegistryCredential
	}
	return registryCredentialsMap, nil
}

func (impl *PluginInputVariableParserImpl) getPluginRegistryCredentialsByRegistryName(registryName string) (*plugin.RegistryCredentials, error) {
	registryCredentials, err := impl.dockerRegistryConfig.FetchOneDockerAccount(registryName)
	if err != nil {
		impl.logger.Errorw("error in fetching registry details by registry name", "err", err)
		if err == pg.ErrNoRows {
			return nil, fmt.Errorf("invalid registry name: registry details not found in global container registries")
		}
		return nil, err
	}
	return &plugin.RegistryCredentials{
		RegistryType:       string(registryCredentials.RegistryType),
		RegistryURL:        registryCredentials.RegistryURL,
		Username:           registryCredentials.Username,
		Password:           registryCredentials.Password,
		AWSRegion:          registryCredentials.AWSRegion,
		AWSSecretAccessKey: registryCredentials.AWSSecretAccessKey,
		AWSAccessKeyId:     registryCredentials.AWSAccessKeyId,
	}, nil
}

func (impl *PluginInputVariableParserImpl) getRegistryDestinationImageMapping(
	registryRepoMapping map[string][]string,
	dockerImageTag string,
	registryCredentials map[string]plugin.RegistryCredentials) map[string][]string {

	// creating map with registry as key and list of destination images in that registry
	registryDestinationImageMapping := make(map[string][]string)
	for registry, destinationRepositories := range registryRepoMapping {
		registryCredential := registryCredentials[registry]
		var destinationImages []string
		for _, repo := range destinationRepositories {
			destinationImage := fmt.Sprintf("%s/%s:%s", registryCredential.RegistryURL, repo, dockerImageTag)
			destinationImages = append(destinationImages, destinationImage)
		}
		registryDestinationImageMapping[registry] = destinationImages
	}

	return registryDestinationImageMapping
}

func (impl *PluginInputVariableParserImpl) createEcrRepoIfRequired(registryCredentials map[string]plugin.RegistryCredentials, registryRepoMapping map[string][]string) error {
	for registry, registryCredential := range registryCredentials {
		if registryCredential.RegistryType == dockerRegistryRepository.REGISTRYTYPE_ECR {
			repositories := registryRepoMapping[registry]
			for _, dockerRepo := range repositories {
				err := util.CreateEcrRepo(dockerRepo, registryCredential.AWSRegion, registryCredential.AWSAccessKeyId, registryCredential.AWSSecretAccessKey)
				if err != nil {
					if errors1.IsAlreadyExists(err) {
						impl.logger.Warnw("this repo already exists!!, skipping repo creation", "repo", dockerRepo)
					} else {
						impl.logger.Errorw("ecr repo creation failed, it might be due to authorization or any other external "+
							"dependency. please create repo manually before triggering ci", "repo", dockerRepo, "err", err)
						return err
					}
				}
			}
		}
	}
	return nil
}
