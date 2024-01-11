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

func GetIdentifierValue(identifierType models.IdentifierType, appNameToIdMap map[string]int, identifierName string, envNameToIdMap map[string]int, clusterNameToIdMap map[string]int) (int, error) {
	var found bool
	var identifierValue int
	if identifierType == models.ApplicationName {
		identifierValue, found = appNameToIdMap[identifierName]
		if !found {
			return 0, fmt.Errorf("ApplicationName mapping not found %s", identifierName)
		}
	} else if identifierType == models.EnvName {
		identifierValue, found = envNameToIdMap[identifierName]
		if !found {
			return 0, fmt.Errorf("EnvName mapping not found %s", identifierName)
		}
	} else if identifierType == models.ClusterName {
		identifierValue, found = clusterNameToIdMap[identifierName]
		if !found {
			return 0, fmt.Errorf("ClusterName mapping not found %s", identifierName)
		}
	} else {
		return 0, fmt.Errorf("invalid identifierType")
	}
	return identifierValue, nil
}
