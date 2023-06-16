package globalPolicy

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	bean2 "github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/history"
	bean3 "github.com/devtron-labs/devtron/pkg/globalPolicy/history/bean"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type GlobalPolicyService interface {
	GetById(id int) (*bean.GlobalPolicyDto, error)
	GetAllGlobalPolicies(policyOf bean.GlobalPolicyType, policyVersion bean.GlobalPolicyVersion) ([]*bean.GlobalPolicyDto, error)
	GetPolicyOffendingPipelinesWfTree(policyId int) (*bean.PolicyOffendingPipelineWfTreeObject, error)
	GetOnlyBlockageStateForCiPipeline(ciPipelineId int, branchValues []string) (bool, bool, *bean.ConsequenceDto, error)
	GetBlockageStateForACIPipelineTrigger(ciPipelineId int, parentCiPipelineId int, branchValues []string, toOnlyGetBlockedStatePolicies bool) (bool, bool, *bean.ConsequenceDto, error)
	GetMandatoryPluginsForACiPipeline(ciPipelineId, appId int, branchValues []string, toOnlyGetBlockedStatePolicies bool) (*bean.MandatoryPluginDto, map[string]*bean.ConsequenceDto, error)
	CreateOrUpdateGlobalPolicy(policy *bean.GlobalPolicyDto) error
	DeleteGlobalPolicy(policyId int, userId int32) error
}

type GlobalPolicyServiceImpl struct {
	logger                                *zap.SugaredLogger
	globalPolicyRepository                repository.GlobalPolicyRepository
	globalPolicySearchableFieldRepository repository.GlobalPolicySearchableFieldRepository
	devtronResourceService                devtronResource.DevtronResourceService
	ciPipelineRepository                  pipelineConfig.CiPipelineRepository
	pipelineRepository                    pipelineConfig.PipelineRepository
	appWorkflowRepository                 appWorkflow.AppWorkflowRepository
	pipelineStageRepository               repository2.PipelineStageRepository
	appRepository                         app.AppRepository
	globalPolicyHistoryService            history.GlobalPolicyHistoryService
	ciPipelineMaterialRepository          pipelineConfig.CiPipelineMaterialRepository
	gitMaterialRepository                 pipelineConfig.MaterialRepository
}

func NewGlobalPolicyServiceImpl(logger *zap.SugaredLogger,
	globalPolicyRepository repository.GlobalPolicyRepository,
	globalPolicySearchableFieldRepository repository.GlobalPolicySearchableFieldRepository,
	devtronResourceService devtronResource.DevtronResourceService,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	pipelineStageRepository repository2.PipelineStageRepository,
	appRepository app.AppRepository,
	globalPolicyHistoryService history.GlobalPolicyHistoryService,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	gitMaterialRepository pipelineConfig.MaterialRepository) *GlobalPolicyServiceImpl {
	return &GlobalPolicyServiceImpl{
		logger:                                logger,
		globalPolicyRepository:                globalPolicyRepository,
		globalPolicySearchableFieldRepository: globalPolicySearchableFieldRepository,
		devtronResourceService:                devtronResourceService,
		ciPipelineRepository:                  ciPipelineRepository,
		pipelineStageRepository:               pipelineStageRepository,
		appRepository:                         appRepository,
		globalPolicyHistoryService:            globalPolicyHistoryService,
		ciPipelineMaterialRepository:          ciPipelineMaterialRepository,
		pipelineRepository:                    pipelineRepository,
		appWorkflowRepository:                 appWorkflowRepository,
		gitMaterialRepository:                 gitMaterialRepository,
	}
}

func (impl *GlobalPolicyServiceImpl) GetById(id int) (*bean.GlobalPolicyDto, error) {
	//getting global policy entry
	globalPolicy, err := impl.globalPolicyRepository.GetById(id)
	if err != nil {
		impl.logger.Errorw("error in getting policy by id", "err", err, "id", id)
		return nil, err
	}
	globalPolicyDto, err := globalPolicy.GetGlobalPolicyDto()
	if err != nil {
		impl.logger.Errorw("error in getting globalPolicyDto", "err", err, "policyId", globalPolicy.Id)
		return nil, err
	}
	return globalPolicyDto, nil
}

func (impl *GlobalPolicyServiceImpl) GetAllGlobalPolicies(policyOf bean.GlobalPolicyType, policyVersion bean.GlobalPolicyVersion) ([]*bean.GlobalPolicyDto, error) {
	//getting all global policy entries
	globalPolicies, err := impl.globalPolicyRepository.GetAllByPolicyOfAndVersion(policyOf, policyVersion)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting all policies", "err", err, "policyOf", policyOf, "policyVersion", policyVersion)
		return nil, err
	}
	globalPolicyDtos, err := impl.getGlobalPolicyDtos(globalPolicies)
	if err != nil {
		impl.logger.Errorw("error in getting globalPolicyDtos", "err", err, "globalPolicies", globalPolicies)
		return nil, err
	}
	return globalPolicyDtos, nil
}

func (impl *GlobalPolicyServiceImpl) CreateOrUpdateGlobalPolicy(policy *bean.GlobalPolicyDto) error {
	//creating global policy entry
	err := impl.createOrUpdateGlobalPolicyInDb(policy)
	if err != nil {
		impl.logger.Errorw("error in creating global policy entry, CreateOrUpdateGlobalPolicy", "err", err, "policy", policy)
		return err
	}
	//creating global policy searchable fields entries
	err = impl.createGlobalPolicySearchableFieldsInDbIfNeeded(policy)
	if err != nil {
		impl.logger.Errorw("error in creating global policy searchable field entry, CreateOrUpdateGlobalPolicy", "err", err, "policy", policy)
		return err
	}
	return nil
}

func (impl *GlobalPolicyServiceImpl) DeleteGlobalPolicy(policyId int, userId int32) error {
	policyModel, err := impl.globalPolicyRepository.GetById(policyId)
	if err != nil {
		impl.logger.Errorw("error in getting global policy by id", "err", err, "policyId", policyId)
		return err
	}
	err = impl.deleteGlobalPolicyAndSearchableFields(policyId, userId)
	if err != nil {
		impl.logger.Errorw("error, deleteGlobalPolicyAndSearchableFields", "err", err, "policyId", policyId)
		return err
	}
	//creating history entry
	err = impl.globalPolicyHistoryService.CreateHistoryEntry(policyModel, bean3.HISTORY_OF_ACTION_DELETE)
	if err != nil {
		impl.logger.Warnw("error in creating global policy history", "err", err, "policyId", policyId)
	}
	return nil
}

func (impl *GlobalPolicyServiceImpl) deleteGlobalPolicyAndSearchableFields(policyId int, userId int32) error {
	//initiating transaction
	dbConnection := impl.globalPolicyRepository.GetDbConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in initiating transaction", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	//mark global policy entry deleted
	err = impl.globalPolicyRepository.MarkDeletedById(policyId, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in marking global policy entry deleted", "err", err, "policyId", policyId)
		return err
	}
	//deleting global policy searchable field entries deleted
	err = impl.globalPolicySearchableFieldRepository.DeleteByPolicyId(policyId, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting searchable fields entry by global policy id", "err", err, "globalPolicyId", policyId)
		return err
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return err
	}
	return nil
}

func (impl *GlobalPolicyServiceImpl) GetPolicyOffendingPipelinesWfTree(policyId int) (*bean.PolicyOffendingPipelineWfTreeObject, error) {
	//get all policies
	globalPolicy, err := impl.globalPolicyRepository.GetById(policyId)
	if err != nil {
		impl.logger.Errorw("error in getting policy by Id", "err", err, "policyId", policyId)
		return nil, err
	}
	offendingPipelineWfTree := &bean.PolicyOffendingPipelineWfTreeObject{
		PolicyId: policyId,
	}
	if globalPolicy.Enabled { //getting workflows only when policy is enabled, because disabled policies are not enforced and would not be having any offending pipelines
		var globalPolicyDetailDto bean.GlobalPolicyDetailDto
		err = json.Unmarshal([]byte(globalPolicy.PolicyJson), &globalPolicyDetailDto)
		if err != nil {
			impl.logger.Errorw("error in un-marshaling global policy json", "err", err, "policyJson", globalPolicy.PolicyJson)
			return nil, err
		}
		allProjects, allClusters, branchList, isAnyEnvSelectorPresent, isProductionEnvFlag,
			projectAppNameMap, clusterEnvNameMap := getAllAppEnvBranchDetailsFromGlobalPolicyDetail(&globalPolicyDetailDto)
		var ciPipelineProjectAppNameObjs []*pipelineConfig.CiPipelineAppProject
		if len(allProjects) > 0 {
			ciPipelineProjectAppNameObjs, err = impl.ciPipelineRepository.GetAllCIAppAndProjectByProjectNames(allProjects)
			if err != nil {
				impl.logger.Errorw("error in getting all ci pipelines by project names", "err", err)
				return nil, err
			}
			if len(ciPipelineProjectAppNameObjs) == 0 {
				//no pipelines found in given projects, no possible offending pipelines
				return offendingPipelineWfTree, nil
			}
		}
		//getting pipelines to be filtered for project and app names
		ciPipelinesToBeFiltered := getFilteredCiPipelinesByProjectAppObjs(ciPipelineProjectAppNameObjs, projectAppNameMap)

		if isAnyEnvSelectorPresent {
			var ciPipelineClusterEnvNameObjs []*pipelineConfig.CiPipelineEnvCluster
			if isProductionEnvFlag {
				ciPipelineClusterEnvNameObjs, err = impl.ciPipelineRepository.GetAllCIsClusterAndEnvForAllProductionEnvCD(ciPipelinesToBeFiltered)
				if err != nil {
					impl.logger.Errorw("error in getting all ci pipelines by cluster names", "err", err)
					return nil, err
				}
			} else if len(allClusters) > 0 {
				ciPipelineClusterEnvNameObjs, err = impl.ciPipelineRepository.GetAllCIsClusterAndEnvByCDClusterNames(allClusters, ciPipelinesToBeFiltered)
				if err != nil {
					impl.logger.Errorw("error in getting all ci pipelines by cluster names", "err", err)
					return nil, err
				}
			}
			//resetting filter pipelines, now will be updating on basis of cluster match
			ciPipelinesToBeFiltered = nil

			//getting pipelines to be filtered for project and app names
			ciPipelinesToBeFiltered = getFilteredCiPipelinesByClusterAndEnvObjs(ciPipelineClusterEnvNameObjs, isProductionEnvFlag, clusterEnvNameMap)

			if len(ciPipelinesToBeFiltered) == 0 {
				//no pipelines found in given environment selector, no possible offending pipelines
				return offendingPipelineWfTree, nil
			}
		}

		var ciPipelinesForConfiguredPlugins []int
		ciPipelineParentChildMap := make(map[int][]int) //map of parent ciPipelineId and (all linked ones + self)
		ciPipelineMaterialMap := make(map[int][]*pipelineConfig.CiPipelineMaterial)
		if len(branchList) != 0 {
			var ciPipelineMaterials []*pipelineConfig.CiPipelineMaterial
			if len(ciPipelinesToBeFiltered) > 0 {
				ciPipelineMaterials, err = impl.ciPipelineMaterialRepository.GetByCiPipelineIdsExceptUnsetRegexBranch(ciPipelinesToBeFiltered)
				if err != nil {
					impl.logger.Errorw("error in GetByCiPipelineIdsExceptUnsetRegexBranch", "err", err, "ciPipelineIds", ciPipelinesToBeFiltered)
					return nil, err
				}
			} else {
				ciPipelineMaterials, err = impl.ciPipelineMaterialRepository.GetAllExceptUnsetRegexBranch()
				if err != nil {
					impl.logger.Errorw("error in GetAllExceptUnsetRegexBranch", "err", err)
					return nil, err
				}
			}
			ciPipelinesFinalMap := make(map[int]bool, len(ciPipelineMaterials))
			for _, ciPipelineMaterial := range ciPipelineMaterials {
				ciPipelineId := ciPipelineMaterial.CiPipelineId
				parentCiPipelineId := ciPipelineMaterial.ParentCiPipeline
				ciPipelineMaterialMap[ciPipelineId] = append(ciPipelineMaterialMap[ciPipelineId], ciPipelineMaterial)
				pipelineTobeUsedToFetchConfiguredPlugins := 0
				if parentCiPipelineId != 0 {
					pipelineTobeUsedToFetchConfiguredPlugins = parentCiPipelineId
				} else {
					pipelineTobeUsedToFetchConfiguredPlugins = ciPipelineId
				}
				if _, ok := ciPipelinesFinalMap[pipelineTobeUsedToFetchConfiguredPlugins]; !ok {
					for _, branch := range branchList {
						isBranchMatched, err := isBranchValueMatched(branch, ciPipelineMaterial.Value)
						if err != nil {
							impl.logger.Errorw("error in checking if branch value matched or not", "err", err, "branch", branch, "branchValue", ciPipelineMaterial.Value)
							return nil, err
						}
						if isBranchMatched {
							ciPipelinesFinalMap[pipelineTobeUsedToFetchConfiguredPlugins] = true
							ciPipelinesForConfiguredPlugins = append(ciPipelinesForConfiguredPlugins, pipelineTobeUsedToFetchConfiguredPlugins)
							ciPipelineParentChildMap[pipelineTobeUsedToFetchConfiguredPlugins] =
								append(ciPipelineParentChildMap[pipelineTobeUsedToFetchConfiguredPlugins], ciPipelineId)

						}
					}
				} else {
					//adding request ci pipeline again to this because it might be possible
					//that we have entry of parent ci pipeline through one linked, but we need to append this ciPipeline too (might be linked) too
					ciPipelineParentChildMap[pipelineTobeUsedToFetchConfiguredPlugins] =
						append(ciPipelineParentChildMap[pipelineTobeUsedToFetchConfiguredPlugins], ciPipelineId)
				}
			}
		} else {
			ciPipelinesForConfiguredPlugins = ciPipelinesToBeFiltered
			if len(ciPipelinesForConfiguredPlugins) > 0 { //getting all materials by ci pipeline Ids
				ciPipelineMaterials, err := impl.ciPipelineMaterialRepository.FindByCiPipelineIdsIn(ciPipelinesForConfiguredPlugins)
				if err != nil {
					impl.logger.Errorw("error in getting ciPipeline material by ciPipelineIds", "err", err, "ciPipelineIds", ciPipelinesForConfiguredPlugins)
					return nil, err
				}

				for _, ciPipelineMaterial := range ciPipelineMaterials {
					ciPipelineId := ciPipelineMaterial.CiPipelineId
					ciPipelineMaterialMap[ciPipelineId] = append(ciPipelineMaterialMap[ciPipelineId], ciPipelineMaterial)
				}
			}
		}
		var configuredPlugins []*repository2.PipelineStageStep
		if len(ciPipelinesForConfiguredPlugins) > 0 {
			configuredPlugins, err = impl.pipelineStageRepository.GetConfiguredPluginsForCIPipelines(ciPipelinesForConfiguredPlugins)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error, GetConfiguredPluginsForCIPipelines", "err", err, "ciPipelineIds", ciPipelinesForConfiguredPlugins)
				return nil, err
			}
		}
		//map of {ciPipelineId: map of {pluginIdStage : bool} }
		ciPipelineConfiguredPluginMap := make(map[int]map[string]bool)
		for _, configuredPlugin := range configuredPlugins {
			ciPipelineId := configuredPlugin.PipelineStage.CiPipelineId
			pluginIdStr := fmt.Sprintf("%d", configuredPlugin.RefPluginId)
			var pluginConfiguredInStage bean.PluginApplyStage
			switch configuredPlugin.PipelineStage.Type {
			case repository2.PIPELINE_STAGE_TYPE_PRE_CI:
				pluginConfiguredInStage = bean.PLUGIN_APPLY_STAGE_PRE_CI
			case repository2.PIPELINE_STAGE_TYPE_POST_CI:
				pluginConfiguredInStage = bean.PLUGIN_APPLY_STAGE_POST_CI
			}
			pluginIdApplyStage := getSlashSeparatedString(pluginIdStr, pluginConfiguredInStage.ToString())
			pluginIdPreOrPostCiStage := getSlashSeparatedString(pluginIdStr, bean.PLUGIN_APPLY_STAGE_PRE_OR_POST_CI.ToString())
			if _, ok := ciPipelineConfiguredPluginMap[ciPipelineId]; !ok {
				ciPipelineConfiguredPluginMap[ciPipelineId] = make(map[string]bool)
			}
			ciPipelineConfiguredPluginMap[ciPipelineId][pluginIdApplyStage] = true
			ciPipelineConfiguredPluginMap[ciPipelineId][pluginIdPreOrPostCiStage] = true

		}

		offendingCiPipelineIds := make([]int, 0, len(ciPipelineConfiguredPluginMap))
		for _, ciPipelineId := range ciPipelinesForConfiguredPlugins {
			if configuredPluginMap, ok := ciPipelineConfiguredPluginMap[ciPipelineId]; ok {
				for _, mandatoryPlugin := range globalPolicyDetailDto.Definitions {
					data := mandatoryPlugin.Data
					pluginIdStr := fmt.Sprintf("%d", data.PluginId)
					pluginApplyStage := data.ApplyToStage
					pluginIdApplyStage := getSlashSeparatedString(pluginIdStr, pluginApplyStage.ToString())

					if _, ok2 := configuredPluginMap[pluginIdApplyStage]; !ok2 {
						//all linked and self pipeline id fetch for this ci pipeline id
						selfAndAllLinkedPipelineIds := ciPipelineParentChildMap[ciPipelineId]
						offendingCiPipelineIds = append(offendingCiPipelineIds, selfAndAllLinkedPipelineIds...)
						break
					}
				}
			} else {
				offendingCiPipelineIds = append(offendingCiPipelineIds, ciPipelineId)
			}
		}

		wfComponents, err := impl.findAllWorkflowsComponentDetailsForCiPipelineIds(offendingCiPipelineIds, ciPipelineMaterialMap)
		if err != nil {
			impl.logger.Errorw("error, findAllWorkflowsComponentDetailsForCiPipelineIds", "err", err, "ciPipelineIds", offendingCiPipelineIds)
			return nil, err
		}
		offendingPipelineWfTree.Workflows = wfComponents
	}
	return offendingPipelineWfTree, nil
}

func (impl *GlobalPolicyServiceImpl) GetOnlyBlockageStateForCiPipeline(ciPipelineId int, branchValues []string) (bool, bool, *bean.ConsequenceDto, error) {
	ciPipeline, err := impl.ciPipelineRepository.FindById(ciPipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting ci pipeline by id", "err", err, "ciPipelineId", ciPipeline)
		return false, false, nil, err
	}
	return impl.GetBlockageStateForACIPipelineTrigger(ciPipeline.Id, ciPipeline.ParentCiPipeline, branchValues, false)
}

func (impl *GlobalPolicyServiceImpl) GetBlockageStateForACIPipelineTrigger(ciPipelineId int, parentCiPipelineId int, branchValues []string, toOnlyGetBlockedStatePolicies bool) (bool, bool, *bean.ConsequenceDto, error) {
	ciPipelineIdToGetBlockageState := 0
	if parentCiPipelineId != 0 {
		ciPipelineIdToGetBlockageState = parentCiPipelineId
	} else {
		ciPipelineIdToGetBlockageState = ciPipelineId
	}
	isOffendingMandatoryPlugin := false
	isCIPipelineTriggerBlocked := false
	//getting all mandatory plugins for a ci pipeline
	mandatoryPlugins, mandatoryPluginsBlockageState, err := impl.GetMandatoryPluginsForACiPipeline(ciPipelineIdToGetBlockageState, 0, branchValues, toOnlyGetBlockedStatePolicies)
	if err != nil {
		impl.logger.Errorw("error in getting mandatory plugins for a ci", "err", err, "ciPipelineId", ciPipelineId)
		return isOffendingMandatoryPlugin, isCIPipelineTriggerBlocked, nil, err
	}

	var definitions []*bean.MandatoryPluginDefinitionDto
	if mandatoryPlugins != nil {
		definitions = mandatoryPlugins.Definitions
	}
	if len(definitions) == 0 {
		//no mandatory plugins found
		return isOffendingMandatoryPlugin, isCIPipelineTriggerBlocked, nil, nil
	}
	//getting all configured plugins for ci pipeline
	configuredPlugins, err := impl.pipelineStageRepository.GetConfiguredPluginsForCIPipelines([]int{ciPipelineIdToGetBlockageState})
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error, GetConfiguredPluginsForCIPipelines", "err", err, "ciPipelineId", ciPipelineIdToGetBlockageState)
		return isOffendingMandatoryPlugin, isCIPipelineTriggerBlocked, nil, err
	}
	configuredPluginMap := make(map[string]bool)
	for _, configuredPlugin := range configuredPlugins {
		pluginIdStr := fmt.Sprintf("%d", configuredPlugin.RefPluginId)
		var pluginConfiguredInStage bean.PluginApplyStage
		switch configuredPlugin.PipelineStage.Type {
		case repository2.PIPELINE_STAGE_TYPE_PRE_CI:
			pluginConfiguredInStage = bean.PLUGIN_APPLY_STAGE_PRE_CI
		case repository2.PIPELINE_STAGE_TYPE_POST_CI:
			pluginConfiguredInStage = bean.PLUGIN_APPLY_STAGE_POST_CI
		}
		pluginIdApplyStage := getSlashSeparatedString(pluginIdStr, pluginConfiguredInStage.ToString())
		pluginIdPreOrPostCiStage := getSlashSeparatedString(pluginIdStr, bean.PLUGIN_APPLY_STAGE_PRE_OR_POST_CI.ToString())
		configuredPluginMap[pluginIdApplyStage] = true
		configuredPluginMap[pluginIdPreOrPostCiStage] = true
	}

	var blockageStateFinal *bean.ConsequenceDto
	for _, mandatoryPlugin := range definitions {
		data := mandatoryPlugin.Data
		pluginIdStr := fmt.Sprintf("%d", data.PluginId)
		applyToStage := data.ApplyToStage
		pluginIdApplyStage := getSlashSeparatedString(pluginIdStr, applyToStage.ToString())
		if _, ok := configuredPluginMap[pluginIdApplyStage]; !ok {
			blockageState := mandatoryPluginsBlockageState[pluginIdApplyStage]
			if blockageStateFinal == nil {
				blockageStateFinal = blockageState
			} else {
				if sev := blockageStateFinal.GetSeverity(blockageState); sev == bean.SEVERITY_MORE_SEVERE {
					blockageStateFinal = blockageState
				}
			}
			//mandatory plugin not found in configured plugin, marking block state as true
			isOffendingMandatoryPlugin = true
			isBlockingConsequence := checkIfConsequenceIsBlocking(blockageState)
			if isBlockingConsequence {
				isCIPipelineTriggerBlocked = true
			}
		}
	}
	return isOffendingMandatoryPlugin, isCIPipelineTriggerBlocked, blockageStateFinal, nil
}

func (impl *GlobalPolicyServiceImpl) GetMandatoryPluginsForACiPipeline(ciPipelineId, appId int, branchValues []string,
	toOnlyGetBlockedStatePolicies bool) (*bean.MandatoryPluginDto, map[string]*bean.ConsequenceDto, error) {
	var ciPipelineAppProjectObjs []*pipelineConfig.CiPipelineAppProject
	var err error
	if ciPipelineId != 0 {
		//getting all linked ci pipelines and ciPipelineId along with appName and projectName
		ciPipelineAppProjectObjs, err = impl.ciPipelineRepository.GetAppAndProjectNameForParentAndAllLinkedCI(ciPipelineId)
		if err != nil {
			impl.logger.Errorw("error in getting appName and projectName of parent and all linked ci", "err", err, "ciPipelineId", ciPipelineId)
			return nil, nil, err
		}
		if len(ciPipelineAppProjectObjs) == 0 {
			//at least 1 obj is expected of the given ci pipeline
			return nil, nil, &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "bad request, ci pipeline not found"}
		}
	} else if appId != 0 { //ciPipeline is zero(ciPipeline create request), assuming ciPipelineId as zero and moving ahead
		app, err := impl.appRepository.FindAppAndProjectByAppId(appId)
		if err != nil {
			impl.logger.Errorw("error in getting app by id", "err", err, "appId", appId)
			return nil, nil, err
		}
		ciPipelineAppProjectObjs = []*pipelineConfig.CiPipelineAppProject{
			{
				CiPipelineId: 0,
				AppName:      app.AppName,
				ProjectName:  app.Team.Name,
			},
		}
	} else {
		return nil, nil, &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "bad request, need ciPipelineId or appId"}
	}
	allCiPipelineIds := make([]int, 0, len(ciPipelineAppProjectObjs))
	ciPipelineIdNameMap := make(map[int]string, len(ciPipelineAppProjectObjs))
	allProjectAppNames := make([]string, 0, len(ciPipelineAppProjectObjs))
	projectMap := make(map[string]bool, len(ciPipelineAppProjectObjs))
	ciPipelineIdProjectAppNameMap := make(map[int]*bean.PluginSourceCiPipelineAppDetailDto, len(ciPipelineAppProjectObjs))

	for _, ciPipelineAppProjectObj := range ciPipelineAppProjectObjs {
		ciPipelineIdInObj := ciPipelineAppProjectObj.CiPipelineId
		allCiPipelineIds = append(allCiPipelineIds, ciPipelineIdInObj)
		projectName := ciPipelineAppProjectObj.ProjectName
		appName := ciPipelineAppProjectObj.AppName
		ciPipelineIdNameMap[ciPipelineIdInObj] = ciPipelineAppProjectObj.CiPipelineName
		projectAppName := getSlashSeparatedString(projectName, appName)
		allProjectAppNames = append(allProjectAppNames, projectAppName)
		projectMap[projectName] = true
		ciPipelineIdProjectAppNameMap[ciPipelineIdInObj] = getCIPipelineAppDetailDto(projectName, appName)
	}
	var ciPipelineEnvClusterObjs []*pipelineConfig.CiPipelineEnvCluster
	if len(allCiPipelineIds) > 0 {
		//getting all envName and clusterName of all linked and parent ci
		ciPipelineEnvClusterObjs, err = impl.ciPipelineRepository.GetAllCDsEnvAndClusterNameByCiPipelineIds(allCiPipelineIds)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting envName and clusterName by ciPipelineIds", "err", err, "ciPipelineIds", allCiPipelineIds)
			return nil, nil, err
		}
	}
	haveAnyProductionEnv := false
	//map of ciPipelineId with all productionEnv present in that ciPipeline's workflow
	ciPipelineIdProductionEnvDetailMap := make(map[int][]*bean.PluginSourceCiPipelineEnvDetailDto)
	ciPipelineIdEnvDetailMap := make(map[int][]*bean.PluginSourceCiPipelineEnvDetailDto)
	allClusterEnvNames := make([]string, 0, len(ciPipelineEnvClusterObjs))
	clusterMap := make(map[string]bool, len(ciPipelineEnvClusterObjs))
	for _, ciPipelineEnvClusterObj := range ciPipelineEnvClusterObjs {
		ciPipelineIdInObj := ciPipelineEnvClusterObj.CiPipelineId
		clusterName := ciPipelineEnvClusterObj.ClusterName
		envName := ciPipelineEnvClusterObj.EnvName
		clusterMap[clusterName] = true
		envDetailDto := getCIPipelineEnvDetailDto(clusterName, envName)
		if ciPipelineEnvClusterObj.IsProductionEnv {
			haveAnyProductionEnv = true //setting flag for having at least one production env to true to filter global policies query using this
			ciPipelineIdProductionEnvDetailMap[ciPipelineIdInObj] = append(ciPipelineIdProductionEnvDetailMap[ciPipelineIdInObj],
				envDetailDto)
		}
		ciPipelineIdEnvDetailMap[ciPipelineIdInObj] = append(ciPipelineIdEnvDetailMap[ciPipelineIdInObj], envDetailDto)
		clusterEnvName := getSlashSeparatedString(clusterName, envName)
		allClusterEnvNames = append(allClusterEnvNames, clusterEnvName)
	}

	searchableKeyNameIdMap := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()
	//getting searchable keyId and value map for filtering all searchable fields in db
	searchableKeyIdValueMapWhereOrGroup, searchableKeyIdValueMapWhereAndGroup := getSearchableKeyIdValueMapForFilter(allProjectAppNames, allClusterEnvNames, branchValues,
		haveAnyProductionEnv, toOnlyGetBlockedStatePolicies, searchableKeyNameIdMap)

	searchableFieldModels, err := impl.globalPolicySearchableFieldRepository.GetSearchableFields(searchableKeyIdValueMapWhereOrGroup, searchableKeyIdValueMapWhereAndGroup)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting searchable fields", "err", err)
		return nil, nil, err
	}
	globalPolicyIds, err := impl.getFilteredGlobalPolicyIdsFromSearchableFields(searchableFieldModels, projectMap, clusterMap, branchValues)
	if err != nil {
		impl.logger.Errorw("error, getFilteredGlobalPolicyIdsFromSearchableFields", "err", err, "searchableFieldModels", searchableFieldModels)
		return nil, nil, err
	}
	var globalPolicies []*repository.GlobalPolicy
	if len(globalPolicyIds) > 0 { //getting global policies by ids
		globalPolicies, err = impl.globalPolicyRepository.GetEnabledPoliciesByIds(globalPolicyIds)
		if err != nil {
			impl.logger.Errorw("error in getting global policies by ids", "err", err, "globalPolicyIds", globalPolicyIds)
			return nil, nil, err
		}
	}
	//map of "pluginId/pluginApplyStage" and its sources
	mandatoryPluginDefinitionMap := make(map[string][]*bean.DefinitionSourceDto)
	mandatoryPluginBlockageMap := make(map[string]*bean.ConsequenceDto)
	for _, globalPolicy := range globalPolicies {
		var globalPolicyDetailDto bean.GlobalPolicyDetailDto
		err = json.Unmarshal([]byte(globalPolicy.PolicyJson), &globalPolicyDetailDto)
		if err != nil {
			impl.logger.Errorw("error in un-marshaling global policy json", "err", err, "policyJson", globalPolicy.PolicyJson)
			return nil, nil, err
		}
		consequence := globalPolicyDetailDto.Consequences[0] //hard coding to get only one consequence of all since only one supported in plugin
		isConsequenceBlocking := checkIfConsequenceIsBlocking(consequence)
		if isConsequenceBlocking { //consequence not blocking, skipping this policy
			consequence = &bean.ConsequenceDto{
				Action: bean.CONSEQUENCE_ACTION_BLOCK,
			}
		} else if toOnlyGetBlockedStatePolicies {
			continue //consequence is not blocking, and we need only blocked policies, so skipping
		}
		//map of plugins for which consequence is more or same severe than already present, and need to be included in this policy
		pluginIdApplyStageMap := getPluginIdApplyStageAndPluginBlockageMaps(globalPolicyDetailDto.Definitions, consequence, mandatoryPluginBlockageMap)
		definitionSourceDtos, err := impl.getDefinitionSourceDtos(globalPolicyDetailDto, allCiPipelineIds, ciPipelineId,
			ciPipelineIdProjectAppNameMap, ciPipelineIdEnvDetailMap, ciPipelineIdProductionEnvDetailMap, branchValues, globalPolicy.Name, ciPipelineIdNameMap)
		if err != nil {
			impl.logger.Errorw("error in getting definitionSourceDtos", "err", err, "globalPolicy", globalPolicy)
			return nil, nil, err
		}

		if len(definitionSourceDtos) != 0 {
			updateMandatoryPluginDefinitionMap(pluginIdApplyStageMap, mandatoryPluginDefinitionMap, definitionSourceDtos)
		}
	}
	mandatoryPluginsDefinitions := make([]*bean.MandatoryPluginDefinitionDto, 0)
	for pluginIdStage, definitionSources := range mandatoryPluginDefinitionMap {
		mandatoryPluginDefinition, err := getMandatoryPluginDefinition(pluginIdStage, definitionSources)
		if err != nil {
			impl.logger.Errorw("error, getMandatoryPluginDefinition", "err", err)
			return nil, nil, err
		}
		mandatoryPluginsDefinitions = append(mandatoryPluginsDefinitions, mandatoryPluginDefinition)
	}
	mandatoryPlugins := &bean.MandatoryPluginDto{
		Definitions: mandatoryPluginsDefinitions,
	}
	return mandatoryPlugins, mandatoryPluginBlockageMap, nil
}

func (impl *GlobalPolicyServiceImpl) createOrUpdateGlobalPolicyInDb(policy *bean.GlobalPolicyDto) error {
	//getting policy entry by name
	oldEntry, err := impl.globalPolicyRepository.GetByName(policy.Name)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting global policy by name", "err", err, "name", policy.Name)
		return err
	}
	if oldEntry != nil && oldEntry.Id > 0 && policy.Id == 0 {
		//bad create request since name already exists
		impl.logger.Errorw("error in creating global policy, policy already exists by name", "err", err, "name", policy.Name)
		return &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "bad request, policy name already exists"}
	}
	//getting policy detail json
	policyDetailJson, err := json.Marshal(policy.GlobalPolicyDetailDto)
	if err != nil {
		impl.logger.Errorw("error in marshaling globalPolicyDetailDto", "err", err, "policyId", policy.Id)
		return err
	}
	globalPolicyModel := globalPolicyDbAdapter(policy, string(policyDetailJson), nil)
	var historyAction bean3.HistoryOfAction
	if policy.Id == 0 {
		//create entry
		err = impl.globalPolicyRepository.Create(globalPolicyModel)
		if err != nil {
			impl.logger.Errorw("error, createOrUpdateGlobalPolicyInDb", "err", err, "policyDto", policy)
			return err
		}
		//setting id in policy dto
		policy.Id = globalPolicyModel.Id
		historyAction = bean3.HISTORY_OF_ACTION_CREATE
	} else {
		//update entry
		err = impl.globalPolicyRepository.Update(globalPolicyModel)
		if err != nil {
			impl.logger.Errorw("error, createOrUpdateGlobalPolicyInDb", "err", err, "policyDto", policy)
			return err
		}
		historyAction = bean3.HISTORY_OF_ACTION_UPDATE
	}
	//creating history entry
	err = impl.globalPolicyHistoryService.CreateHistoryEntry(globalPolicyModel, historyAction)
	if err != nil {
		impl.logger.Warnw("error in creating global policy history", "err", err, "policyId", globalPolicyModel.Id)
	}
	return nil
}

func (impl *GlobalPolicyServiceImpl) createGlobalPolicySearchableFieldsInDbIfNeeded(policy *bean.GlobalPolicyDto) error {
	//initiating transaction
	dbConnection := impl.globalPolicyRepository.GetDbConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in initiating transaction", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	//first deleting old searchable entries, doing so even if policy is disabled to free indexes
	err = impl.globalPolicySearchableFieldRepository.DeleteByPolicyId(policy.Id, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting global policy searchable fields entries", "err", err, "policy", policy)
		return err
	}
	if policy.Enabled {
		//storing searchable fields only when policy is enabled; since no use of storing searchable fields
		//when policy is disabled and not going to be enforced at all

		//creating global policy searchable fields entries
		err = impl.createGlobalPolicySearchableFieldsInDb(policy, tx)
		if err != nil {
			impl.logger.Errorw("error in creating global policy searchable field entry, CreateOrUpdateGlobalPolicy", "err", err, "policy", policy)
			return err
		}
	}
	//committing transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return err
	}
	return nil
}

func (impl *GlobalPolicyServiceImpl) createGlobalPolicySearchableFieldsInDb(policy *bean.GlobalPolicyDto, tx *pg.Tx) error {
	searchableKeyNameIdMap := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()
	searchableKeyEntriesTotal := make([]*repository.GlobalPolicySearchableField, 0)

	//no searchable fields in definitions for plugin story, TODO add if needed in further global policies

	//getting searchable fields for selectors
	selectorSearchableKeyEntries := impl.getSearchableKeyIdValueEntriesFromSelectors(policy, searchableKeyNameIdMap)

	//getting searchable fields for consequences
	consequenceSearchableKeyEntries := impl.getSearchableKeyIdValueEntriesFromConsequences(policy, searchableKeyNameIdMap)

	//adding all searchable key entries
	searchableKeyEntriesTotal = append(searchableKeyEntriesTotal, selectorSearchableKeyEntries...)
	searchableKeyEntriesTotal = append(searchableKeyEntriesTotal, consequenceSearchableKeyEntries...)

	//saving entries in db
	err := impl.globalPolicySearchableFieldRepository.CreateInBatchWithTxn(searchableKeyEntriesTotal, tx)
	if err != nil {
		impl.logger.Errorw("error in creating global policy searchable fields entry", "err", err, "policy", policy)
		return err
	}
	return nil
}

func (impl *GlobalPolicyServiceImpl) getSearchableKeyIdValueEntriesFromSelectors(policy *bean.GlobalPolicyDto,
	searchableKeyNameIdMap map[bean2.DevtronResourceSearchableKeyName]int) []*repository.GlobalPolicySearchableField {
	searchableFieldEntries := make([]*repository.GlobalPolicySearchableField, 0)
	//getting attributes for plugin policy selectors from global var and then making searchable entries around them
	//TODO: remove and derive from db when more policies are supported
	for _, attribute := range bean.GlobalPluginPolicySelectorAttributes {
		searchableFieldEntriesForAttribute := impl.getSearchableKeyIdValueEntriesForASelectorAttribute(policy, attribute, searchableKeyNameIdMap)
		searchableFieldEntries = append(searchableFieldEntries, searchableFieldEntriesForAttribute...)
	}

	return searchableFieldEntries
}

func (impl *GlobalPolicyServiceImpl) getSearchableKeyIdValueEntriesFromConsequences(policy *bean.GlobalPolicyDto,
	searchableKeyNameIdMap map[bean2.DevtronResourceSearchableKeyName]int) []*repository.GlobalPolicySearchableField {
	searchableFieldEntries := make([]*repository.GlobalPolicySearchableField, 0)
	for _, consequence := range policy.Consequences {
		searchableKeyId := searchableKeyNameIdMap[bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_TRIGGER_ACTION]
		//only passing action to searchable field for now and not metadata field
		//since only plugin policies are supported as of now and only one of it has metadata field
		searchableFieldEntries = append(searchableFieldEntries, globalPolicySearchableFieldDbAdapter(policy.Id,
			searchableKeyId, consequence.Action.ToString(), false, policy.UserId, bean.GLOBAL_POLICY_COMPONENT_CONSEQUENCE))
	}
	return searchableFieldEntries
}

func (impl *GlobalPolicyServiceImpl) getSearchableKeyIdValueEntriesForASelectorAttribute(policy *bean.GlobalPolicyDto,
	attribute bean2.DevtronResourceAttributeName, searchableKeyNameIdMap map[bean2.DevtronResourceSearchableKeyName]int) []*repository.GlobalPolicySearchableField {
	policyId := policy.Id
	userId := policy.UserId
	selectors := &bean.SelectorDto{}
	if policy != nil && policy.Selectors != nil {
		selectors = policy.Selectors
	}
	var searchableFieldEntries []*repository.GlobalPolicySearchableField
	switch attribute {
	case bean2.DEVTRON_RESOURCE_ATTRIBUTE_APP_NAME:
		searchableKeyId := searchableKeyNameIdMap[bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME]
		applicationSelectors := make([]*bean.ProjectAppDto, 0)
		if selectors != nil {
			applicationSelectors = selectors.ApplicationSelector
		}
		for _, appSelector := range applicationSelectors {
			if len(appSelector.AppNames) == 0 {
				projectAppName := getSlashSeparatedString(appSelector.ProjectName, bean.POLICY_ALL_OBJECTS_PLACEHOLDER)
				searchableFieldEntries = append(searchableFieldEntries,
					globalPolicySearchableFieldDbAdapter(policyId, searchableKeyId, projectAppName, true, userId, bean.GLOBAL_POLICY_COMPONENT_SELECTOR))
			} else {
				for _, appName := range appSelector.AppNames {
					projectAppName := getSlashSeparatedString(appSelector.ProjectName, appName)
					searchableFieldEntries = append(searchableFieldEntries,
						globalPolicySearchableFieldDbAdapter(policyId, searchableKeyId, projectAppName, false, userId, bean.GLOBAL_POLICY_COMPONENT_SELECTOR))
				}
			}
		}
	case bean2.DEVTRON_RESOURCE_ATTRIBUTE_ENVIRONMENT_IS_PRODUCTION:
		allProductionEnvironments := false
		if selectors != nil && selectors.EnvironmentSelector != nil {
			allProductionEnvironments = policy.Selectors.EnvironmentSelector.AllProductionEnvironments
		}
		if allProductionEnvironments {
			searchableKeyId := searchableKeyNameIdMap[bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV]
			searchableFieldEntries = append(searchableFieldEntries,
				globalPolicySearchableFieldDbAdapter(policyId, searchableKeyId, bean.TRUE_STRING, false, userId, bean.GLOBAL_POLICY_COMPONENT_SELECTOR))
		}
	case bean2.DEVTRON_RESOURCE_ATTRIBUTE_ENVIRONMENT_NAME:
		clusterEnvList := make([]*bean.ClusterEnvDto, 0)
		if selectors != nil && selectors.EnvironmentSelector != nil {
			if selectors.EnvironmentSelector.ClusterEnvList != nil {
				clusterEnvList = selectors.EnvironmentSelector.ClusterEnvList
			}
		}
		searchableKeyId := searchableKeyNameIdMap[bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME]
		for _, clusterEnvObj := range clusterEnvList {
			if len(clusterEnvObj.EnvNames) == 0 {
				clusterEnvName := getSlashSeparatedString(clusterEnvObj.ClusterName, bean.POLICY_ALL_OBJECTS_PLACEHOLDER)
				searchableFieldEntries = append(searchableFieldEntries,
					globalPolicySearchableFieldDbAdapter(policyId, searchableKeyId, clusterEnvName, true, userId, bean.GLOBAL_POLICY_COMPONENT_SELECTOR))
			} else {
				for _, envName := range clusterEnvObj.EnvNames {
					clusterEnvName := getSlashSeparatedString(clusterEnvObj.ClusterName, envName)
					searchableFieldEntries = append(searchableFieldEntries,
						globalPolicySearchableFieldDbAdapter(policyId, searchableKeyId, clusterEnvName, false, userId, bean.GLOBAL_POLICY_COMPONENT_SELECTOR))
				}
			}
		}

	case bean2.DEVTRON_RESOURCE_ATTRIBUTE_CI_PIPELINE_BRANCH_VALUE:
		searchableKeyId := searchableKeyNameIdMap[bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH]
		branchList := make([]*bean.BranchDto, 0)
		if selectors != nil {
			branchList = policy.Selectors.BranchList
		}
		for _, branchObj := range branchList {
			var isRegexTypeBranchBool bool
			if branchObj.BranchValueType == bean2.VALUE_TYPE_REGEX {
				isRegexTypeBranchBool = true
			} else if branchObj.BranchValueType == bean2.VALUE_TYPE_FIXED {
				isRegexTypeBranchBool = false
			}
			searchableFieldEntries = append(searchableFieldEntries,
				globalPolicySearchableFieldDbAdapter(policyId, searchableKeyId, branchObj.Value, isRegexTypeBranchBool, userId, bean.GLOBAL_POLICY_COMPONENT_SELECTOR))
		}

	default:
		impl.logger.Errorw("invalid attribute found, getSearchableKeyIdValueEntriesForASelectorAttribute", "policy", policy)
		return nil
	}
	return searchableFieldEntries
}

func globalPolicyDbAdapter(policyDto *bean.GlobalPolicyDto, policyDetailJson string, oldEntry *repository.GlobalPolicy) *repository.GlobalPolicy {
	globalPolicyDbObj := &repository.GlobalPolicy{
		Id:          policyDto.Id,
		Name:        policyDto.Name,
		PolicyJson:  policyDetailJson,
		PolicyOf:    policyDto.PolicyOf.ToString(),
		Version:     policyDto.PolicyVersion.ToString(),
		Description: policyDto.Description,
		Enabled:     policyDto.Enabled,
		Deleted:     false,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: policyDto.UserId,
			UpdatedOn: time.Now(),
			UpdatedBy: policyDto.UserId,
		},
	}
	if oldEntry != nil {
		globalPolicyDbObj.CreatedOn = oldEntry.CreatedOn
		globalPolicyDbObj.CreatedBy = oldEntry.CreatedBy
	}
	return globalPolicyDbObj
}

func globalPolicySearchableFieldDbAdapter(policyId, searchableKeyId int, value string, isRegex bool,
	userId int32, policyComponent bean.GlobalPolicyComponent) *repository.GlobalPolicySearchableField {
	return &repository.GlobalPolicySearchableField{
		GlobalPolicyId:  policyId,
		SearchableKeyId: searchableKeyId,
		Value:           value,
		IsRegex:         isRegex,
		PolicyComponent: policyComponent,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
}

func getCIPipelineAppDetailDto(projectName, appName string) *bean.PluginSourceCiPipelineAppDetailDto {
	return &bean.PluginSourceCiPipelineAppDetailDto{
		ProjectName: projectName,
		AppName:     appName,
	}
}

func getCIPipelineEnvDetailDto(clusterName, envName string) *bean.PluginSourceCiPipelineEnvDetailDto {
	return &bean.PluginSourceCiPipelineEnvDetailDto{
		EnvName:     envName,
		ClusterName: clusterName,
	}
}

func getSlashSeparatedString(strings ...string) string {
	result := ""
	for i, s := range strings {
		if i == 0 {
			result = s
		} else {
			result = fmt.Sprintf("%s/%s", result, s)
		}
	}
	return result
}

func getSearchableKeyIdValueMapForFilter(allProjectAppNames, allClusterEnvNames, branchValues []string,
	haveAnyProductionEnv, toOnlyGetBlockedStatePolicies bool, searchableKeyNameIdMap map[bean2.DevtronResourceSearchableKeyName]int) (map[int][]string, map[int][]string) {
	searchableKeyIdValueMapWhereOrGroup, searchableKeyIdValueMapWhereAndGroup := make(map[int][]string), make(map[int][]string)

	if len(allProjectAppNames) > 0 {
		//for app
		appSearchableId := searchableKeyNameIdMap[bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME]
		searchableKeyIdValueMapWhereOrGroup[appSearchableId] = allProjectAppNames
	}
	if len(allClusterEnvNames) > 0 {
		//for env
		envSearchableId := searchableKeyNameIdMap[bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME]
		searchableKeyIdValueMapWhereOrGroup[envSearchableId] = allClusterEnvNames
	}

	if haveAnyProductionEnv {
		//for production env
		productionEnvSearchableId := searchableKeyNameIdMap[bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV]
		searchableKeyIdValueMapWhereOrGroup[productionEnvSearchableId] = []string{bean.TRUE_STRING}
	}

	if len(branchValues) > 0 {
		//for branch
		branchSearchableId := searchableKeyNameIdMap[bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH]
		searchableKeyIdValueMapWhereOrGroup[branchSearchableId] = branchValues
	}

	if toOnlyGetBlockedStatePolicies {
		//setting blocking actions in filter map
		ciPipelineTriggerActionSearchableId := searchableKeyNameIdMap[bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_TRIGGER_ACTION]
		searchableKeyIdValueMapWhereAndGroup[ciPipelineTriggerActionSearchableId] = []string{bean.CONSEQUENCE_ACTION_BLOCK.ToString(),
			bean.CONSEQUENCE_ACTION_ALLOW_UNTIL_TIME.ToString()}
	}
	return searchableKeyIdValueMapWhereOrGroup, searchableKeyIdValueMapWhereAndGroup
}

func (impl *GlobalPolicyServiceImpl) getFilteredGlobalPolicyIdsFromSearchableFields(searchableFieldsModels []*repository.GlobalPolicySearchableField,
	projectMap, clusterMap map[string]bool, branchValues []string) ([]int, error) {
	searchableKeyIdNameMap := impl.devtronResourceService.GetAllSearchableKeyIdNameMap()
	globalPolicyIds := make([]int, 0)
	globalPolicyIdsMap := make(map[int]bool)
	var err error
	for _, searchableFieldsModel := range searchableFieldsModels {
		searchableKeyName := searchableKeyIdNameMap[searchableFieldsModel.SearchableKeyId]
		globalPolicyId := searchableFieldsModel.GlobalPolicyId
		if _, ok := globalPolicyIdsMap[globalPolicyId]; ok {
			//policy already present, no need to process further
			continue
		}
		value := searchableFieldsModel.Value
		vals := strings.Split(value, "/")
		if searchableFieldsModel.IsRegex {
			switch searchableKeyName {
			case bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME:
				//checking if project for this needed by us or not (we always keep project value, so if it matches then probably this is an entry for all apps)
				if len(vals) > 0 {
					if projectMap[vals[0]] {
						globalPolicyIds = append(globalPolicyIds, globalPolicyId)
					}
				}
			case bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME:
				//checking if cluster for this needed by us or not (we always keep cluster value, so if it matches then probably this is an entry for all envs)
				if len(vals) > 0 {
					if clusterMap[vals[0]] {
						globalPolicyIds = append(globalPolicyIds, globalPolicyId)
					}
				}
			case bean2.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH:
				isAnyBranchMatched := false
				for _, branchValue := range branchValues {
					if len(branchValue) != 0 {
						isAnyBranchMatched, err = regexp.MatchString(value, branchValue)
						if err != nil {
							impl.logger.Errorw("error in matching branch regex", "err", err, "regexExpr", value, "valueToBeMatched", branchValue)
							return nil, err
						}
					} else {
						isAnyBranchMatched = true
					}
					if isAnyBranchMatched {
						globalPolicyIds = append(globalPolicyIds, globalPolicyId)
						break
					}
				}
			}
		} else {
			//add global policy id directly
			globalPolicyIds = append(globalPolicyIds, globalPolicyId)
		}
	}
	return globalPolicyIds, nil
}

func getPluginIdApplyStageAndPluginBlockageMaps(definitions []*bean.DefinitionDto, consequence *bean.ConsequenceDto,
	mandatoryPluginBlockageMap map[string]*bean.ConsequenceDto) map[string]bean.Severity {
	pluginIdApplyStageMap := make(map[string]bean.Severity)
	for _, definition := range definitions {
		pluginIdApplyStage := getSlashSeparatedString(fmt.Sprintf("%d", definition.Data.PluginId), definition.Data.ApplyToStage.ToString())
		if definition.AttributeType == bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN {
			blockageStagePlugin, ok := mandatoryPluginBlockageMap[pluginIdApplyStage]
			if ok {
				severity := blockageStagePlugin.GetSeverity(consequence)
				if severity == bean.SEVERITY_MORE_SEVERE || severity == bean.SEVERITY_SAME_SEVERE {
					pluginIdApplyStageMap[pluginIdApplyStage] = severity
					mandatoryPluginBlockageMap[pluginIdApplyStage] = consequence
				}
			} else {
				mandatoryPluginBlockageMap[pluginIdApplyStage] = consequence
				pluginIdApplyStageMap[pluginIdApplyStage] = bean.SEVERITY_SAME_SEVERE
			}
		}
	}
	return pluginIdApplyStageMap
}

func (impl *GlobalPolicyServiceImpl) getMatchedBranchList(branchList []*bean.BranchDto, branchValues []string) ([]string, error) {
	matchedBranchList := make([]string, 0, len(branchList)*len(branchValues))
	for _, branch := range branchList {
		for _, branchValue := range branchValues {
			isBranchMatched, err := isBranchValueMatched(branch, branchValue)
			if err != nil {
				impl.logger.Errorw("error in checking if branch value matched or not", "err", err, "branch", branch, "branchValue", branchValue)
				return nil, err
			}
			if isBranchMatched {
				matchedBranchList = append(matchedBranchList, branchValue)
			}
		}
	}
	return matchedBranchList, nil
}
func isBranchValueMatched(branch *bean.BranchDto, branchValue string) (bool, error) {
	isBranchMatched := false
	var err error
	if len(branchValue) != 0 {
		if branch.BranchValueType == bean2.VALUE_TYPE_REGEX {
			isBranchMatched, err = regexp.MatchString(branch.Value, branchValue)
			if err != nil {
				return isBranchMatched, err
			}
		} else {
			if branch.Value == branchValue {
				isBranchMatched = true
			}
		}
	} else {
		isBranchMatched = true
	}
	return isBranchMatched, nil
}

func getAppSelectorMap(projectAppList []*bean.ProjectAppDto) map[string]bool {
	appSelectorMap := make(map[string]bool)
	for _, selector := range projectAppList {
		projectName := selector.ProjectName
		if len(selector.AppNames) > 0 {
			for _, appName := range selector.AppNames {
				projectAppName := getSlashSeparatedString(projectName, appName)
				appSelectorMap[projectAppName] = true
			}
		} else {
			projectAppName := getSlashSeparatedString(projectName, bean.POLICY_ALL_OBJECTS_PLACEHOLDER)
			appSelectorMap[projectAppName] = true
		}
	}
	return appSelectorMap
}

func getEnvSelectorMap(clusterEnvList []*bean.ClusterEnvDto) map[string]bool {
	envSelectorMap := make(map[string]bool)
	for _, selector := range clusterEnvList {
		clusterName := selector.ClusterName
		if len(selector.EnvNames) > 0 {
			for _, envName := range selector.EnvNames {
				clusterEnvName := getSlashSeparatedString(clusterName, envName)
				envSelectorMap[clusterEnvName] = true
			}
		} else {
			clusterEnvName := getSlashSeparatedString(clusterName, bean.POLICY_ALL_OBJECTS_PLACEHOLDER)
			envSelectorMap[clusterEnvName] = true
		}
	}
	return envSelectorMap
}

func checkAppSelectorAndGetDefinitionSourceInfo(appSelectorMap map[string]bool, projectName, appName string) (string, string, bool) {
	projectAppName := getSlashSeparatedString(projectName, appName)
	projectAllApp := getSlashSeparatedString(projectName, bean.POLICY_ALL_OBJECTS_PLACEHOLDER)
	definitionSourceProjectName := ""
	definitionSourceAppName := ""
	toContinue := false
	if appSelectorMap[projectAppName] {
		definitionSourceProjectName = projectName
		definitionSourceAppName = appName
	} else if appSelectorMap[projectAllApp] {
		definitionSourceProjectName = projectName
	} else {
		// app is not matched and needed to be matched, continuing for next policy
		toContinue = true
	}
	return definitionSourceProjectName, definitionSourceAppName, toContinue
}

func checkEnvSelectorAndGetDefinitionSourceInfo(allProductionEnvsFlag bool, envSelectorMap map[string]bool,
	ciPipelineIdProductionEnvDetailMap, ciPipelineIdEnvDetailMap map[int][]*bean.PluginSourceCiPipelineEnvDetailDto,
	ciPipelineId int, definitionSourceTemplate bean.DefinitionSourceDto) []*bean.DefinitionSourceDto {
	var definitionSourceTemplateForEnvs []*bean.DefinitionSourceDto
	if allProductionEnvsFlag {
		allProductionEnvInCI := ciPipelineIdProductionEnvDetailMap[ciPipelineId]
		for _, productionEnvCI := range allProductionEnvInCI {
			definitionSourceTemplateForEnv := definitionSourceTemplate
			definitionSourceTemplateForEnv.IsDueToProductionEnvironment = true
			definitionSourceTemplateForEnv.EnvironmentName = productionEnvCI.EnvName
			definitionSourceTemplateForEnv.ClusterName = productionEnvCI.ClusterName
			definitionSourceTemplateForEnvs = append(definitionSourceTemplateForEnvs, &definitionSourceTemplateForEnv)
		}
	} else {
		allEnvInCI := ciPipelineIdEnvDetailMap[ciPipelineId]
		for _, envCI := range allEnvInCI {
			definitionSourceTemplateForEnv := definitionSourceTemplate
			definitionSourceTemplateForEnv.IsDueToProductionEnvironment = false
			clusterName := envCI.ClusterName
			envName := envCI.EnvName
			clusterEnvName := getSlashSeparatedString(clusterName, envName)
			clusterAllEnv := getSlashSeparatedString(clusterName, bean.POLICY_ALL_OBJECTS_PLACEHOLDER)
			if envSelectorMap[clusterEnvName] {
				definitionSourceTemplateForEnv.EnvironmentName = envCI.EnvName
				definitionSourceTemplateForEnv.ClusterName = envCI.ClusterName
			} else if envSelectorMap[clusterAllEnv] {
				definitionSourceTemplateForEnv.ClusterName = envCI.ClusterName
			} else {
				// env is not matched and needed to be matched, continuing for next environment
				continue
			}
			definitionSourceTemplateForEnvs = append(definitionSourceTemplateForEnvs, &definitionSourceTemplateForEnv)
		}
	}
	return definitionSourceTemplateForEnvs
}

func (impl *GlobalPolicyServiceImpl) getDefinitionSourceDtos(globalPolicyDetailDto bean.GlobalPolicyDetailDto, allCiPipelineIds []int,
	ciPipelineId int, ciPipelineIdProjectAppNameMap map[int]*bean.PluginSourceCiPipelineAppDetailDto,
	ciPipelineIdEnvDetailMap, ciPipelineIdProductionEnvDetailMap map[int][]*bean.PluginSourceCiPipelineEnvDetailDto,
	branchValues []string, globalPolicyName string, ciPipelineIdNameMap map[int]string) ([]*bean.DefinitionSourceDto, error) {
	branchList := globalPolicyDetailDto.Selectors.BranchList
	matchedBranchList, err := impl.getMatchedBranchList(branchList, branchValues)
	if err != nil {
		impl.logger.Errorw("error, getMatchedBranchList", "err", err, "branchList", branchList, "branchValues", branchValues)
		return nil, err
	}

	if len(branchList) > 0 && len(matchedBranchList) == 0 {
		//we have some branch configured in policy, but we have got no matches so skipping this policy
		return nil, nil
	}
	selectors := &bean.SelectorDto{}
	if globalPolicyDetailDto.Selectors != nil {
		selectors = globalPolicyDetailDto.Selectors
	}
	var projectAppList []*bean.ProjectAppDto
	var clusterEnvList []*bean.ClusterEnvDto
	var allProductionEnvsFlag bool
	if selectors != nil && selectors.ApplicationSelector != nil {
		projectAppList = globalPolicyDetailDto.Selectors.ApplicationSelector
	}
	if selectors != nil && selectors.EnvironmentSelector != nil {
		allProductionEnvsFlag = globalPolicyDetailDto.Selectors.EnvironmentSelector.AllProductionEnvironments
		if selectors.EnvironmentSelector.ClusterEnvList != nil {
			clusterEnvList = globalPolicyDetailDto.Selectors.EnvironmentSelector.ClusterEnvList
		}
	}
	needToCheckAppSelector := len(projectAppList) != 0
	needToCheckEnvSelector := allProductionEnvsFlag || (len(clusterEnvList) != 0)

	appSelectorMap := getAppSelectorMap(projectAppList)
	envSelectorMap := getEnvSelectorMap(clusterEnvList)
	var definitionSourceDtos []*bean.DefinitionSourceDto
	toContinue := false
	for _, ciPipelineIdInObj := range allCiPipelineIds {
		isLinkedCi := ciPipelineIdInObj != ciPipelineId
		definitionSourceAppName := ""
		definitionSourceProjectName := ""
		appProjectDto := ciPipelineIdProjectAppNameMap[ciPipelineIdInObj]
		projectName := appProjectDto.ProjectName
		appName := appProjectDto.AppName
		if needToCheckAppSelector {
			definitionSourceProjectName, definitionSourceAppName, toContinue = checkAppSelectorAndGetDefinitionSourceInfo(appSelectorMap, projectName, appName)
			if toContinue { // app is not matched and needed to be matched, skipping this policy
				continue
			}
		}
		definitionSourceTemplate := bean.DefinitionSourceDto{
			ProjectName:           definitionSourceProjectName,
			AppName:               definitionSourceAppName,
			BranchNames:           matchedBranchList,
			IsDueToLinkedPipeline: isLinkedCi,
			CiPipelineName:        ciPipelineIdNameMap[ciPipelineIdInObj],
			PolicyName:            globalPolicyName,
		}

		if needToCheckEnvSelector {
			definitionSourceTemplateForEnvs := checkEnvSelectorAndGetDefinitionSourceInfo(allProductionEnvsFlag, envSelectorMap,
				ciPipelineIdProductionEnvDetailMap, ciPipelineIdEnvDetailMap, ciPipelineIdInObj, definitionSourceTemplate)
			if len(definitionSourceTemplateForEnvs) > 0 {
				definitionSourceDtos = append(definitionSourceDtos, definitionSourceTemplateForEnvs...)
			} else { //no env matched but needed to be matched, skipping this policy
				continue
			}
		} else { //environment is not needed to be checked in policy, then only enforce on basis of app/branch
			definitionSourceDtos = append(definitionSourceDtos, &definitionSourceTemplate)
		}
	}
	return definitionSourceDtos, nil
}

func getMandatoryPluginDefinition(pluginIdStage string, definitionSources []*bean.DefinitionSourceDto) (*bean.MandatoryPluginDefinitionDto, error) {
	vals := strings.Split(pluginIdStage, "/")
	pluginId, err := strconv.Atoi(vals[0])
	if err != nil {
		return nil, err
	}
	applyToStage := vals[1]
	return &bean.MandatoryPluginDefinitionDto{
		DefinitionDto: &bean.DefinitionDto{
			AttributeType: bean2.DEVTRON_RESOURCE_ATTRIBUTE_TYPE_PLUGIN,
			Data: bean.DefinitionDataDto{
				PluginId:     pluginId,
				ApplyToStage: bean.PluginApplyStage(applyToStage),
			},
		},
		DefinitionSources: definitionSources,
	}, nil
}

func updateMandatoryPluginDefinitionMap(pluginIdApplyStageMap map[string]bean.Severity,
	mandatoryPluginDefinitionMap map[string][]*bean.DefinitionSourceDto, definitionSourceDtos []*bean.DefinitionSourceDto) {
	for pluginIdStage, severity := range pluginIdApplyStageMap {
		if severity == bean.SEVERITY_SAME_SEVERE {
			mandatoryPluginDefinitionMap[pluginIdStage] = append(mandatoryPluginDefinitionMap[pluginIdStage], definitionSourceDtos...)
		} else if severity == bean.SEVERITY_MORE_SEVERE {
			mandatoryPluginDefinitionMap[pluginIdStage] = definitionSourceDtos
		}
	}
}

func checkIfConsequenceIsBlocking(consequence *bean.ConsequenceDto) bool {
	isBlocking := true
	minimumSevereConsequenceToBlock := &bean.ConsequenceDto{
		Action:        bean.CONSEQUENCE_ACTION_ALLOW_UNTIL_TIME,
		MetadataField: time.Now(),
	}
	sev := minimumSevereConsequenceToBlock.GetSeverity(consequence)
	if sev != bean.SEVERITY_MORE_SEVERE { //if time for allowing is not passed then skip this policy for getting mandatory plugins in blockage state
		isBlocking = false
	}
	return isBlocking
}

func getAllAppEnvBranchDetailsFromGlobalPolicyDetail(globalPolicyDetail *bean.GlobalPolicyDetailDto) ([]string, []string, []*bean.BranchDto, bool, bool, map[string]bool, map[string]bool) {
	allProjects, allClusters := make([]string, 0), make([]string, 0)
	branchList := make([]*bean.BranchDto, 0)
	isProductionEnvFlag, isAnyEnvSelectorPresent := false, false
	projectAppNamePolicyIdsMap, clusterEnvNamePolicyIdsMap := make(map[string]bool, 0), make(map[string]bool, 0)
	selectors := &bean.SelectorDto{}
	if globalPolicyDetail != nil && globalPolicyDetail.Selectors != nil {
		selectors = globalPolicyDetail.Selectors
	}
	applicationSelectors := make([]*bean.ProjectAppDto, 0)
	if selectors != nil {
		applicationSelectors = selectors.ApplicationSelector
		for _, applicationSelector := range applicationSelectors {
			projectName := applicationSelector.ProjectName
			allProjects = append(allProjects, projectName)
			if len(applicationSelector.AppNames) == 0 {
				projectAppName := getSlashSeparatedString(projectName, bean.POLICY_ALL_OBJECTS_PLACEHOLDER)
				projectAppNamePolicyIdsMap[projectAppName] = true
			} else {
				for _, appName := range applicationSelector.AppNames {
					projectAppName := getSlashSeparatedString(projectName, appName)
					projectAppNamePolicyIdsMap[projectAppName] = true
				}
			}
		}

		if selectors.EnvironmentSelector != nil {
			isProductionEnvFlag = selectors.EnvironmentSelector.AllProductionEnvironments
			clusterEnvList := selectors.EnvironmentSelector.ClusterEnvList
			isAnyEnvSelectorPresent = isProductionEnvFlag || (len(clusterEnvList) != 0)
			for _, clusterEnv := range clusterEnvList {
				clusterName := clusterEnv.ClusterName
				allClusters = append(allClusters, clusterName)
				if len(clusterEnv.EnvNames) == 0 {
					clusterEnvName := getSlashSeparatedString(clusterName, bean.POLICY_ALL_OBJECTS_PLACEHOLDER)
					clusterEnvNamePolicyIdsMap[clusterEnvName] = true
				} else {
					for _, envName := range clusterEnv.EnvNames {
						clusterEnvName := getSlashSeparatedString(clusterName, envName)
						clusterEnvNamePolicyIdsMap[clusterEnvName] = true
					}
				}

			}
		}
		branchList = selectors.BranchList
	}

	return allProjects, allClusters, branchList, isProductionEnvFlag, isAnyEnvSelectorPresent, projectAppNamePolicyIdsMap, clusterEnvNamePolicyIdsMap
}

func (impl *GlobalPolicyServiceImpl) findAllWorkflowsComponentDetailsForCiPipelineIds(ciPipelineIds []int,
	ciPipelineMaterialMap map[int][]*pipelineConfig.CiPipelineMaterial) ([]*bean.WorkflowTreeComponentDto, error) {
	var appWorkflowMappings []*appWorkflow.AppWorkflowMapping
	var err error
	if len(ciPipelineIds) > 0 {
		appWorkflowMappings, err = impl.appWorkflowRepository.FindMappingsOfWfWithSpecificCIPipelineIds(ciPipelineIds)
		if err != nil {
			impl.logger.Errorw("error in getting appWorkflowMappings by ciPipeline Ids", "err", err, "ciPipelineIds", ciPipelineIds)
			return nil, err
		}
	}
	var wfComponentDetails []*bean.WorkflowTreeComponentDto
	wfIdAndComponentDtoIndexMap := make(map[int]int)
	var cdPipelineIds []int
	var appIds []int
	for _, appWfMapping := range appWorkflowMappings {
		appId := appWfMapping.AppWorkflow.AppId
		appIds = append(appIds, appId)
		if appWfMapping.Type == appWorkflow.CDPIPELINE {
			cdPipelineIds = append(cdPipelineIds, appWfMapping.ComponentId)
		}
		appWorkflowId := appWfMapping.AppWorkflowId
		if _, ok := wfIdAndComponentDtoIndexMap[appWorkflowId]; !ok {
			wfComponentDetail := &bean.WorkflowTreeComponentDto{
				Id:    appWorkflowId,
				AppId: appId,
				Name:  appWfMapping.AppWorkflow.Name,
			}
			wfComponentDetails = append(wfComponentDetails, wfComponentDetail)
			wfIdAndComponentDtoIndexMap[appWorkflowId] = len(wfComponentDetails) - 1 //index of whComponentDetail latest addition
		}
	}
	appIdGitMaterialMap := make(map[int][]*bean.Material)
	if len(appIds) > 0 {
		gitMaterials, err := impl.gitMaterialRepository.FindByAppIds(appIds)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in fetching git materials by appIds", "err", err, "appIds", appIds)
			return nil, err
		}
		for _, g := range gitMaterials {
			material := &bean.Material{
				GitMaterialId: g.Id,
				MaterialName:  g.Name[strings.Index(g.Name, "-")+1:],
			}
			appIdGitMaterialMap[g.AppId] = append(appIdGitMaterialMap[g.AppId], material)
		}
	}

	var ciPipelines []*pipelineConfig.CiPipeline
	if len(ciPipelineIds) > 0 { //getting all ciPipelines by ids
		ciPipelines, err = impl.ciPipelineRepository.FindByIdsIn(ciPipelineIds)
		if err != nil {
			impl.logger.Errorw("error in getting ciPipelines by ids", "err", err, "ids", ciPipelineIds)
			return nil, err
		}
	}
	ciPipelineIdNameMap := make(map[int]string, len(ciPipelines))
	for _, ciPipeline := range ciPipelines {
		ciPipelineIdNameMap[ciPipeline.Id] = ciPipeline.Name
	}
	var cdPipelines []*pipelineConfig.Pipeline
	//getting all ciPipelines by appId
	if len(cdPipelineIds) > 0 {
		cdPipelines, err = impl.pipelineRepository.FindByIdsIn(cdPipelineIds)
		if err != nil {
			impl.logger.Errorw("error in getting cdPipelines by ids", "err", err, "ids", cdPipelineIds)
			return nil, err
		}
	}
	cdPipelineIdNameMap := make(map[int]string, len(cdPipelines))
	for _, cdPipeline := range cdPipelines {
		cdPipelineIdNameMap[cdPipeline.Id] = cdPipeline.Environment.Name
	}

	for _, appWfMapping := range appWorkflowMappings {
		if index, ok := wfIdAndComponentDtoIndexMap[appWfMapping.AppWorkflowId]; ok {
			wfComponentDetails[index].GitMaterials = appIdGitMaterialMap[appWfMapping.AppWorkflow.AppId]

			if appWfMapping.Type == appWorkflow.CIPIPELINE {
				ciPipelineId := appWfMapping.ComponentId
				//getting all materials from map for this ci pipeline
				ciPipelineMaterials := ciPipelineMaterialMap[ciPipelineId]
				wfComponentDetails[index].CiPipelineId = ciPipelineId
				wfComponentDetails[index].CiMaterials = ciPipelineMaterials
				if name, ok1 := ciPipelineIdNameMap[ciPipelineId]; ok1 {
					wfComponentDetails[index].CiPipelineName = name
				}
			} else if appWfMapping.Type == appWorkflow.CDPIPELINE {
				if envName, ok1 := cdPipelineIdNameMap[appWfMapping.ComponentId]; ok1 {
					wfComponentDetails[index].CdPipelines = append(wfComponentDetails[index].CdPipelines, envName)
				}
			}
		}
	}
	return wfComponentDetails, nil
}

func (impl *GlobalPolicyServiceImpl) getGlobalPolicyDtos(globalPolicies []*repository.GlobalPolicy) ([]*bean.GlobalPolicyDto, error) {
	globalPolicyDtos := make([]*bean.GlobalPolicyDto, 0, len(globalPolicies))
	for _, globalPolicy := range globalPolicies {
		globalPolicyDto, err := globalPolicy.GetGlobalPolicyDto()
		if err != nil {
			impl.logger.Errorw("error in getting globalPolicyDto", "err", err, "policyId", globalPolicy.Id)
			return nil, err
		}
		globalPolicyDtos = append(globalPolicyDtos, globalPolicyDto)
	}
	return globalPolicyDtos, nil
}

func getFilteredCiPipelinesByProjectAppObjs(ciPipelineProjectAppNameObjs []*pipelineConfig.CiPipelineAppProject,
	projectAppNameMap map[string]bool) []int {
	ciPipelinesToBeFiltered := make([]int, 0, len(ciPipelineProjectAppNameObjs))
	for _, ciPipelineProjectAppNameObj := range ciPipelineProjectAppNameObjs {
		ciPipelineId := ciPipelineProjectAppNameObj.CiPipelineId
		projectAppName := getSlashSeparatedString(ciPipelineProjectAppNameObj.ProjectName, ciPipelineProjectAppNameObj.AppName)
		projectAllApp := getSlashSeparatedString(ciPipelineProjectAppNameObj.ProjectName, bean.POLICY_ALL_OBJECTS_PLACEHOLDER)
		if projectAppNameMap[projectAllApp] || projectAppNameMap[projectAppName] {
			ciPipelinesToBeFiltered = append(ciPipelinesToBeFiltered, ciPipelineId)
		}
	}
	return ciPipelinesToBeFiltered
}

func getFilteredCiPipelinesByClusterAndEnvObjs(ciPipelineClusterEnvNameObjs []*pipelineConfig.CiPipelineEnvCluster,
	isProductionEnvFlag bool, clusterEnvNameMap map[string]bool) []int {
	ciPipelinesToBeFiltered := make([]int, 0, len(ciPipelineClusterEnvNameObjs))
	for _, ciPipelineClusterEnvNameObj := range ciPipelineClusterEnvNameObjs {
		clusterEnvName := getSlashSeparatedString(ciPipelineClusterEnvNameObj.ClusterName, ciPipelineClusterEnvNameObj.EnvName)
		clusterAllEnv := getSlashSeparatedString(ciPipelineClusterEnvNameObj.ClusterName, bean.POLICY_ALL_OBJECTS_PLACEHOLDER)
		isEntryMatched := (ciPipelineClusterEnvNameObj.IsProductionEnv == isProductionEnvFlag) || clusterEnvNameMap[clusterEnvName] || clusterEnvNameMap[clusterAllEnv]
		if isEntryMatched {
			ciPipelinesToBeFiltered = append(ciPipelinesToBeFiltered, ciPipelineClusterEnvNameObj.CiPipelineId)
		}
	}
	return ciPipelinesToBeFiltered
}
