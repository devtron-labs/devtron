package adapter

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
)

func EnvOverrideDBToDTO(dbObj *chartConfig.EnvConfigOverride) *bean.EnvConfigOverride {

	envOverride := &bean.EnvConfigOverride{
		Id:                        dbObj.Id,
		ChartId:                   dbObj.ChartId,
		TargetEnvironment:         dbObj.TargetEnvironment,
		EnvOverrideValues:         dbObj.EnvOverrideValues,
		Status:                    dbObj.Status,
		ManualReviewed:            dbObj.ManualReviewed,
		Active:                    dbObj.Active,
		Namespace:                 dbObj.Namespace,
		Environment:               dbObj.Environment,
		Latest:                    dbObj.Latest,
		Previous:                  dbObj.Previous,
		IsOverride:                dbObj.IsOverride,
		IsBasicViewLocked:         dbObj.IsBasicViewLocked,
		CurrentViewEditor:         dbObj.CurrentViewEditor,
		CreatedOn:                 dbObj.CreatedOn,
		CreatedBy:                 dbObj.CreatedBy,
		UpdatedOn:                 dbObj.UpdatedOn,
		UpdatedBy:                 dbObj.UpdatedBy,
		ResolvedEnvOverrideValues: dbObj.ResolvedEnvOverrideValues,
		VariableSnapshot:          dbObj.VariableSnapshot,
		VariableSnapshotForCM:     dbObj.VariableSnapshotForCM,
		VariableSnapshotForCS:     dbObj.VariableSnapshotForCS,
		Chart:                     dbObj.Chart,
		MergeStrategy:             dbObj.MergeStrategy,
	}
	return envOverride
}

func EnvOverrideDTOToDB(DTO *bean.EnvConfigOverride) *chartConfig.EnvConfigOverride {

	envOverride := &chartConfig.EnvConfigOverride{
		Id:                DTO.Id,
		ChartId:           DTO.ChartId,
		TargetEnvironment: DTO.TargetEnvironment,
		EnvOverrideValues: DTO.EnvOverrideValues,
		Status:            DTO.Status,
		ManualReviewed:    DTO.ManualReviewed,
		Active:            DTO.Active,
		Namespace:         DTO.Namespace,
		Environment:       DTO.Environment,
		Latest:            DTO.Latest,
		Previous:          DTO.Previous,
		IsOverride:        DTO.IsOverride,
		IsBasicViewLocked: DTO.IsBasicViewLocked,
		CurrentViewEditor: DTO.CurrentViewEditor,
		AuditLog: sql.AuditLog{
			CreatedOn: DTO.CreatedOn,
			CreatedBy: DTO.CreatedBy,
			UpdatedOn: DTO.UpdatedOn,
			UpdatedBy: DTO.UpdatedBy,
		},
		ResolvedEnvOverrideValues: DTO.ResolvedEnvOverrideValues,
		VariableSnapshot:          DTO.VariableSnapshot,
		VariableSnapshotForCM:     DTO.VariableSnapshotForCM,
		VariableSnapshotForCS:     DTO.VariableSnapshotForCS,
		MergeStrategy:             DTO.MergeStrategy,
	}
	envOverride.Chart = DTO.Chart
	return envOverride
}
