package globalPolicy

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	mocks6 "github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow/mocks"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	mocks4 "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	mocks3 "github.com/devtron-labs/devtron/pkg/devtronResource/mocks"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/globalPolicy/history/bean"
	mocks2 "github.com/devtron-labs/devtron/pkg/globalPolicy/history/mocks"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/repository"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/repository/mocks"
	repository3 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	mocks5 "github.com/devtron-labs/devtron/pkg/pipeline/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestCheckIfConsequenceIsBlocking(t *testing.T) {
	//t.SkipNow()
	testCases := []struct {
		consequence *bean.ConsequenceDto
		want        bool
	}{
		{
			consequence: &bean.ConsequenceDto{
				Action:        bean.CONSEQUENCE_ACTION_BLOCK,
				MetadataField: time.Now(),
			},
			want: true},

		{
			consequence: &bean.ConsequenceDto{
				Action:        bean.CONSEQUENCE_ACTION_ALLOW_FOREVER,
				MetadataField: time.Now(),
			},
			want: false},
		{
			consequence: &bean.ConsequenceDto{
				Action:        bean.CONSEQUENCE_ACTION_ALLOW_UNTIL_TIME,
				MetadataField: time.Now().Add(time.Hour),
			},
			want: false},
		{
			consequence: &bean.ConsequenceDto{
				Action:        bean.CONSEQUENCE_ACTION_ALLOW_UNTIL_TIME,
				MetadataField: time.Now().Add(-time.Hour),
			},
			want: true},
		{
			consequence: &bean.ConsequenceDto{
				Action:        "INVALID_ACTION",
				MetadataField: time.Now(),
			},
			want: false,
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("checkIfConsequenceIsBlocking-test-%d", i+1), func(t *testing.T) {
			isBlocking := checkIfConsequenceIsBlocking(testCase.consequence)
			assert.Equal(t, testCase.want, isBlocking)
		})
	}
}

func TestGetMandatoryPluginDefinition(t *testing.T) {
	//t.SkipNow()
	testCases := []struct {
		pluginIdStage     string
		definitionSources []*bean.DefinitionSourceDto
		wantResult        *bean.MandatoryPluginDefinitionDto
		wantErr           bool
	}{
		{
			pluginIdStage: "1234/PRE_CI",
			definitionSources: []*bean.DefinitionSourceDto{
				{
					ProjectName:                  "Project1",
					AppName:                      "App1",
					ClusterName:                  "Cluster1",
					EnvironmentName:              "Environment1",
					BranchNames:                  []string{"Branch1", "Branch2"},
					IsDueToProductionEnvironment: true,
					IsDueToLinkedPipeline:        true,
					CiPipelineName:               "Pipeline1",
					PolicyName:                   "Policy1",
				},
				{
					ProjectName:                  "Project2",
					AppName:                      "App2",
					IsDueToProductionEnvironment: false,
					IsDueToLinkedPipeline:        true,
					CiPipelineName:               "Pipeline2",
					PolicyName:                   "Policy2",
				},
			},
			wantResult: &bean.MandatoryPluginDefinitionDto{
				DefinitionDto: &bean.DefinitionDto{
					AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
					Data: bean.DefinitionDataDto{
						PluginId:     1234,
						ApplyToStage: bean.PluginApplyStage("PRE_CI"),
					},
				},
			},
			wantErr: false,
		},
		{
			pluginIdStage: "invalidId/POST_CI",
			definitionSources: []*bean.DefinitionSourceDto{
				{
					ClusterName:                  "Cluster1",
					EnvironmentName:              "Environment1",
					BranchNames:                  []string{"Branch1", "Branch2"},
					IsDueToProductionEnvironment: true,
					CiPipelineName:               "Pipeline1",
					PolicyName:                   "Policy1",
				},
				{
					ProjectName:                  "Project2",
					AppName:                      "App2",
					IsDueToProductionEnvironment: false,
					IsDueToLinkedPipeline:        true,
					CiPipelineName:               "Pipeline2",
					PolicyName:                   "Policy2",
				},
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			pluginIdStage: "28",
			definitionSources: []*bean.DefinitionSourceDto{
				{
					ClusterName:                  "Cluster1",
					EnvironmentName:              "Environment1",
					BranchNames:                  []string{"Branch1", "Branch2"},
					IsDueToProductionEnvironment: true,
					CiPipelineName:               "Pipeline1",
					PolicyName:                   "Policy1",
				},
				{
					ProjectName:                  "Project2",
					AppName:                      "App2",
					IsDueToProductionEnvironment: false,
					IsDueToLinkedPipeline:        true,
					CiPipelineName:               "Pipeline2",
					PolicyName:                   "Policy2",
				},
			},
			wantResult: nil,
			wantErr:    true,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getMandatoryPluginDefinition-test-%d", i+1), func(t *testing.T) {
			if testCase.wantResult != nil {
				testCase.wantResult.DefinitionSources = testCase.definitionSources
			}
			result, err := getMandatoryPluginDefinition(testCase.pluginIdStage, testCase.definitionSources)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.wantResult, result)
		})
	}
}

func TestIsBranchValueMatched(t *testing.T) {
	//t.SkipNow()
	testCases := []struct {
		branchDto   *bean.BranchDto
		branchValue string
		wantResult  bool
		wantErr     bool
	}{
		{
			branchDto: &bean.BranchDto{
				BranchValueType: bean2.VALUE_TYPE_REGEX,
				Value:           "^feature/.*$",
			},
			branchValue: "feature/branch1",
			wantResult:  true,
			wantErr:     false,
		},
		{
			branchDto: &bean.BranchDto{
				BranchValueType: bean2.VALUE_TYPE_REGEX,
				Value:           "^feature/.*$",
			},
			branchValue: "bugfix/branch1",
			wantResult:  false,
			wantErr:     false,
		},
		{
			branchDto: &bean.BranchDto{
				BranchValueType: bean2.VALUE_TYPE_FIXED,
				Value:           "main",
			},
			branchValue: "main",
			wantResult:  true,
			wantErr:     false,
		},
		{
			branchDto: &bean.BranchDto{
				BranchValueType: bean2.VALUE_TYPE_FIXED,
				Value:           "develop",
			},
			branchValue: "feature/branch1",
			wantResult:  false,
			wantErr:     false,
		},
		{
			branchDto: &bean.BranchDto{
				BranchValueType: bean2.VALUE_TYPE_REGEX,
				Value:           "^feature/.*$",
			},
			branchValue: "",
			wantResult:  false,
			wantErr:     true,
		},

		{
			branchDto: &bean.BranchDto{
				BranchValueType: bean2.VALUE_TYPE_REGEX,
				Value:           "[a-z+",
			},
			branchValue: "feature/branch1",
			wantResult:  false,
			wantErr:     true,
		},
		{
			branchDto: &bean.BranchDto{
				BranchValueType: bean2.VALUE_TYPE_REGEX,
				Value:           "",
			},
			branchValue: "",
			wantResult:  false,
			wantErr:     true,
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("isBranchValueMatched-test-%d", i+1), func(t *testing.T) {
			result, err := isBranchValueMatched(testCase.branchDto, testCase.branchValue)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.wantResult, result)
		})
	}
}

func TestGetAppProjectDetailsFromCiPipelineProjectAppObjs(t *testing.T) {

	testCases := []struct {
		ciPipelineAppProjectObjs          []*pipelineConfig.CiPipelineAppProject
		wantAllCiPipelineIds              []int
		wantCiPipelineIdNameMap           map[int]string
		wantAllProjectAppNames            []string
		wantProjectMap                    map[string]bool
		wantCiPipelineIdProjectAppNameMap map[int]*bean.PluginSourceCiPipelineAppDetailDto
	}{
		{
			ciPipelineAppProjectObjs: []*pipelineConfig.CiPipelineAppProject{
				{
					CiPipelineId:   1,
					CiPipelineName: "Pipeline1",
					AppName:        "App1",
					ProjectName:    "Project1",
				},
				{
					CiPipelineId:   2,
					CiPipelineName: "Pipeline2",
					AppName:        "App2",
					ProjectName:    "Project2",
				},
			},
			wantAllCiPipelineIds: []int{1, 2},
			wantCiPipelineIdNameMap: map[int]string{
				1: "Pipeline1",
				2: "Pipeline2",
			},
			wantAllProjectAppNames: []string{"Project1/App1", "Project2/App2"},
			wantProjectMap: map[string]bool{
				"Project1": true,
				"Project2": true,
			},
			wantCiPipelineIdProjectAppNameMap: map[int]*bean.PluginSourceCiPipelineAppDetailDto{
				1: {ProjectName: "Project1", AppName: "App1"},
				2: {ProjectName: "Project2", AppName: "App2"},
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getAppProjectDetailsFromCiPipelineProjectAppObjs-test-%d", i+1), func(t *testing.T) {
			allCiPipelineIds, ciPipelineIdNameMap, allProjectAppNames, projectMap, ciPipelineIdProjectAppNameMap :=
				getAppProjectDetailsFromCiPipelineProjectAppObjs(testCase.ciPipelineAppProjectObjs)

			assert.Len(t, allCiPipelineIds, len(testCase.wantAllCiPipelineIds))
			for _, allCiPipelineId := range allCiPipelineIds {
				assert.Contains(t, testCase.wantAllCiPipelineIds, allCiPipelineId)
			}

			assert.Len(t, ciPipelineIdNameMap, len(testCase.wantCiPipelineIdNameMap))
			for key, value := range ciPipelineIdNameMap {
				assert.Equal(t, value, testCase.wantCiPipelineIdNameMap[key])
			}

			assert.Len(t, allProjectAppNames, len(testCase.wantAllProjectAppNames))
			for _, allProjectAppName := range allProjectAppNames {
				assert.Contains(t, testCase.wantAllProjectAppNames, allProjectAppName)
			}

			assert.Len(t, projectMap, len(testCase.wantProjectMap))
			for key, value := range projectMap {
				assert.Equal(t, value, testCase.wantProjectMap[key])
			}

			assert.Len(t, ciPipelineIdProjectAppNameMap, len(testCase.wantCiPipelineIdProjectAppNameMap))
			for key, value := range ciPipelineIdProjectAppNameMap {
				assert.Equal(t, value, testCase.wantCiPipelineIdProjectAppNameMap[key])
			}
		})
	}
}

func TestGetEnvClusterDetailsFromCIPipelineEnvClusterObjs(t *testing.T) {
	testCases := []struct {
		ciPipelineEnvClusterObjs               []*pipelineConfig.CiPipelineEnvCluster
		wantHaveAnyProductionEnv               bool
		wantCiPipelineIdProductionEnvDetailMap map[int][]*bean.PluginSourceCiPipelineEnvDetailDto
		wantCiPipelineIdEnvDetailMap           map[int][]*bean.PluginSourceCiPipelineEnvDetailDto
		wantAllClusterEnvNames                 []string
		wantClusterMap                         map[string]bool
	}{
		{
			ciPipelineEnvClusterObjs: []*pipelineConfig.CiPipelineEnvCluster{
				{
					CiPipelineId:    1,
					ClusterName:     "cluster1",
					EnvName:         "env1",
					IsProductionEnv: true,
				},
				{
					CiPipelineId:    1,
					ClusterName:     "cluster1",
					EnvName:         "env2",
					IsProductionEnv: false,
				},
				{
					CiPipelineId:    2,
					ClusterName:     "cluster2",
					EnvName:         "env3",
					IsProductionEnv: true,
				},
				{
					CiPipelineId:    2,
					ClusterName:     "cluster2",
					EnvName:         "env4",
					IsProductionEnv: false,
				},
			},
			wantHaveAnyProductionEnv: true,
			wantCiPipelineIdProductionEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{
				1: {
					{
						ClusterName: "cluster1",
						EnvName:     "env1",
					},
				},
				2: {
					{
						ClusterName: "cluster2",
						EnvName:     "env3",
					},
				},
			},
			wantCiPipelineIdEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{
				1: {
					{
						ClusterName: "cluster1",
						EnvName:     "env1",
					},
					{
						ClusterName: "cluster1",
						EnvName:     "env2",
					},
				},
				2: {
					{
						ClusterName: "cluster2",
						EnvName:     "env3",
					},
					{
						ClusterName: "cluster2",
						EnvName:     "env4",
					},
				},
			},
			wantAllClusterEnvNames: []string{"cluster1/env1", "cluster1/env2", "cluster2/env3", "cluster2/env4"},
			wantClusterMap: map[string]bool{
				"cluster1": true,
				"cluster2": true,
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getEnvClusterDetailsFromCIPipelineEnvClusterObjs-test-%d", i+1), func(t *testing.T) {
			haveAnyProductionEnv, ciPipelineIdProductionEnvDetailMap, ciPipelineIdEnvDetailMap, allClusterEnvNames, clusterMap :=
				getEnvClusterDetailsFromCIPipelineEnvClusterObjs(testCase.ciPipelineEnvClusterObjs)

			assert.Equal(t, testCase.wantHaveAnyProductionEnv, haveAnyProductionEnv)

			assert.Len(t, ciPipelineIdProductionEnvDetailMap, len(testCase.wantCiPipelineIdProductionEnvDetailMap))
			for key, value := range ciPipelineIdProductionEnvDetailMap {
				assert.Equal(t, value, testCase.wantCiPipelineIdProductionEnvDetailMap[key])
			}

			assert.Len(t, ciPipelineIdEnvDetailMap, len(testCase.wantCiPipelineIdEnvDetailMap))
			for key, value := range ciPipelineIdEnvDetailMap {
				assert.Equal(t, value, testCase.wantCiPipelineIdEnvDetailMap[key])
			}

			assert.Len(t, allClusterEnvNames, len(testCase.wantAllClusterEnvNames))
			for _, allProjectEnvName := range allClusterEnvNames {
				assert.Contains(t, testCase.wantAllClusterEnvNames, allProjectEnvName)
			}

			assert.Len(t, clusterMap, len(testCase.wantClusterMap))
			for key, value := range clusterMap {
				assert.Equal(t, value, testCase.wantClusterMap[key])
			}
		})
	}
}

func TestGetMatchedBranchList(t *testing.T) {
	globalPolicyService, _, _, _, _, _, _, _, _, _, _, err := getGlobalPolicyService(t)
	assert.NoError(t, err)

	testCases := []struct {
		branchList            []*bean.BranchDto
		branchValues          []string
		wantMatchedBranchList []string
		wantErr               bool
	}{
		{
			branchList: []*bean.BranchDto{
				{
					BranchValueType: bean2.VALUE_TYPE_REGEX,
					Value:           "feature/.*",
				},
				{
					BranchValueType: bean2.VALUE_TYPE_FIXED,
					Value:           "main",
				},
				{
					BranchValueType: bean2.VALUE_TYPE_FIXED,
					Value:           "develop",
				},
			},
			branchValues:          []string{"feature/branch1", "develop", "feature/branch2", "release/v1.0"},
			wantMatchedBranchList: []string{"feature/branch1", "develop", "feature/branch2"},
			wantErr:               false,
		},
		{
			branchList: []*bean.BranchDto{
				{
					BranchValueType: bean2.VALUE_TYPE_REGEX,
					Value:           "",
				},
			},
			branchValues:          []string{"feature/branch1", "develop", "feature/branch2", "release/v1.0"},
			wantMatchedBranchList: []string{},
			wantErr:               true,
		},
		{
			branchList:            []*bean.BranchDto{},
			branchValues:          []string{"feature/branch1", "develop"},
			wantMatchedBranchList: []string{},
			wantErr:               false,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getMatchedBranchList-test-%d", i+1), func(t *testing.T) {
			matchedBranches, err := globalPolicyService.getMatchedBranchList(testCase.branchList, testCase.branchValues)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Len(t, matchedBranches, len(testCase.wantMatchedBranchList))
			for _, matchedBranch := range matchedBranches {
				assert.Contains(t, testCase.wantMatchedBranchList, matchedBranch)
			}
		})
	}
}

func TestGetDefinitionSourceDtos(t *testing.T) {
	globalPolicyService, _, _, _, _, _, _, _, _, _, _, err := getGlobalPolicyService(t)
	assert.NoError(t, err)

	testCases := []struct {
		globalPolicyDetailDto              bean.GlobalPolicyDetailDto
		allCiPipelineIds                   []int
		ciPipelineId                       int
		ciPipelineIdProjectAppNameMap      map[int]*bean.PluginSourceCiPipelineAppDetailDto
		ciPipelineIdEnvDetailMap           map[int][]*bean.PluginSourceCiPipelineEnvDetailDto
		ciPipelineIdProductionEnvDetailMap map[int][]*bean.PluginSourceCiPipelineEnvDetailDto
		branchValues                       []string
		globalPolicyName                   string
		ciPipelineIdNameMap                map[int]string
		wantResult                         []*bean.DefinitionSourceDto
		wantError                          bool
	}{
		{
			globalPolicyDetailDto: bean.GlobalPolicyDetailDto{
				Selectors: &bean.SelectorDto{
					ApplicationSelector: []*bean.ProjectAppDto{
						{
							ProjectName: "Project1",
						},
					},
				},
			},
			allCiPipelineIds: []int{1},
			ciPipelineId:     1,
			ciPipelineIdProjectAppNameMap: map[int]*bean.PluginSourceCiPipelineAppDetailDto{
				1: {
					ProjectName: "Project1",
					AppName:     "App1",
				},
			},
			ciPipelineIdEnvDetailMap:           make(map[int][]*bean.PluginSourceCiPipelineEnvDetailDto),
			ciPipelineIdProductionEnvDetailMap: make(map[int][]*bean.PluginSourceCiPipelineEnvDetailDto),
			branchValues:                       []string{"main"},
			globalPolicyName:                   "GlobalPolicy1",
			ciPipelineIdNameMap: map[int]string{
				1: "CIPipeline1",
			},
			wantResult: []*bean.DefinitionSourceDto{
				{
					ProjectName:           "Project1",
					AppName:               "",
					BranchNames:           []string{},
					IsDueToLinkedPipeline: false,
					CiPipelineName:        "CIPipeline1",
					PolicyName:            "GlobalPolicy1",
				},
			},
			wantError: false,
		},
		{
			globalPolicyDetailDto: bean.GlobalPolicyDetailDto{
				Selectors: &bean.SelectorDto{
					EnvironmentSelector: &bean.EnvironmentSelectorDto{
						AllProductionEnvironments: true,
					},
				},
			},
			allCiPipelineIds: []int{1},
			ciPipelineId:     1,
			ciPipelineIdProjectAppNameMap: map[int]*bean.PluginSourceCiPipelineAppDetailDto{
				1: {
					ProjectName: "Project1",
					AppName:     "App1",
				},
			},
			ciPipelineIdEnvDetailMap:           make(map[int][]*bean.PluginSourceCiPipelineEnvDetailDto),
			ciPipelineIdProductionEnvDetailMap: make(map[int][]*bean.PluginSourceCiPipelineEnvDetailDto),
			branchValues:                       []string{"main"},
			globalPolicyName:                   "GlobalPolicy1",
			ciPipelineIdNameMap: map[int]string{
				1: "CIPipeline1",
			},
			wantResult: []*bean.DefinitionSourceDto(nil),
			wantError:  false,
		},
		{
			globalPolicyDetailDto: bean.GlobalPolicyDetailDto{
				Selectors: &bean.SelectorDto{
					EnvironmentSelector: &bean.EnvironmentSelectorDto{
						ClusterEnvList: []*bean.ClusterEnvDto{
							{
								ClusterName: "DefaultCluster",
								EnvNames:    []string{"devtron-demo", "devtron-test"},
							},
						},
					},
				},
			},
			allCiPipelineIds: []int{1},
			ciPipelineId:     1,
			ciPipelineIdProjectAppNameMap: map[int]*bean.PluginSourceCiPipelineAppDetailDto{
				1: {
					ProjectName: "Project1",
					AppName:     "App1",
				},
			},
			ciPipelineIdEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{
				1: {
					{
						ClusterName: "DefaultCluster",
						EnvName:     "devtron-demo",
					},
				},
			},
			ciPipelineIdProductionEnvDetailMap: make(map[int][]*bean.PluginSourceCiPipelineEnvDetailDto),
			branchValues:                       []string{"main"},
			globalPolicyName:                   "GlobalPolicy1",
			ciPipelineIdNameMap: map[int]string{
				1: "CIPipeline1",
			},
			wantResult: []*bean.DefinitionSourceDto{
				{
					ClusterName:           "DefaultCluster",
					EnvironmentName:       "devtron-demo",
					BranchNames:           []string{},
					IsDueToLinkedPipeline: false,
					CiPipelineName:        "CIPipeline1",
					PolicyName:            "GlobalPolicy1",
				},
			},
			wantError: false,
		},
		{
			globalPolicyDetailDto: bean.GlobalPolicyDetailDto{
				Selectors: &bean.SelectorDto{
					BranchList: []*bean.BranchDto{
						{
							Value:           "main",
							BranchValueType: bean2.VALUE_TYPE_FIXED,
						},
					},
				},
			},
			allCiPipelineIds: []int{1},
			ciPipelineId:     1,
			ciPipelineIdProjectAppNameMap: map[int]*bean.PluginSourceCiPipelineAppDetailDto{
				1: {
					ProjectName: "Project1",
					AppName:     "App1",
				},
			},
			ciPipelineIdEnvDetailMap:           make(map[int][]*bean.PluginSourceCiPipelineEnvDetailDto),
			ciPipelineIdProductionEnvDetailMap: make(map[int][]*bean.PluginSourceCiPipelineEnvDetailDto),
			branchValues:                       []string{"main"},
			globalPolicyName:                   "GlobalPolicy1",
			ciPipelineIdNameMap: map[int]string{
				1: "CIPipeline1",
			},
			wantResult: []*bean.DefinitionSourceDto{
				{
					BranchNames:           []string{"main"},
					IsDueToLinkedPipeline: false,
					CiPipelineName:        "CIPipeline1",
					PolicyName:            "GlobalPolicy1",
				},
			},
			wantError: false,
		},
		{
			globalPolicyDetailDto: bean.GlobalPolicyDetailDto{
				Selectors: &bean.SelectorDto{
					BranchList: []*bean.BranchDto{
						{
							Value:           "main",
							BranchValueType: bean2.VALUE_TYPE_FIXED,
						},
					},
				},
			},
			allCiPipelineIds: []int{1},
			ciPipelineId:     1,
			ciPipelineIdProjectAppNameMap: map[int]*bean.PluginSourceCiPipelineAppDetailDto{
				1: {
					ProjectName: "Project1",
					AppName:     "App1",
				},
			},
			ciPipelineIdEnvDetailMap:           make(map[int][]*bean.PluginSourceCiPipelineEnvDetailDto),
			ciPipelineIdProductionEnvDetailMap: make(map[int][]*bean.PluginSourceCiPipelineEnvDetailDto),
			branchValues:                       []string{""},
			globalPolicyName:                   "GlobalPolicy1",
			ciPipelineIdNameMap: map[int]string{
				1: "CIPipeline1",
			},
			wantResult: []*bean.DefinitionSourceDto(nil),
			wantError:  true,
		},
		{
			globalPolicyDetailDto: bean.GlobalPolicyDetailDto{
				Selectors: &bean.SelectorDto{
					BranchList: []*bean.BranchDto{
						{
							Value:           "feature",
							BranchValueType: bean2.VALUE_TYPE_FIXED,
						},
					},
				},
			},
			allCiPipelineIds: []int{1},
			ciPipelineId:     1,
			ciPipelineIdProjectAppNameMap: map[int]*bean.PluginSourceCiPipelineAppDetailDto{
				1: {
					ProjectName: "Project1",
					AppName:     "App1",
				},
			},
			ciPipelineIdEnvDetailMap:           make(map[int][]*bean.PluginSourceCiPipelineEnvDetailDto),
			ciPipelineIdProductionEnvDetailMap: make(map[int][]*bean.PluginSourceCiPipelineEnvDetailDto),
			branchValues:                       []string{"main"},
			globalPolicyName:                   "GlobalPolicy1",
			ciPipelineIdNameMap: map[int]string{
				1: "CIPipeline1",
			},
			wantResult: nil,
			wantError:  false,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getDefinitionSourceDtos-test-%d", i+1), func(t *testing.T) {
			definitionSourceDtos, err := globalPolicyService.getDefinitionSourceDtos(testCase.globalPolicyDetailDto, testCase.allCiPipelineIds, testCase.ciPipelineId,
				testCase.ciPipelineIdProjectAppNameMap, testCase.ciPipelineIdEnvDetailMap, testCase.ciPipelineIdProductionEnvDetailMap,
				testCase.branchValues, testCase.globalPolicyName, testCase.ciPipelineIdNameMap)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.wantResult, definitionSourceDtos)
		})
	}
}

func TestGetPluginIdApplyStageAndPluginBlockageMaps(t *testing.T) {
	testCases := []struct {
		definitions                []*bean.DefinitionDto
		consequence                *bean.ConsequenceDto
		mandatoryPluginBlockageMap map[string]*bean.ConsequenceDto
		wantPluginIdApplyStageMap  map[string]bean.Severity
	}{
		{
			definitions: []*bean.DefinitionDto{
				{
					AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
					Data: bean.DefinitionDataDto{
						PluginId:     1,
						ApplyToStage: bean.PluginApplyStage("stage1"),
					},
				},
				{
					AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
					Data: bean.DefinitionDataDto{
						PluginId:     2,
						ApplyToStage: bean.PluginApplyStage("stage2"),
					},
				},
			},
			consequence: &bean.ConsequenceDto{
				Action:        bean.CONSEQUENCE_ACTION_ALLOW_UNTIL_TIME,
				MetadataField: time.Now(),
			},
			mandatoryPluginBlockageMap: map[string]*bean.ConsequenceDto{
				"1/stage1": {
					Action:        bean.CONSEQUENCE_ACTION_BLOCK,
					MetadataField: time.Now(),
				},
				"2/stage2": {
					Action: bean.CONSEQUENCE_ACTION_ALLOW_FOREVER,
				},
			},
			wantPluginIdApplyStageMap: map[string]bean.Severity{
				"2/stage2": bean.SEVERITY_MORE_SEVERE,
			},
		},
		{
			definitions: []*bean.DefinitionDto{
				{
					AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
					Data: bean.DefinitionDataDto{
						PluginId:     1,
						ApplyToStage: bean.PluginApplyStage("stage1"),
					},
				},
				{
					AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
					Data: bean.DefinitionDataDto{
						PluginId:     2,
						ApplyToStage: bean.PluginApplyStage("stage2"),
					},
				},
			},
			consequence: &bean.ConsequenceDto{
				Action: bean.CONSEQUENCE_ACTION_BLOCK,
			},
			mandatoryPluginBlockageMap: map[string]*bean.ConsequenceDto{
				"1/stage1": {
					Action:        bean.CONSEQUENCE_ACTION_BLOCK,
					MetadataField: time.Now(),
				},
				"2/stage2": {
					Action: bean.CONSEQUENCE_ACTION_ALLOW_FOREVER,
				},
			},
			wantPluginIdApplyStageMap: map[string]bean.Severity{
				"1/stage1": bean.SEVERITY_SAME_SEVERE,
				"2/stage2": bean.SEVERITY_MORE_SEVERE,
			},
		},
		{
			definitions: []*bean.DefinitionDto{
				{
					AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
					Data: bean.DefinitionDataDto{
						PluginId:     1,
						ApplyToStage: bean.PluginApplyStage("stage1"),
					},
				},
				{
					AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
					Data: bean.DefinitionDataDto{
						PluginId:     2,
						ApplyToStage: bean.PluginApplyStage("stage2"),
					},
				},
			},
			consequence: &bean.ConsequenceDto{
				Action: bean.CONSEQUENCE_ACTION_BLOCK,
			},
			mandatoryPluginBlockageMap: map[string]*bean.ConsequenceDto{
				"1/stage1": {
					Action:        bean.CONSEQUENCE_ACTION_BLOCK,
					MetadataField: time.Now(),
				},
			},
			wantPluginIdApplyStageMap: map[string]bean.Severity{
				"1/stage1": bean.SEVERITY_SAME_SEVERE,
				"2/stage2": bean.SEVERITY_SAME_SEVERE,
			},
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getPluginIdApplyStageAndPluginBlockageMaps-test-%d", i+1), func(t *testing.T) {
			pluginIdApplyStageMap := getPluginIdApplyStageAndPluginBlockageMaps(testCase.definitions, testCase.consequence, testCase.mandatoryPluginBlockageMap)

			assert.Len(t, pluginIdApplyStageMap, len(testCase.wantPluginIdApplyStageMap))
			for key, value := range testCase.wantPluginIdApplyStageMap {
				assert.Equal(t, value, pluginIdApplyStageMap[key])
			}
		})
	}
}

func TestCheckAppSelectorAndGetDefinitionSourceInfo(t *testing.T) {

	testCases := []struct {
		appSelectorMap                  map[string]bool
		projectName                     string
		appName                         string
		wantDefinitionSourceAppName     string
		wantDefinitionSourceProjectName string
		wantToContinue                  bool
	}{
		{
			appSelectorMap: map[string]bool{
				"project1/app1": true,
				"project2/app2": true,
			},
			projectName:                     "project1",
			appName:                         "app1",
			wantDefinitionSourceProjectName: "project1",
			wantDefinitionSourceAppName:     "app1",
			wantToContinue:                  false,
		},
		{
			appSelectorMap: map[string]bool{
				"project1/app1": true,
				"project2/app2": true,
			},
			projectName:                     "project3",
			appName:                         "app3",
			wantDefinitionSourceProjectName: "",
			wantDefinitionSourceAppName:     "",
			wantToContinue:                  true,
		},
		{
			appSelectorMap: map[string]bool{
				"project1/*":    true,
				"project2/app2": true,
			},
			projectName:                     "project1",
			appName:                         "app3",
			wantDefinitionSourceProjectName: "project1",
			wantDefinitionSourceAppName:     "",
			wantToContinue:                  false,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("checkAppSelectorAndGetDefinitionSourceInfo-test-%d", i+1), func(t *testing.T) {
			definitionSourceProjectName, definitionSourceAppName, toContinue :=
				checkAppSelectorAndGetDefinitionSourceInfo(testCase.appSelectorMap, testCase.projectName, testCase.appName)
			assert.Equal(t, testCase.wantDefinitionSourceProjectName, definitionSourceProjectName)
			assert.Equal(t, testCase.wantDefinitionSourceAppName, definitionSourceAppName)
			assert.Equal(t, testCase.wantToContinue, toContinue)
		})
	}
}

func TestGetSlashSeparatedString(t *testing.T) {
	testCases := []struct {
		inputStrings []string
		wantString   string
	}{
		{
			inputStrings: []string{"project1", "app1"},
			wantString:   "project1/app1",
		},
		{
			inputStrings: []string{"project2", "app2", "cluster2", "env2"},
			wantString:   "project2/app2/cluster2/env2",
		},
		{
			inputStrings: []string{"project3"},
			wantString:   "project3",
		},
		{
			inputStrings: []string{},
			wantString:   "",
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getSlashSeparatedString-test-%d", i+1), func(t *testing.T) {
			result := getSlashSeparatedString(testCase.inputStrings...)
			assert.Equal(t, testCase.wantString, result)
		})
	}
}

func TestCheckEnvSelectorAndGetDefinitionSourceInfo(t *testing.T) {

	testCases := []struct {
		allProductionEnvsFlag               bool
		envSelectorMap                      map[string]bool
		ciPipelineIdProductionEnvDetailMap  map[int][]*bean.PluginSourceCiPipelineEnvDetailDto
		ciPipelineIdEnvDetailMap            map[int][]*bean.PluginSourceCiPipelineEnvDetailDto
		ciPipelineId                        int
		definitionSourceTemplate            bean.DefinitionSourceDto
		wantDefinitionSourceTemplateForEnvs []*bean.DefinitionSourceDto
	}{
		{
			allProductionEnvsFlag: false,
			envSelectorMap: map[string]bool{
				"cluster1/env1": true,
				"cluster2/env2": true,
			},
			ciPipelineIdProductionEnvDetailMap: nil,
			ciPipelineIdEnvDetailMap:           nil,
			ciPipelineId:                       123,
			definitionSourceTemplate: bean.DefinitionSourceDto{
				ProjectName:           "project1",
				AppName:               "app1",
				CiPipelineName:        "pipeline1",
				PolicyName:            "policy1",
				IsDueToLinkedPipeline: false,
			},
			wantDefinitionSourceTemplateForEnvs: nil,
		},
		{
			allProductionEnvsFlag: true,
			envSelectorMap: map[string]bool{
				"cluster1/env1": true,
				"cluster2/env2": true,
			},
			ciPipelineIdProductionEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{
				123: {
					{
						ClusterName: "cluster1",
						EnvName:     "env1",
					},
					{
						ClusterName: "cluster2",
						EnvName:     "env2",
					},
				},
			},
			ciPipelineIdEnvDetailMap: nil,
			ciPipelineId:             123,
			definitionSourceTemplate: bean.DefinitionSourceDto{
				ProjectName:           "project1",
				AppName:               "app1",
				CiPipelineName:        "pipeline1",
				PolicyName:            "policy1",
				IsDueToLinkedPipeline: false,
			},
			wantDefinitionSourceTemplateForEnvs: []*bean.DefinitionSourceDto{
				{
					ProjectName:                  "project1",
					AppName:                      "app1",
					ClusterName:                  "cluster1",
					EnvironmentName:              "env1",
					BranchNames:                  nil,
					IsDueToProductionEnvironment: true,
					IsDueToLinkedPipeline:        false,
					CiPipelineName:               "pipeline1",
					PolicyName:                   "policy1",
				},
				{
					ProjectName:                  "project1",
					AppName:                      "app1",
					ClusterName:                  "cluster2",
					EnvironmentName:              "env2",
					BranchNames:                  nil,
					IsDueToProductionEnvironment: true,
					IsDueToLinkedPipeline:        false,
					CiPipelineName:               "pipeline1",
					PolicyName:                   "policy1",
				},
			},
		},
		{
			allProductionEnvsFlag: false,
			envSelectorMap: map[string]bool{
				"cluster1/env1": true,
				"cluster2/env2": true,
			},
			ciPipelineIdProductionEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{
				123: {
					{
						ClusterName: "cluster1",
						EnvName:     "env1",
					},
					{
						ClusterName: "cluster2",
						EnvName:     "env2",
					},
				},
			},
			ciPipelineIdEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{
				123: {
					{
						ClusterName: "cluster1",
						EnvName:     "env1",
					},
				},
			},
			ciPipelineId: 123,
			definitionSourceTemplate: bean.DefinitionSourceDto{
				ProjectName:           "project1",
				AppName:               "app1",
				CiPipelineName:        "pipeline1",
				PolicyName:            "policy1",
				IsDueToLinkedPipeline: false,
			},
			wantDefinitionSourceTemplateForEnvs: []*bean.DefinitionSourceDto{
				{
					ProjectName:                  "project1",
					AppName:                      "app1",
					ClusterName:                  "cluster1",
					EnvironmentName:              "env1",
					BranchNames:                  nil,
					IsDueToProductionEnvironment: false,
					IsDueToLinkedPipeline:        false,
					CiPipelineName:               "pipeline1",
					PolicyName:                   "policy1",
				},
			},
		},
		{
			allProductionEnvsFlag: false,
			envSelectorMap: map[string]bool{
				"cluster1/*":    true,
				"cluster2/env2": true,
			},
			ciPipelineIdProductionEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{
				123: {
					{
						ClusterName: "cluster1",
						EnvName:     "env1",
					},
					{
						ClusterName: "cluster2",
						EnvName:     "env2",
					},
				},
			},
			ciPipelineIdEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{
				123: {
					{
						ClusterName: "cluster1",
						EnvName:     "env1",
					},
				},
			},
			ciPipelineId: 123,
			definitionSourceTemplate: bean.DefinitionSourceDto{
				ProjectName:           "project1",
				AppName:               "app1",
				CiPipelineName:        "pipeline1",
				PolicyName:            "policy1",
				IsDueToLinkedPipeline: false,
			},
			wantDefinitionSourceTemplateForEnvs: []*bean.DefinitionSourceDto{
				{
					ProjectName:                  "project1",
					AppName:                      "app1",
					ClusterName:                  "cluster1",
					BranchNames:                  nil,
					IsDueToProductionEnvironment: false,
					IsDueToLinkedPipeline:        false,
					CiPipelineName:               "pipeline1",
					PolicyName:                   "policy1",
				},
			},
		},
		{
			allProductionEnvsFlag: false,
			envSelectorMap: map[string]bool{
				"cluster1/*":    true,
				"cluster2/env2": true,
			},
			ciPipelineIdProductionEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{
				123: {
					{
						ClusterName: "cluster1",
						EnvName:     "env1",
					},
					{
						ClusterName: "cluster2",
						EnvName:     "env2",
					},
				},
			},
			ciPipelineIdEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{
				123: {
					{
						ClusterName: "cluster1",
						EnvName:     "env1",
					},
					{
						ClusterName: "cluster3",
						EnvName:     "env3",
					},
				},
			},
			ciPipelineId: 123,
			definitionSourceTemplate: bean.DefinitionSourceDto{
				ProjectName:           "project1",
				AppName:               "app1",
				CiPipelineName:        "pipeline1",
				PolicyName:            "policy1",
				IsDueToLinkedPipeline: false,
			},
			wantDefinitionSourceTemplateForEnvs: []*bean.DefinitionSourceDto{
				{
					ProjectName:                  "project1",
					AppName:                      "app1",
					ClusterName:                  "cluster1",
					BranchNames:                  nil,
					IsDueToProductionEnvironment: false,
					IsDueToLinkedPipeline:        false,
					CiPipelineName:               "pipeline1",
					PolicyName:                   "policy1",
				},
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("checkEnvSelectorAndGetDefinitionSourceInfo-test-%d", i+1), func(t *testing.T) {
			definitionSourceTemplateForEnvs := checkEnvSelectorAndGetDefinitionSourceInfo(testCase.allProductionEnvsFlag, testCase.envSelectorMap,
				testCase.ciPipelineIdProductionEnvDetailMap, testCase.ciPipelineIdEnvDetailMap, testCase.ciPipelineId, testCase.definitionSourceTemplate)
			assert.Len(t, definitionSourceTemplateForEnvs, len(testCase.wantDefinitionSourceTemplateForEnvs))
			for _, wantDefinitionSourceTemplateForEnv := range testCase.wantDefinitionSourceTemplateForEnvs {
				assert.Contains(t, definitionSourceTemplateForEnvs, wantDefinitionSourceTemplateForEnv)
			}

		})
	}
}

func TestUpdateMandatoryPluginDefinitionMap(t *testing.T) {

	testCases := []struct {
		pluginIdApplyStageMap            map[string]bean.Severity
		mandatoryPluginDefinitionMap     map[string][]*bean.DefinitionSourceDto
		definitionSourceDtos             []*bean.DefinitionSourceDto
		wantMandatoryPluginDefinitionMap map[string][]*bean.DefinitionSourceDto
	}{
		{
			pluginIdApplyStageMap: map[string]bean.Severity{
				"plugin1/stage1": bean.SEVERITY_SAME_SEVERE,
				"plugin2/stage2": bean.SEVERITY_MORE_SEVERE,
			},
			mandatoryPluginDefinitionMap: map[string][]*bean.DefinitionSourceDto{
				"plugin1/stage1": {
					{
						ProjectName:                  "project1",
						AppName:                      "app1",
						ClusterName:                  "cluster1",
						EnvironmentName:              "env1",
						BranchNames:                  []string{"branch1", "branch2"},
						IsDueToProductionEnvironment: true,
						IsDueToLinkedPipeline:        false,
						CiPipelineName:               "pipeline1",
						PolicyName:                   "policy1",
					},
				},
			},
			definitionSourceDtos: []*bean.DefinitionSourceDto{
				{
					ProjectName:                  "project2",
					AppName:                      "app2",
					ClusterName:                  "cluster2",
					EnvironmentName:              "env2",
					BranchNames:                  []string{"branch3", "branch4"},
					IsDueToProductionEnvironment: false,
					IsDueToLinkedPipeline:        true,
					CiPipelineName:               "pipeline2",
					PolicyName:                   "policy2",
				},
			},
			wantMandatoryPluginDefinitionMap: map[string][]*bean.DefinitionSourceDto{
				"plugin1/stage1": {
					{
						ProjectName:                  "project1",
						AppName:                      "app1",
						ClusterName:                  "cluster1",
						EnvironmentName:              "env1",
						BranchNames:                  []string{"branch1", "branch2"},
						IsDueToProductionEnvironment: true,
						IsDueToLinkedPipeline:        false,
						CiPipelineName:               "pipeline1",
						PolicyName:                   "policy1",
					},
					{
						ProjectName:                  "project2",
						AppName:                      "app2",
						ClusterName:                  "cluster2",
						EnvironmentName:              "env2",
						BranchNames:                  []string{"branch3", "branch4"},
						IsDueToProductionEnvironment: false,
						IsDueToLinkedPipeline:        true,
						CiPipelineName:               "pipeline2",
						PolicyName:                   "policy2",
					},
				},
				"plugin2/stage2": {
					{
						ProjectName:                  "project2",
						AppName:                      "app2",
						ClusterName:                  "cluster2",
						EnvironmentName:              "env2",
						BranchNames:                  []string{"branch3", "branch4"},
						IsDueToProductionEnvironment: false,
						IsDueToLinkedPipeline:        true,
						CiPipelineName:               "pipeline2",
						PolicyName:                   "policy2",
					},
				},
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("updateMandatoryPluginDefinitionMap-test-%d", i+1), func(t *testing.T) {
			updateMandatoryPluginDefinitionMap(testCase.pluginIdApplyStageMap, testCase.mandatoryPluginDefinitionMap, testCase.definitionSourceDtos)
			assert.Equal(t, testCase.wantMandatoryPluginDefinitionMap, testCase.mandatoryPluginDefinitionMap)
		})
	}
}

func TestGetAllAppEnvBranchDetailsFromGlobalPolicyDetail(t *testing.T) {
	testCases := []struct {
		globalPolicyDetail             *bean.GlobalPolicyDetailDto
		wantAllProjects                []string
		wantAllClusters                []string
		wantBranchList                 []*bean.BranchDto
		wantIsProductionEnvFlag        bool
		wantIsAnyEnvSelectorPresent    bool
		wantProjectAppNamePolicyIdsMap map[string]bool
		wantClusterEnvNamePolicyIdsMap map[string]bool
	}{
		{
			globalPolicyDetail: &bean.GlobalPolicyDetailDto{
				Selectors: &bean.SelectorDto{
					ApplicationSelector: []*bean.ProjectAppDto{
						{
							ProjectName: "project1",
							AppNames:    []string{"app1", "app2"},
						},
						{
							ProjectName: "project2",
							AppNames:    nil,
						},
					},
					EnvironmentSelector: &bean.EnvironmentSelectorDto{
						AllProductionEnvironments: true,
						ClusterEnvList: []*bean.ClusterEnvDto{
							{
								ClusterName: "cluster1",
								EnvNames:    []string{"env1", "env2"},
							},
							{
								ClusterName: "cluster2",
								EnvNames:    nil,
							},
						},
					},
					BranchList: []*bean.BranchDto{
						{
							BranchValueType: bean2.VALUE_TYPE_REGEX,
							Value:           "branch1",
						},
						{
							BranchValueType: bean2.VALUE_TYPE_FIXED,
							Value:           "branch2",
						},
					},
				},
			},
			wantAllProjects: []string{"project1", "project2"},
			wantAllClusters: []string{"cluster1", "cluster2"},
			wantBranchList: []*bean.BranchDto{
				{
					BranchValueType: bean2.VALUE_TYPE_REGEX,
					Value:           "branch1",
				},
				{
					BranchValueType: bean2.VALUE_TYPE_FIXED,
					Value:           "branch2",
				},
			},
			wantIsProductionEnvFlag:     true,
			wantIsAnyEnvSelectorPresent: true,
			wantProjectAppNamePolicyIdsMap: map[string]bool{
				"project1/app1": true,
				"project1/app2": true,
				"project2/*":    true,
			},
			wantClusterEnvNamePolicyIdsMap: map[string]bool{
				"cluster1/env1": true,
				"cluster1/env2": true,
				"cluster2/*":    true,
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getAllAppEnvBranchDetailsFromGlobalPolicyDetail-test-%d", i+1), func(t *testing.T) {
			allProjects, allClusters, branchList, isProductionEnvFlag,
				isAnyEnvSelectorPresent, projectAppNamePolicyIdsMap, clusterEnvNamePolicyIdsMap :=
				getAllAppEnvBranchDetailsFromGlobalPolicyDetail(testCase.globalPolicyDetail)
			assert.Equal(t, testCase.wantAllProjects, allProjects)
			assert.Equal(t, testCase.wantAllClusters, allClusters)
			assert.Equal(t, testCase.wantBranchList, branchList)
			assert.Equal(t, testCase.wantIsProductionEnvFlag, isProductionEnvFlag)
			assert.Equal(t, testCase.wantIsAnyEnvSelectorPresent, isAnyEnvSelectorPresent)
			assert.Equal(t, testCase.wantProjectAppNamePolicyIdsMap, projectAppNamePolicyIdsMap)
			assert.Equal(t, testCase.wantClusterEnvNamePolicyIdsMap, clusterEnvNamePolicyIdsMap)
		})
	}
}

func TestGetFilteredCiPipelinesByProjectAppObjs(t *testing.T) {

	testCases := []struct {
		ciPipelineProjectAppNameObjs []*pipelineConfig.CiPipelineAppProject
		projectAppNameMap            map[string]bool
		wanFilteredCiPipelines       []int
	}{
		{
			ciPipelineProjectAppNameObjs: []*pipelineConfig.CiPipelineAppProject{
				{
					CiPipelineId: 1,
					ProjectName:  "project1",
					AppName:      "app1",
				},
				{
					CiPipelineId: 2,
					ProjectName:  "project1",
					AppName:      "app2",
				},
				{
					CiPipelineId: 3,
					ProjectName:  "project2",
					AppName:      "app3",
				},
			},
			projectAppNameMap: map[string]bool{
				"project1/app1": true,
				"project2/app2": true,
			},
			wanFilteredCiPipelines: []int{1},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getFilteredCiPipelinesByProjectAppObjs-test-%d", i+1), func(t *testing.T) {
			filteredCiPipelines := getFilteredCiPipelinesByProjectAppObjs(testCase.ciPipelineProjectAppNameObjs, testCase.projectAppNameMap)
			assert.Equal(t, testCase.wanFilteredCiPipelines, filteredCiPipelines)
		})
	}

}

func TestGetFilteredCiPipelinesByClusterAndEnvObjs(t *testing.T) {
	testCases := []struct {
		ciPipelineClusterEnvNameObjs []*pipelineConfig.CiPipelineEnvCluster
		isProductionEnvFlag          bool
		clusterEnvNameMap            map[string]bool
		wanFilteredCiPipelines       []int
	}{
		{
			ciPipelineClusterEnvNameObjs: []*pipelineConfig.CiPipelineEnvCluster{
				{
					CiPipelineId:    1,
					ClusterName:     "cluster1",
					EnvName:         "env1",
					IsProductionEnv: true,
				},
				{
					CiPipelineId:    2,
					ClusterName:     "cluster1",
					EnvName:         "env2",
					IsProductionEnv: false,
				},
				{
					CiPipelineId:    3,
					ClusterName:     "cluster2",
					EnvName:         "env3",
					IsProductionEnv: true,
				},
			},
			isProductionEnvFlag: true,
			clusterEnvNameMap: map[string]bool{
				"cluster1/env1": true,
				"cluster2/env3": true,
			},
			wanFilteredCiPipelines: []int{1, 3},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getFilteredCiPipelinesByClusterAndEnvObjs-test-%d", i+1), func(t *testing.T) {
			filteredCiPipelines := getFilteredCiPipelinesByClusterAndEnvObjs(testCase.ciPipelineClusterEnvNameObjs, testCase.isProductionEnvFlag, testCase.clusterEnvNameMap)
			assert.Equal(t, testCase.wanFilteredCiPipelines, filteredCiPipelines)
		})
	}
}

func TestGetSearchableKeyIdValueEntriesForASelectorAttribute(t *testing.T) {
	globalPolicy, _, _, _, _, _, _, _, _, _, _, err := getGlobalPolicyService(t)
	assert.NoError(t, err)
	testCases := []struct {
		policy                     *bean.GlobalPolicyDto
		attribute                  bean2.DevtronResourceAttributeName
		searchableKeyNameIdMap     map[bean2.DevtronResourceSearchableKeyName]int
		wantSearchableFieldEntries []*repository.GlobalPolicySearchableField
	}{
		{
			policy: &bean.GlobalPolicyDto{
				Id:     1,
				UserId: 123,
				GlobalPolicyDetailDto: &bean.GlobalPolicyDetailDto{
					Selectors: &bean.SelectorDto{
						ApplicationSelector: []*bean.ProjectAppDto{
							{
								ProjectName: "project1",
								AppNames:    []string{"app1", "app2"},
							},
							{
								ProjectName: "project2",
								AppNames:    nil,
							},
						},
						EnvironmentSelector: &bean.EnvironmentSelectorDto{
							AllProductionEnvironments: true,
							ClusterEnvList: []*bean.ClusterEnvDto{
								{
									ClusterName: "cluster1",
									EnvNames:    []string{"env1", "env2"},
								},
								{
									ClusterName: "cluster2",
									EnvNames:    nil,
								},
							},
						},
						BranchList: []*bean.BranchDto{
							{
								Value:           "branch1",
								BranchValueType: bean2.VALUE_TYPE_REGEX,
							},
							{
								Value:           "branch2",
								BranchValueType: bean2.VALUE_TYPE_FIXED,
							},
						},
					},
				},
			},
			attribute: bean2.DEVTRON_RESOURCE_ATTRIBUTE_APP_NAME,
			searchableKeyNameIdMap: map[bean2.DevtronResourceSearchableKeyName]int{
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME:      1,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV: 2,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME:      3,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH:    4,
			},
			wantSearchableFieldEntries: []*repository.GlobalPolicySearchableField{
				{
					GlobalPolicyId:  1,
					SearchableKeyId: 1,
					Value:           "project1/app1",
					IsRegex:         false,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
				{
					GlobalPolicyId:  1,
					SearchableKeyId: 1,
					Value:           "project1/app2",
					IsRegex:         false,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
				{
					GlobalPolicyId:  1,
					SearchableKeyId: 1,
					Value:           "project2/*",
					IsRegex:         true,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
			},
		},
		{
			policy: &bean.GlobalPolicyDto{
				Id:     1,
				UserId: 123,
				GlobalPolicyDetailDto: &bean.GlobalPolicyDetailDto{
					Selectors: &bean.SelectorDto{
						ApplicationSelector: []*bean.ProjectAppDto{
							{
								ProjectName: "project1",
								AppNames:    []string{"app1", "app2"},
							},
							{
								ProjectName: "project2",
								AppNames:    nil,
							},
						},
						EnvironmentSelector: &bean.EnvironmentSelectorDto{
							AllProductionEnvironments: true,
							ClusterEnvList: []*bean.ClusterEnvDto{
								{
									ClusterName: "cluster1",
									EnvNames:    []string{"env1", "env2"},
								},
								{
									ClusterName: "cluster2",
									EnvNames:    nil,
								},
							},
						},
						BranchList: []*bean.BranchDto{
							{
								Value:           "branch1",
								BranchValueType: bean2.VALUE_TYPE_REGEX,
							},
							{
								Value:           "branch2",
								BranchValueType: bean2.VALUE_TYPE_FIXED,
							},
						},
					},
				},
			},
			attribute: bean2.DEVTRON_RESOURCE_ATTRIBUTE_ENVIRONMENT_NAME,
			searchableKeyNameIdMap: map[bean2.DevtronResourceSearchableKeyName]int{
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME:      1,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV: 2,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME:      3,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH:    4,
			},
			wantSearchableFieldEntries: []*repository.GlobalPolicySearchableField{
				{
					GlobalPolicyId:  1,
					SearchableKeyId: 3,
					Value:           "cluster1/*",
					IsRegex:         true,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
				{
					GlobalPolicyId:  1,
					SearchableKeyId: 3,
					Value:           "cluster1/env1",
					IsRegex:         false,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
				{
					GlobalPolicyId:  1,
					SearchableKeyId: 3,
					Value:           "cluster1/env2",
					IsRegex:         false,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
				{
					GlobalPolicyId:  1,
					SearchableKeyId: 3,
					Value:           "cluster2/*",
					IsRegex:         true,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
			},
		},
		{
			policy: &bean.GlobalPolicyDto{
				Id:     1,
				UserId: 123,
				GlobalPolicyDetailDto: &bean.GlobalPolicyDetailDto{
					Selectors: &bean.SelectorDto{
						ApplicationSelector: []*bean.ProjectAppDto{
							{
								ProjectName: "project1",
								AppNames:    []string{"app1", "app2"},
							},
							{
								ProjectName: "project2",
								AppNames:    nil,
							},
						},
						EnvironmentSelector: &bean.EnvironmentSelectorDto{
							AllProductionEnvironments: true,
							ClusterEnvList: []*bean.ClusterEnvDto{
								{
									ClusterName: "cluster1",
									EnvNames:    []string{"env1", "env2"},
								},
								{
									ClusterName: "cluster2",
									EnvNames:    nil,
								},
							},
						},
						BranchList: []*bean.BranchDto{
							{
								Value:           "branch1",
								BranchValueType: bean2.VALUE_TYPE_REGEX,
							},
							{
								Value:           "branch2",
								BranchValueType: bean2.VALUE_TYPE_FIXED,
							},
						},
					},
				},
			},
			attribute: bean2.DEVTRON_RESOURCE_ATTRIBUTE_ENVIRONMENT_IS_PRODUCTION,
			searchableKeyNameIdMap: map[bean2.DevtronResourceSearchableKeyName]int{
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME:      1,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV: 2,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME:      3,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH:    4,
			},
			wantSearchableFieldEntries: []*repository.GlobalPolicySearchableField{
				{
					GlobalPolicyId:  1,
					SearchableKeyId: 2,
					Value:           "true",
					IsRegex:         false,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
			},
		},
		{
			policy: &bean.GlobalPolicyDto{
				Id:     1,
				UserId: 123,
				GlobalPolicyDetailDto: &bean.GlobalPolicyDetailDto{
					Selectors: &bean.SelectorDto{
						ApplicationSelector: []*bean.ProjectAppDto{
							{
								ProjectName: "project1",
								AppNames:    []string{"app1", "app2"},
							},
							{
								ProjectName: "project2",
								AppNames:    nil,
							},
						},
						EnvironmentSelector: &bean.EnvironmentSelectorDto{
							AllProductionEnvironments: true,
							ClusterEnvList: []*bean.ClusterEnvDto{
								{
									ClusterName: "cluster1",
									EnvNames:    []string{"env1", "env2"},
								},
								{
									ClusterName: "cluster2",
									EnvNames:    nil,
								},
							},
						},
						BranchList: []*bean.BranchDto{
							{
								Value:           "branch1",
								BranchValueType: bean2.VALUE_TYPE_REGEX,
							},
							{
								Value:           "branch2",
								BranchValueType: bean2.VALUE_TYPE_FIXED,
							},
						},
					},
				},
			},
			attribute: bean2.DEVTRON_RESOURCE_ATTRIBUTE_CI_PIPELINE_BRANCH_VALUE,
			searchableKeyNameIdMap: map[bean2.DevtronResourceSearchableKeyName]int{
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME:      1,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV: 2,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME:      3,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH:    4,
			},
			wantSearchableFieldEntries: []*repository.GlobalPolicySearchableField{
				{
					GlobalPolicyId:  1,
					SearchableKeyId: 4,
					Value:           "branch1",
					IsRegex:         true,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
				{
					GlobalPolicyId:  1,
					SearchableKeyId: 4,
					Value:           "branch2",
					IsRegex:         false,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getSearchableKeyIdValueEntriesForASelectorAttribute-test-%d", i+1), func(t *testing.T) {
			searchableFieldEntries := globalPolicy.getSearchableKeyIdValueEntriesForASelectorAttribute(testCase.policy, testCase.attribute, testCase.searchableKeyNameIdMap)
			for _, searchableFieldEntry := range searchableFieldEntries {
				searchableFieldEntry.AuditLog = sql.AuditLog{}
				assert.Contains(t, testCase.wantSearchableFieldEntries, searchableFieldEntry)
			}

		})
	}
}

func TestGetEnvSelectorMap(t *testing.T) {

	testCases := []struct {
		clusterEnvList     []*bean.ClusterEnvDto
		wantEnvSelectorMap map[string]bool
	}{
		{
			clusterEnvList: []*bean.ClusterEnvDto{
				{
					ClusterName: "cluster1",
					EnvNames:    []string{"env1", "env2"},
				},
				{
					ClusterName: "cluster2",
					EnvNames:    []string{"env3"},
				},
				{
					ClusterName: "cluster3",
					EnvNames:    []string{},
				},
			},
			wantEnvSelectorMap: map[string]bool{
				"cluster1/env1": true,
				"cluster1/env2": true,
				"cluster2/env3": true,
				"cluster3/*":    true,
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getEnvSelectorMap-test-%d", i+1), func(t *testing.T) {
			envSelectorMap := getEnvSelectorMap(testCase.clusterEnvList)
			assert.Equal(t, testCase.wantEnvSelectorMap, envSelectorMap)
		})
	}
}

func TestGetAppSelectorMap(t *testing.T) {

	testCases := []struct {
		projectAppList     []*bean.ProjectAppDto
		wantAppSelectorMap map[string]bool
	}{
		{
			projectAppList: []*bean.ProjectAppDto{
				{
					ProjectName: "project1",
					AppNames:    []string{"app1", "app2"},
				},
				{
					ProjectName: "project2",
					AppNames:    []string{"app3"},
				},
				{
					ProjectName: "project3",
					AppNames:    []string{},
				},
			},
			wantAppSelectorMap: map[string]bool{
				"project1/app1": true,
				"project1/app2": true,
				"project2/app3": true,
				"project3/*":    true,
			},
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getAppSelectorMap-test-%d", i+1), func(t *testing.T) {
			appSelectorMap := getAppSelectorMap(testCase.projectAppList)
			assert.Equal(t, testCase.wantAppSelectorMap, appSelectorMap)
		})
	}
}

func TestGetSearchableKeyIdValueMapForFilter(t *testing.T) {
	searchableKeyNameIdMap := map[bean2.DevtronResourceSearchableKeyName]int{
		bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME:           1,
		bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME:           2,
		bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV:      3,
		bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH:         4,
		bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_TRIGGER_ACTION: 5,
	}

	testCases := []struct {
		allProjectAppNames            []string
		allClusterEnvNames            []string
		branchValues                  []string
		haveAnyProductionEnv          bool
		toOnlyGetBlockedStatePolicies bool
		wantWhereOrGroup              map[int][]string
		wantWhereAndGroup             map[int][]string
	}{
		{
			allProjectAppNames:            []string{"app1", "app2"},
			allClusterEnvNames:            []string{"env1", "env2"},
			branchValues:                  []string{"branch1", "branch2"},
			haveAnyProductionEnv:          true,
			toOnlyGetBlockedStatePolicies: true,
			wantWhereOrGroup: map[int][]string{
				1: {"app1", "app2"},
				2: {"env1", "env2"},
				3: {"true"},
				4: {"branch1", "branch2"},
			},
			wantWhereAndGroup: map[int][]string{
				5: {bean.CONSEQUENCE_ACTION_BLOCK.ToString(), bean.CONSEQUENCE_ACTION_ALLOW_UNTIL_TIME.ToString()},
			},
		},
		{
			allProjectAppNames:            []string{},
			allClusterEnvNames:            []string{},
			branchValues:                  []string{},
			haveAnyProductionEnv:          false,
			toOnlyGetBlockedStatePolicies: false,
			wantWhereOrGroup:              map[int][]string{},
			wantWhereAndGroup:             map[int][]string{},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getSearchableKeyIdValueMapForFilter-test-%d", i+1), func(t *testing.T) {
			whereOrGroup, whereAndGroup := getSearchableKeyIdValueMapForFilter(testCase.allProjectAppNames, testCase.allClusterEnvNames,
				testCase.branchValues, testCase.haveAnyProductionEnv, testCase.toOnlyGetBlockedStatePolicies, searchableKeyNameIdMap)
			assert.Equal(t, testCase.wantWhereOrGroup, whereOrGroup)
			assert.Equal(t, testCase.wantWhereAndGroup, whereAndGroup)
		})
	}
}

func TestGetPipelineTobeUsedToFetchConfiguredPlugins(t *testing.T) {
	testCases := []struct {
		ciPipelineId       int
		parentCiPipelineId int
		wantPipelineToUse  int
	}{
		{
			ciPipelineId:       1,
			parentCiPipelineId: 2,
			wantPipelineToUse:  2,
		},
		{
			ciPipelineId:       1,
			parentCiPipelineId: 0,
			wantPipelineToUse:  1,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getPipelineTobeUsedToFetchConfiguredPlugins-test-%d", i+1), func(t *testing.T) {
			pipelineToUse := getPipelineTobeUsedToFetchConfiguredPlugins(&pipelineConfig.CiPipelineMaterial{
				CiPipelineId:     testCase.ciPipelineId,
				ParentCiPipeline: testCase.parentCiPipelineId,
			})
			assert.Equal(t, testCase.wantPipelineToUse, pipelineToUse)
		})
	}
}

func TestGetCIPipelineConfiguredPluginMap(t *testing.T) {
	testCases := []struct {
		configuredPlugins                   []*repository3.PipelineStageStep
		wantCiPipelineConfiguredPipelineMap map[int]map[string]bool
	}{
		{
			configuredPlugins: []*repository3.PipelineStageStep{
				{
					PipelineStage: &repository3.PipelineStage{
						CiPipelineId: 1,
						Type:         repository3.PIPELINE_STAGE_TYPE_PRE_CI,
					},
					RefPluginId: 1,
				},
				{
					PipelineStage: &repository3.PipelineStage{
						CiPipelineId: 1,
						Type:         repository3.PIPELINE_STAGE_TYPE_POST_CI,
					},
					RefPluginId: 2,
				},
			},
			wantCiPipelineConfiguredPipelineMap: map[int]map[string]bool{
				1: {
					"1/PRE_CI":         true,
					"1/PRE_OR_POST_CI": true,
					"2/POST_CI":        true,
					"2/PRE_OR_POST_CI": true,
				},
			},
		},
		{
			configuredPlugins: []*repository3.PipelineStageStep{
				{
					PipelineStage: &repository3.PipelineStage{
						CiPipelineId: 1,
						Type:         repository3.PIPELINE_STAGE_TYPE_PRE_CI,
					},
					RefPluginId: 1,
				},
				{
					PipelineStage: &repository3.PipelineStage{
						CiPipelineId: 2,
						Type:         repository3.PIPELINE_STAGE_TYPE_PRE_CI,
					},
					RefPluginId: 2,
				},
				{
					PipelineStage: &repository3.PipelineStage{
						CiPipelineId: 2,
						Type:         repository3.PIPELINE_STAGE_TYPE_POST_CI,
					},
					RefPluginId: 3,
				},
			},
			wantCiPipelineConfiguredPipelineMap: map[int]map[string]bool{
				1: {
					"1/PRE_CI":         true,
					"1/PRE_OR_POST_CI": true,
				},
				2: {
					"2/PRE_CI":         true,
					"2/PRE_OR_POST_CI": true,
					"3/POST_CI":        true,
					"3/PRE_OR_POST_CI": true,
				},
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getPipelineTobeUsedToFetchConfiguredPlugins-test-%d", i+1), func(t *testing.T) {
			ciPipelineConfiguredPipelineMap := getCIPipelineConfiguredPluginMap(testCase.configuredPlugins)
			assert.Equal(t, testCase.wantCiPipelineConfiguredPipelineMap, ciPipelineConfiguredPipelineMap)
		})
	}
}

func TestGetOffendingCIPipelineIds(t *testing.T) {
	testCases := []struct {
		ciPipelinesForConfiguredPlugins []int
		ciPipelineConfiguredPluginMap   map[int]map[string]bool
		globalPolicyDetailDto           bean.GlobalPolicyDetailDto
		ciPipelineParentChildMap        map[int][]int
		wantOffendingCiPipelineIds      []int
	}{
		{
			ciPipelinesForConfiguredPlugins: []int{1, 2, 3},
			ciPipelineConfiguredPluginMap: map[int]map[string]bool{
				1: {
					"1/PRE_CI": true,
				},
				2: {
					"2/PRE_CI": true,
				},
				3: {
					"3/PRE_CI": true,
				},
			},
			globalPolicyDetailDto: bean.GlobalPolicyDetailDto{
				Definitions: []*bean.DefinitionDto{
					{
						Data: bean.DefinitionDataDto{
							PluginId:     1,
							ApplyToStage: bean.PLUGIN_APPLY_STAGE_PRE_CI,
						},
					},
				},
			},
			ciPipelineParentChildMap:   map[int][]int{},
			wantOffendingCiPipelineIds: []int{},
		},
		{
			ciPipelinesForConfiguredPlugins: []int{1, 2, 3},
			ciPipelineConfiguredPluginMap: map[int]map[string]bool{
				1: {
					"1/PRE_CI": true,
				},
				3: {
					"3/PRE_CI": true,
				},
			},
			globalPolicyDetailDto: bean.GlobalPolicyDetailDto{
				Definitions: []*bean.DefinitionDto{
					{
						Data: bean.DefinitionDataDto{
							PluginId:     1,
							ApplyToStage: bean.PLUGIN_APPLY_STAGE_PRE_CI,
						},
					},
				},
			},
			ciPipelineParentChildMap:   map[int][]int{},
			wantOffendingCiPipelineIds: []int{2},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getOffendingCIPipelineIds-test-%d", i+1), func(t *testing.T) {
			offendingCiPipelineIds := getOffendingCIPipelineIds(testCase.ciPipelinesForConfiguredPlugins, testCase.ciPipelineConfiguredPluginMap, testCase.globalPolicyDetailDto, testCase.ciPipelineParentChildMap)
			assert.Equal(t, testCase.wantOffendingCiPipelineIds, offendingCiPipelineIds)
		})
	}
}

func TestGetConfiguredPluginMap(t *testing.T) {
	testCases := []struct {
		configuredPlugins       []*repository3.PipelineStageStep
		wantConfiguredPluginMap map[string]bool
	}{
		{
			configuredPlugins: []*repository3.PipelineStageStep{
				{
					RefPluginId: 1,
					PipelineStage: &repository3.PipelineStage{
						Type: repository3.PIPELINE_STAGE_TYPE_PRE_CI,
					},
				},
			},
			wantConfiguredPluginMap: map[string]bool{
				"1/PRE_CI":         true,
				"1/PRE_OR_POST_CI": true,
			},
		},
		{
			configuredPlugins: []*repository3.PipelineStageStep{
				{
					RefPluginId: 1,
					PipelineStage: &repository3.PipelineStage{
						Type: repository3.PIPELINE_STAGE_TYPE_PRE_CI,
					},
				},
				{
					RefPluginId: 2,
					PipelineStage: &repository3.PipelineStage{
						Type: repository3.PIPELINE_STAGE_TYPE_POST_CI,
					},
				},
			},
			wantConfiguredPluginMap: map[string]bool{
				"1/PRE_CI":         true,
				"1/PRE_OR_POST_CI": true,
				"2/POST_CI":        true,
				"2/PRE_OR_POST_CI": true,
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getConfiguredPluginMap-test-%d", i+1), func(t *testing.T) {
			configuredPluginMap := getConfiguredPluginMap(testCase.configuredPlugins)
			assert.Equal(t, testCase.wantConfiguredPluginMap, configuredPluginMap)
		})
	}
}

func TestGetBlockageStateDetails(t *testing.T) {
	testCases := []struct {
		definitions              []*bean.MandatoryPluginDefinitionDto
		configuredPluginMap      map[string]bool
		mandatoryPluginsBlockage map[string]*bean.ConsequenceDto
		wantOffendingPlugin      bool
		wantCIPipelineBlocked    bool
		wantBlockageState        *bean.ConsequenceDto
	}{
		{
			definitions: []*bean.MandatoryPluginDefinitionDto{
				{
					DefinitionDto: &bean.DefinitionDto{
						AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
						Data: bean.DefinitionDataDto{
							PluginId:     1,
							ApplyToStage: bean.PLUGIN_APPLY_STAGE_PRE_CI,
						},
					},
				},
			},
			configuredPluginMap: map[string]bool{
				"1/PRE_CI": true,
			},
			mandatoryPluginsBlockage: map[string]*bean.ConsequenceDto{
				"1/PRE_CI": {
					Action: bean.CONSEQUENCE_ACTION_BLOCK,
				},
			},
			wantOffendingPlugin:   false,
			wantCIPipelineBlocked: false,
			wantBlockageState:     nil,
		},
		{
			definitions: []*bean.MandatoryPluginDefinitionDto{
				{
					DefinitionDto: &bean.DefinitionDto{
						AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
						Data: bean.DefinitionDataDto{
							PluginId:     1,
							ApplyToStage: bean.PLUGIN_APPLY_STAGE_PRE_CI,
						},
					},
				},
				{
					DefinitionDto: &bean.DefinitionDto{
						AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
						Data: bean.DefinitionDataDto{
							PluginId:     3,
							ApplyToStage: bean.PLUGIN_APPLY_STAGE_PRE_OR_POST_CI,
						},
					},
				},
			},
			configuredPluginMap: map[string]bool{
				"2/PRE_CI": true,
			},
			mandatoryPluginsBlockage: map[string]*bean.ConsequenceDto{
				"1/PRE_CI": {
					Action:        bean.CONSEQUENCE_ACTION_ALLOW_UNTIL_TIME,
					MetadataField: time.Now().Add(-time.Hour),
				},
				"3/PRE_OR_POST_CI": {
					Action: bean.CONSEQUENCE_ACTION_BLOCK,
				},
			},
			wantOffendingPlugin:   true,
			wantCIPipelineBlocked: true,
			wantBlockageState: &bean.ConsequenceDto{
				Action: bean.CONSEQUENCE_ACTION_BLOCK,
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getBlockageStateDetails-test-%d", i+1), func(t *testing.T) {
			offendingPlugin, ciPipelineBlocked, blockageState := getBlockageStateDetails(testCase.definitions, testCase.configuredPluginMap, testCase.mandatoryPluginsBlockage)
			assert.Equal(t, testCase.wantOffendingPlugin, offendingPlugin)
			assert.Equal(t, testCase.wantCIPipelineBlocked, ciPipelineBlocked)
			assert.Equal(t, testCase.wantBlockageState, blockageState)
		})
	}
}

func TestGlobalPolicyDbAdapter(t *testing.T) {
	testCases := []struct {
		policyDto            *bean.GlobalPolicyDto
		policyDetailJson     string
		oldEntry             *repository.GlobalPolicy
		wantToCheckCreatedOn bool
		wantCreatedOn        time.Time
		wantCreatedBy        int32
		wantUpdatedOn        time.Time
		wantUpdatedBy        int32
	}{
		{
			policyDto: &bean.GlobalPolicyDto{
				Id:            1,
				Name:          "Policy 1",
				PolicyOf:      bean.GLOBAL_POLICY_TYPE_PLUGIN,
				PolicyVersion: bean.GLOBAL_POLICY_VERSION_V1,
				Description:   "Test policy",
				Enabled:       true,
				UserId:        1,
			},
			policyDetailJson: `{"policyData": "test data"}`,
			oldEntry: &repository.GlobalPolicy{
				Id:   1,
				Name: "Old Policy",
				AuditLog: sql.AuditLog{
					CreatedOn: func() time.Time {
						now := time.Now()
						return now.Truncate(time.Second).Add(-709133000)
					}(),
					CreatedBy: 2,
					UpdatedOn: time.Now(),
					UpdatedBy: 1,
				},
			},
			wantToCheckCreatedOn: true,
			wantCreatedOn: func() time.Time {
				now := time.Now()
				return now.Truncate(time.Second).Add(-709133000)
			}(),
			wantCreatedBy: 2,
			wantUpdatedOn: func() time.Time {
				now := time.Now()
				return now.Truncate(time.Second)
			}(),
			wantUpdatedBy: 1,
		},
		{
			policyDto: &bean.GlobalPolicyDto{
				Id:            1,
				Name:          "Policy 1",
				PolicyOf:      bean.GLOBAL_POLICY_TYPE_PLUGIN,
				PolicyVersion: bean.GLOBAL_POLICY_VERSION_V1,
				Description:   "Test policy",
				Enabled:       true,
				UserId:        1,
			},
			policyDetailJson:     `{"policyData": "test data"}`,
			oldEntry:             nil,
			wantToCheckCreatedOn: false,
			wantCreatedOn: func() time.Time {
				now := time.Now()
				return now.Truncate(time.Second)
			}(),
			wantCreatedBy: 1,
			wantUpdatedOn: func() time.Time {
				now := time.Now()
				return now.Truncate(time.Second)
			}(),
			wantUpdatedBy: 1,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("globalPolicyDbAdapter-test-%d", i+1), func(t *testing.T) {
			wantGlobalPolicyDbObj := &repository.GlobalPolicy{
				Id:          testCase.policyDto.Id,
				Name:        testCase.policyDto.Name,
				PolicyJson:  testCase.policyDetailJson,
				PolicyOf:    testCase.policyDto.PolicyOf.ToString(),
				Version:     testCase.policyDto.PolicyVersion.ToString(),
				Description: testCase.policyDto.Description,
				Enabled:     testCase.policyDto.Enabled,
				Deleted:     false,
				AuditLog: sql.AuditLog{
					CreatedOn: testCase.wantCreatedOn,
					CreatedBy: testCase.wantCreatedBy,
					UpdatedOn: testCase.wantUpdatedOn,
					UpdatedBy: testCase.wantUpdatedBy,
				},
			}
			actualGlobalPolicyDbObj := globalPolicyDbAdapter(testCase.policyDto, testCase.policyDetailJson, testCase.oldEntry)
			actualGlobalPolicyDbObj.UpdatedOn = func() time.Time {
				now := time.Now()
				return now.Truncate(time.Second)
			}()
			if !testCase.wantToCheckCreatedOn {
				actualGlobalPolicyDbObj.CreatedOn = func() time.Time {
					now := time.Now()
					return now.Truncate(time.Second)
				}()
			}
			assert.Equal(t, wantGlobalPolicyDbObj, actualGlobalPolicyDbObj)
		})
	}
}

func TestGetWfComponentDetailsAndMap(t *testing.T) {
	testCases := []struct {
		appWorkflowMappings    []*appWorkflow.AppWorkflowMapping
		wantWfComponentDetails []*bean.WorkflowTreeComponentDto
		wantWfIdComponentMap   map[int]int
		wantAppIds             []int
		wantCdPipelineIds      []int
	}{
		{
			appWorkflowMappings: []*appWorkflow.AppWorkflowMapping{
				{
					AppWorkflowId: 1,
					AppWorkflow: &appWorkflow.AppWorkflow{
						AppId: 101,
						Name:  "Workflow 1",
					},
					ComponentId: 200,
					Type:        appWorkflow.CIPIPELINE,
				},
				{
					AppWorkflowId: 1,
					AppWorkflow: &appWorkflow.AppWorkflow{
						AppId: 101,
						Name:  "Workflow 1",
					},
					ComponentId: 201,
					Type:        appWorkflow.CDPIPELINE,
					ParentId:    200,
					ParentType:  appWorkflow.CIPIPELINE,
				},
			},
			wantWfComponentDetails: []*bean.WorkflowTreeComponentDto{
				{
					Id:    1,
					AppId: 101,
					Name:  "Workflow 1",
				},
			},
			wantWfIdComponentMap: map[int]int{
				1: 0,
			},
			wantAppIds:        []int{101, 101}, //for
			wantCdPipelineIds: []int{201},
		},
		{
			appWorkflowMappings: []*appWorkflow.AppWorkflowMapping{
				{
					AppWorkflowId: 1,
					AppWorkflow: &appWorkflow.AppWorkflow{
						AppId: 101,
						Name:  "Workflow 1",
					},
					ComponentId: 200,
					Type:        appWorkflow.CIPIPELINE,
				},
				{
					AppWorkflowId: 2,
					AppWorkflow: &appWorkflow.AppWorkflow{
						AppId: 102,
						Name:  "Workflow 2",
					},
					ComponentId: 201,
					Type:        appWorkflow.CDPIPELINE,
				},
			},
			wantWfComponentDetails: []*bean.WorkflowTreeComponentDto{
				{
					Id:    1,
					AppId: 101,
					Name:  "Workflow 1",
				},
				{
					Id:    2,
					AppId: 102,
					Name:  "Workflow 2",
				},
			},
			wantWfIdComponentMap: map[int]int{
				1: 0,
				2: 1,
			},
			wantAppIds:        []int{101, 102},
			wantCdPipelineIds: []int{201},
		},
		{
			appWorkflowMappings:    []*appWorkflow.AppWorkflowMapping(nil),
			wantWfComponentDetails: []*bean.WorkflowTreeComponentDto(nil),
			wantWfIdComponentMap:   map[int]int{},
			wantAppIds:             []int(nil),
			wantCdPipelineIds:      []int(nil),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getWfComponentDetailsAndMap-test-%d", i+1), func(t *testing.T) {
			actualWfComponentDetails, actualWfIdComponentMap, actualAppIds, actualCdPipelineIds :=
				getWfComponentDetailsAndMap(testCase.appWorkflowMappings)

			assert.Equal(t, testCase.wantWfComponentDetails, actualWfComponentDetails)
			assert.Equal(t, testCase.wantWfIdComponentMap, actualWfIdComponentMap)
			assert.Equal(t, testCase.wantAppIds, actualAppIds)
			assert.Equal(t, testCase.wantCdPipelineIds, actualCdPipelineIds)
		})
	}
}

func TestGetUpdatedWfComponentDetailsWithCIAndCDInfo(t *testing.T) {
	testCases := []struct {
		appWorkflowMappings         []*appWorkflow.AppWorkflowMapping
		wfIdAndComponentDtoIndexMap map[int]int
		appIdGitMaterialMap         map[int][]*bean.Material
		wfComponentDetails          []*bean.WorkflowTreeComponentDto
		ciPipelineMaterialMap       map[int][]*pipelineConfig.CiPipelineMaterial
		ciPipelineIdNameMap         map[int]string
		cdPipelineIdNameMap         map[int]string
		wantWfComponentDetails      []*bean.WorkflowTreeComponentDto
	}{
		{
			appWorkflowMappings: []*appWorkflow.AppWorkflowMapping{
				{
					AppWorkflowId: 1,
					AppWorkflow: &appWorkflow.AppWorkflow{
						AppId: 101,
					},
					ComponentId: 201,
					Type:        appWorkflow.CIPIPELINE,
				},
			},
			wfIdAndComponentDtoIndexMap: map[int]int{
				1: 0,
			},
			appIdGitMaterialMap: map[int][]*bean.Material{
				101: {
					{
						MaterialName: "Git Material 1",
					},
				},
			},
			wfComponentDetails: []*bean.WorkflowTreeComponentDto{
				{
					Id:    1,
					AppId: 101,
				},
			},
			ciPipelineMaterialMap: map[int][]*pipelineConfig.CiPipelineMaterial{
				201: {
					{
						Id: 1,
					},
				},
			},
			ciPipelineIdNameMap: map[int]string{
				201: "CIPipeline1",
			},
			cdPipelineIdNameMap: map[int]string{},
			wantWfComponentDetails: []*bean.WorkflowTreeComponentDto{
				{
					Id:             1,
					AppId:          101,
					GitMaterials:   []*bean.Material{{MaterialName: "Git Material 1"}},
					CiPipelineId:   201,
					CiMaterials:    []*pipelineConfig.CiPipelineMaterial{{Id: 1}},
					CiPipelineName: "CIPipeline1",
				},
			},
		},
		{
			appWorkflowMappings: []*appWorkflow.AppWorkflowMapping{
				{
					AppWorkflowId: 1,
					AppWorkflow: &appWorkflow.AppWorkflow{
						AppId: 101,
					},
					ComponentId: 301,
					Type:        appWorkflow.CDPIPELINE,
				},
			},
			wfIdAndComponentDtoIndexMap: map[int]int{
				1: 0,
			},
			appIdGitMaterialMap: map[int][]*bean.Material{
				101: {
					{
						MaterialName: "Git Material 1",
					},
				},
			},
			wfComponentDetails: []*bean.WorkflowTreeComponentDto{
				{
					Id:    1,
					AppId: 101,
				},
			},
			ciPipelineMaterialMap: map[int][]*pipelineConfig.CiPipelineMaterial{},
			ciPipelineIdNameMap:   map[int]string{},
			cdPipelineIdNameMap: map[int]string{
				301: "CD Pipeline 1",
			},
			wantWfComponentDetails: []*bean.WorkflowTreeComponentDto{
				{
					Id:           1,
					AppId:        101,
					GitMaterials: []*bean.Material{{MaterialName: "Git Material 1"}},
					CdPipelines:  []string{"CD Pipeline 1"},
				},
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getUpdatedWfComponentDetailsWithCIAndCDInfo-test-%d", i+1), func(t *testing.T) {
			actualWfComponentDetails := getUpdatedWfComponentDetailsWithCIAndCDInfo(testCase.appWorkflowMappings, testCase.wfIdAndComponentDtoIndexMap,
				testCase.appIdGitMaterialMap, testCase.wfComponentDetails, testCase.ciPipelineMaterialMap, testCase.ciPipelineIdNameMap, testCase.cdPipelineIdNameMap)
			assert.Equal(t, testCase.wantWfComponentDetails, actualWfComponentDetails)
		})
	}
}

func TestGetDefinitionSourceDtosForACIPipeline(t *testing.T) {
	testCases := []struct {
		ciPipelineIdInObj             int
		ciPipelineId                  int
		matchedBranchList             []string
		ciPipelineIdNameMap           map[int]string
		globalPolicyName              string
		ciPipelineIdProjectAppNameMap map[int]*bean.PluginSourceCiPipelineAppDetailDto
		needToCheckAppSelector        bool
		needToCheckEnvSelector        bool
		allProductionEnvsFlag         bool
		appSelectorMap                map[string]bool
		envSelectorMap                map[string]bool
		ciPipelineIdProductionEnvDetailMap,
		ciPipelineIdEnvDetailMap map[int][]*bean.PluginSourceCiPipelineEnvDetailDto
		wantDefinitionSourceDtos []*bean.DefinitionSourceDto
	}{
		{
			ciPipelineIdInObj: 1,
			ciPipelineId:      1,
			matchedBranchList: []string{"main"},
			ciPipelineIdNameMap: map[int]string{
				1: "CIPipeline1",
			},
			globalPolicyName: "GlobalPolicy1",
			ciPipelineIdProjectAppNameMap: map[int]*bean.PluginSourceCiPipelineAppDetailDto{
				1: {
					ProjectName: "Project1",
					AppName:     "App1",
				},
			},
			needToCheckAppSelector: true,
			needToCheckEnvSelector: false,
			allProductionEnvsFlag:  false,
			appSelectorMap: map[string]bool{
				"Project1/App1": true,
			},
			envSelectorMap:                     map[string]bool{},
			ciPipelineIdProductionEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{},
			ciPipelineIdEnvDetailMap:           map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{},
			wantDefinitionSourceDtos: []*bean.DefinitionSourceDto{
				{
					ProjectName:           "Project1",
					AppName:               "App1",
					BranchNames:           []string{"main"},
					IsDueToLinkedPipeline: false,
					CiPipelineName:        "CIPipeline1",
					PolicyName:            "GlobalPolicy1",
				},
			},
		},
		{
			ciPipelineIdInObj: 1,
			ciPipelineId:      1,
			matchedBranchList: []string{"main"},
			ciPipelineIdNameMap: map[int]string{
				1: "CIPipeline1",
			},
			globalPolicyName: "GlobalPolicy1",
			ciPipelineIdProjectAppNameMap: map[int]*bean.PluginSourceCiPipelineAppDetailDto{
				1: {
					ProjectName: "Project1",
					AppName:     "App1",
				},
			},
			needToCheckAppSelector: true,
			needToCheckEnvSelector: false,
			allProductionEnvsFlag:  false,
			appSelectorMap: map[string]bool{
				"Project2/App2": true,
			},
			envSelectorMap:                     map[string]bool{},
			ciPipelineIdProductionEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{},
			ciPipelineIdEnvDetailMap:           map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{},
			wantDefinitionSourceDtos:           nil,
		},
		{
			ciPipelineIdInObj: 1,
			ciPipelineId:      1,
			matchedBranchList: []string{"main"},
			ciPipelineIdNameMap: map[int]string{
				1: "CIPipeline1",
			},
			globalPolicyName: "GlobalPolicy1",
			ciPipelineIdProjectAppNameMap: map[int]*bean.PluginSourceCiPipelineAppDetailDto{
				1: {
					ProjectName: "Project1",
					AppName:     "App1",
				},
			},
			needToCheckAppSelector: false,
			needToCheckEnvSelector: true,
			allProductionEnvsFlag:  true,
			appSelectorMap:         map[string]bool{},
			envSelectorMap: map[string]bool{
				"test/devtron-demo": true,
			},
			ciPipelineIdProductionEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{
				1: {
					{
						ClusterName: "test",
						EnvName:     "devtron-demo",
					},
				},
			},
			ciPipelineIdEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{
				1: {
					{
						ClusterName: "staging",
						EnvName:     "devtron-demo",
					},
				},
			},
			wantDefinitionSourceDtos: []*bean.DefinitionSourceDto{
				{
					ClusterName:                  "test",
					EnvironmentName:              "devtron-demo",
					BranchNames:                  []string{"main"},
					IsDueToProductionEnvironment: true,
					IsDueToLinkedPipeline:        false,
					CiPipelineName:               "CIPipeline1",
					PolicyName:                   "GlobalPolicy1",
				},
			},
		},
		{
			ciPipelineIdInObj: 1,
			ciPipelineId:      1,
			matchedBranchList: []string{"main"},
			ciPipelineIdNameMap: map[int]string{
				1: "CIPipeline1",
			},
			globalPolicyName: "GlobalPolicy1",
			ciPipelineIdProjectAppNameMap: map[int]*bean.PluginSourceCiPipelineAppDetailDto{
				1: {
					ProjectName: "Project1",
					AppName:     "App1",
				},
			},
			needToCheckAppSelector: false,
			needToCheckEnvSelector: true,
			allProductionEnvsFlag:  false,
			appSelectorMap:         map[string]bool{},
			envSelectorMap: map[string]bool{
				"test/devtron-cd": true,
			},
			ciPipelineIdProductionEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{
				1: {
					{
						ClusterName: "test",
						EnvName:     "devtron-demo",
					},
				},
			},
			ciPipelineIdEnvDetailMap: map[int][]*bean.PluginSourceCiPipelineEnvDetailDto{},
			wantDefinitionSourceDtos: []*bean.DefinitionSourceDto(nil),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getDefinitionSourceDtosForACIPipeline-test-%d", i+1), func(t *testing.T) {
			actualDefinitionSourceDtos := getDefinitionSourceDtosForACIPipeline(testCase.ciPipelineIdInObj, testCase.ciPipelineId,
				testCase.matchedBranchList, testCase.ciPipelineIdNameMap, testCase.globalPolicyName, testCase.ciPipelineIdProjectAppNameMap,
				testCase.needToCheckAppSelector, testCase.needToCheckEnvSelector, testCase.allProductionEnvsFlag, testCase.appSelectorMap,
				testCase.envSelectorMap, testCase.ciPipelineIdProductionEnvDetailMap, testCase.ciPipelineIdEnvDetailMap)

			assert.Equal(t, testCase.wantDefinitionSourceDtos, actualDefinitionSourceDtos)
		})
	}
}

func TestGetFilteredGlobalPolicyIdsFromSearchableFields(t *testing.T) {
	testCases := []struct {
		searchableFieldsModels []*repository.GlobalPolicySearchableField
		projectMap             map[string]bool
		clusterMap             map[string]bool
		branchValues           []string
		wantResult             []int
		wantError              bool
	}{
		{
			searchableFieldsModels: []*repository.GlobalPolicySearchableField{
				{
					SearchableKeyId: 1,
					GlobalPolicyId:  1,
					Value:           "Project1/App1",
					IsRegex:         false,
				},
			},
			projectMap: map[string]bool{
				"Project1": true,
			},
			clusterMap:   make(map[string]bool),
			branchValues: []string{"main"},
			wantResult:   []int{1},
			wantError:    false,
		},
		{
			searchableFieldsModels: []*repository.GlobalPolicySearchableField{
				{
					SearchableKeyId: 1,
					GlobalPolicyId:  1,
					Value:           "Project1/App1",
					IsRegex:         false,
				},
				{
					SearchableKeyId: 3,
					GlobalPolicyId:  1,
					Value:           "Cluster1/Env1",
					IsRegex:         false,
				},
			},
			projectMap: map[string]bool{
				"Project1": true,
			},
			clusterMap:   make(map[string]bool),
			branchValues: []string{"main"},
			wantResult:   []int{1},
			wantError:    false,
		},
		{
			searchableFieldsModels: []*repository.GlobalPolicySearchableField{
				{
					SearchableKeyId: 1,
					GlobalPolicyId:  1,
					Value:           "Project1/*",
					IsRegex:         true,
				},
			},
			projectMap: map[string]bool{
				"Project1": true,
			},
			clusterMap:   make(map[string]bool),
			branchValues: []string{"main"},
			wantResult:   []int{1},
			wantError:    false,
		},
		{
			searchableFieldsModels: []*repository.GlobalPolicySearchableField{
				{
					SearchableKeyId: 2,
					GlobalPolicyId:  2,
					Value:           "true",
					IsRegex:         false,
				},
			},
			projectMap: make(map[string]bool),
			clusterMap: map[string]bool{
				"Cluster1": true,
			},
			branchValues: []string{"main"},
			wantResult:   []int{2},
			wantError:    false,
		},
		{
			searchableFieldsModels: []*repository.GlobalPolicySearchableField{
				{
					SearchableKeyId: 3,
					GlobalPolicyId:  3,
					Value:           "Cluster1/Env1",
					IsRegex:         false,
				},
			},
			projectMap: make(map[string]bool),
			clusterMap: map[string]bool{
				"Cluster1": true,
			},
			branchValues: []string{"main"},
			wantResult:   []int{3},
			wantError:    false,
		},
		{
			searchableFieldsModels: []*repository.GlobalPolicySearchableField{
				{
					SearchableKeyId: 3,
					GlobalPolicyId:  3,
					Value:           "Cluster1/*",
					IsRegex:         true,
				},
			},
			projectMap: make(map[string]bool),
			clusterMap: map[string]bool{
				"Cluster1": true,
			},
			branchValues: []string{"main"},
			wantResult:   []int{3},
			wantError:    false,
		},
		{
			searchableFieldsModels: []*repository.GlobalPolicySearchableField{
				{
					SearchableKeyId: 4,
					GlobalPolicyId:  4,
					Value:           "^feature/.*",
					IsRegex:         true,
				},
			},
			projectMap:   make(map[string]bool),
			clusterMap:   make(map[string]bool),
			branchValues: []string{"feature/123", "main"},
			wantResult:   []int{4},
			wantError:    false,
		},
		{
			searchableFieldsModels: []*repository.GlobalPolicySearchableField{
				{
					SearchableKeyId: 4,
					GlobalPolicyId:  4,
					Value:           "a(b",
					IsRegex:         true,
				},
			},
			projectMap:   make(map[string]bool),
			clusterMap:   make(map[string]bool),
			branchValues: []string{"main"},
			wantResult:   nil,
			wantError:    true,
		},
	}
	searchableKeyIdNameMap := map[int]bean2.DevtronResourceSearchableKeyName{
		1: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME,
		2: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV,
		3: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME,
		4: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH,
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getOffendingCIPipelineIds-test-%d", i+1), func(t *testing.T) {
			globalPolicyIds, err := getFilteredGlobalPolicyIdsFromSearchableFields(testCase.searchableFieldsModels,
				testCase.projectMap, testCase.clusterMap, testCase.branchValues, searchableKeyIdNameMap)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.wantResult, globalPolicyIds)
		})
	}
}

func TestGetById(t *testing.T) {
	testCases := []struct {
		id         int
		wantPolicy *repository.GlobalPolicy
		wantDto    *bean.GlobalPolicyDto
		repoError  error
		wantError  bool
	}{
		{
			id: 1,
			wantPolicy: &repository.GlobalPolicy{
				Id:          1,
				Name:        "Test Policy",
				PolicyOf:    "PLUGIN",
				Version:     "V1",
				Description: "Test policy",
				PolicyJson:  "{\"definitions\":[{\"attributeType\":\"PLUGIN\",\"data\":{\"pluginId\":1,\"applyToStage\":\"PRE_CI\"}}],\"selectors\":{\"application\":[{\"projectName\":\"project-1\",\"appNames\":[\"app-1\",\"app-2\"]}]},\"consequences\":[]}",
				Enabled:     true,
				Deleted:     false,
			},
			wantDto: &bean.GlobalPolicyDto{
				Id:            1,
				Name:          "Test Policy",
				PolicyOf:      "PLUGIN",
				PolicyVersion: "V1",
				Description:   "Test policy",
				GlobalPolicyDetailDto: &bean.GlobalPolicyDetailDto{
					Definitions: []*bean.DefinitionDto{
						{
							AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
							Data: bean.DefinitionDataDto{
								PluginId:     1,
								ApplyToStage: bean.PLUGIN_APPLY_STAGE_PRE_CI,
							},
						},
					},
					Selectors: &bean.SelectorDto{
						ApplicationSelector: []*bean.ProjectAppDto{
							{
								ProjectName: "project-1",
								AppNames:    []string{"app-1", "app-2"},
							},
						},
						EnvironmentSelector: nil,
						BranchList:          nil,
					},
					Consequences: []*bean.ConsequenceDto{},
				},
				Enabled: true,
			},
			repoError: nil,
			wantError: false,
		},
		{
			id: 1,
			wantPolicy: &repository.GlobalPolicy{
				Id:          1,
				Name:        "Test Policy",
				PolicyOf:    "PLUGIN",
				Version:     "V1",
				Description: "Test policy",
				PolicyJson:  "",
				Enabled:     true,
				Deleted:     false,
			},
			wantDto:   nil,
			repoError: nil,
			wantError: true,
		},
		{
			id:         100,
			wantPolicy: nil,
			wantDto:    nil,
			repoError:  pg.ErrNoRows,
			wantError:  true,
		},
		// Add more test cases as needed
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("GetById-test-%d", i+1), func(t *testing.T) {
			globalPolicyService, globalPolicyRepository, _, _, _, _, _, _, _, _, _, err := getGlobalPolicyService(t)
			assert.NoError(t, err)
			globalPolicyRepository.On("GetById", testCase.id).Return(testCase.wantPolicy, testCase.repoError)
			result, err := globalPolicyService.GetById(testCase.id)
			globalPolicyRepository.AssertExpectations(t)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.wantDto, result)
		})
	}
}

func TestGetAllGlobalPolicies(t *testing.T) {
	testCases := []struct {
		name           string
		policyOf       bean.GlobalPolicyType
		policyVersion  bean.GlobalPolicyVersion
		wantPolicies   []*repository.GlobalPolicy
		wantDto        []*bean.GlobalPolicyDto
		repoError      error
		wantError      error
		wantRepoCalled bool
	}{
		{
			name:          "Valid policies",
			policyOf:      bean.GLOBAL_POLICY_TYPE_PLUGIN,
			policyVersion: bean.GLOBAL_POLICY_VERSION_V1,
			wantPolicies: []*repository.GlobalPolicy{
				{
					Id:          1,
					Name:        "Test Policy 1",
					PolicyOf:    "PLUGIN",
					Version:     "V1",
					Description: "Test policy 1",
					PolicyJson:  "{\"definitions\":[{\"attributeType\":\"PLUGIN\",\"data\":{\"pluginId\":1,\"applyToStage\":\"PRE_CI\"}}],\"selectors\":{\"application\":[{\"projectName\":\"project-1\",\"appNames\":[\"app-1\",\"app-2\"]}]},\"consequences\":[]}",
					Enabled:     true,
					Deleted:     false,
				},
				{
					Id:          2,
					Name:        "Test Policy 2",
					PolicyOf:    "PLUGIN",
					Version:     "V1",
					Description: "Test policy 2",
					PolicyJson:  "{\"definitions\":[{\"attributeType\":\"PLUGIN\",\"data\":{\"pluginId\":2,\"applyToStage\":\"POST_CI\"}}],\"selectors\":{\"application\":[{\"projectName\":\"project-2\",\"appNames\":[\"app-3\",\"app-4\"]}]},\"consequences\":[]}",
					Enabled:     true,
					Deleted:     false,
				},
			},
			wantDto: []*bean.GlobalPolicyDto{
				{
					Id:            1,
					Name:          "Test Policy 1",
					PolicyOf:      "PLUGIN",
					PolicyVersion: "V1",
					Description:   "Test policy 1",
					GlobalPolicyDetailDto: &bean.GlobalPolicyDetailDto{
						Definitions: []*bean.DefinitionDto{
							{
								AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
								Data: bean.DefinitionDataDto{
									PluginId:     1,
									ApplyToStage: bean.PLUGIN_APPLY_STAGE_PRE_CI,
								},
							},
						},
						Selectors: &bean.SelectorDto{
							ApplicationSelector: []*bean.ProjectAppDto{
								{
									ProjectName: "project-1",
									AppNames:    []string{"app-1", "app-2"},
								},
							},
							EnvironmentSelector: nil,
							BranchList:          nil,
						},
						Consequences: []*bean.ConsequenceDto{},
					},
					Enabled: true,
				},
				{
					Id:            2,
					Name:          "Test Policy 2",
					PolicyOf:      "PLUGIN",
					PolicyVersion: "V1",
					Description:   "Test policy 2",
					GlobalPolicyDetailDto: &bean.GlobalPolicyDetailDto{
						Definitions: []*bean.DefinitionDto{
							{
								AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
								Data: bean.DefinitionDataDto{
									PluginId:     2,
									ApplyToStage: bean.PLUGIN_APPLY_STAGE_POST_CI,
								},
							},
						},
						Selectors: &bean.SelectorDto{
							ApplicationSelector: []*bean.ProjectAppDto{
								{
									ProjectName: "project-2",
									AppNames:    []string{"app-3", "app-4"},
								},
							},
							EnvironmentSelector: nil,
							BranchList:          nil,
						},
						Consequences: []*bean.ConsequenceDto{},
					},
					Enabled: true,
				},
			},
			repoError:      nil,
			wantError:      nil,
			wantRepoCalled: true,
		},
		// Add more test cases as needed
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("GetAllGlobalPolicies-test-%d", i+1), func(t *testing.T) {
			globalPolicyService, globalPolicyRepository, _, _, _, _, _, _, _, _, _, err := getGlobalPolicyService(t)
			assert.NoError(t, err)
			// Set the expectations on the mock repository
			globalPolicyRepository.On("GetAllByPolicyOfAndVersion", testCase.policyOf, testCase.policyVersion).Return(testCase.wantPolicies, testCase.repoError)

			// Call the GetAllGlobalPolicies method
			result, err := globalPolicyService.GetAllGlobalPolicies(testCase.policyOf, testCase.policyVersion)

			// Assert that the mock repository was called as want
			globalPolicyRepository.AssertExpectations(t)

			// Assert the result and error
			assert.Equal(t, testCase.wantError, err)
			assert.Equal(t, testCase.wantDto, result)
		})
	}
}

func TestCreateOrUpdateGlobalPolicy(t *testing.T) {
	testCases := []struct {
		policy                       *bean.GlobalPolicyDto
		mockGetByName                *repository.GlobalPolicy
		mockGetByNameError           error
		mockCreateOrNot              bool
		mockCreateError              error
		mockUpdateError              error
		mockHistoryOfAction          bean3.HistoryOfAction
		mockHistoryCreateError       error
		mockDeleteError              error
		mockStartTransactionError    error
		mockCommitTransactionError   error
		mockRollBackTransactionError error
		mockSearchableKeyNameIdMap   map[bean2.DevtronResourceSearchableKeyName]int
		mockCreateFieldError         error
		wantError                    bool
	}{
		{
			policy: &bean.GlobalPolicyDto{
				Id:            0,
				Name:          "TestPolicy",
				Enabled:       true,
				PolicyOf:      bean.GLOBAL_POLICY_TYPE_PLUGIN,
				PolicyVersion: bean.GLOBAL_POLICY_VERSION_V1,
				GlobalPolicyDetailDto: &bean.GlobalPolicyDetailDto{
					Definitions: []*bean.DefinitionDto{
						{
							AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
							Data: bean.DefinitionDataDto{
								PluginId:     2,
								ApplyToStage: bean.PLUGIN_APPLY_STAGE_POST_CI,
							},
						},
					},
					Selectors: &bean.SelectorDto{
						ApplicationSelector: []*bean.ProjectAppDto{
							{
								ProjectName: "project-2",
								AppNames:    []string{"app-3", "app-4"},
							},
						},
						EnvironmentSelector: nil,
						BranchList:          nil,
					},
					Consequences: []*bean.ConsequenceDto{},
				},
			},
			mockHistoryOfAction:          bean3.HISTORY_OF_ACTION_CREATE,
			mockStartTransactionError:    nil,
			mockCommitTransactionError:   nil,
			mockRollBackTransactionError: fmt.Errorf("error in rolling back transaction"),
			mockCreateOrNot:              true,
			mockGetByName:                nil,
			mockCreateError:              nil,
			mockDeleteError:              nil,
			mockCreateFieldError:         nil,
			wantError:                    false,
		},
	}

	// Run test cases
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("CreateOrUpdateGlobalPolicy-test-%d", i+1), func(t *testing.T) {
			globalPolicyService, globalPolicyRepository, globalPolicySearchableFieldRepository, devtronResourceService, globalPolicyHistoryService, _, _, _, _, _, _, err := getGlobalPolicyService(t)
			assert.NoError(t, err)
			// Set up mock expectations
			globalPolicyRepository.On("GetByName", testCase.policy.Name).Return(testCase.mockGetByName, testCase.mockGetByNameError)
			if testCase.mockCreateOrNot {
				globalPolicyRepository.On("Create", mock.AnythingOfType("*repository.GlobalPolicy")).Return(testCase.mockCreateError)
			} else {
				globalPolicyRepository.On("Update", mock.AnythingOfType("repository.GlobalPolicy")).Return(testCase.mockUpdateError)
			}
			globalPolicyRepository.On("GetDbTransaction").Return(&pg.Tx{}, testCase.mockStartTransactionError)
			globalPolicyRepository.On("CommitTransaction", &pg.Tx{}).Return(testCase.mockCommitTransactionError)
			globalPolicyRepository.On("RollBackTransaction", &pg.Tx{}).Return(testCase.mockRollBackTransactionError)
			globalPolicySearchableFieldRepository.On("DeleteByPolicyId", testCase.policy.Id, &pg.Tx{}).Return(testCase.mockDeleteError)
			globalPolicyHistoryService.On("CreateHistoryEntry", mock.AnythingOfType("*repository.GlobalPolicy"), testCase.mockHistoryOfAction).Return(testCase.mockHistoryCreateError)
			devtronResourceService.On("GetAllSearchableKeyNameIdMap").Return(testCase.mockSearchableKeyNameIdMap)
			globalPolicySearchableFieldRepository.On("CreateInBatchWithTxn", mock.AnythingOfType("[]*repository.GlobalPolicySearchableField"), &pg.Tx{}).Return(testCase.mockCreateFieldError)
			err = globalPolicyService.CreateOrUpdateGlobalPolicy(testCase.policy)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteGlobalPolicy(t *testing.T) {
	testCases := []struct {
		policyId                     int
		userId                       int32
		mockGetById                  *repository.GlobalPolicy
		mockGetByIdError             error
		mockHistoryOfAction          bean3.HistoryOfAction
		mockHistoryCreateError       error
		mockMarkDeletedByIdError     error
		mockStartTransactionError    error
		mockCommitTransactionError   error
		mockRollBackTransactionError error
		mockDeleteByPolicyIdError    error
		wantError                    bool
	}{
		{
			policyId: 1,
			userId:   1,
			mockGetById: &repository.GlobalPolicy{
				Id:          1,
				Name:        "Test Policy",
				PolicyOf:    "PLUGIN",
				Version:     "V1",
				Description: "Test policy",
				PolicyJson:  "{\"definitions\":[{\"attributeType\":\"PLUGIN\",\"data\":{\"pluginId\":1,\"applyToStage\":\"PRE_CI\"}}],\"selectors\":{\"application\":[{\"projectName\":\"project-1\",\"appNames\":[\"app-1\",\"app-2\"]}]},\"consequences\":[]}",
				Enabled:     true,
				Deleted:     false,
			},
			mockHistoryOfAction:          bean3.HISTORY_OF_ACTION_DELETE,
			mockHistoryCreateError:       nil,
			mockStartTransactionError:    nil,
			mockCommitTransactionError:   nil,
			mockRollBackTransactionError: fmt.Errorf("error in rolling back transaction"),
			mockMarkDeletedByIdError:     nil,
			mockDeleteByPolicyIdError:    nil,
			wantError:                    false,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("DeleteGlobalPolicy-test-%d", i+1), func(t *testing.T) {
			globalPolicyService, globalPolicyRepository, globalPolicySearchableFieldRepository, _, globalPolicyHistoryService, _, _, _, _, _, _, err := getGlobalPolicyService(t)
			assert.NoError(t, err)
			globalPolicyRepository.On("GetById", testCase.policyId).Return(testCase.mockGetById, testCase.mockGetByIdError)
			globalPolicyRepository.On("GetDbTransaction").Return(&pg.Tx{}, testCase.mockStartTransactionError)
			globalPolicyRepository.On("CommitTransaction", &pg.Tx{}).Return(testCase.mockCommitTransactionError)
			globalPolicyRepository.On("RollBackTransaction", &pg.Tx{}).Return(testCase.mockRollBackTransactionError)
			globalPolicyRepository.On("MarkDeletedById", testCase.policyId, testCase.userId, &pg.Tx{}).Return(testCase.mockMarkDeletedByIdError)
			globalPolicySearchableFieldRepository.On("DeleteByPolicyId", testCase.policyId, &pg.Tx{}).Return(testCase.mockDeleteByPolicyIdError)
			globalPolicyHistoryService.On("CreateHistoryEntry", mock.AnythingOfType("*repository.GlobalPolicy"), testCase.mockHistoryOfAction).Return(testCase.mockHistoryCreateError)
			err = globalPolicyService.DeleteGlobalPolicy(testCase.policyId, testCase.userId)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetMandatoryPluginsForACiPipeline(t *testing.T) {
	testCases := []struct {
		ciPipelineId                      int
		appId                             int
		branchValues                      []string
		toOnlyGetBlockedStatePolicies     bool
		mockCiPipelineAppProjectObjs      []*pipelineConfig.CiPipelineAppProject
		mockCiPipelineAppProjectObjsError error
		allCiPipelineIds                  []int
		mockCiPipelineEnvClusterObjs      []*pipelineConfig.CiPipelineEnvCluster
		mockCiPipelineEnvClusterObjsError error
		mockSearchableKeyNameIdMap        map[bean2.DevtronResourceSearchableKeyName]int
		mockSearchableFieldModels         []*repository.GlobalPolicySearchableField
		mockSearchableFieldModelsError    error
		mockSearchableKeyIdNameMap        map[int]bean2.DevtronResourceSearchableKeyName
		mockGlobalPolicyIds               []int
		mockGlobalPolicies                []*repository.GlobalPolicy
		mockGlobalPoliciesError           error
		wantMandatoryPlugins              *bean.MandatoryPluginDto
		wantMandatoryPluginBlockageMap    map[string]*bean.ConsequenceDto
		wantError                         bool
	}{
		{
			ciPipelineId:                  1,
			appId:                         0,
			branchValues:                  []string{"main", "test"},
			toOnlyGetBlockedStatePolicies: false,
			mockCiPipelineAppProjectObjs: []*pipelineConfig.CiPipelineAppProject{
				{
					CiPipelineId:   1,
					CiPipelineName: "CIPipeline1",
					AppName:        "App1",
					ProjectName:    "Project1",
				},
			},
			mockCiPipelineAppProjectObjsError: nil,
			allCiPipelineIds:                  []int{1},
			mockCiPipelineEnvClusterObjs: []*pipelineConfig.CiPipelineEnvCluster{
				{
					CiPipelineId:    1,
					IsProductionEnv: true,
					EnvName:         "Prod1",
					ClusterName:     "Cluster1",
				},
				{
					CiPipelineId:    1,
					IsProductionEnv: false,
					EnvName:         "NonProd2",
					ClusterName:     "Cluster2",
				},
			},
			mockCiPipelineEnvClusterObjsError: nil,
			mockSearchableKeyNameIdMap: map[bean2.DevtronResourceSearchableKeyName]int{
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME:           1,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME:           2,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV:      3,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH:         4,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_TRIGGER_ACTION: 5,
			},
			mockSearchableFieldModels: []*repository.GlobalPolicySearchableField{
				{
					Id:              1,
					GlobalPolicyId:  1,
					SearchableKeyId: 1,
					Value:           "Project1/App1",
					IsRegex:         false,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
				{
					Id:              1,
					GlobalPolicyId:  1,
					SearchableKeyId: 3,
					Value:           "true",
					IsRegex:         false,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
			},
			mockSearchableFieldModelsError: nil,
			mockSearchableKeyIdNameMap: map[int]bean2.DevtronResourceSearchableKeyName{
				1: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME,
				2: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME,
				3: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV,
				4: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH,
				5: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_TRIGGER_ACTION,
			},
			mockGlobalPolicyIds: []int{1},
			mockGlobalPolicies: []*repository.GlobalPolicy{
				{
					Id:       1,
					Name:     "Test Policy",
					PolicyOf: bean.GLOBAL_POLICY_TYPE_PLUGIN.ToString(),
					Version:  bean.GLOBAL_POLICY_VERSION_V1.ToString(),
					PolicyJson: `{
								"definitions":[{"attributeType":"PLUGIN","data":{"pluginId":1,"applyToStage":"PRE_CI"}}],
								"selectors":{
									"application":[{"projectName":"Project1","appNames":["App1"]}],
									"environment" : { "allProductionEnvironments": true}},"consequences":[{"action":"BLOCK"}]}`,
					Enabled: true,
					Deleted: false,
				},
			},
			mockGlobalPoliciesError: nil,
			wantMandatoryPlugins: &bean.MandatoryPluginDto{
				Definitions: []*bean.MandatoryPluginDefinitionDto{
					{
						DefinitionDto: &bean.DefinitionDto{
							AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
							Data: bean.DefinitionDataDto{
								PluginId:     1,
								ApplyToStage: bean.PLUGIN_APPLY_STAGE_PRE_CI,
							},
						},
						DefinitionSources: []*bean.DefinitionSourceDto{
							{
								ProjectName:                  "Project1",
								AppName:                      "App1",
								ClusterName:                  "Cluster1",
								EnvironmentName:              "Prod1",
								BranchNames:                  []string{},
								IsDueToProductionEnvironment: true,
								IsDueToLinkedPipeline:        false,
								CiPipelineName:               "CIPipeline1",
								PolicyName:                   "Test Policy",
							},
						},
					},
				},
			},
			wantMandatoryPluginBlockageMap: map[string]*bean.ConsequenceDto{
				"1/PRE_CI": {
					Action: bean.CONSEQUENCE_ACTION_BLOCK,
				},
			},
			wantError: false,
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("GetMandatoryPluginsForACiPipeline-test-%d", i+1), func(t *testing.T) {
			globalPolicyService, globalPolicyRepository, globalPolicySearchableFieldRepository, devtronResourceService, _, ciPipelineRepository, _, _, _, _, _, err := getGlobalPolicyService(t)
			assert.NoError(t, err)

			ciPipelineRepository.On("GetAppAndProjectNameForParentAndAllLinkedCI", testCase.ciPipelineId).Return(testCase.mockCiPipelineAppProjectObjs, testCase.mockCiPipelineAppProjectObjsError)
			ciPipelineRepository.On("GetAllCDsEnvAndClusterNameByCiPipelineIds", testCase.allCiPipelineIds).Return(testCase.mockCiPipelineEnvClusterObjs, testCase.mockCiPipelineEnvClusterObjsError)
			devtronResourceService.On("GetAllSearchableKeyNameIdMap").Return(testCase.mockSearchableKeyNameIdMap)
			globalPolicySearchableFieldRepository.On("GetSearchableFields", mock.Anything, mock.Anything).Return(testCase.mockSearchableFieldModels, testCase.mockSearchableFieldModelsError)
			devtronResourceService.On("GetAllSearchableKeyIdNameMap").Return(testCase.mockSearchableKeyIdNameMap)
			globalPolicyRepository.On("GetEnabledPoliciesByIds", mock.Anything).Return(testCase.mockGlobalPolicies, testCase.mockGlobalPoliciesError)
			mandatoryPlugins, mandatoryBlockageMap, err := globalPolicyService.GetMandatoryPluginsForACiPipeline(testCase.ciPipelineId, testCase.appId, testCase.branchValues, testCase.toOnlyGetBlockedStatePolicies)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.wantMandatoryPlugins, mandatoryPlugins)
			assert.Equal(t, testCase.wantMandatoryPluginBlockageMap, mandatoryBlockageMap)
		})
	}
}

func TestGetBlockageStateForACIPipelineTrigger(t *testing.T) {
	testCases := []struct {
		ciPipelineId                      int
		parentCiPipelineId                int
		branchValues                      []string
		toOnlyGetBlockedStatePolicies     bool
		mockCiPipelineAppProjectObjs      []*pipelineConfig.CiPipelineAppProject
		mockCiPipelineAppProjectObjsError error
		allCiPipelineIds                  []int
		mockCiPipelineEnvClusterObjs      []*pipelineConfig.CiPipelineEnvCluster
		mockCiPipelineEnvClusterObjsError error
		mockSearchableKeyNameIdMap        map[bean2.DevtronResourceSearchableKeyName]int
		mockSearchableFieldModels         []*repository.GlobalPolicySearchableField
		mockSearchableFieldModelsError    error
		mockSearchableKeyIdNameMap        map[int]bean2.DevtronResourceSearchableKeyName
		mockGlobalPolicyIds               []int
		mockGlobalPolicies                []*repository.GlobalPolicy
		mockGlobalPoliciesError           error
		mockConfiguredPlugins             []*repository3.PipelineStageStep
		mockGetConfiguredPluginsError     error
		wantIsOffendingMandatoryPlugin    bool
		wantIsCIPipelineTriggerBlocked    bool
		wantBlockageStateFinal            *bean.ConsequenceDto
		wantError                         bool
	}{
		{
			ciPipelineId:                  1,
			parentCiPipelineId:            0,
			branchValues:                  []string{"main", "test"},
			toOnlyGetBlockedStatePolicies: false,
			mockCiPipelineAppProjectObjs: []*pipelineConfig.CiPipelineAppProject{
				{
					CiPipelineId:   1,
					CiPipelineName: "CIPipeline1",
					AppName:        "App1",
					ProjectName:    "Project1",
				},
			},
			mockCiPipelineAppProjectObjsError: nil,
			allCiPipelineIds:                  []int{1},
			mockCiPipelineEnvClusterObjs: []*pipelineConfig.CiPipelineEnvCluster{
				{
					CiPipelineId:    1,
					IsProductionEnv: true,
					EnvName:         "Prod1",
					ClusterName:     "Cluster1",
				},
				{
					CiPipelineId:    1,
					IsProductionEnv: false,
					EnvName:         "NonProd2",
					ClusterName:     "Cluster2",
				},
			},
			mockCiPipelineEnvClusterObjsError: nil,
			mockSearchableKeyNameIdMap: map[bean2.DevtronResourceSearchableKeyName]int{
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME:           1,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME:           2,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV:      3,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH:         4,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_TRIGGER_ACTION: 5,
			},
			mockSearchableFieldModels: []*repository.GlobalPolicySearchableField{
				{
					Id:              1,
					GlobalPolicyId:  1,
					SearchableKeyId: 1,
					Value:           "Project1/App1",
					IsRegex:         false,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
				{
					Id:              1,
					GlobalPolicyId:  1,
					SearchableKeyId: 3,
					Value:           "true",
					IsRegex:         false,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
			},
			mockSearchableFieldModelsError: nil,
			mockSearchableKeyIdNameMap: map[int]bean2.DevtronResourceSearchableKeyName{
				1: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME,
				2: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME,
				3: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV,
				4: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH,
				5: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_TRIGGER_ACTION,
			},
			mockGlobalPolicyIds: []int{1},
			mockGlobalPolicies: []*repository.GlobalPolicy{
				{
					Id:       1,
					Name:     "Test Policy",
					PolicyOf: bean.GLOBAL_POLICY_TYPE_PLUGIN.ToString(),
					Version:  bean.GLOBAL_POLICY_VERSION_V1.ToString(),
					PolicyJson: `{
								"definitions":[{"attributeType":"PLUGIN","data":{"pluginId":1,"applyToStage":"PRE_CI"}}],
								"selectors":{
									"application":[{"projectName":"Project1","appNames":["App1"]}],
									"environment" : { "allProductionEnvironments": true}},"consequences":[{"action":"BLOCK"}]}`,
					Enabled: true,
					Deleted: false,
				},
			},
			mockGlobalPoliciesError: nil,
			mockConfiguredPlugins: []*repository3.PipelineStageStep{
				{
					RefPluginId: 2,
					PipelineStage: &repository3.PipelineStage{
						Type: repository3.PIPELINE_STAGE_TYPE_PRE_CI,
					},
				},
			},
			mockGetConfiguredPluginsError:  nil,
			wantIsOffendingMandatoryPlugin: true,
			wantIsCIPipelineTriggerBlocked: true,
			wantBlockageStateFinal:         &bean.ConsequenceDto{Action: bean.CONSEQUENCE_ACTION_BLOCK},

			wantError: false,
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("GetBlockageStateForACIPipelineTrigger-test-%d", i+1), func(t *testing.T) {
			globalPolicyService, globalPolicyRepository, globalPolicySearchableFieldRepository, devtronResourceService, _, ciPipelineRepository, _, pipelineStageRepository, _, _, _, err := getGlobalPolicyService(t)
			assert.NoError(t, err)

			ciPipelineRepository.On("GetAppAndProjectNameForParentAndAllLinkedCI", testCase.ciPipelineId).Return(testCase.mockCiPipelineAppProjectObjs, testCase.mockCiPipelineAppProjectObjsError)
			ciPipelineRepository.On("GetAllCDsEnvAndClusterNameByCiPipelineIds", testCase.allCiPipelineIds).Return(testCase.mockCiPipelineEnvClusterObjs, testCase.mockCiPipelineEnvClusterObjsError)
			devtronResourceService.On("GetAllSearchableKeyNameIdMap").Return(testCase.mockSearchableKeyNameIdMap)
			globalPolicySearchableFieldRepository.On("GetSearchableFields", mock.Anything, mock.Anything).Return(testCase.mockSearchableFieldModels, testCase.mockSearchableFieldModelsError)
			devtronResourceService.On("GetAllSearchableKeyIdNameMap").Return(testCase.mockSearchableKeyIdNameMap)
			globalPolicyRepository.On("GetEnabledPoliciesByIds", mock.Anything).Return(testCase.mockGlobalPolicies, testCase.mockGlobalPoliciesError)
			pipelineStageRepository.On("GetConfiguredPluginsForCIPipelines", mock.Anything).Return(testCase.mockConfiguredPlugins, testCase.mockGetConfiguredPluginsError)
			isOffendingMandatoryPlugin, isCIPipelineTriggerBlocked, blockageStateFinal, err := globalPolicyService.GetBlockageStateForACIPipelineTrigger(testCase.ciPipelineId, testCase.parentCiPipelineId, testCase.branchValues, testCase.toOnlyGetBlockedStatePolicies)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.wantIsOffendingMandatoryPlugin, isOffendingMandatoryPlugin)
			assert.Equal(t, testCase.wantIsCIPipelineTriggerBlocked, isCIPipelineTriggerBlocked)
			assert.Equal(t, testCase.wantBlockageStateFinal, blockageStateFinal)
		})
	}
}

func TestGetOnlyBlockageStateForCiPipeline(t *testing.T) {
	testCases := []struct {
		ciPipelineId                      int
		mockCiPipeline                    *pipelineConfig.CiPipeline
		mockCiPipelineError               error
		branchValues                      []string
		toOnlyGetBlockedStatePolicies     bool
		mockCiPipelineAppProjectObjs      []*pipelineConfig.CiPipelineAppProject
		mockCiPipelineAppProjectObjsError error
		allCiPipelineIds                  []int
		mockCiPipelineEnvClusterObjs      []*pipelineConfig.CiPipelineEnvCluster
		mockCiPipelineEnvClusterObjsError error
		mockSearchableKeyNameIdMap        map[bean2.DevtronResourceSearchableKeyName]int
		mockSearchableFieldModels         []*repository.GlobalPolicySearchableField
		mockSearchableFieldModelsError    error
		mockSearchableKeyIdNameMap        map[int]bean2.DevtronResourceSearchableKeyName
		mockGlobalPolicyIds               []int
		mockGlobalPolicies                []*repository.GlobalPolicy
		mockGlobalPoliciesError           error
		mockConfiguredPlugins             []*repository3.PipelineStageStep
		mockGetConfiguredPluginsError     error
		wantBlockageStateFinal            *bean.ConsequenceDto
		wantError                         bool
	}{
		{
			ciPipelineId: 1,
			mockCiPipeline: &pipelineConfig.CiPipeline{
				Id:               1,
				ParentCiPipeline: 0,
			},
			branchValues:                  []string{"main", "test"},
			toOnlyGetBlockedStatePolicies: false,
			mockCiPipelineAppProjectObjs: []*pipelineConfig.CiPipelineAppProject{
				{
					CiPipelineId:   1,
					CiPipelineName: "CIPipeline1",
					AppName:        "App1",
					ProjectName:    "Project1",
				},
			},
			mockCiPipelineAppProjectObjsError: nil,
			allCiPipelineIds:                  []int{1},
			mockCiPipelineEnvClusterObjs: []*pipelineConfig.CiPipelineEnvCluster{
				{
					CiPipelineId:    1,
					IsProductionEnv: true,
					EnvName:         "Prod1",
					ClusterName:     "Cluster1",
				},
				{
					CiPipelineId:    1,
					IsProductionEnv: false,
					EnvName:         "NonProd2",
					ClusterName:     "Cluster2",
				},
			},
			mockCiPipelineEnvClusterObjsError: nil,
			mockSearchableKeyNameIdMap: map[bean2.DevtronResourceSearchableKeyName]int{
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME:           1,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME:           2,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV:      3,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH:         4,
				bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_TRIGGER_ACTION: 5,
			},
			mockSearchableFieldModels: []*repository.GlobalPolicySearchableField{
				{
					Id:              1,
					GlobalPolicyId:  1,
					SearchableKeyId: 1,
					Value:           "Project1/App1",
					IsRegex:         false,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
				{
					Id:              1,
					GlobalPolicyId:  1,
					SearchableKeyId: 3,
					Value:           "true",
					IsRegex:         false,
					PolicyComponent: bean.GLOBAL_POLICY_COMPONENT_SELECTOR,
				},
			},
			mockSearchableFieldModelsError: nil,
			mockSearchableKeyIdNameMap: map[int]bean2.DevtronResourceSearchableKeyName{
				1: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME,
				2: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME,
				3: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV,
				4: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH,
				5: bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_TRIGGER_ACTION,
			},
			mockGlobalPolicyIds: []int{1},
			mockGlobalPolicies: []*repository.GlobalPolicy{
				{
					Id:       1,
					Name:     "Test Policy",
					PolicyOf: bean.GLOBAL_POLICY_TYPE_PLUGIN.ToString(),
					Version:  bean.GLOBAL_POLICY_VERSION_V1.ToString(),
					PolicyJson: `{
								"definitions":[{"attributeType":"PLUGIN","data":{"pluginId":1,"applyToStage":"PRE_CI"}}],
								"selectors":{
									"application":[{"projectName":"Project1","appNames":["App1"]}],
									"environment" : { "allProductionEnvironments": true}},"consequences":[{"action":"BLOCK"}]}`,
					Enabled: true,
					Deleted: false,
				},
			},
			mockGlobalPoliciesError: nil,
			mockConfiguredPlugins: []*repository3.PipelineStageStep{
				{
					RefPluginId: 2,
					PipelineStage: &repository3.PipelineStage{
						Type: repository3.PIPELINE_STAGE_TYPE_PRE_CI,
					},
				},
			},
			mockGetConfiguredPluginsError: nil,
			wantBlockageStateFinal:        &bean.ConsequenceDto{Action: bean.CONSEQUENCE_ACTION_BLOCK},
			wantError:                     false,
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("GetOnlyBlockageStateForCiPipeline-test-%d", i+1), func(t *testing.T) {
			globalPolicyService, globalPolicyRepository, globalPolicySearchableFieldRepository, devtronResourceService, _, ciPipelineRepository, _, pipelineStageRepository, _, _, _, err := getGlobalPolicyService(t)
			assert.NoError(t, err)
			ciPipelineRepository.On("FindById", testCase.ciPipelineId).
				Return(testCase.mockCiPipeline, testCase.mockCiPipelineError)
			ciPipelineRepository.On("GetAppAndProjectNameForParentAndAllLinkedCI", testCase.ciPipelineId).Return(testCase.mockCiPipelineAppProjectObjs, testCase.mockCiPipelineAppProjectObjsError)
			ciPipelineRepository.On("GetAllCDsEnvAndClusterNameByCiPipelineIds", testCase.allCiPipelineIds).Return(testCase.mockCiPipelineEnvClusterObjs, testCase.mockCiPipelineEnvClusterObjsError)
			devtronResourceService.On("GetAllSearchableKeyNameIdMap").Return(testCase.mockSearchableKeyNameIdMap)
			globalPolicySearchableFieldRepository.On("GetSearchableFields", mock.Anything, mock.Anything).Return(testCase.mockSearchableFieldModels, testCase.mockSearchableFieldModelsError)
			devtronResourceService.On("GetAllSearchableKeyIdNameMap").Return(testCase.mockSearchableKeyIdNameMap)
			globalPolicyRepository.On("GetEnabledPoliciesByIds", mock.Anything).Return(testCase.mockGlobalPolicies, testCase.mockGlobalPoliciesError)
			pipelineStageRepository.On("GetConfiguredPluginsForCIPipelines", mock.Anything).Return(testCase.mockConfiguredPlugins, testCase.mockGetConfiguredPluginsError)
			_, _, blockageStateFinal, err := globalPolicyService.GetOnlyBlockageStateForCiPipeline(testCase.ciPipelineId, testCase.branchValues)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.wantBlockageStateFinal, blockageStateFinal)
		})
	}
}

func TestGetAppIdGitMaterialMap(t *testing.T) {
	testCases := []struct {
		appIds                []int
		mockGitMaterials      []*pipelineConfig.GitMaterial
		mockGitMaterialsError error
		wantError             bool
	}{
		{
			appIds: []int{1, 2, 3},
			mockGitMaterials: []*pipelineConfig.GitMaterial{
				{
					Id:    1,
					AppId: 1,
					Name:  "1-devtron",
				},
				{
					Id:    2,
					AppId: 1,
					Name:  "2-devtron",
				},
				{
					Id:    3,
					AppId: 2,
					Name:  "3-devtron",
				},
				{
					Id:    4,
					AppId: 3,
					Name:  "4-devtron",
				},
			},
			mockGitMaterialsError: nil,
			wantError:             false,
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getAppIdGitMaterialMap-test-%d", i+1), func(t *testing.T) {
			globalPolicyService, _, _, _, _, _, gitMaterialRepository, _, _, _, _, err := getGlobalPolicyService(t)
			assert.NoError(t, err)
			gitMaterialRepository.On("FindByAppIds", testCase.appIds).Return(testCase.mockGitMaterials, testCase.mockGitMaterialsError)
			appIdGitMaterialMap, err := globalPolicyService.getAppIdGitMaterialMap(testCase.appIds)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, len(testCase.appIds), len(appIdGitMaterialMap))
		})
	}
}

func TestGetCIPipelineMaterialsByFilteredCIPipelineIds(t *testing.T) {
	testCases := []struct {
		ciPipelinesToBeFiltered      []int
		mockCiPipelineMaterials      []*pipelineConfig.CiPipelineMaterial
		mockCiPipelineMaterialsError error
		wantCiPipelineMaterials      []*pipelineConfig.CiPipelineMaterial
		wantError                    bool
	}{
		{
			ciPipelinesToBeFiltered: []int{1, 2},
			mockCiPipelineMaterials: []*pipelineConfig.CiPipelineMaterial{
				{
					Id:           1,
					CiPipelineId: 1,
					Value:        "main",
				},
				{
					Id:               2,
					CiPipelineId:     2,
					ParentCiPipeline: 1,
					Value:            "main",
				},
			},
			mockCiPipelineMaterialsError: nil,
			wantCiPipelineMaterials: []*pipelineConfig.CiPipelineMaterial{
				{
					Id:           1,
					CiPipelineId: 1,
					Value:        "main",
				},
				{
					Id:               2,
					CiPipelineId:     2,
					ParentCiPipeline: 1,
					Value:            "main",
				},
			},
			wantError: false,
		},
		{
			ciPipelinesToBeFiltered: []int{},
			mockCiPipelineMaterials: []*pipelineConfig.CiPipelineMaterial{
				{
					Id:           1,
					CiPipelineId: 1,
					Value:        "main",
				},
				{
					Id:               2,
					CiPipelineId:     2,
					ParentCiPipeline: 1,
					Value:            "main",
				},
				{
					Id:           3,
					CiPipelineId: 3,
					Value:        "test",
				},
			},
			mockCiPipelineMaterialsError: nil,
			wantCiPipelineMaterials: []*pipelineConfig.CiPipelineMaterial{
				{
					Id:           1,
					CiPipelineId: 1,
					Value:        "main",
				},
				{
					Id:               2,
					CiPipelineId:     2,
					ParentCiPipeline: 1,
					Value:            "main",
				},
				{
					Id:           3,
					CiPipelineId: 3,
					Value:        "test",
				},
			},
			wantError: false,
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getCIPipelineMaterialsByFilteredCIPipelineIds-test-%d", i+1), func(t *testing.T) {
			globalPolicyService, _, _, _, _, _, _, _, ciPipelineMaterialRepository, _, _, err := getGlobalPolicyService(t)
			assert.NoError(t, err)
			if len(testCase.ciPipelinesToBeFiltered) == 0 {
				ciPipelineMaterialRepository.On("GetAllExceptUnsetRegexBranch").Return(testCase.mockCiPipelineMaterials, testCase.mockCiPipelineMaterialsError)
			} else {
				ciPipelineMaterialRepository.On("GetByCiPipelineIdsExceptUnsetRegexBranch", testCase.ciPipelinesToBeFiltered).Return(testCase.mockCiPipelineMaterials, testCase.mockCiPipelineMaterialsError)
			}
			ciPipelineMaterials, err := globalPolicyService.getCIPipelineMaterialsByFilteredCIPipelineIds(testCase.ciPipelinesToBeFiltered)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.wantCiPipelineMaterials, ciPipelineMaterials)
		})
	}
}

func TestGetCIPipelinesForConfiguredPluginsForBranchSelector(t *testing.T) {
	testCases := []struct {
		ciPipelinesToBeFiltered             []int
		branchList                          []*bean.BranchDto
		mockCiPipelineMaterials             []*pipelineConfig.CiPipelineMaterial
		mockCiPipelineMaterialsError        error
		wantCiPipelinesForConfiguredPlugins []int
		wantCiPipelineParentChildMap        map[int][]int
		wantCiPipelineMaterialMap           map[int][]*pipelineConfig.CiPipelineMaterial
		wantError                           bool
	}{
		{
			ciPipelinesToBeFiltered: []int{1, 2},
			branchList: []*bean.BranchDto{
				{
					BranchValueType: bean2.VALUE_TYPE_FIXED,
					Value:           "main",
				},
				{
					BranchValueType: bean2.VALUE_TYPE_REGEX,
					Value:           "^t.*",
				},
				{
					BranchValueType: bean2.VALUE_TYPE_REGEX,
					Value:           "^mai.*",
				},
			},
			mockCiPipelineMaterials: []*pipelineConfig.CiPipelineMaterial{
				{
					Id:           1,
					CiPipelineId: 1,
					Value:        "main",
				},
				{
					Id:               2,
					CiPipelineId:     3,
					ParentCiPipeline: 2,
					Value:            "test",
				},
				{
					Id:           3,
					CiPipelineId: 4,
					Value:        "master",
				},
				{
					Id:           4,
					CiPipelineId: 1,
					Value:        "test",
				},
			},
			mockCiPipelineMaterialsError:        nil,
			wantCiPipelinesForConfiguredPlugins: []int{1, 2},
			wantCiPipelineParentChildMap: map[int][]int{
				1: {1, 1},
				2: {3},
			},
			wantCiPipelineMaterialMap: map[int][]*pipelineConfig.CiPipelineMaterial{
				1: {
					{
						Id:           1,
						CiPipelineId: 1,
						Value:        "main",
					},
					{
						Id:           4,
						CiPipelineId: 1,
						Value:        "test",
					},
				},

				3: {
					{
						Id:               2,
						CiPipelineId:     3,
						ParentCiPipeline: 2,
						Value:            "test",
					},
				},
				4: {
					{
						Id:           3,
						CiPipelineId: 4,
						Value:        "master",
					},
				},
			},
			wantError: false,
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getCIPipelinesForConfiguredPluginsForBranchSelector-test-%d", i+1), func(t *testing.T) {
			globalPolicyService, _, _, _, _, _, _, _, ciPipelineMaterialRepository, _, _, err := getGlobalPolicyService(t)
			assert.NoError(t, err)
			ciPipelineMaterialRepository.On("GetByCiPipelineIdsExceptUnsetRegexBranch", testCase.ciPipelinesToBeFiltered).Return(testCase.mockCiPipelineMaterials, testCase.mockCiPipelineMaterialsError)
			ciPipelinesForConfiguredPlugins, ciPipelineParentChildMap, ciPipelineMaterialMap, err := globalPolicyService.getCIPipelinesForConfiguredPluginsForBranchSelector(testCase.ciPipelinesToBeFiltered, testCase.branchList)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.wantCiPipelinesForConfiguredPlugins, ciPipelinesForConfiguredPlugins)
			assert.Equal(t, testCase.wantCiPipelineParentChildMap, ciPipelineParentChildMap)
			assert.Equal(t, testCase.wantCiPipelineMaterialMap, ciPipelineMaterialMap)
		})
	}
}

func TestGetFilteredCIPipelinesForEnvSelector(t *testing.T) {
	testCases := []struct {
		ciPipelinesToBeFiltered               []int
		isProductionEnvFlag                   bool
		allClusters                           []string
		clusterEnvNameMap                     map[string]bool
		mockCiPipelineClusterEnvNameObjs      []*pipelineConfig.CiPipelineEnvCluster
		mockCiPipelineClusterEnvNameObjsError error
		wantFilteredCiPipelines               []int
		wantError                             bool
	}{
		{
			ciPipelinesToBeFiltered: []int{1, 2},
			isProductionEnvFlag:     true,
			allClusters:             nil,
			clusterEnvNameMap:       nil,
			mockCiPipelineClusterEnvNameObjs: []*pipelineConfig.CiPipelineEnvCluster{
				{
					CiPipelineId:    1,
					IsProductionEnv: false,
					EnvName:         "env1",
					ClusterName:     "cluster1",
				},
				{
					CiPipelineId:    1,
					IsProductionEnv: true,
					EnvName:         "prodEnv1",
					ClusterName:     "cluster1",
				},
				{
					CiPipelineId:    1,
					IsProductionEnv: true,
					EnvName:         "prodEnv2",
					ClusterName:     "cluster1",
				},
				{
					CiPipelineId:    2,
					IsProductionEnv: false,
					EnvName:         "env1",
					ClusterName:     "cluster1",
				},
				{
					CiPipelineId:    2,
					IsProductionEnv: false,
					EnvName:         "env2",
					ClusterName:     "cluster1",
				},
			},
			mockCiPipelineClusterEnvNameObjsError: nil,
			wantFilteredCiPipelines:               []int{1, 1},
			wantError:                             false,
		},
		{
			ciPipelinesToBeFiltered: []int{1, 2},
			isProductionEnvFlag:     false,
			allClusters:             []string{"cluster1", "cluster2"},
			clusterEnvNameMap: map[string]bool{
				"cluster1/*":    true,
				"cluster2/env3": true,
			},
			mockCiPipelineClusterEnvNameObjs: []*pipelineConfig.CiPipelineEnvCluster{
				{
					CiPipelineId:    1,
					IsProductionEnv: false,
					EnvName:         "env1",
					ClusterName:     "cluster1",
				},
				{
					CiPipelineId:    1,
					IsProductionEnv: true,
					EnvName:         "prodEnv1",
					ClusterName:     "cluster1",
				},
				{
					CiPipelineId:    2,
					IsProductionEnv: false,
					EnvName:         "env3",
					ClusterName:     "cluster2",
				},
				{
					CiPipelineId:    2,
					IsProductionEnv: false,
					EnvName:         "env4",
					ClusterName:     "cluster2",
				},
			},
			mockCiPipelineClusterEnvNameObjsError: nil,
			wantFilteredCiPipelines:               []int{1, 1, 2},
			wantError:                             false,
		},
		{
			ciPipelinesToBeFiltered:               []int{1, 2},
			isProductionEnvFlag:                   false,
			allClusters:                           nil,
			clusterEnvNameMap:                     nil,
			mockCiPipelineClusterEnvNameObjs:      []*pipelineConfig.CiPipelineEnvCluster{},
			mockCiPipelineClusterEnvNameObjsError: nil,
			wantFilteredCiPipelines:               []int{},
			wantError:                             false,
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getFilteredCIPipelinesForEnvSelector-test-%d", i+1), func(t *testing.T) {
			globalPolicyService, _, _, _, _, ciPipelineRepository, _, _, _, _, _, err := getGlobalPolicyService(t)
			assert.NoError(t, err)
			if testCase.isProductionEnvFlag {
				ciPipelineRepository.On("GetAllCIsClusterAndEnvForAllProductionEnvCD", testCase.ciPipelinesToBeFiltered).
					Return(testCase.mockCiPipelineClusterEnvNameObjs, testCase.mockCiPipelineClusterEnvNameObjsError)
			} else if len(testCase.allClusters) > 0 {
				ciPipelineRepository.On("GetAllCIsClusterAndEnvByCDClusterNames", testCase.allClusters, testCase.ciPipelinesToBeFiltered).
					Return(testCase.mockCiPipelineClusterEnvNameObjs, testCase.mockCiPipelineClusterEnvNameObjsError)
			}
			filteredCiPipelines, err := globalPolicyService.getFilteredCIPipelinesForEnvSelector(testCase.ciPipelinesToBeFiltered, testCase.isProductionEnvFlag, testCase.allClusters, testCase.clusterEnvNameMap)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.wantFilteredCiPipelines, filteredCiPipelines)
		})
	}
}

func TestGetCIPipelinesForConfiguredPlugins(t *testing.T) {
	testCases := []struct {
		globalPolicyDetailDto                 bean.GlobalPolicyDetailDto
		wantProjectAppMock                    bool
		mockCiPipelineProjectAppNameObjs      []*pipelineConfig.CiPipelineAppProject
		mockCiPipelineProjectAppNameObjsError error
		mockCiPipelinesToBeFiltered           []int
		mockAllClusters                       []string
		mockCiPipelineClusterEnvNameObjs      []*pipelineConfig.CiPipelineEnvCluster
		mockCiPipelineClusterEnvNameObjsError error
		wantBranchListMock                    bool
		wantFilteredCiPipelines               []int
		mockCiPipelineMaterials               []*pipelineConfig.CiPipelineMaterial
		mockCiPipelineMaterialsError          error
		wantError                             bool
	}{
		{
			globalPolicyDetailDto: bean.GlobalPolicyDetailDto{
				Selectors: &bean.SelectorDto{
					ApplicationSelector: []*bean.ProjectAppDto{
						{
							ProjectName: "Project1",
						},
					},
					EnvironmentSelector: &bean.EnvironmentSelectorDto{
						ClusterEnvList: []*bean.ClusterEnvDto{
							{
								ClusterName: "default-cluster",
							},
						},
					},
					BranchList: []*bean.BranchDto{
						{
							BranchValueType: bean2.VALUE_TYPE_FIXED,
							Value:           "main",
						},
					},
				},
			},
			wantProjectAppMock: true,
			mockCiPipelineProjectAppNameObjs: []*pipelineConfig.CiPipelineAppProject{
				{
					CiPipelineId: 1,
					AppName:      "App1",
					ProjectName:  "Project1",
				},
				{
					CiPipelineId: 2,
					AppName:      "App1",
					ProjectName:  "Project1",
				},
				{
					CiPipelineId: 3,
					AppName:      "App2",
					ProjectName:  "Project2",
				},
			},
			mockCiPipelineProjectAppNameObjsError: nil,
			mockCiPipelinesToBeFiltered:           []int{1, 2},
			mockAllClusters:                       []string{"default-cluster"},
			mockCiPipelineClusterEnvNameObjs: []*pipelineConfig.CiPipelineEnvCluster{
				{
					CiPipelineId:    1,
					IsProductionEnv: false,
					EnvName:         "env1",
					ClusterName:     "default-cluster",
				},
				{
					CiPipelineId:    1,
					IsProductionEnv: true,
					EnvName:         "prodEnv1",
					ClusterName:     "cluster1",
				},
				{
					CiPipelineId:    1,
					IsProductionEnv: true,
					EnvName:         "prodEnv2",
					ClusterName:     "cluster1",
				},
				{
					CiPipelineId:    2,
					IsProductionEnv: false,
					EnvName:         "env1",
					ClusterName:     "default-cluster",
				},
				{
					CiPipelineId:    2,
					IsProductionEnv: false,
					EnvName:         "env2",
					ClusterName:     "cluster1",
				},
			},
			mockCiPipelineClusterEnvNameObjsError: nil,
			wantFilteredCiPipelines:               []int{1},
			wantBranchListMock:                    true,
			mockCiPipelineMaterials: []*pipelineConfig.CiPipelineMaterial{
				{
					Id:           1,
					CiPipelineId: 1,
					Value:        "main",
				},
				{
					Id:               2,
					CiPipelineId:     2,
					ParentCiPipeline: 1,
					Value:            "main",
				},
			},
			mockCiPipelineMaterialsError: nil,
			wantError:                    false,
		},
		{
			globalPolicyDetailDto: bean.GlobalPolicyDetailDto{
				Selectors: &bean.SelectorDto{
					ApplicationSelector: []*bean.ProjectAppDto{
						{
							ProjectName: "Project1",
						},
					},
					EnvironmentSelector: &bean.EnvironmentSelectorDto{
						ClusterEnvList: []*bean.ClusterEnvDto{
							{
								ClusterName: "default-cluster",
							},
						},
					},
				},
			},
			wantProjectAppMock: true,
			mockCiPipelineProjectAppNameObjs: []*pipelineConfig.CiPipelineAppProject{
				{
					CiPipelineId: 1,
					AppName:      "App1",
					ProjectName:  "Project1",
				},
				{
					CiPipelineId: 2,
					AppName:      "App1",
					ProjectName:  "Project1",
				},
				{
					CiPipelineId: 3,
					AppName:      "App2",
					ProjectName:  "Project2",
				},
			},
			mockCiPipelineProjectAppNameObjsError: nil,
			mockCiPipelinesToBeFiltered:           []int{1, 2},
			mockAllClusters:                       []string{"default-cluster"},
			mockCiPipelineClusterEnvNameObjs: []*pipelineConfig.CiPipelineEnvCluster{
				{
					CiPipelineId:    1,
					IsProductionEnv: false,
					EnvName:         "env1",
					ClusterName:     "default-cluster",
				},
				{
					CiPipelineId:    1,
					IsProductionEnv: true,
					EnvName:         "prodEnv1",
					ClusterName:     "cluster1",
				},
				{
					CiPipelineId:    1,
					IsProductionEnv: true,
					EnvName:         "prodEnv2",
					ClusterName:     "cluster1",
				},
				{
					CiPipelineId:    2,
					IsProductionEnv: false,
					EnvName:         "env1",
					ClusterName:     "default-cluster",
				},
				{
					CiPipelineId:    2,
					IsProductionEnv: false,
					EnvName:         "env2",
					ClusterName:     "cluster1",
				},
			},
			mockCiPipelineClusterEnvNameObjsError: nil,
			wantFilteredCiPipelines:               []int{1, 2},
			wantBranchListMock:                    false,
			mockCiPipelineMaterials: []*pipelineConfig.CiPipelineMaterial{
				{
					Id:           1,
					CiPipelineId: 1,
					Value:        "main",
				},
				{
					Id:               2,
					CiPipelineId:     2,
					ParentCiPipeline: 1,
					Value:            "main",
				},
			},
			mockCiPipelineMaterialsError: nil,
			wantError:                    false,
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("getCIPipelinesForConfiguredPlugins-test-%d", i+1), func(t *testing.T) {
			globalPolicyService, _, _, _, _, ciPipelineRepository, _, _, ciPipelineMaterialRepository, _, _, err := getGlobalPolicyService(t)
			assert.NoError(t, err)
			if testCase.wantProjectAppMock {
				ciPipelineRepository.On("GetAllCIAppAndProjectByProjectNames", mock.Anything).
					Return(testCase.mockCiPipelineProjectAppNameObjs, testCase.mockCiPipelineProjectAppNameObjsError)
			}
			if len(testCase.mockAllClusters) > 0 {
				ciPipelineRepository.On("GetAllCIsClusterAndEnvByCDClusterNames", testCase.mockAllClusters, testCase.mockCiPipelinesToBeFiltered).
					Return(testCase.mockCiPipelineClusterEnvNameObjs, testCase.mockCiPipelineClusterEnvNameObjsError)
			}
			if testCase.wantBranchListMock {
				ciPipelineMaterialRepository.On("GetByCiPipelineIdsExceptUnsetRegexBranch", testCase.mockCiPipelinesToBeFiltered).
					Return(testCase.mockCiPipelineMaterials, testCase.mockCiPipelineMaterialsError)
			} else {
				ciPipelineMaterialRepository.On("FindByCiPipelineIdsIn", mock.Anything).
					Return(testCase.mockCiPipelineMaterials, testCase.mockCiPipelineMaterialsError)
			}
			ciPipelinesForConfiguredPlugins, _, _, err := globalPolicyService.getCIPipelinesForConfiguredPlugins(testCase.globalPolicyDetailDto)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.wantFilteredCiPipelines, ciPipelinesForConfiguredPlugins)
		})
	}
}

func TestFindAllWorkflowsComponentDetailsForCiPipelineIds(t *testing.T) {
	testCases := []struct {
		ciPipelineIds                []int
		ciPipelineMaterialMap        map[int][]*pipelineConfig.CiPipelineMaterial
		mockAppWorkflowMappings      []*appWorkflow.AppWorkflowMapping
		mockAppWorkflowMappingsError error
		mockAppIdsForGitMaterial     []int
		mockGitMaterials             []*pipelineConfig.GitMaterial
		mockAppIdGitMaterialError    error
		mockCiPipelines              []*pipelineConfig.CiPipeline
		mockCiPipelinesError         error
		mockCdPipelines              []*pipelineConfig.Pipeline
		mockCdPipelinesError         error
		wantError                    bool
	}{
		{
			ciPipelineIds: []int{1, 2},
			ciPipelineMaterialMap: map[int][]*pipelineConfig.CiPipelineMaterial{
				1: {
					{
						Id:            1,
						CiPipelineId:  1,
						GitMaterialId: 1,
					},
					{
						Id:            2,
						CiPipelineId:  1,
						GitMaterialId: 2,
					},
				},
				2: {
					{
						Id:            3,
						CiPipelineId:  2,
						GitMaterialId: 3,
					},
				},
			},
			mockAppIdsForGitMaterial: []int{1, 1, 2},
			mockAppWorkflowMappings: []*appWorkflow.AppWorkflowMapping{
				{
					AppWorkflowId: 1,
					Type:          appWorkflow.CIPIPELINE,
					ComponentId:   1,
					AppWorkflow: &appWorkflow.AppWorkflow{
						AppId: 1,
						Name:  "wf-1",
					},
				},
				{
					AppWorkflowId: 1,
					Type:          appWorkflow.CDPIPELINE,
					ComponentId:   1,
					AppWorkflow: &appWorkflow.AppWorkflow{
						AppId: 1,
						Name:  "wf-1",
					},
				},
				{
					AppWorkflowId: 2,
					Type:          appWorkflow.CIPIPELINE,
					ComponentId:   2,
					AppWorkflow: &appWorkflow.AppWorkflow{
						AppId: 2,
						Name:  "wf-2",
					},
				},
			},
			mockGitMaterials: []*pipelineConfig.GitMaterial{
				{
					Id:   1,
					Name: "1-test1",
				},
				{
					Id:   2,
					Name: "2-test2",
				},

				{
					Id:   3,
					Name: "3-test3",
				},
			},
			mockAppIdGitMaterialError: nil,
			mockCiPipelines: []*pipelineConfig.CiPipeline{
				{
					Id:   1,
					Name: "CI1",
				},
				{
					Id:   2,
					Name: "CI2",
				},
			},
			mockCdPipelines: []*pipelineConfig.Pipeline{
				{
					Id: 1,
					Environment: repository2.Environment{
						Name: "devtron-demo",
					},
				},
			},
			wantError: false,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("findAllWorkflowsComponentDetailsForCiPipelineIds-test-%d", i+1), func(t *testing.T) {
			globalPolicyService, _, _, _, _, ciPipelineRepository, gitMaterialRepository, _, _,
				pipelineRepository, appWorkflowRepository, err := getGlobalPolicyService(t)
			assert.NoError(t, err)
			appWorkflowRepository.On("FindMappingsOfWfWithSpecificCIPipelineIds", testCase.ciPipelineIds).
				Return(testCase.mockAppWorkflowMappings, testCase.mockAppWorkflowMappingsError)
			gitMaterialRepository.On("FindByAppIds", testCase.mockAppIdsForGitMaterial).
				Return(testCase.mockGitMaterials, testCase.mockAppIdGitMaterialError)
			ciPipelineRepository.On("FindByIdsIn", mock.Anything).
				Return(testCase.mockCiPipelines, testCase.mockCiPipelinesError)
			pipelineRepository.On("FindByIdsIn", mock.Anything).
				Return(testCase.mockCdPipelines, testCase.mockCdPipelinesError)
			appIdGitMaterialMap, err := globalPolicyService.findAllWorkflowsComponentDetailsForCiPipelineIds(testCase.ciPipelineIds, testCase.ciPipelineMaterialMap)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, len(testCase.ciPipelineIds), len(appIdGitMaterialMap))
		})
	}
}

func TestGetPolicyOffendingPipelinesWfTree(t *testing.T) {
	testCases := []struct {
		policyId              int
		mockGlobalPolicy      *repository.GlobalPolicy
		mockGlobalPolicyError error
		wantError             bool
	}{
		{
			policyId: 1,
			mockGlobalPolicy: &repository.GlobalPolicy{
				Id:       1,
				PolicyOf: bean.GLOBAL_POLICY_TYPE_PLUGIN.ToString(),
				Version:  bean.GLOBAL_POLICY_VERSION_V1.ToString(),
				Enabled:  false,
			},
			mockGlobalPolicyError: nil,
			wantError:             false,
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("GetPolicyOffendingPipelinesWfTree-test-%d", i+1), func(t *testing.T) {
			globalPolicyService, globalPolicyRepository, _, _, _, _, _, _, _, _, _, err := getGlobalPolicyService(t)
			assert.NoError(t, err)
			globalPolicyRepository.On("GetById", testCase.policyId).
				Return(testCase.mockGlobalPolicy, testCase.mockGlobalPolicyError)
			_, err = globalPolicyService.GetPolicyOffendingPipelinesWfTree(testCase.policyId)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func getGlobalPolicyService(t *testing.T) (globalPolicyService *GlobalPolicyServiceImpl, globalPolicyRepository *mocks.GlobalPolicyRepository, globalPolicySearchableFieldRepository *mocks.GlobalPolicySearchableFieldRepository, devtronResourceService *mocks3.DevtronResourceService, globalPolicyHistoryService *mocks2.GlobalPolicyHistoryService, ciPipelineRepository *mocks4.CiPipelineRepository, gitMaterialRepository *mocks4.MaterialRepository, pipelineStageRepository *mocks5.PipelineStageRepository, ciPipelineMaterialRepository *mocks4.CiPipelineMaterialRepository, pipelineRepository *mocks4.PipelineRepository, appWorkflowRepository *mocks6.AppWorkflowRepository, err error) {
	logger, err := util.NewSugardLogger()
	if err != nil {
		return nil, globalPolicyRepository, globalPolicySearchableFieldRepository, devtronResourceService, globalPolicyHistoryService, ciPipelineRepository, gitMaterialRepository, pipelineStageRepository, ciPipelineMaterialRepository, pipelineRepository, appWorkflowRepository, err
	}
	var dbConnection *pg.DB = nil
	globalPolicyRepository = mocks.NewGlobalPolicyRepository(t)
	globalPolicySearchableFieldRepository = mocks.NewGlobalPolicySearchableFieldRepository(t)
	devtronResourceService = mocks3.NewDevtronResourceService(t)
	ciPipelineRepository = mocks4.NewCiPipelineRepository(t)
	pipelineRepository = mocks4.NewPipelineRepository(t)
	appWorkflowRepository = mocks6.NewAppWorkflowRepository(t)
	pipelineStageRepository = mocks5.NewPipelineStageRepository(t)
	appRepository := app.NewAppRepositoryImpl(dbConnection, logger)
	globalPolicyHistoryService = mocks2.NewGlobalPolicyHistoryService(t)
	ciPipelineMaterialRepository = mocks4.NewCiPipelineMaterialRepository(t)
	gitMaterialRepository = mocks4.NewMaterialRepository(t)
	globalPolicyService = NewGlobalPolicyServiceImpl(logger, globalPolicyRepository,
		globalPolicySearchableFieldRepository, devtronResourceService, ciPipelineRepository, pipelineRepository,
		appWorkflowRepository, pipelineStageRepository, appRepository, globalPolicyHistoryService, ciPipelineMaterialRepository, gitMaterialRepository)
	return globalPolicyService, globalPolicyRepository, globalPolicySearchableFieldRepository, devtronResourceService, globalPolicyHistoryService, ciPipelineRepository, gitMaterialRepository, pipelineStageRepository, ciPipelineMaterialRepository, pipelineRepository, appWorkflowRepository, nil
}
