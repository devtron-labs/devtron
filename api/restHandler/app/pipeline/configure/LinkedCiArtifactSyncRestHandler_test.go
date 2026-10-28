/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package configure

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	appWorkflowRepo "github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	appWorkflowService "github.com/devtron-labs/devtron/pkg/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/bean"
	buildBean "github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeSyncUserService struct {
	user.UserService
	userId int32
}

func (f *fakeSyncUserService) GetLoggedInUser(r *http.Request) (int32, error) {
	if f.userId == 0 {
		return 0, errors.New("no user")
	}
	return f.userId, nil
}

// fakeSyncEnforcer allows "resource|action|object" keys present in allowed
type fakeSyncEnforcer struct {
	casbin.Enforcer
	allowed map[string]bool
}

func (f *fakeSyncEnforcer) Enforce(token, resource, action, object string) bool {
	return f.allowed[resource+"|"+action+"|"+object]
}
func (f *fakeSyncEnforcer) EnforceInBatch(token, resource, action string, vals []string) map[string]bool {
	result := make(map[string]bool, len(vals))
	for _, v := range vals {
		result[v] = f.allowed[resource+"|"+action+"|"+v]
	}
	return result
}

type fakeSyncEnforcerUtil struct{ rbac.EnforcerUtil }

func (f *fakeSyncEnforcerUtil) GetAppRBACName(appName string) string { return "team/" + appName }
func (f *fakeSyncEnforcerUtil) GetRbacObjectsByEnvIdsAndAppId(envIds []int, appId int) (map[int]string, map[string]string) {
	objects := make(map[int]string)
	for _, envId := range envIds {
		if envId > 0 {
			objects[envId] = fmt.Sprintf("env-%d/app-%d", envId, appId)
		}
	}
	return objects, nil
}

type fakeSyncCiPipelineRepo struct {
	pipelineConfig.CiPipelineRepository
	pipelines map[int]*pipelineConfig.CiPipeline
	err       error
}

func (f *fakeSyncCiPipelineRepo) FindOneWithMinData(id int) (*pipelineConfig.CiPipeline, error) {
	if f.err != nil {
		return nil, f.err
	}
	if p, ok := f.pipelines[id]; ok {
		return p, nil
	}
	return nil, pg.ErrNoRows
}

// fakeSyncWorkflowService maps ci pipeline -> workflow -> cd pipeline ids
type fakeSyncWorkflowService struct {
	appWorkflowService.AppWorkflowService
	workflowOfCi map[int]int
	cdsOfWf      map[int][]int
	err          error
}

func (f *fakeSyncWorkflowService) FindWFMappingByComponent(componentType string, componentId int) (*appWorkflowRepo.AppWorkflowMapping, error) {
	if f.err != nil {
		return nil, f.err
	}
	wfId, ok := f.workflowOfCi[componentId]
	if !ok || componentType != appWorkflowRepo.CIPIPELINE {
		return nil, pg.ErrNoRows
	}
	return &appWorkflowRepo.AppWorkflowMapping{AppWorkflowId: wfId}, nil
}
func (f *fakeSyncWorkflowService) FindWFCDMappingsByWorkflowId(appWorkflowId int) ([]*appWorkflowRepo.AppWorkflowMapping, error) {
	var mappings []*appWorkflowRepo.AppWorkflowMapping
	for _, cdId := range f.cdsOfWf[appWorkflowId] {
		mappings = append(mappings, &appWorkflowRepo.AppWorkflowMapping{ComponentId: cdId})
	}
	return mappings, nil
}

type fakeSyncPipelineRepo struct {
	pipelineConfig.PipelineRepository
}

func (f *fakeSyncPipelineRepo) FindByIdsIn(ids []int) ([]*pipelineConfig.Pipeline, error) {
	var pipelines []*pipelineConfig.Pipeline
	for _, id := range ids {
		pipelines = append(pipelines, &pipelineConfig.Pipeline{Id: id, AppId: 500, EnvironmentId: id * 10})
	}
	return pipelines, nil
}

type fakeSyncOrchestrator struct {
	pipeline.CiCdPipelineOrchestrator
	syncErr         error
	syncedIds       []int
	syncedUserIds   []int32
	children        []*pipelineConfig.CiPipeline
	bulkErr         error
	bulkCalledFor   []int
	accessDecisions map[int]bool
}

func (f *fakeSyncOrchestrator) SyncLinkedCiArtifacts(ciPipelineId int, userId int32) (*buildBean.LinkedCiArtifactSyncResponse, error) {
	f.syncedIds = append(f.syncedIds, ciPipelineId)
	f.syncedUserIds = append(f.syncedUserIds, userId)
	if f.syncErr != nil {
		return nil, f.syncErr
	}
	return &buildBean.LinkedCiArtifactSyncResponse{CiPipelineId: ciPipelineId, ParentCiPipelineId: 1, AppId: 5, AppName: "child", CopiedCount: 4}, nil
}
func (f *fakeSyncOrchestrator) SyncLinkedCiArtifactsForSource(sourceCiPipelineId int, userId int32, hasAccess func(*pipelineConfig.CiPipeline) bool) (*buildBean.LinkedCiArtifactBulkSyncResponse, error) {
	f.bulkCalledFor = append(f.bulkCalledFor, sourceCiPipelineId)
	if f.bulkErr != nil {
		return nil, f.bulkErr
	}
	f.accessDecisions = make(map[int]bool)
	resp := &buildBean.LinkedCiArtifactBulkSyncResponse{SourceCiPipelineId: sourceCiPipelineId, Results: []*buildBean.LinkedCiArtifactSyncResponse{}}
	for _, child := range f.children {
		allowed := hasAccess(child)
		f.accessDecisions[child.Id] = allowed
		if !allowed {
			resp.UnauthorizedCount++
			continue
		}
		resp.Results = append(resp.Results, &buildBean.LinkedCiArtifactSyncResponse{CiPipelineId: child.Id, CopiedCount: 2})
	}
	return resp, nil
}

type syncHandlerFixture struct {
	userId       int32
	allowed      map[string]bool
	pipelines    map[int]*pipelineConfig.CiPipeline
	repoErr      error
	workflowOfCi map[int]int
	cdsOfWf      map[int][]int
	workflowErr  error
	orchestrator *fakeSyncOrchestrator
}

func (fx *syncHandlerFixture) handler() *PipelineConfigRestHandlerImpl {
	if fx.orchestrator == nil {
		fx.orchestrator = &fakeSyncOrchestrator{}
	}
	return &PipelineConfigRestHandlerImpl{
		Logger:                   zap.NewNop().Sugar(),
		userAuthService:          &fakeSyncUserService{userId: fx.userId},
		enforcer:                 &fakeSyncEnforcer{allowed: fx.allowed},
		enforcerUtil:             &fakeSyncEnforcerUtil{},
		ciPipelineRepository:     &fakeSyncCiPipelineRepo{pipelines: fx.pipelines, err: fx.repoErr},
		appWorkflowService:       &fakeSyncWorkflowService{workflowOfCi: fx.workflowOfCi, cdsOfWf: fx.cdsOfWf, err: fx.workflowErr},
		pipelineRepository:       &fakeSyncPipelineRepo{},
		ciCdPipelineOrchestrator: fx.orchestrator,
	}
}

func init() {
	_, _ = util.InitLogger() // the shared response writer logs through the global logger
}

func callSyncHandler(h http.HandlerFunc, ciPipelineId string) (int, map[string]interface{}) {
	r := httptest.NewRequest(http.MethodPost, "/orchestrator/app/ci-pipeline/"+ciPipelineId+"/linked-ci/sync-artifacts", nil)
	r.Header.Set("token", "t")
	r = mux.SetURLVars(r, map[string]string{"ciPipelineId": ciPipelineId})
	w := httptest.NewRecorder()
	h(w, r)
	body := map[string]interface{}{}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return w.Code, body
}

func syncTestPipeline(id int, appName string) *pipelineConfig.CiPipeline {
	return &pipelineConfig.CiPipeline{Id: id, AppId: id * 100, ParentCiPipeline: 1, PipelineType: "LINKED", Active: true, App: &app.App{AppName: appName}}
}

const (
	appAdmin = casbin.ResourceApplications + "|" + casbin.ActionCreate + "|team/"
	appView  = casbin.ResourceApplications + "|" + casbin.ActionGet + "|team/"
	envEdit  = casbin.ResourceEnvironment + "|" + casbin.ActionUpdate + "|"
)

func TestSyncLinkedCiArtifactsHandler(t *testing.T) {
	tests := []struct {
		name          string
		fx            syncHandlerFixture
		id            string
		wantStatus    int
		wantSynced    bool
		wantResultIds bool
	}{
		{name: "no user", fx: syncHandlerFixture{}, id: "7", wantStatus: http.StatusUnauthorized},
		{name: "non numeric id", fx: syncHandlerFixture{userId: 2}, id: "abc", wantStatus: http.StatusBadRequest},
		{name: "pipeline not found", fx: syncHandlerFixture{userId: 2}, id: "7", wantStatus: http.StatusNotFound},
		{name: "pipeline lookup fails", fx: syncHandlerFixture{userId: 2, repoErr: errors.New("db")}, id: "7", wantStatus: http.StatusInternalServerError},
		{name: "no CD in workflow and not app admin", fx: syncHandlerFixture{userId: 2, pipelines: map[int]*pipelineConfig.CiPipeline{7: syncTestPipeline(7, "child")},
			allowed: map[string]bool{appView + "child": true}}, id: "7", wantStatus: http.StatusForbidden},
		{name: "no CD in workflow and app admin", fx: syncHandlerFixture{userId: 2, pipelines: map[int]*pipelineConfig.CiPipeline{7: syncTestPipeline(7, "child")},
			allowed: map[string]bool{appAdmin + "child": true}}, id: "7", wantStatus: http.StatusOK, wantSynced: true, wantResultIds: true},
		{name: "workflow has CDs and user edits one of their envs", fx: syncHandlerFixture{userId: 2, pipelines: map[int]*pipelineConfig.CiPipeline{7: syncTestPipeline(7, "child")},
			workflowOfCi: map[int]int{7: 70}, cdsOfWf: map[int][]int{70: {3, 4}}, allowed: map[string]bool{envEdit + "env-40/app-500": true}}, id: "7", wantStatus: http.StatusOK, wantSynced: true, wantResultIds: true},
		{name: "workflow has CDs and user edits none of their envs", fx: syncHandlerFixture{userId: 2, pipelines: map[int]*pipelineConfig.CiPipeline{7: syncTestPipeline(7, "child")},
			workflowOfCi: map[int]int{7: 70}, cdsOfWf: map[int][]int{70: {3, 4}}, allowed: map[string]bool{appView + "child": true, envEdit + "env-99/app-500": true}}, id: "7", wantStatus: http.StatusForbidden},
		{name: "workflow lookup error fails closed", fx: syncHandlerFixture{userId: 2, pipelines: map[int]*pipelineConfig.CiPipeline{7: syncTestPipeline(7, "child")},
			workflowErr: errors.New("db"), allowed: map[string]bool{appAdmin + "child": true}}, id: "7", wantStatus: http.StatusForbidden},
		{name: "pipeline without app fails closed", fx: syncHandlerFixture{userId: 2, pipelines: map[int]*pipelineConfig.CiPipeline{7: {Id: 7}},
			allowed: map[string]bool{appAdmin: true}}, id: "7", wantStatus: http.StatusForbidden},
		{name: "service says not linked", fx: syncHandlerFixture{userId: 2, pipelines: map[int]*pipelineConfig.CiPipeline{7: syncTestPipeline(7, "child")},
			allowed: map[string]bool{appAdmin + "child": true}, orchestrator: &fakeSyncOrchestrator{syncErr: util.NewApiError(http.StatusBadRequest, "not linked", "not linked")}}, id: "7", wantStatus: http.StatusBadRequest, wantSynced: true},
		{name: "service says not found", fx: syncHandlerFixture{userId: 2, pipelines: map[int]*pipelineConfig.CiPipeline{7: syncTestPipeline(7, "child")},
			allowed: map[string]bool{appAdmin + "child": true}, orchestrator: &fakeSyncOrchestrator{syncErr: util.NewApiError(http.StatusNotFound, "gone", "gone")}}, id: "7", wantStatus: http.StatusNotFound, wantSynced: true},
		{name: "service fails", fx: syncHandlerFixture{userId: 2, pipelines: map[int]*pipelineConfig.CiPipeline{7: syncTestPipeline(7, "child")},
			allowed: map[string]bool{appAdmin + "child": true}, orchestrator: &fakeSyncOrchestrator{syncErr: errors.New("boom")}}, id: "7", wantStatus: http.StatusInternalServerError, wantSynced: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.fx.handler()
			status, body := callSyncHandler(h.SyncLinkedCiArtifacts, tt.id)
			assert.Equal(t, tt.wantStatus, status)
			if tt.wantSynced {
				assert.Equal(t, []int{7}, tt.fx.orchestrator.syncedIds)
				assert.Equal(t, []int32{2}, tt.fx.orchestrator.syncedUserIds)
			} else {
				assert.Empty(t, tt.fx.orchestrator.syncedIds, "nothing is synced")
			}
			if tt.wantResultIds {
				result, ok := body["result"].(map[string]interface{})
				require.True(t, ok, "body: %v", body)
				assert.EqualValues(t, 7, result["ciPipelineId"])
				assert.EqualValues(t, 1, result["parentCiPipelineId"])
				assert.EqualValues(t, 5, result["appId"])
				assert.Equal(t, "child", result["appName"])
				assert.EqualValues(t, 4, result["copiedCount"])
				_, hasError := result["error"]
				assert.False(t, hasError, "error is omitted on success")
			}
		})
	}
}

func TestSyncLinkedCiArtifactsForSourceHandler(t *testing.T) {
	source := syncTestPipeline(1, "source")
	children := []*pipelineConfig.CiPipeline{syncTestPipeline(11, "admin-child"), syncTestPipeline(12, "locked-child"), syncTestPipeline(13, "env-child"), {Id: 14}}
	tests := []struct {
		name       string
		fx         syncHandlerFixture
		id         string
		wantStatus int
		wantBulk   bool
	}{
		{name: "no user", fx: syncHandlerFixture{}, id: "1", wantStatus: http.StatusUnauthorized},
		{name: "non numeric id", fx: syncHandlerFixture{userId: 2}, id: "x", wantStatus: http.StatusBadRequest},
		{name: "source not found", fx: syncHandlerFixture{userId: 2}, id: "1", wantStatus: http.StatusNotFound},
		{name: "source lookup fails", fx: syncHandlerFixture{userId: 2, repoErr: errors.New("db")}, id: "1", wantStatus: http.StatusInternalServerError},
		{name: "cannot view source", fx: syncHandlerFixture{userId: 2, pipelines: map[int]*pipelineConfig.CiPipeline{1: source},
			allowed: map[string]bool{appAdmin + "admin-child": true}}, id: "1", wantStatus: http.StatusForbidden},
		{name: "service fails", fx: syncHandlerFixture{userId: 2, pipelines: map[int]*pipelineConfig.CiPipeline{1: source},
			allowed: map[string]bool{appView + "source": true}, orchestrator: &fakeSyncOrchestrator{bulkErr: errors.New("boom")}}, id: "1", wantStatus: http.StatusInternalServerError, wantBulk: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.fx.handler()
			status, _ := callSyncHandler(h.SyncLinkedCiArtifactsForSource, tt.id)
			assert.Equal(t, tt.wantStatus, status)
			if tt.wantBulk {
				assert.Equal(t, []int{1}, tt.fx.orchestrator.bulkCalledFor)
			} else {
				assert.Empty(t, tt.fx.orchestrator.bulkCalledFor)
			}
		})
	}

	t.Run("checks each child with the CI edit rule", func(t *testing.T) {
		fx := syncHandlerFixture{
			userId:       2,
			pipelines:    map[int]*pipelineConfig.CiPipeline{1: source},
			workflowOfCi: map[int]int{13: 130},
			cdsOfWf:      map[int][]int{130: {6}},
			allowed: map[string]bool{
				appView + "source":         true,
				appAdmin + "admin-child":   true,
				appView + "locked-child":   true,
				envEdit + "env-60/app-500": true,
			},
			orchestrator: &fakeSyncOrchestrator{children: children},
		}
		status, body := callSyncHandler(fx.handler().SyncLinkedCiArtifactsForSource, "1")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, map[int]bool{11: true, 12: false, 13: true, 14: false}, fx.orchestrator.accessDecisions)
		result := body["result"].(map[string]interface{})
		assert.EqualValues(t, 1, result["sourceCiPipelineId"])
		assert.EqualValues(t, 2, result["unauthorizedCount"])
		assert.Len(t, result["results"], 2)
	})
}

// regression: the CI edit RBAC lookup was split into a reusable helper
func TestGetCdPipelinesForCIPatchRbac_Unchanged(t *testing.T) {
	fx := syncHandlerFixture{workflowOfCi: map[int]int{7: 70, 8: 80}, cdsOfWf: map[int][]int{70: {3, 4}}}
	h := fx.handler()
	ids := func(p []*pipelineConfig.Pipeline) []int {
		out := []int{}
		for _, x := range p {
			out = append(out, x.Id)
		}
		return out
	}
	tests := []struct {
		name    string
		req     bean.CiPatchRequest
		wantIds []int
		wantErr bool
	}{
		{name: "create has no CDs", req: bean.CiPatchRequest{Action: bean.CREATE, CiPipeline: &bean.CiPipeline{Id: 7}}, wantIds: []int{}},
		{name: "update finds the workflow's CDs", req: bean.CiPatchRequest{Action: bean.UPDATE_SOURCE, CiPipeline: &bean.CiPipeline{Id: 7}}, wantIds: []int{3, 4}},
		{name: "update of a workflow without CDs", req: bean.CiPatchRequest{Action: bean.UPDATE_SOURCE, CiPipeline: &bean.CiPipeline{Id: 8}}, wantIds: []int{}},
		{name: "switch uses the source pipeline", req: bean.CiPatchRequest{Action: bean.CREATE, CiPipeline: &bean.CiPipeline{Id: 99}, SwitchFromCiPipelineId: 7}, wantIds: []int{3, 4}},
		{name: "missing workflow mapping is an error", req: bean.CiPatchRequest{Action: bean.UPDATE_SOURCE, CiPipeline: &bean.CiPipeline{Id: 5}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.getCdPipelinesForCIPatchRbac(&tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantIds, ids(got))
		})
	}
}
