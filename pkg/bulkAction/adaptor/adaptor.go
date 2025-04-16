package adaptor

import "github.com/devtron-labs/devtron/pkg/bulkAction/bean"

func GetCmAndSecretBulkUpdateResponseForOneApp(appId int, appName string, envId int, names []string, message string) *bean.CmAndSecretBulkUpdateResponseForOneApp {
	return &bean.CmAndSecretBulkUpdateResponseForOneApp{
		AppId:   appId,
		AppName: appName,
		EnvId:   envId,
		Names:   names,
		Message: message,
	}
}
