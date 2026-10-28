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

package pipeline

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/build/pipeline/bean/common"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeLinkedCiTxManager struct {
	startErr, commitErr          error
	started, committed, rolledBk int
}

func (f *fakeLinkedCiTxManager) StartTx() (*pg.Tx, error) {
	f.started++
	return nil, f.startErr
}
func (f *fakeLinkedCiTxManager) RollbackTx(tx *pg.Tx) error { f.rolledBk++; return nil }
func (f *fakeLinkedCiTxManager) CommitTx(tx *pg.Tx) error {
	f.committed++
	return f.commitErr
}

type fakeLinkedCiArtifactRepo struct {
	repository.CiArtifactRepository
	lockErr, oldestErr, latestErr, saveErr error
	oldestParentId                         int
	parentArtifacts                        []repository.CiArtifact
	saved                                  []*repository.CiArtifact
	lockedIds, oldestIds, latestIds        []int
	latestLimits                           []int
	failSaveForChild                       int
}

func (f *fakeLinkedCiArtifactRepo) LockLinkedCiArtifactCopy(tx *pg.Tx, ciPipelineId int) error {
	f.lockedIds = append(f.lockedIds, ciPipelineId)
	return f.lockErr
}
func (f *fakeLinkedCiArtifactRepo) GetOldestParentCiArtifactId(tx *pg.Tx, ciPipelineId int) (int, error) {
	f.oldestIds = append(f.oldestIds, ciPipelineId)
	return f.oldestParentId, f.oldestErr
}
func (f *fakeLinkedCiArtifactRepo) GetLatestOwnBuildArtifactsByCiPipelineId(tx *pg.Tx, ciPipelineId, limit int) ([]repository.CiArtifact, error) {
	f.latestIds = append(f.latestIds, ciPipelineId)
	f.latestLimits = append(f.latestLimits, limit)
	return f.parentArtifacts, f.latestErr
}
func (f *fakeLinkedCiArtifactRepo) SaveAllInTx(tx *pg.Tx, artifacts []*repository.CiArtifact) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	if f.failSaveForChild > 0 && len(artifacts) > 0 && artifacts[0].PipelineId == f.failSaveForChild {
		return errors.New("save failed")
	}
	f.saved = append(f.saved, artifacts...)
	return nil
}

type fakeLinkedCiPipelineRepo struct {
	pipelineConfig.CiPipelineRepository
	pipeline    *pipelineConfig.CiPipeline
	findErr     error
	children    []*pipelineConfig.CiPipeline
	childrenErr error
}

func (f *fakeLinkedCiPipelineRepo) FindOneWithMinData(id int) (*pipelineConfig.CiPipeline, error) {
	return f.pipeline, f.findErr
}
func (f *fakeLinkedCiPipelineRepo) FindLinkedCiByParentCiPipelineId(parentCiPipelineId int) ([]*pipelineConfig.CiPipeline, error) {
	return f.children, f.childrenErr
}

func newLinkedCiTestOrchestrator(artifactRepo *fakeLinkedCiArtifactRepo, pipelineRepo *fakeLinkedCiPipelineRepo, txManager *fakeLinkedCiTxManager, limit int) *CiCdPipelineOrchestratorImpl {
	return &CiCdPipelineOrchestratorImpl{
		logger:               zap.NewNop().Sugar(),
		CiArtifactRepository: artifactRepo,
		ciPipelineRepository: pipelineRepo,
		transactionManager:   txManager,
		globalEnvVariables:   &util2.GlobalEnvVariables{LinkedCiArtifactCopyLimit: limit},
	}
}

// parentBuilds returns own-build rows for the given ids, newest first like the repository
func parentBuilds(ids ...int) []repository.CiArtifact {
	artifacts := make([]repository.CiArtifact, 0, len(ids))
	for i := len(ids) - 1; i >= 0; i-- {
		artifacts = append(artifacts, repository.CiArtifact{Id: ids[i], Image: "img:" + string(rune('a'+i))})
	}
	return artifacts
}

func copiedParentIds(artifacts []*repository.CiArtifact) []int {
	ids := make([]int, 0, len(artifacts))
	for _, a := range artifacts {
		ids = append(ids, a.ParentCiArtifact)
	}
	return ids
}

func TestBuildLinkedCiArtifactCopies_Window(t *testing.T) {
	tests := []struct {
		name           string
		parent         []repository.CiArtifact
		oldestParentId int
		want           []int
	}{
		{"no parent images", nil, 0, []int{}},
		{"no parent images, child has images", nil, 50, []int{}},
		{"empty child copies all, oldest first", parentBuilds(10, 20, 30), 0, []int{10, 20, 30}},
		{"child holds newer images, copies only older ones", parentBuilds(10, 20, 30, 40), 30, []int{10, 20}},
		{"boundary: the child's oldest image itself is not copied again", parentBuilds(10, 20, 30), 20, []int{10}},
		{"child already has everything in the window", parentBuilds(10, 20, 30), 10, []int{}},
		{"child older than the whole window", parentBuilds(10, 20, 30), 5, []int{}},
		{"child newer than the whole window", parentBuilds(10, 20, 30), 99, []int{10, 20, 30}},
		{"unsorted input still inserts oldest first", []repository.CiArtifact{{Id: 20}, {Id: 40}, {Id: 10}, {Id: 30}}, 0, []int{10, 20, 30, 40}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildLinkedCiArtifactCopies(tt.parent, tt.oldestParentId, 7, 1)
			assert.Equal(t, tt.want, copiedParentIds(got))
		})
	}
}

func TestBuildLinkedCiArtifactCopies_FieldMapping(t *testing.T) {
	builtAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	wfId := 99
	parent := repository.CiArtifact{
		Id: 11, Image: "repo/app:v1", ImageDigest: "sha256:abc", MaterialInfo: `[{"x":1}]`, DataSource: "CI-RUNNER",
		PipelineId: 3, WorkflowId: &wfId, IsArtifactUploaded: true, ScanEnabled: true, Scanned: true,
		TargetPlatforms: "linux/amd64", CredentialsSourceType: "global_container_registry", CredentialSourceValue: "reg",
		ComponentId: 3, ExternalCiPipelineId: 4,
	}
	parent.CreatedOn = builtAt
	before := time.Now()
	got := buildLinkedCiArtifactCopies([]repository.CiArtifact{parent}, 0, 7, 42)
	require.Len(t, got, 1)
	c := got[0]
	assert.Equal(t, 0, c.Id)
	assert.Equal(t, "repo/app:v1", c.Image)
	assert.Equal(t, "sha256:abc", c.ImageDigest)
	assert.Equal(t, `[{"x":1}]`, c.MaterialInfo)
	assert.Equal(t, "CI-RUNNER", c.DataSource)
	assert.Equal(t, 7, c.PipelineId)
	assert.Equal(t, 11, c.ParentCiArtifact)
	assert.True(t, c.IsArtifactUploaded)
	assert.True(t, c.ScanEnabled)
	assert.True(t, c.Scanned)
	assert.Equal(t, "linux/amd64", c.TargetPlatforms)
	assert.Nil(t, c.WorkflowId, "copies never own a CI workflow")
	assert.Empty(t, c.CredentialsSourceType)
	assert.Empty(t, c.CredentialSourceValue)
	assert.Zero(t, c.ComponentId)
	assert.Zero(t, c.ExternalCiPipelineId)
	assert.Equal(t, builtAt, c.CreatedOn, "keeps the original build time")
	assert.Equal(t, int32(42), c.CreatedBy)
	assert.Equal(t, int32(42), c.UpdatedBy)
	assert.False(t, c.UpdatedOn.Before(before))
}

func TestBuildLinkedCiArtifactCopies_ScannedOnlyWhenScanEnabled(t *testing.T) {
	tests := []struct {
		scanEnabled, scanned, want bool
	}{
		{false, false, false},
		{false, true, false},
		{true, false, false},
		{true, true, true},
	}
	for _, tt := range tests {
		got := buildLinkedCiArtifactCopies([]repository.CiArtifact{{Id: 1, ScanEnabled: tt.scanEnabled, Scanned: tt.scanned}}, 0, 2, 1)
		require.Len(t, got, 1)
		assert.Equal(t, tt.want, got[0].Scanned, "scanEnabled=%v scanned=%v", tt.scanEnabled, tt.scanned)
	}
}

func TestCopyArtifactsFromParentCiPipeline(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name          string
		limit         int
		repo          *fakeLinkedCiArtifactRepo
		tx            *fakeLinkedCiTxManager
		wantCount     int
		wantErr       error
		wantStarted   int
		wantCommitted int
		wantSaved     []int
	}{
		{name: "limit zero skips everything", limit: 0, repo: &fakeLinkedCiArtifactRepo{parentArtifacts: parentBuilds(1)}, tx: &fakeLinkedCiTxManager{}},
		{name: "negative limit skips everything", limit: -3, repo: &fakeLinkedCiArtifactRepo{parentArtifacts: parentBuilds(1)}, tx: &fakeLinkedCiTxManager{}},
		{name: "start tx fails", limit: 10, repo: &fakeLinkedCiArtifactRepo{}, tx: &fakeLinkedCiTxManager{startErr: boom}, wantErr: boom, wantStarted: 1},
		{name: "lock fails", limit: 10, repo: &fakeLinkedCiArtifactRepo{lockErr: boom, parentArtifacts: parentBuilds(1)}, tx: &fakeLinkedCiTxManager{}, wantErr: boom, wantStarted: 1},
		{name: "oldest lookup fails", limit: 10, repo: &fakeLinkedCiArtifactRepo{oldestErr: boom, parentArtifacts: parentBuilds(1)}, tx: &fakeLinkedCiTxManager{}, wantErr: boom, wantStarted: 1},
		{name: "parent lookup fails", limit: 10, repo: &fakeLinkedCiArtifactRepo{latestErr: boom}, tx: &fakeLinkedCiTxManager{}, wantErr: boom, wantStarted: 1},
		{name: "parent lookup no rows is not an error", limit: 10, repo: &fakeLinkedCiArtifactRepo{latestErr: pg.ErrNoRows}, tx: &fakeLinkedCiTxManager{}, wantStarted: 1},
		{name: "nothing missing commits nothing", limit: 10, repo: &fakeLinkedCiArtifactRepo{parentArtifacts: parentBuilds(5, 6), oldestParentId: 5}, tx: &fakeLinkedCiTxManager{}, wantStarted: 1},
		{name: "save fails", limit: 10, repo: &fakeLinkedCiArtifactRepo{saveErr: boom, parentArtifacts: parentBuilds(5, 6)}, tx: &fakeLinkedCiTxManager{}, wantErr: boom, wantStarted: 1},
		{name: "commit fails", limit: 10, repo: &fakeLinkedCiArtifactRepo{parentArtifacts: parentBuilds(5, 6)}, tx: &fakeLinkedCiTxManager{commitErr: boom}, wantErr: boom, wantStarted: 1, wantCommitted: 1, wantSaved: []int{5, 6}},
		{name: "copies missing images oldest first", limit: 10, repo: &fakeLinkedCiArtifactRepo{parentArtifacts: parentBuilds(5, 6, 7, 8), oldestParentId: 7}, tx: &fakeLinkedCiTxManager{}, wantCount: 2, wantStarted: 1, wantCommitted: 1, wantSaved: []int{5, 6}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := newLinkedCiTestOrchestrator(tt.repo, &fakeLinkedCiPipelineRepo{}, tt.tx, tt.limit)
			count, err := impl.copyArtifactsFromParentCiPipeline(3, 9, 1)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Zero(t, count)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, count)
			}
			assert.Equal(t, tt.wantStarted, tt.tx.started)
			assert.Equal(t, tt.wantCommitted, tt.tx.committed)
			wantRolledBack := tt.wantStarted
			if tt.tx.startErr != nil {
				wantRolledBack = 0
			}
			assert.Equal(t, wantRolledBack, tt.tx.rolledBk, "deferred rollback runs whenever a tx was opened")
			if tt.wantSaved == nil {
				tt.wantSaved = []int{}
			}
			assert.Equal(t, tt.wantSaved, copiedParentIds(tt.repo.saved))
			if tt.wantStarted == 1 && tt.tx.startErr == nil {
				assert.Equal(t, []int{9}, tt.repo.lockedIds, "locks the child, not the parent")
			}
			if len(tt.repo.latestIds) > 0 {
				assert.Equal(t, []int{3}, tt.repo.latestIds, "reads the parent's images")
				assert.Equal(t, []int{tt.limit}, tt.repo.latestLimits)
				assert.Equal(t, []int{9}, tt.repo.oldestIds, "reads the child's oldest image")
			}
		})
	}
}

func linkedCiPipeline(id, parentId int, pipelineType string) *pipelineConfig.CiPipeline {
	return &pipelineConfig.CiPipeline{Id: id, AppId: id * 10, ParentCiPipeline: parentId, PipelineType: pipelineType, Active: true,
		App: &app.App{AppName: "app-" + string(rune('a'+id%26))}}
}

func TestSyncLinkedCiArtifacts_Validation(t *testing.T) {
	boom := errors.New("db down")
	tests := []struct {
		name       string
		pipeline   *pipelineConfig.CiPipeline
		findErr    error
		wantStatus int
		wantErr    error
	}{
		{name: "not found", findErr: pg.ErrNoRows, wantStatus: http.StatusNotFound},
		{name: "inactive", pipeline: func() *pipelineConfig.CiPipeline { p := linkedCiPipeline(5, 3, "LINKED"); p.Active = false; return p }(), wantStatus: http.StatusNotFound},
		{name: "db error passes through", findErr: boom, wantErr: boom},
		{name: "regular CI", pipeline: linkedCiPipeline(5, 0, string(common.CI_BUILD)), wantStatus: http.StatusBadRequest},
		{name: "regular CI with a stray parent id", pipeline: linkedCiPipeline(5, 3, string(common.CI_BUILD)), wantStatus: http.StatusBadRequest},
		{name: "linked CD", pipeline: linkedCiPipeline(5, 3, string(common.LINKED_CD)), wantStatus: http.StatusBadRequest},
		{name: "job", pipeline: linkedCiPipeline(5, 0, string(common.CI_JOB)), wantStatus: http.StatusBadRequest},
		{name: "linked type without parent", pipeline: linkedCiPipeline(5, 0, string(common.LINKED)), wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			artifactRepo := &fakeLinkedCiArtifactRepo{parentArtifacts: parentBuilds(1)}
			txManager := &fakeLinkedCiTxManager{}
			impl := newLinkedCiTestOrchestrator(artifactRepo, &fakeLinkedCiPipelineRepo{pipeline: tt.pipeline, findErr: tt.findErr}, txManager, 10)
			resp, err := impl.SyncLinkedCiArtifacts(5, 1)
			assert.Nil(t, resp)
			require.Error(t, err)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				var apiErr *util.ApiError
				require.ErrorAs(t, err, &apiErr)
				assert.Equal(t, tt.wantStatus, apiErr.HttpStatusCode)
			}
			assert.Zero(t, txManager.started, "nothing is copied for an invalid pipeline")
		})
	}
}

func TestSyncLinkedCiArtifacts_Success(t *testing.T) {
	artifactRepo := &fakeLinkedCiArtifactRepo{parentArtifacts: parentBuilds(1, 2, 3)}
	impl := newLinkedCiTestOrchestrator(artifactRepo, &fakeLinkedCiPipelineRepo{pipeline: linkedCiPipeline(5, 3, "LINKED")}, &fakeLinkedCiTxManager{}, 10)
	resp, err := impl.SyncLinkedCiArtifacts(5, 1)
	require.NoError(t, err)
	assert.Equal(t, 5, resp.CiPipelineId)
	assert.Equal(t, 3, resp.ParentCiPipelineId)
	assert.Equal(t, 50, resp.AppId)
	assert.Equal(t, "app-f", resp.AppName)
	assert.Equal(t, 3, resp.CopiedCount)
	assert.Empty(t, resp.Error)
	assert.Equal(t, []int{3}, artifactRepo.latestIds)
	assert.Equal(t, []int{5}, artifactRepo.lockedIds)
}

func TestSyncLinkedCiArtifacts_CopyErrorIsReturned(t *testing.T) {
	boom := errors.New("boom")
	impl := newLinkedCiTestOrchestrator(&fakeLinkedCiArtifactRepo{lockErr: boom}, &fakeLinkedCiPipelineRepo{pipeline: linkedCiPipeline(5, 3, "LINKED")}, &fakeLinkedCiTxManager{}, 10)
	resp, err := impl.SyncLinkedCiArtifacts(5, 1)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, boom)
}

func TestSyncLinkedCiArtifacts_ZeroLimitCopiesNothing(t *testing.T) {
	txManager := &fakeLinkedCiTxManager{}
	impl := newLinkedCiTestOrchestrator(&fakeLinkedCiArtifactRepo{parentArtifacts: parentBuilds(1)}, &fakeLinkedCiPipelineRepo{pipeline: linkedCiPipeline(5, 3, "LINKED")}, txManager, 0)
	resp, err := impl.SyncLinkedCiArtifacts(5, 1)
	require.NoError(t, err)
	assert.Zero(t, resp.CopiedCount)
	assert.Zero(t, txManager.started)
}

func TestSyncLinkedCiArtifactsForSource(t *testing.T) {
	t.Run("children lookup fails", func(t *testing.T) {
		boom := errors.New("boom")
		impl := newLinkedCiTestOrchestrator(&fakeLinkedCiArtifactRepo{}, &fakeLinkedCiPipelineRepo{childrenErr: boom}, &fakeLinkedCiTxManager{}, 10)
		resp, err := impl.SyncLinkedCiArtifactsForSource(3, 1, func(*pipelineConfig.CiPipeline) bool { return true })
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, boom)
	})
	t.Run("no children gives an empty list, not null", func(t *testing.T) {
		for _, childrenErr := range []error{nil, pg.ErrNoRows} {
			impl := newLinkedCiTestOrchestrator(&fakeLinkedCiArtifactRepo{}, &fakeLinkedCiPipelineRepo{childrenErr: childrenErr}, &fakeLinkedCiTxManager{}, 10)
			resp, err := impl.SyncLinkedCiArtifactsForSource(3, 1, func(*pipelineConfig.CiPipeline) bool { return true })
			require.NoError(t, err)
			assert.Equal(t, 3, resp.SourceCiPipelineId)
			assert.NotNil(t, resp.Results)
			assert.Empty(t, resp.Results)
			assert.Zero(t, resp.UnauthorizedCount)
		}
	})
	t.Run("skips children without access and reports per child results", func(t *testing.T) {
		children := []*pipelineConfig.CiPipeline{linkedCiPipeline(5, 3, "LINKED"), linkedCiPipeline(6, 3, "LINKED"), linkedCiPipeline(7, 3, "LINKED"), linkedCiPipeline(8, 3, "LINKED")}
		artifactRepo := &fakeLinkedCiArtifactRepo{parentArtifacts: parentBuilds(1, 2), failSaveForChild: 7}
		impl := newLinkedCiTestOrchestrator(artifactRepo, &fakeLinkedCiPipelineRepo{children: children}, &fakeLinkedCiTxManager{}, 10)
		var asked []int
		resp, err := impl.SyncLinkedCiArtifactsForSource(3, 1, func(p *pipelineConfig.CiPipeline) bool {
			asked = append(asked, p.Id)
			return p.Id != 6
		})
		require.NoError(t, err)
		assert.Equal(t, []int{5, 6, 7, 8}, asked)
		assert.Equal(t, 1, resp.UnauthorizedCount)
		require.Len(t, resp.Results, 3)
		assert.Equal(t, 5, resp.Results[0].CiPipelineId)
		assert.Equal(t, 2, resp.Results[0].CopiedCount)
		assert.Empty(t, resp.Results[0].Error)
		assert.Equal(t, 7, resp.Results[1].CiPipelineId)
		assert.Zero(t, resp.Results[1].CopiedCount)
		assert.Equal(t, "failed to sync images", resp.Results[1].Error)
		assert.Equal(t, 8, resp.Results[2].CiPipelineId)
		assert.Equal(t, 2, resp.Results[2].CopiedCount)
		assert.Equal(t, []int{5, 7, 8}, artifactRepo.lockedIds, "no copy for the unauthorized child")
	})
	t.Run("all unauthorized", func(t *testing.T) {
		children := []*pipelineConfig.CiPipeline{linkedCiPipeline(5, 3, "LINKED"), linkedCiPipeline(6, 3, "LINKED")}
		txManager := &fakeLinkedCiTxManager{}
		impl := newLinkedCiTestOrchestrator(&fakeLinkedCiArtifactRepo{parentArtifacts: parentBuilds(1)}, &fakeLinkedCiPipelineRepo{children: children}, txManager, 10)
		resp, err := impl.SyncLinkedCiArtifactsForSource(3, 1, func(*pipelineConfig.CiPipeline) bool { return false })
		require.NoError(t, err)
		assert.Equal(t, 2, resp.UnauthorizedCount)
		assert.Empty(t, resp.Results)
		assert.Zero(t, txManager.started)
	})
}

type fakeOrderCiPipelineRepo struct {
	pipelineConfig.CiPipelineRepository
	pipelines []*pipelineConfig.CiPipeline
	err       error
	calls     [][]int
}

func (f *fakeOrderCiPipelineRepo) FindByIdsIn(ids []int) ([]*pipelineConfig.CiPipeline, error) {
	f.calls = append(f.calls, ids)
	return f.pipelines, f.err
}

func beanIds(artifacts []bean3.CiArtifactBean) []int {
	ids := make([]int, 0, len(artifacts))
	for _, a := range artifacts {
		ids = append(ids, a.Id)
	}
	return ids
}

func TestSortArtifactsNewestFirst(t *testing.T) {
	linked := []*pipelineConfig.CiPipeline{{Id: 7, PipelineType: "LINKED"}, {Id: 8, PipelineType: "CI_BUILD"}, {Id: 9, PipelineType: "LINKED_CD"}}
	// child 7: rows 100 (v7), 101 (v8) from parent builds 70, 80; rows 200..202 backfilled for parent builds 30, 40, 50
	linkedChild := []bean3.CiArtifactBean{
		{Id: 200, CiPipelineId: 7, ParentCiArtifact: 30},
		{Id: 101, CiPipelineId: 7, ParentCiArtifact: 80},
		{Id: 202, CiPipelineId: 7, ParentCiArtifact: 50},
		{Id: 100, CiPipelineId: 7, ParentCiArtifact: 70},
		{Id: 201, CiPipelineId: 7, ParentCiArtifact: 40},
	}
	tests := []struct {
		name             string
		artifacts        []bean3.CiArtifactBean
		useLinkedCiOrder bool
		repoErr          error
		want             []int
		wantCalls        int
	}{
		{name: "linked child follows parent builds", artifacts: linkedChild, useLinkedCiOrder: true, want: []int{101, 100, 202, 201, 200}, wantCalls: 1},
		{name: "linked order off keeps row id order", artifacts: linkedChild, useLinkedCiOrder: false, want: []int{202, 201, 200, 101, 100}},
		{name: "lookup error falls back to row id order", artifacts: linkedChild, useLinkedCiOrder: true, repoErr: errors.New("db"), want: []int{202, 201, 200, 101, 100}, wantCalls: 1},
		{name: "regular CI with plugin copies keeps row id order", useLinkedCiOrder: true, wantCalls: 1,
			artifacts: []bean3.CiArtifactBean{{Id: 10, CiPipelineId: 8}, {Id: 12, CiPipelineId: 8, ParentCiArtifact: 11}, {Id: 11, CiPipelineId: 8}, {Id: 13, CiPipelineId: 8, ParentCiArtifact: 10}},
			want:      []int{13, 12, 11, 10}},
		{name: "linked CD keeps row id order", useLinkedCiOrder: true, wantCalls: 1,
			artifacts: []bean3.CiArtifactBean{{Id: 50, CiPipelineId: 9, ParentCiArtifact: 80}, {Id: 51, CiPipelineId: 9, ParentCiArtifact: 20}},
			want:      []int{51, 50}},
		{name: "no copied rows skips the lookup", useLinkedCiOrder: true,
			artifacts: []bean3.CiArtifactBean{{Id: 1, CiPipelineId: 8}, {Id: 3}, {Id: 2, CiPipelineId: 8}},
			want:      []int{3, 2, 1}},
		{name: "rows without a pipeline id are ordered by row id", useLinkedCiOrder: true, wantCalls: 1,
			artifacts: []bean3.CiArtifactBean{{Id: 300, ParentCiArtifact: 5}, {Id: 101, CiPipelineId: 7, ParentCiArtifact: 80}},
			want:      []int{300, 101}},
		{name: "same source build: newer row first", useLinkedCiOrder: true, wantCalls: 1,
			artifacts: []bean3.CiArtifactBean{{Id: 120, CiPipelineId: 7, ParentCiArtifact: 80}, {Id: 130, CiPipelineId: 7, ParentCiArtifact: 80}, {Id: 80, CiPipelineId: 8}},
			want:      []int{130, 120, 80}},
		{name: "empty list", useLinkedCiOrder: true, artifacts: []bean3.CiArtifactBean{}, want: []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeOrderCiPipelineRepo{pipelines: linked, err: tt.repoErr}
			impl := &AppArtifactManagerImpl{logger: zap.NewNop().Sugar(), CiPipelineRepository: repo}
			artifacts := append([]bean3.CiArtifactBean{}, tt.artifacts...)
			impl.sortArtifactsNewestFirst(artifacts, tt.useLinkedCiOrder)
			assert.Equal(t, tt.want, beanIds(artifacts))
			assert.Len(t, repo.calls, tt.wantCalls)
		})
	}
}

func TestFindLinkedCiPipelineIds(t *testing.T) {
	pipelines := []*pipelineConfig.CiPipeline{{Id: 7, PipelineType: "LINKED"}, {Id: 8, PipelineType: "CI_BUILD"}, {Id: 9, PipelineType: "LINKED_CD"}, {Id: 10, PipelineType: "CI_JOB"}}
	t.Run("no ids skips the lookup", func(t *testing.T) {
		repo := &fakeOrderCiPipelineRepo{pipelines: pipelines}
		impl := &AppArtifactManagerImpl{logger: zap.NewNop().Sugar(), CiPipelineRepository: repo}
		assert.Empty(t, impl.findLinkedCiPipelineIds(nil))
		assert.Empty(t, repo.calls)
	})
	t.Run("only linked CI counts", func(t *testing.T) {
		impl := &AppArtifactManagerImpl{logger: zap.NewNop().Sugar(), CiPipelineRepository: &fakeOrderCiPipelineRepo{pipelines: pipelines}}
		assert.Equal(t, map[int]bool{7: true}, impl.findLinkedCiPipelineIds([]int{7, 8, 9, 10}))
	})
	t.Run("lookup error means no linked order", func(t *testing.T) {
		impl := &AppArtifactManagerImpl{logger: zap.NewNop().Sugar(), CiPipelineRepository: &fakeOrderCiPipelineRepo{pipelines: pipelines, err: errors.New("db")}}
		assert.Empty(t, impl.findLinkedCiPipelineIds([]int{7}))
	})
}

func TestGetLinkedCiPipelineIds_DeduplicatesLookup(t *testing.T) {
	repo := &fakeOrderCiPipelineRepo{pipelines: []*pipelineConfig.CiPipeline{{Id: 7, PipelineType: "LINKED"}}}
	impl := &AppArtifactManagerImpl{logger: zap.NewNop().Sugar(), CiPipelineRepository: repo}
	got := impl.getLinkedCiPipelineIds([]bean3.CiArtifactBean{{Id: 1, CiPipelineId: 7, ParentCiArtifact: 2}, {Id: 3, CiPipelineId: 7, ParentCiArtifact: 4}, {Id: 5, CiPipelineId: 6}})
	assert.Equal(t, map[int]bool{7: true}, got)
	assert.Equal(t, [][]int{{7}}, repo.calls)
}
