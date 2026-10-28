//go:build integration

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
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Runs against a migrated Postgres: PG_ADDR, PG_PORT, PG_USER, PG_PASSWORD, PG_DATABASE.
type linkedCiFixture struct {
	t      *testing.T
	db     *pg.DB
	teamId int
	seq    int
}

func newLinkedCiFixture(t *testing.T) *linkedCiFixture {
	db := pg.Connect(&pg.Options{
		Addr:     os.Getenv("PG_ADDR") + ":" + os.Getenv("PG_PORT"),
		User:     os.Getenv("PG_USER"),
		Password: os.Getenv("PG_PASSWORD"),
		Database: os.Getenv("PG_DATABASE"),
		PoolSize: 30,
	})
	t.Cleanup(func() { db.Close() })
	f := &linkedCiFixture{t: t, db: db}
	_, err := db.QueryOne(pg.Scan(&f.teamId), "INSERT INTO team (name, active, created_on, created_by, updated_on, updated_by) VALUES (?, true, now(), 1, now(), 1) RETURNING id",
		fmt.Sprintf("it-linked-ci-%d", time.Now().UnixNano()))
	require.NoError(t, err)
	return f
}

func (f *linkedCiFixture) ciPipeline(pipelineType string, parentId int, active, deleted bool) int {
	f.seq++
	var appId, id int
	_, err := f.db.QueryOne(pg.Scan(&appId), "INSERT INTO app (app_name, active, team_id, created_on, created_by, updated_on, updated_by) VALUES (?, true, ?, now(), 1, now(), 1) RETURNING id",
		fmt.Sprintf("it-app-%d-%d", f.teamId, f.seq), f.teamId)
	require.NoError(f.t, err)
	_, err = f.db.QueryOne(pg.Scan(&id), "INSERT INTO ci_pipeline (app_id, name, active, deleted, external, parent_ci_pipeline, ci_pipeline_type, created_on, created_by, updated_on, updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, now(), 1, now(), 1) RETURNING id",
		appId, fmt.Sprintf("it-ci-%d", f.seq), active, deleted, parentId > 0, parentId, pipelineType)
	require.NoError(f.t, err)
	return id
}

func (f *linkedCiFixture) artifact(pipelineId, parentArtifact int, dataSource string, createdOn time.Time) int {
	var id int
	var parent interface{}
	if parentArtifact > 0 {
		parent = parentArtifact
	}
	credType, credValue := "", ""
	if dataSource == "post_ci" {
		credType, credValue = "global_container_registry", "other-registry"
	}
	_, err := f.db.QueryOne(pg.Scan(&id), "INSERT INTO ci_artifact (image, image_digest, material_info, data_source, pipeline_id, parent_ci_artifact, component_id, credentials_source_type, credentials_source_value, scan_enabled, scanned, created_on, created_by, updated_on, updated_by) VALUES (?, ?, '[]', ?, ?, ?, ?, ?, ?, false, false, ?, 1, now(), 1) RETURNING id",
		fmt.Sprintf("repo/img:%d-%d", pipelineId, f.seq), "sha256:x", dataSource, pipelineId, parent, pipelineId, credType, credValue, createdOn)
	require.NoError(f.t, err)
	f.seq++
	return id
}

// parentWithBuilds creates a source pipeline with n builds, each followed by a post-CI plugin copy; returns build ids oldest first
func (f *linkedCiFixture) parentWithBuilds(n int) (int, []int) {
	parentId := f.ciPipeline("CI_BUILD", 0, true, false)
	builds := make([]int, 0, n)
	for i := 0; i < n; i++ {
		buildId := f.artifact(parentId, 0, "CI-RUNNER", time.Date(2026, 1, 1+i, 0, 0, 0, 0, time.UTC))
		f.artifact(parentId, buildId, "post_ci", time.Now())
		builds = append(builds, buildId)
	}
	return parentId, builds
}

type childRow struct {
	Id, ParentCiArtifact                                      int
	DataSource, CredentialsSourceType, CredentialsSourceValue string
	CreatedOn                                                 time.Time
	CiWorkflowId                                              *int
}

func (f *linkedCiFixture) childRows(childId int) []childRow {
	var rows []childRow
	_, err := f.db.Query(&rows, "SELECT id, parent_ci_artifact, data_source, credentials_source_type, credentials_source_value, created_on, ci_workflow_id FROM ci_artifact WHERE pipeline_id = ? ORDER BY id", childId)
	require.NoError(f.t, err)
	return rows
}

func parentIdsOf(rows []childRow) []int {
	ids := make([]int, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ParentCiArtifact)
	}
	return ids
}

func (f *linkedCiFixture) orchestrator(limit int) *CiCdPipelineOrchestratorImpl {
	logger := zap.NewNop().Sugar()
	return &CiCdPipelineOrchestratorImpl{
		logger:               logger,
		CiArtifactRepository: repository.NewCiArtifactRepositoryImpl(f.db, logger),
		ciPipelineRepository: pipelineConfig.NewCiPipelineRepositoryImpl(f.db, logger, sql.NewTransactionUtilImpl(f.db)),
		transactionManager:   sql.NewTransactionUtilImpl(f.db),
		globalEnvVariables:   &util2.GlobalEnvVariables{LinkedCiArtifactCopyLimit: limit},
	}
}

func TestIT_LinkedCiCopy_EmptyChildGetsLatestOwnBuildsOldestFirst(t *testing.T) {
	f := newLinkedCiFixture(t)
	parentId, builds := f.parentWithBuilds(12)
	childId := f.ciPipeline("LINKED", parentId, true, false)
	impl := f.orchestrator(10)

	count, err := impl.copyArtifactsFromParentCiPipeline(parentId, childId, 5)
	require.NoError(t, err)
	assert.Equal(t, 10, count)
	rows := f.childRows(childId)
	assert.Equal(t, builds[2:], parentIdsOf(rows), "latest 10 builds, oldest first, no plugin copies")
	for i, r := range rows {
		assert.Equal(t, "CI-RUNNER", r.DataSource)
		assert.Empty(t, r.CredentialsSourceType)
		assert.Empty(t, r.CredentialsSourceValue)
		assert.Nil(t, r.CiWorkflowId)
		assert.True(t, r.CreatedOn.Equal(time.Date(2026, 1, 3+i, 0, 0, 0, 0, time.UTC)), "keeps the parent's build time, got %v", r.CreatedOn)
	}

	again, err := impl.copyArtifactsFromParentCiPipeline(parentId, childId, 5)
	require.NoError(t, err)
	assert.Zero(t, again, "second sync is a no-op")
	assert.Len(t, f.childRows(childId), 10)
}

func TestIT_LinkedCiCopy_ChildWithNewerImagesGetsOnlyTheGap(t *testing.T) {
	f := newLinkedCiFixture(t)
	parentId, builds := f.parentWithBuilds(12)
	childId := f.ciPipeline("LINKED", parentId, true, false)
	for _, b := range builds[9:] {
		f.artifact(childId, b, "CI-RUNNER", time.Now()) // what the parent build fan-out writes
	}
	count, err := f.orchestrator(10).copyArtifactsFromParentCiPipeline(parentId, childId, 5)
	require.NoError(t, err)
	assert.Equal(t, 7, count)
	got := parentIdsOf(f.childRows(childId))
	sort.Ints(got)
	assert.Equal(t, builds[2:], got, "child now holds exactly the parent's latest 10")
}

func TestIT_LinkedCiCopy_NoOpCases(t *testing.T) {
	f := newLinkedCiFixture(t)
	t.Run("child already has the whole window", func(t *testing.T) {
		parentId, builds := f.parentWithBuilds(12)
		childId := f.ciPipeline("LINKED", parentId, true, false)
		for _, b := range builds[2:] {
			f.artifact(childId, b, "CI-RUNNER", time.Now())
		}
		count, err := f.orchestrator(10).copyArtifactsFromParentCiPipeline(parentId, childId, 5)
		require.NoError(t, err)
		assert.Zero(t, count)
	})
	t.Run("child is older than the whole window", func(t *testing.T) {
		parentId, builds := f.parentWithBuilds(12)
		childId := f.ciPipeline("LINKED", parentId, true, false)
		f.artifact(childId, builds[0], "CI-RUNNER", time.Now())
		count, err := f.orchestrator(10).copyArtifactsFromParentCiPipeline(parentId, childId, 5)
		require.NoError(t, err)
		assert.Zero(t, count, "a child linked long ago already got these from parent builds")
	})
	t.Run("parent has no builds", func(t *testing.T) {
		parentId := f.ciPipeline("CI_BUILD", 0, true, false)
		childId := f.ciPipeline("LINKED", parentId, true, false)
		count, err := f.orchestrator(10).copyArtifactsFromParentCiPipeline(parentId, childId, 5)
		require.NoError(t, err)
		assert.Zero(t, count)
	})
	t.Run("parent has only plugin copies", func(t *testing.T) {
		parentId := f.ciPipeline("CI_BUILD", 0, true, false)
		otherBuild := f.artifact(parentId, 0, "CI-RUNNER", time.Now())
		_, err := f.db.Exec("UPDATE ci_artifact SET pipeline_id = NULL WHERE id = ?", otherBuild)
		require.NoError(t, err)
		f.artifact(parentId, otherBuild, "post_ci", time.Now())
		childId := f.ciPipeline("LINKED", parentId, true, false)
		count, err := f.orchestrator(10).copyArtifactsFromParentCiPipeline(parentId, childId, 5)
		require.NoError(t, err)
		assert.Zero(t, count)
	})
	t.Run("parent is deleted", func(t *testing.T) {
		parentId, _ := f.parentWithBuilds(3)
		childId := f.ciPipeline("LINKED", parentId, true, false)
		_, err := f.db.Exec("UPDATE ci_pipeline SET deleted = true WHERE id = ?", parentId)
		require.NoError(t, err)
		count, err := f.orchestrator(10).copyArtifactsFromParentCiPipeline(parentId, childId, 5)
		require.NoError(t, err)
		assert.Zero(t, count)
	})
	t.Run("limit zero", func(t *testing.T) {
		parentId, _ := f.parentWithBuilds(3)
		childId := f.ciPipeline("LINKED", parentId, true, false)
		count, err := f.orchestrator(0).copyArtifactsFromParentCiPipeline(parentId, childId, 5)
		require.NoError(t, err)
		assert.Zero(t, count)
		assert.Empty(t, f.childRows(childId))
	})
}

func TestIT_LinkedCiCopy_LimitAboveParentBuilds(t *testing.T) {
	f := newLinkedCiFixture(t)
	parentId, builds := f.parentWithBuilds(4)
	childId := f.ciPipeline("LINKED", parentId, true, false)
	count, err := f.orchestrator(10).copyArtifactsFromParentCiPipeline(parentId, childId, 5)
	require.NoError(t, err)
	assert.Equal(t, 4, count)
	assert.Equal(t, builds, parentIdsOf(f.childRows(childId)))
}

func TestIT_LinkedCiCopy_ChildFromOldBuggyCopy(t *testing.T) {
	f := newLinkedCiFixture(t)
	parentId, builds := f.parentWithBuilds(12)
	childId := f.ciPipeline("LINKED", parentId, true, false)
	var pluginOfLast int
	_, err := f.db.QueryOne(pg.Scan(&pluginOfLast), "SELECT id FROM ci_artifact WHERE parent_ci_artifact = ? AND pipeline_id = ?", builds[11], parentId)
	require.NoError(t, err)
	f.artifact(childId, pluginOfLast, "post_ci", time.Now()) // the old copy also cloned plugin rows
	count, err := f.orchestrator(10).copyArtifactsFromParentCiPipeline(parentId, childId, 5)
	require.NoError(t, err)
	assert.Equal(t, 10, count, "the build images are still missing, so all 10 are copied")
}

func TestIT_LinkedCiCopy_ConcurrentSyncsNeverDuplicate(t *testing.T) {
	f := newLinkedCiFixture(t)
	parentId, builds := f.parentWithBuilds(12)
	childId := f.ciPipeline("LINKED", parentId, true, false)
	impl := f.orchestrator(10)
	var wg sync.WaitGroup
	var mu sync.Mutex
	total := 0
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count, err := impl.copyArtifactsFromParentCiPipeline(parentId, childId, 5)
			assert.NoError(t, err)
			mu.Lock()
			total += count
			mu.Unlock()
		}()
	}
	wg.Wait()
	assert.Equal(t, 10, total)
	assert.Equal(t, builds[2:], parentIdsOf(f.childRows(childId)))
}

func TestIT_LinkedCiCopy_LockBlocksSecondCopyUntilCommit(t *testing.T) {
	f := newLinkedCiFixture(t)
	parentId, _ := f.parentWithBuilds(2)
	childId := f.ciPipeline("LINKED", parentId, true, false)
	artifactRepo := repository.NewCiArtifactRepositoryImpl(f.db, zap.NewNop().Sugar())
	tx, err := f.db.Begin()
	require.NoError(t, err)
	require.NoError(t, artifactRepo.LockLinkedCiArtifactCopy(tx, childId))
	done := make(chan time.Time, 1)
	go func() {
		_, err := f.orchestrator(10).copyArtifactsFromParentCiPipeline(parentId, childId, 5)
		assert.NoError(t, err)
		done <- time.Now()
	}()
	time.Sleep(500 * time.Millisecond)
	released := time.Now()
	require.NoError(t, tx.Rollback())
	finished := <-done
	assert.True(t, finished.After(released), "the copy waited for the lock")

	otherChild := f.ciPipeline("LINKED", parentId, true, false)
	tx2, err := f.db.Begin()
	require.NoError(t, err)
	defer tx2.Rollback()
	require.NoError(t, artifactRepo.LockLinkedCiArtifactCopy(tx2, childId))
	start := time.Now()
	_, err = f.orchestrator(10).copyArtifactsFromParentCiPipeline(parentId, otherChild, 5)
	require.NoError(t, err)
	assert.Less(t, time.Since(start), 400*time.Millisecond, "locks are per child")
}

func TestIT_LinkedCiCopy_ChildrenAreIndependent(t *testing.T) {
	f := newLinkedCiFixture(t)
	parentId, _ := f.parentWithBuilds(3)
	childA := f.ciPipeline("LINKED", parentId, true, false)
	childB := f.ciPipeline("LINKED", parentId, true, false)
	_, err := f.orchestrator(10).copyArtifactsFromParentCiPipeline(parentId, childA, 5)
	require.NoError(t, err)
	assert.Len(t, f.childRows(childA), 3)
	assert.Empty(t, f.childRows(childB))
}

func TestIT_LinkedCiRepository(t *testing.T) {
	f := newLinkedCiFixture(t)
	repo := repository.NewCiArtifactRepositoryImpl(f.db, zap.NewNop().Sugar())
	parentId, builds := f.parentWithBuilds(5)
	childId := f.ciPipeline("LINKED", parentId, true, false)
	tx, err := f.db.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	oldest, err := repo.GetOldestParentCiArtifactId(tx, childId)
	require.NoError(t, err)
	assert.Zero(t, oldest, "empty child")
	_, err = tx.Exec("INSERT INTO ci_artifact (data_source, pipeline_id, parent_ci_artifact, created_on, created_by, updated_on, updated_by) VALUES ('CI-RUNNER', ?, ?, now(), 1, now(), 1), ('CI-RUNNER', ?, ?, now(), 1, now(), 1), ('CI-RUNNER', ?, NULL, now(), 1, now(), 1)",
		childId, builds[3], childId, builds[1], childId)
	require.NoError(t, err)
	oldest, err = repo.GetOldestParentCiArtifactId(tx, childId)
	require.NoError(t, err)
	assert.Equal(t, builds[1], oldest, "ignores rows without a parent")

	latest, err := repo.GetLatestOwnBuildArtifactsByCiPipelineId(tx, parentId, 3)
	require.NoError(t, err)
	ids := make([]int, 0, len(latest))
	for _, a := range latest {
		ids = append(ids, a.Id)
	}
	assert.Equal(t, []int{builds[4], builds[3], builds[2]}, ids, "newest first, plugin rows excluded, limited")

	none, err := repo.GetLatestOwnBuildArtifactsByCiPipelineId(tx, childId+100000, 3)
	require.NoError(t, err)
	assert.Empty(t, none)
}

func TestIT_SyncLinkedCiArtifacts(t *testing.T) {
	f := newLinkedCiFixture(t)
	parentId, builds := f.parentWithBuilds(3)
	impl := f.orchestrator(10)

	t.Run("linked child is synced", func(t *testing.T) {
		childId := f.ciPipeline("LINKED", parentId, true, false)
		resp, err := impl.SyncLinkedCiArtifacts(childId, 5)
		require.NoError(t, err)
		assert.Equal(t, childId, resp.CiPipelineId)
		assert.Equal(t, parentId, resp.ParentCiPipelineId)
		assert.NotZero(t, resp.AppId)
		assert.Contains(t, resp.AppName, "it-app-")
		assert.Equal(t, 3, resp.CopiedCount)
		assert.Equal(t, builds, parentIdsOf(f.childRows(childId)))
	})
	cases := []struct {
		name       string
		id         func() int
		wantStatus int
	}{
		{"missing pipeline", func() int { return 99999999 }, http.StatusNotFound},
		{"deleted linked pipeline", func() int { return f.ciPipeline("LINKED", parentId, true, true) }, http.StatusNotFound},
		{"inactive linked pipeline", func() int { return f.ciPipeline("LINKED", parentId, false, false) }, http.StatusNotFound},
		{"regular CI", func() int { return parentId }, http.StatusBadRequest},
		{"linked CD", func() int { return f.ciPipeline("LINKED_CD", 12345, true, false) }, http.StatusBadRequest},
		{"job", func() int { return f.ciPipeline("CI_JOB", 0, true, false) }, http.StatusBadRequest},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp, err := impl.SyncLinkedCiArtifacts(c.id(), 5)
			assert.Nil(t, resp)
			var apiErr *util.ApiError
			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, c.wantStatus, apiErr.HttpStatusCode)
		})
	}
	assert.Len(t, parentIdsOf(f.childRows(parentId)), 6, "the parent itself is never written to")
}

func TestIT_SyncLinkedCiArtifactsForSource(t *testing.T) {
	f := newLinkedCiFixture(t)
	parentId, _ := f.parentWithBuilds(2)
	otherParent, _ := f.parentWithBuilds(1)
	childA := f.ciPipeline("LINKED", parentId, true, false)
	childB := f.ciPipeline("LINKED", parentId, true, false)
	inactive := f.ciPipeline("LINKED", parentId, false, false)
	deleted := f.ciPipeline("LINKED", parentId, true, true)
	linkedCd := f.ciPipeline("LINKED_CD", parentId, true, false)
	otherChild := f.ciPipeline("LINKED", otherParent, true, false)
	noAccess := f.ciPipeline("LINKED", parentId, true, false)

	resp, err := f.orchestrator(10).SyncLinkedCiArtifactsForSource(parentId, 5, func(p *pipelineConfig.CiPipeline) bool {
		require.NotNil(t, p.App, "app is loaded for RBAC")
		return p.Id != noAccess
	})
	require.NoError(t, err)
	assert.Equal(t, parentId, resp.SourceCiPipelineId)
	assert.Equal(t, 1, resp.UnauthorizedCount)
	require.Len(t, resp.Results, 2)
	assert.Equal(t, childA, resp.Results[0].CiPipelineId)
	assert.Equal(t, childB, resp.Results[1].CiPipelineId)
	for _, r := range resp.Results {
		assert.Equal(t, 2, r.CopiedCount)
		assert.Empty(t, r.Error)
		assert.NotEmpty(t, r.AppName)
	}
	for _, untouched := range []int{inactive, deleted, linkedCd, otherChild, noAccess} {
		assert.Empty(t, f.childRows(untouched), "pipeline %d must not be synced", untouched)
	}
}
