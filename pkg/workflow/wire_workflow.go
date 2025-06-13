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

package workflow

import (
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	"github.com/devtron-labs/devtron/pkg/workflow/status"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/hook"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/repository"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/service"
	"github.com/google/wire"
)

var WorkflowWireSet = wire.NewSet(
	cd.CdWorkflowWireSet,
	status.WorkflowStatusWireSet,
	hook.NewTriggerAuditHookImpl,
	wire.Bind(new(hook.TriggerAuditHook), new(*hook.TriggerAuditHookImpl)),
	service.NewWorkflowTriggerAuditServiceImpl,
	wire.Bind(new(service.WorkflowTriggerAuditService), new(*service.WorkflowTriggerAuditServiceImpl)),
	repository.NewWorkflowConfigSnapshotRepositoryImpl,
	wire.Bind(new(repository.WorkflowConfigSnapshotRepository), new(*repository.WorkflowConfigSnapshotRepositoryImpl)),
)
