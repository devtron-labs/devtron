package utils

import (
	"github.com/devtron-labs/devtron/pkg/variables/models"
)

func ManifestToPayload(manifest models.ScopedVariableManifest, userId int32) models.Payload {

	variableList := make([]*models.Variables, 0)

	for _, spec := range manifest.Spec {
		attributes := make([]models.AttributeValue, 0)
		for _, value := range spec.Values {
			attribute := models.AttributeValue{
				VariableValue: models.VariableValue{Value: value.Value},
				AttributeType: value.Category,
			}

			if value.Selectors != nil && value.Selectors.AttributeSelectors != nil {
				attribute.AttributeParams = value.Selectors.AttributeSelectors
			}
			attributes = append(attributes, attribute)
		}
		variable := models.Variables{
			Definition: models.Definition{
				VarName:          spec.Name,
				DataType:         models.PRIMITIVE_TYPE,
				VarType:          models.PUBLIC,
				Description:      spec.Documentation,
				ShortDescription: spec.Description,
			},
			AttributeValues: attributes,
		}
		if spec.IsSensitive {
			variable.Definition.VarType = models.PRIVATE
		}
		variableList = append(variableList, &variable)
	}
	payload := models.Payload{
		Variables: variableList,
		UserId:    userId,
	}
	return payload
}

func PayloadToManifest(payload models.Payload) models.ScopedVariableManifest {
	manifest := models.ScopedVariableManifest{
		ApiVersion: "devtron.ai/v1beta1",
		Kind:       "Variable",
		Spec:       make([]models.VariableSpec, 0),
	}
	for _, variable := range payload.Variables {
		spec := models.VariableSpec{
			Name:          variable.Definition.VarName,
			Documentation: variable.Definition.Description,
			Description:   variable.Definition.ShortDescription,
			Values:        make([]models.VariableValueSpec, 0),
			IsSensitive:   variable.Definition.VarType.IsTypeSensitive(),
		}
		for _, attribute := range variable.AttributeValues {
			valueSpec := models.VariableValueSpec{
				Value:    attribute.VariableValue.Value,
				Category: attribute.AttributeType,
			}
			if attribute.AttributeParams != nil {
				valueSpec.Selectors = &models.Selector{AttributeSelectors: attribute.AttributeParams}
			}
			spec.Values = append(spec.Values, valueSpec)
		}
		manifest.Spec = append(manifest.Spec, spec)
	}
	return manifest
}
