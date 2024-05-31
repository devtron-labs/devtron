/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
