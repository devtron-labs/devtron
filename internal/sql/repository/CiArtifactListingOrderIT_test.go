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

package repository

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Runs against a migrated Postgres: PG_ADDR, PG_PORT, PG_USER, PG_PASSWORD, PG_DATABASE.
type orderFixture struct {
	t      *testing.T
	db     *pg.DB
	teamId int
	envId  int
	seq    int
}

func newOrderFixture(t *testing.T) *orderFixture {
	db := pg.Connect(&pg.Options{
		Addr:     os.Getenv("PG_ADDR") + ":" + os.Getenv("PG_PORT"),
		User:     os.Getenv("PG_USER"),
		Password: os.Getenv("PG_PASSWORD"),
		Database: os.Getenv("PG_DATABASE"),
	})
	t.Cleanup(func() { db.Close() })
	f := &orderFixture{t: t, db: db}
	_, err := db.QueryOne(pg.Scan(&f.teamId), "INSERT INTO team (name, active, created_on, created_by, updated_on, updated_by) VALUES (?, true, now(), 1, now(), 1) RETURNING id",
		fmt.Sprintf("it-order-%d", time.Now().UnixNano()))
	require.NoError(t, err)
	_, err = db.QueryOne(pg.Scan(&f.envId), "SELECT id FROM environment ORDER BY id LIMIT 1")
	require.NoError(t, err)
	return f
}

func (f *orderFixture) ciPipeline(pipelineType string, parentId int) (ciPipelineId, appId int) {
	f.seq++
	_, err := f.db.QueryOne(pg.Scan(&appId), "INSERT INTO app (app_name, active, team_id, created_on, created_by, updated_on, updated_by) VALUES (?, true, ?, now(), 1, now(), 1) RETURNING id",
		fmt.Sprintf("it-order-app-%d-%d", f.teamId, f.seq), f.teamId)
	require.NoError(f.t, err)
	_, err = f.db.QueryOne(pg.Scan(&ciPipelineId), "INSERT INTO ci_pipeline (app_id, name, active, deleted, external, parent_ci_pipeline, ci_pipeline_type, created_on, created_by, updated_on, updated_by) VALUES (?, ?, true, false, ?, ?, ?, now(), 1, now(), 1) RETURNING id",
		appId, fmt.Sprintf("it-order-ci-%d", f.seq), parentId > 0, parentId, pipelineType)
	require.NoError(f.t, err)
	return ciPipelineId, appId
}

func (f *orderFixture) cdPipeline(ciPipelineId, appId int) int {
	var id int
	_, err := f.db.QueryOne(pg.Scan(&id), "INSERT INTO pipeline (app_id, ci_pipeline_id, environment_id, trigger_type, pipeline_name, deleted, created_on, created_by, updated_on, updated_by) VALUES (?, ?, ?, 'MANUAL', ?, false, now(), 1, now(), 1) RETURNING id",
		appId, ciPipelineId, f.envId, fmt.Sprintf("it-order-cd-%d-%d", f.teamId, ciPipelineId))
	require.NoError(f.t, err)
	return id
}

func (f *orderFixture) artifact(pipelineId, parentArtifact int, dataSource, image string) int {
	var id int
	var parent interface{}
	if parentArtifact > 0 {
		parent = parentArtifact
	}
	_, err := f.db.QueryOne(pg.Scan(&id), "INSERT INTO ci_artifact (image, data_source, pipeline_id, component_id, parent_ci_artifact, created_on, created_by, updated_on, updated_by) VALUES (?, ?, ?, ?, ?, now(), 1, now(), 1) RETURNING id",
		image, dataSource, pipelineId, pipelineId, parent)
	require.NoError(f.t, err)
	return id
}

func (f *orderFixture) idsByIdDesc(pipelineId int) []int {
	var ids []int
	_, err := f.db.Query(&ids, "SELECT id FROM ci_artifact WHERE pipeline_id = ? ORDER BY id DESC", pipelineId)
	require.NoError(f.t, err)
	return ids
}

func artifactIds[T any](artifacts []T, id func(T) int) []int {
	ids := make([]int, 0, len(artifacts))
	for _, a := range artifacts {
		ids = append(ids, id(a))
	}
	return ids
}

func ptrId(a *CiArtifact) int { return a.Id }

type orderScenario struct {
	f                   *orderFixture
	repo                *CiArtifactRepositoryImpl
	parent, parentCd    int
	builds              []int
	child, childCd      int
	childNewestFirst    []int // fan-out rows for builds 7,8 then backfilled 3..6, in parent build order
	linkedCd, linkedCdC int
	linkedCdRows        []int
}

// newOrderScenario: a parent with 8 builds (each with a post-CI plugin copy), a linked child holding
// builds 7-8 from parent runs plus 3-6 backfilled later, and a linked CD child synced out of order.
func newOrderScenario(t *testing.T) *orderScenario {
	f := newOrderFixture(t)
	s := &orderScenario{f: f, repo: NewCiArtifactRepositoryImpl(f.db, zap.NewNop().Sugar())}
	var parentApp int
	s.parent, parentApp = f.ciPipeline("CI_BUILD", 0)
	s.parentCd = f.cdPipeline(s.parent, parentApp)
	for i := 1; i <= 8; i++ {
		b := f.artifact(s.parent, 0, "CI-RUNNER", fmt.Sprintf("repo/app:v%d", i))
		f.artifact(s.parent, b, "post_ci", fmt.Sprintf("other/app:v%d", i))
		s.builds = append(s.builds, b)
	}
	var childApp int
	s.child, childApp = f.ciPipeline("LINKED", s.parent)
	s.childCd = f.cdPipeline(s.child, childApp)
	c7 := f.artifact(s.child, s.builds[6], "CI-RUNNER", "repo/app:v7")
	c8 := f.artifact(s.child, s.builds[7], "CI-RUNNER", "repo/app:v8")
	backfilled := make([]int, 0, 4)
	for i := 3; i <= 6; i++ {
		backfilled = append(backfilled, f.artifact(s.child, s.builds[i-1], "CI-RUNNER", fmt.Sprintf("repo/app:v%d", i)))
	}
	s.childNewestFirst = []int{c8, c7, backfilled[3], backfilled[2], backfilled[1], backfilled[0]}

	var linkedCdApp int
	s.linkedCd, linkedCdApp = f.ciPipeline("LINKED_CD", s.parentCd)
	s.linkedCdC = f.cdPipeline(s.linkedCd, linkedCdApp)
	l1 := f.artifact(s.linkedCd, s.builds[7], "CI-RUNNER", "repo/app:v8")
	l2 := f.artifact(s.linkedCd, s.builds[1], "CI-RUNNER", "repo/app:v2") // parent env rolled back to v2
	s.linkedCdRows = []int{l2, l1}
	return s
}

func (s *orderScenario) cdList(t *testing.T, cdPipelineId int, mutate func(*bean.ArtifactsListFilterOptions)) ([]int, int) {
	// the caller turns the linked order on only for a linked CI source
	opts := &bean.ArtifactsListFilterOptions{PipelineId: cdPipelineId, ParentStageType: bean.CI_WORKFLOW_TYPE, StageType: bean.CD_WORKFLOW_TYPE_DEPLOY,
		SearchString: EmptyLikeRegex, Limit: 50, UseLinkedCiOrder: cdPipelineId == s.childCd}
	if mutate != nil {
		mutate(opts)
	}
	artifacts, total, err := s.repo.GetArtifactsByCDPipelineV3(opts)
	require.NoError(t, err)
	return artifactIds(artifacts, ptrId), total
}

func TestIT_CdImageList_LinkedChildFollowsParentBuildOrder(t *testing.T) {
	s := newOrderScenario(t)
	got, total := s.cdList(t, s.childCd, nil)
	assert.Equal(t, s.childNewestFirst, got, "v8, v7, then backfilled v6..v3")
	assert.Equal(t, 6, total)
	assert.NotEqual(t, s.f.idsByIdDesc(s.child), got, "plain row id order would put the backfilled v6 on top")

}

func TestIT_CdImageList_PagingSearchAndExclude(t *testing.T) {
	s := newOrderScenario(t)
	var pages []int
	for offset := 0; offset < 6; offset += 2 {
		page, total := s.cdList(t, s.childCd, func(o *bean.ArtifactsListFilterOptions) { o.Limit, o.Offset = 2, offset })
		assert.Equal(t, 6, total)
		pages = append(pages, page...)
	}
	assert.Equal(t, s.childNewestFirst, pages, "pages are stable and complete")

	searched, total := s.cdList(t, s.childCd, func(o *bean.ArtifactsListFilterOptions) { o.SearchString = "%app:v%" })
	assert.Equal(t, s.childNewestFirst, searched)
	assert.Equal(t, 6, total)
	one, _ := s.cdList(t, s.childCd, func(o *bean.ArtifactsListFilterOptions) { o.SearchString = "%app:v5%" })
	assert.Equal(t, []int{s.childNewestFirst[3]}, one)

	excluded, total := s.cdList(t, s.childCd, func(o *bean.ArtifactsListFilterOptions) { o.ExcludeArtifactIds = []int{s.childNewestFirst[0]} })
	assert.Equal(t, s.childNewestFirst[1:], excluded)
	assert.Equal(t, 5, total)
}

func TestIT_CdImageList_RegularCiUnchanged(t *testing.T) {
	s := newOrderScenario(t)
	got, total := s.cdList(t, s.parentCd, nil)
	assert.Equal(t, s.f.idsByIdDesc(s.parent), got, "builds and plugin copies keep row id order")
	assert.Equal(t, 16, total)
}

func TestIT_CdImageList_LinkedCdUnchanged(t *testing.T) {
	s := newOrderScenario(t)
	got, _ := s.cdList(t, s.linkedCdC, nil)
	assert.Equal(t, s.linkedCdRows, got, "the image synced last stays on top even though it is an older build")
}

func TestIT_CdImageList_LinkedOrderIsSafeOnOtherSources(t *testing.T) {
	s := newOrderScenario(t)
	on := func(o *bean.ArtifactsListFilterOptions) { o.UseLinkedCiOrder = true }
	regular, _ := s.cdList(t, s.parentCd, on)
	assert.Equal(t, s.f.idsByIdDesc(s.parent), regular, "even with the flag on, non-linked rows keep row id order")
	linkedCd, _ := s.cdList(t, s.linkedCdC, on)
	assert.Equal(t, s.linkedCdRows, linkedCd)
	off := func(o *bean.ArtifactsListFilterOptions) { o.UseLinkedCiOrder = false }
	child, _ := s.cdList(t, s.childCd, off)
	assert.Equal(t, s.f.idsByIdDesc(s.child), child, "flag off is exactly the old row id order")
}

func TestIT_CdImageList_LinkedChildWithoutBackfillUnchanged(t *testing.T) {
	f := newOrderFixture(t)
	repo := NewCiArtifactRepositoryImpl(f.db, zap.NewNop().Sugar())
	parent, _ := f.ciPipeline("CI_BUILD", 0)
	child, childApp := f.ciPipeline("LINKED", parent)
	cd := f.cdPipeline(child, childApp)
	for i := 1; i <= 5; i++ {
		b := f.artifact(parent, 0, "CI-RUNNER", fmt.Sprintf("repo/x:v%d", i))
		f.artifact(child, b, "CI-RUNNER", fmt.Sprintf("repo/x:v%d", i))
	}
	artifacts, _, err := repo.GetArtifactsByCDPipelineV3(&bean.ArtifactsListFilterOptions{PipelineId: cd, ParentStageType: bean.CI_WORKFLOW_TYPE,
		SearchString: EmptyLikeRegex, Limit: 50, UseLinkedCiOrder: true})
	require.NoError(t, err)
	assert.Equal(t, f.idsByIdDesc(child), artifactIds(artifacts, ptrId), "children filled only by parent builds keep today's order")
}

func TestIT_WebhookImageListUnchanged(t *testing.T) {
	s := newOrderScenario(t)
	query, _ := BuildQueryForParentTypeCIOrWebhook(bean.ArtifactsListFilterOptions{ParentStageType: bean.WEBHOOK_WORKFLOW_TYPE, SearchString: EmptyLikeRegex, Limit: 5})
	assert.Contains(t, query, " ORDER BY cia.id DESC")
	assert.NotContains(t, query, "lcp")
	_ = s
}
