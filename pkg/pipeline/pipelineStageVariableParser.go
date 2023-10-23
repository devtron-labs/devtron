package pipeline

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/pkg"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/plugin"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strings"
)

type SkopeoInputVariable = string
type RefPluginName = string

const (
	SKOPEO RefPluginName = "Skopeo"
)

const (
	DESTINATION_INFO SkopeoInputVariable = "DESTINATION_INFO"
	SOURCE_INFO      SkopeoInputVariable = "SOURCE_INFO"
)

type PluginInputVariableParser interface {
	ParseSkopeoPluginInputVariables(inputVariables []*bean.VariableObject, entityKey int, entityValue string, pluginTriggerImage string, buildConfigurationRegistry string) (map[string][]string, map[string]plugin.RegistryCredentials, error)
}

type PluginInputVariableParserImpl struct {
	logger               *zap.SugaredLogger
	dockerRegistryConfig DockerRegistryConfig
	customTagService     pkg.CustomTagService
}

func NewPluginInputVariableParserImpl(
	logger *zap.SugaredLogger,
	dockerRegistryConfig DockerRegistryConfig,
	customTagService pkg.CustomTagService,
) *PluginInputVariableParserImpl {
	return &PluginInputVariableParserImpl{
		logger:               logger,
		dockerRegistryConfig: dockerRegistryConfig,
		customTagService:     customTagService,
	}
}

func (impl *PluginInputVariableParserImpl) ParseSkopeoPluginInputVariables(inputVariables []*bean.VariableObject, entityKey int, entityValue string, pluginTriggerImage string, buildConfigurationRegistry string) (map[string][]string, map[string]plugin.RegistryCredentials, error) {
	var DestinationInfo, SourceRegistry, SourceImage string
	for _, ipVariable := range inputVariables {
		if ipVariable.Name == DESTINATION_INFO {
			DestinationInfo = ipVariable.Value
		} else if ipVariable.Name == SOURCE_INFO {
			if len(pluginTriggerImage) == 0 {
				if len(ipVariable.Value) == 0 {
					impl.logger.Errorw("No image provided in source or during trigger time")
					return nil, nil, errors.New("no image provided in source or during trigger time")
				}
				SourceInfo := ipVariable.Value
				SourceInfoSplit := strings.Split(SourceInfo, "|")
				SourceImage = SourceInfoSplit[len(SourceInfoSplit)-1]
				SourceRegistry = SourceInfoSplit[0]
			} else {
				SourceImage = pluginTriggerImage
				SourceRegistry = buildConfigurationRegistry
			}
		}
	}
	registryDestinationImageMap, registryCredentialMap, err := impl.getRegistryDetailsAndDestinationImagePathForSkopeo(entityKey, entityValue, SourceImage, SourceRegistry, DestinationInfo)
	if err != nil {
		impl.logger.Errorw("Error in parsing skopeo input variables")
		return nil, nil, err
	}
	return registryDestinationImageMap, registryCredentialMap, nil
}

func (impl *PluginInputVariableParserImpl) getRegistryDetailsAndDestinationImagePathForSkopeo(entityKey int, entityValue, sourceImage string, sourceRegistry string, destinationInfo string) (registryDestinationImageMap map[string][]string, registryCredentialsMap map[string]plugin.RegistryCredentials, err error) {
	registryDestinationImageMap = make(map[string][]string)
	registryCredentialsMap = make(map[string]plugin.RegistryCredentials)

	var sourceImageTag string
	sourceSplit := strings.Split(sourceImage, ":")
	sourceImageTag = sourceSplit[len(sourceSplit)-1]

	//saving source registry credentials
	registryCredentials, err := impl.dockerRegistryConfig.FetchOneDockerAccount(sourceRegistry)
	if err != nil {
		impl.logger.Errorw("error in fetching registry details by registry name", "err", err)
		return registryDestinationImageMap, registryCredentialsMap, err
	}
	registryCredentialsMap["SOURCE_REGISTRY_CREDENTIAL"] = plugin.RegistryCredentials{
		RegistryType:       string(registryCredentials.RegistryType),
		RegistryURL:        registryCredentials.RegistryURL,
		Username:           registryCredentials.Username,
		Password:           registryCredentials.Password,
		AWSRegion:          registryCredentials.AWSRegion,
		AWSSecretAccessKey: registryCredentials.AWSSecretAccessKey,
		AWSAccessKeyId:     registryCredentials.AWSAccessKeyId,
	}

	registryRepoDetails := strings.Split(destinationInfo, "\n")
	for _, detail := range registryRepoDetails {
		registryRepoSplit := strings.Split(detail, "|")
		registryName := strings.Trim(registryRepoSplit[0], " ")
		registryCredentials, err := impl.dockerRegistryConfig.FetchOneDockerAccount(registryName)
		if err != nil {
			impl.logger.Errorw("error in fetching registry details by registry name", "err", err)
			return registryDestinationImageMap, registryCredentialsMap, err
		}
		var destinationImages []string
		repositoryValues := registryRepoSplit[1]
		repositoryValuesSplit := strings.Split(repositoryValues, ",")

		for _, repositoryName := range repositoryValuesSplit {
			repositoryName = strings.Trim(repositoryName, " ")
			customTag, err := impl.customTagService.GetCustomTagByEntityKeyAndValue(entityKey, entityValue)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error in fetching custom tag by entity key and entity value ", "entityKey", entityKey, "entityValue", entityValue)
				return registryDestinationImageMap, registryCredentialsMap, err
			}
			var destinationImage string
			if customTag != nil && customTag.Id == 0 {
				destinationImage = fmt.Sprintf("%s/%s:%s", registryCredentials.RegistryURL, repositoryName, sourceImageTag)
			} else {
				imagePathReservation, err := impl.customTagService.GenerateImagePath(entityKey, entityValue, registryCredentials.RegistryURL, repositoryName)
				if err != nil {
					impl.logger.Errorw("error in reserving image path", "err", err)
					return registryDestinationImageMap, registryCredentialsMap, err
				}
				destinationImage = imagePathReservation.ImagePath
			}
			destinationImages = append(destinationImages, destinationImage)
		}
		registryDestinationImageMap[registryName] = destinationImages
		registryCredentialsMap[registryName] = plugin.RegistryCredentials{
			RegistryType:       string(registryCredentials.RegistryType),
			RegistryURL:        registryCredentials.RegistryURL,
			Username:           registryCredentials.Username,
			Password:           registryCredentials.Password,
			AWSRegion:          registryCredentials.AWSRegion,
			AWSSecretAccessKey: registryCredentials.AWSSecretAccessKey,
			AWSAccessKeyId:     registryCredentials.AWSAccessKeyId,
		}
	}
	//adding source registry details
	return registryDestinationImageMap, registryCredentialsMap, nil
}
