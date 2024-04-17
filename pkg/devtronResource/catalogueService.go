package devtronResource

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/tidwall/gjson"
)

func (impl *DevtronResourceServiceImpl) updateCatalogDataInResourceObj(resourceSchema *repository.DevtronResourceSchema,
	existingResourceObject *repository.DevtronResourceObject, resourceObject *bean.DevtronResourceObjectGetAPIBean) (err error) {
	referencedPaths, schemaWithUpdatedRefData, err := getReferencedPathsAndUpdatedSchema(resourceSchema.Schema)
	if err != nil {
		impl.logger.Errorw("err, getReferencedPathsAndUpdatedSchema", "err", err, "schema", resourceSchema.Schema)
		return err
	}
	schema, err := impl.getUpdatedSchemaWithAllRefObjectValues(schemaWithUpdatedRefData, referencedPaths)
	if err != nil {
		impl.logger.Errorw("error, getUpdatedSchemaWithAllRefObjectValues", "err", err,
			"schemaWithUpdatedRefData", schemaWithUpdatedRefData, "referencedPaths", referencedPaths)
		return err
	}
	resourceObject.Schema = schema
	//checking if resource object exists
	if existingResourceObject != nil && existingResourceObject.Id > 0 {
		//getting metadata out of this object
		metadataObject := gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectMetadataPath)
		resourceObject.CatalogData = metadataObject.Raw
	}
	return nil
}
