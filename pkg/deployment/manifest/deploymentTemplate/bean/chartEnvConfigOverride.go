package bean

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"time"
)

type EnvConfigOverride struct {
	Id                        int
	ChartId                   int
	TargetEnvironment         int //target environment
	EnvOverrideValues         string
	EnvOverridePatchValues    string
	Status                    models.ChartStatus //new, deployment-in-progress, error, rollbacked, su
	ManualReviewed            bool
	Active                    bool
	Namespace                 string
	Chart                     *chartRepoRepository.Chart
	Environment               *repository2.Environment
	Latest                    bool
	Previous                  bool
	IsOverride                bool
	IsBasicViewLocked         bool
	CurrentViewEditor         models.ChartsViewEditorType
	CreatedOn                 time.Time
	CreatedBy                 int32
	UpdatedOn                 time.Time
	UpdatedBy                 int32
	ResolvedEnvOverrideValues string
	MergeStrategy             models.MergeStrategy
	VariableSnapshot          map[string]string
	//ResolvedEnvOverrideValuesForCM string
	VariableSnapshotForCM map[string]string
	//ResolvedEnvOverrideValuesForCS string
	VariableSnapshotForCS map[string]string
}
