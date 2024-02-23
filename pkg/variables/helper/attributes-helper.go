package helper

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/variables/models"
)

func GetAttributeNames(payload models.Payload) ([]string, []string, []string) {
	appNames := make([]string, 0)
	envNames := make([]string, 0)
	clusterNames := make([]string, 0)
	for _, variable := range payload.Variables {
		for _, value := range variable.AttributeValues {
			for identifierType, _ := range value.AttributeParams {
				if identifierType == models.ApplicationName {
					appNames = append(appNames, value.AttributeParams[identifierType])
				} else if identifierType == models.EnvName {
					envNames = append(envNames, value.AttributeParams[identifierType])
				} else if identifierType == models.ClusterName {
					clusterNames = append(clusterNames, value.AttributeParams[identifierType])
				}
			}

		}
	}
	return appNames, envNames, clusterNames
}

type AttributesMappings struct {
	AppNameToId     map[string]int
	EnvNameToId     map[string]int
	ClusterNameToId map[string]int
}

func GetIdentifierValueV1(identifierType models.IdentifierType, identifierName string, attributesMapping *AttributesMappings) (int, error) {
	var found bool
	var identifierValue int
	if identifierType == models.ApplicationName {
		identifierValue, found = attributesMapping.AppNameToId[identifierName]
		if !found {
			return 0, fmt.Errorf("ApplicationName mapping not found %s", identifierName)
		}
	} else if identifierType == models.EnvName {
		identifierValue, found = attributesMapping.EnvNameToId[identifierName]
		if !found {
			return 0, fmt.Errorf("EnvName mapping not found %s", identifierName)
		}
	} else if identifierType == models.ClusterName {
		identifierValue, found = attributesMapping.ClusterNameToId[identifierName]
		if !found {
			return 0, fmt.Errorf("ClusterName mapping not found %s", identifierName)
		}
	} else {
		return 0, fmt.Errorf("invalid identifierType")
	}
	return identifierValue, nil
}

// , appNameToIdMap map[string]int, envNameToIdMap map[string]int, clusterNameToIdMap map[string]int
func GetIdentifierValue(identifierType models.IdentifierType, identifierName string, attributesMapping *AttributesMappings) (int, error) {

	var identifierValue int
	var ok bool
	switch identifierType {
	case models.ApplicationName:
		if identifierValue, ok = attributesMapping.AppNameToId[identifierName]; !ok {
			return 0, fmt.Errorf("ApplicationName mapping not found %s", identifierName)
		}
	case models.EnvName:
		if identifierValue, ok = attributesMapping.EnvNameToId[identifierName]; !ok {
			return 0, fmt.Errorf("EnvName mapping not found %s", identifierName)
		}
	case models.ClusterName:
		if identifierValue, ok = attributesMapping.ClusterNameToId[identifierName]; !ok {
			return 0, fmt.Errorf("ClusterName mapping not found %s", identifierName)
		}
	default:
		return 0, fmt.Errorf("invalid identifierType")
	}
	return identifierValue, nil
}
