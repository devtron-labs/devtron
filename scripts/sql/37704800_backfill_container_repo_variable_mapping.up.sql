/*
 * Copyright (c) 2026. Devtron Inc.
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

-- Backfills variable_entity_mapping for Container Repository fields saved before this
-- release started tracking them (entity_type 10/11). Without this, a variable used only
-- by an already-saved docker_repository value has no mapping row and deletion protection
-- doesn't see it as in use. Mirrors the write-side extraction in
-- CiTemplateService.TrackContainerRepositoryUsage - exactly one full @{{VAR_NAME}} token,
-- same regex as RepoFieldVariableResolver.

INSERT INTO variable_entity_mapping (variable_name, entity_type, entity_id, is_deleted, created_on, created_by, updated_on, updated_by)
SELECT substring(ct.docker_repository from '^@\{\{([A-Za-z0-9_]+)\}\}$'), 10, ct.id, false, now(), 1, now(), 1
FROM ci_template ct
WHERE ct.active = true
  AND ct.docker_repository ~ '^@\{\{[A-Za-z0-9_]+\}\}$'
  AND NOT EXISTS (
      SELECT 1 FROM variable_entity_mapping vem
      WHERE vem.entity_type = 10 AND vem.entity_id = ct.id AND vem.is_deleted = false
  );

INSERT INTO variable_entity_mapping (variable_name, entity_type, entity_id, is_deleted, created_on, created_by, updated_on, updated_by)
SELECT substring(cto.docker_repository from '^@\{\{([A-Za-z0-9_]+)\}\}$'), 11, cto.id, false, now(), 1, now(), 1
FROM ci_template_override cto
WHERE cto.active = true
  AND cto.docker_repository ~ '^@\{\{[A-Za-z0-9_]+\}\}$'
  AND NOT EXISTS (
      SELECT 1 FROM variable_entity_mapping vem
      WHERE vem.entity_type = 11 AND vem.entity_id = cto.id AND vem.is_deleted = false
  );
