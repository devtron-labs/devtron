/*
 * Copyright (c) 2024. Devtron Inc.
 */

package adapter

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
)

func GetResourceObjectAudit(devtronResourceObject *repository.DevtronResourceObject, auditAction repository.AuditOperationType, auditPath []string) *repository.DevtronResourceObjectAudit {
	return &repository.DevtronResourceObjectAudit{
		DevtronResourceObjectId: devtronResourceObject.Id,
		ObjectData:              devtronResourceObject.ObjectData,
		AuditOperation:          auditAction,
		AuditLog: sql.AuditLog{
			CreatedOn: devtronResourceObject.CreatedOn,
			CreatedBy: devtronResourceObject.CreatedBy,
			UpdatedBy: devtronResourceObject.UpdatedBy,
			UpdatedOn: devtronResourceObject.UpdatedOn,
		},
		AuditOperationPath: auditPath,
	}
}
