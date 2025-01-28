package adapter

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/operationAudit/bean"
	"github.com/devtron-labs/devtron/pkg/operationAudit/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
)

func BuildOperationAuditModel(entityId int32, entityType bean2.EntityType, operationType bean2.OperationType,
	permissionsAuditDto *bean.PermissionsAuditDto, userIdForAuditLog int32, schemaFor bean2.SchemaFor) (*repository.OperationAudit, error) {
	permissionsJson, err := json.Marshal(permissionsAuditDto)
	if err != nil {
		errToReturn := fmt.Sprintf("error in marshalling permissions audit dto :%s", err.Error())
		return nil, errors.New(errToReturn)
	}
	return &repository.OperationAudit{
		EntityId:        entityId,
		EntityType:      entityType,
		OperationType:   operationType,
		PermissionsJson: string(permissionsJson),
		SchemaFor:       schemaFor,
		AuditLog:        sql.NewDefaultAuditLog(userIdForAuditLog),
	}, nil

}
