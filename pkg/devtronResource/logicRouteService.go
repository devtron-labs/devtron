package devtronResource

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
)

func getApiResourceKindUIComponentFunc(kind, component string) func(*DevtronResourceServiceImpl, *repository.DevtronResourceSchema,
	*repository.DevtronResourceObject, *bean.DevtronResourceObjectBean) error {
	if f, ok := getApiResourceKindUIComponentFuncMap[getKeyForKindAndUIComponent(kind, component)]; ok {
		return f
	} else {
		return nil
	}
}

var getApiResourceKindUIComponentFuncMap = map[string]func(*DevtronResourceServiceImpl, *repository.DevtronResourceSchema,
	*repository.DevtronResourceObject, *bean.DevtronResourceObjectBean) error{
	getKeyForKindAndUIComponent(bean.DevtronResourceApplication, bean.UIComponentCatalog): (*DevtronResourceServiceImpl).updateCatalogDataInResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceCluster, bean.UIComponentCatalog):     (*DevtronResourceServiceImpl).updateCatalogDataInResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceJob, bean.UIComponentCatalog):         (*DevtronResourceServiceImpl).updateCatalogDataInResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceRelease, bean.UIComponentCatalog):     (*DevtronResourceServiceImpl).updateCatalogDataInResourceObj,
}

func getKeyForKindAndUIComponent[K, C any](kind K, component C) string {
	return fmt.Sprintf("%s-%s", kind, component)
}
