/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package batch

import (
	v1 "github.com/devtron-labs/devtron/pkg/apis/devtron/v1"
	"go.uber.org/zap"
)

type TaskAction interface {
	Execute(task *v1.Task, taskType string) error
}

type TaskActionImpl struct {
	logger *zap.SugaredLogger
}

func NewTaskAction(logger *zap.SugaredLogger) *TaskActionImpl {
	dh := &TaskActionImpl{
		logger: logger,
	}
	return dh
}

var taskExecutor = []func(impl TaskActionImpl, holder *v1.Task, dataType string) error{executeTaskCreate}

func (impl TaskActionImpl) Execute(task *v1.Task, taskType string) error {
	return nil
}

func executeTaskCreate(impl TaskActionImpl, task *v1.Task, dataType string) error {
	if task.Operation != v1.Create {
		return nil
	}

	return nil
}
