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
	"testing"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/stretchr/testify/assert"
)

func TestLinkedCiAwareOrderBy(t *testing.T) {
	assert.Equal(t, "CASE WHEN cia.parent_ci_artifact > 0 AND cia.pipeline_id IN (SELECT id FROM ci_pipeline WHERE ci_pipeline_type = 'LINKED')"+
		" THEN cia.parent_ci_artifact ELSE cia.id END DESC, cia.id DESC", LinkedCiAwareOrderBy("cia"))
}

func TestBuildQueryForParentTypeCIOrWebhook_OrderBy(t *testing.T) {
	regularQuery, _ := BuildQueryForParentTypeCIOrWebhook(bean.ArtifactsListFilterOptions{ParentStageType: bean.CI_WORKFLOW_TYPE, SearchString: EmptyLikeRegex, Limit: 5})
	assert.Contains(t, regularQuery, " ORDER BY cia.id DESC LIMIT ? OFFSET ?", "non-linked sources keep the old query")
	linkedQuery, _ := BuildQueryForParentTypeCIOrWebhook(bean.ArtifactsListFilterOptions{ParentStageType: bean.CI_WORKFLOW_TYPE, SearchString: EmptyLikeRegex, Limit: 5, UseLinkedCiOrder: true})
	assert.Contains(t, linkedQuery, " ORDER BY "+LinkedCiAwareOrderBy("cia")+" LIMIT ? OFFSET ?")
	for _, useLinkedCiOrder := range []bool{false, true} {
		webhookQuery, _ := BuildQueryForParentTypeCIOrWebhook(bean.ArtifactsListFilterOptions{ParentStageType: bean.WEBHOOK_WORKFLOW_TYPE, SearchString: EmptyLikeRegex, Limit: 5, UseLinkedCiOrder: useLinkedCiOrder})
		assert.Contains(t, webhookQuery, " ORDER BY cia.id DESC LIMIT ? OFFSET ?")
	}
}
