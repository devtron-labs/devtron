package bean

import (
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/bean/AppView"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appStore/chartGroup"
	bean4 "github.com/devtron-labs/devtron/pkg/appWorkflow/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/team/bean"
)

type UserResourceResponseDto struct {
	Data interface{} `json:"data"`
}
type ResourceOptionsDto struct {
	TeamsResp            []bean2.TeamRequest
	HelmEnvResp          []*bean.ClusterEnvDto
	ClusterResp          []bean3.ClusterBean
	NameSpaces           []string
	ApiResourcesResp     *k8s.GetAllApiResourcesResponse
	ClusterResourcesResp *k8s.ClusterResourceListMap
	TeamAppResp          []*app.TeamAppBean
	EnvResp              []bean.EnvironmentBean
	ChartGroupResp       *chartGroup.ChartGroupList
	JobsResp             []*AppView.JobContainer
	AppWfsResp           *bean4.WorkflowNamesResponse
}

func NewResourceOptionsDto() *ResourceOptionsDto {
	return &ResourceOptionsDto{}
}
func (r *ResourceOptionsDto) WithTeamsResp(teamsResp []bean2.TeamRequest) *ResourceOptionsDto {
	r.TeamsResp = teamsResp
	return r
}
func (r *ResourceOptionsDto) WithHelmEnvResp(helmEnvResp []*bean.ClusterEnvDto) *ResourceOptionsDto {
	r.HelmEnvResp = helmEnvResp
	return r
}
func (r *ResourceOptionsDto) WithClusterResp(clusterResp []bean3.ClusterBean) *ResourceOptionsDto {
	r.ClusterResp = clusterResp
	return r
}
func (r *ResourceOptionsDto) WithNameSpaces(nameSpaces []string) *ResourceOptionsDto {
	r.NameSpaces = nameSpaces
	return r
}
func (r *ResourceOptionsDto) WithApiResourcesResp(apiResourcesResp *k8s.GetAllApiResourcesResponse) *ResourceOptionsDto {
	r.ApiResourcesResp = apiResourcesResp
	return r

}
func (r *ResourceOptionsDto) WithClusterResourcesResp(clusterResourcesResp *k8s.ClusterResourceListMap) *ResourceOptionsDto {
	r.ClusterResourcesResp = clusterResourcesResp
	return r
}
func (r *ResourceOptionsDto) WithTeamAppResp(teamAppResp []*app.TeamAppBean) *ResourceOptionsDto {
	r.TeamAppResp = teamAppResp
	return r
}

func (r *ResourceOptionsDto) WithEnvResp(envResp []bean.EnvironmentBean) *ResourceOptionsDto {
	r.EnvResp = envResp
	return r
}
func (r *ResourceOptionsDto) WithChartGroupResp(chartGroupResp *chartGroup.ChartGroupList) *ResourceOptionsDto {
	r.ChartGroupResp = chartGroupResp
	return r
}
func (r *ResourceOptionsDto) WithJobsResp(jobsResp []*AppView.JobContainer) *ResourceOptionsDto {
	r.JobsResp = jobsResp
	return r
}
func (r *ResourceOptionsDto) WithAppWfsResp(appWfsResp *bean4.WorkflowNamesResponse) *ResourceOptionsDto {
	r.AppWfsResp = appWfsResp
	return r
}

type Version string
type UserResourceKind string

const (
	KindTeam               UserResourceKind = "team"
	KindEnvironment        UserResourceKind = "environment"
	Application            UserResourceKind = "application"
	KindDevtronApplication UserResourceKind = Application + "/devtron-application"
	KindHelmApplication    UserResourceKind = Application + "/helm-application"
	KindHelmEnvironment    UserResourceKind = "environment/helm"
	KindCluster            UserResourceKind = "cluster"
	KindChartGroup         UserResourceKind = "chartGroup"
	KindJobs               UserResourceKind = "jobs"
	KindWorkflow           UserResourceKind = "workflow"
	ClusterNamespaces      UserResourceKind = "cluster/namespaces"
	ClusterApiResources    UserResourceKind = "cluster/apiResources"
	ClusterResources       UserResourceKind = "cluster/resources"
)

const (
	VersionV1 Version = "v1"
)
