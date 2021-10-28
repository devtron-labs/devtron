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

package validation

import (
	"fmt"

	v1 "github.com/devtron-labs/devtron/pkg/apis/devtron/v1"
	"github.com/devtron-labs/devtron/util"
)

var validatePrePostDeploymentFunc = []func(task *v1.Task) error{validatePrePostDeploymentVersion, validatePrePostDeploymentClone,
	validatePrePostDeploymentCreate, validatePrePostDeploymentEdit, validatePrePostDeploymentAppend, validatePrePostDeploymentDelete}
var validPrePostDeploymentVersions = []string{"app/v1"}

var validateStageFunc = []func(stage *v1.Stage) error{validateStageAppend, validateStageCreate, validateStageEdit, validateStageDelete}

func validatePrePostDeployment(task *v1.Task, props v1.InheritedProps) error {
	errs := make([]string, 0)
	task.UpdateMissingProps(props)

	//source should be same as parent
	errs = util.AppendErrorString(errs, task.CompareSource(props.Source))

	//destination should be same as parent
	errs = util.AppendErrorString(errs, task.CompareDestination(props.Destination))

	//is destination uniquely identifiable
	errs = util.AppendErrorString(errs, validatePrePostDeploymentDestination(task.Destination))

	for _, f := range validatePrePostDeploymentFunc {
		errs = util.AppendErrorString(errs, f(task))
	}

	for _, f := range validateStageFunc {
		for i, _ := range task.Stages {
			errs = util.AppendErrorString(errs, f(&task.Stages[i]))
		}
	}

	return util.GetErrorOrNil(errs)
}

func validatePrePostDeploymentDestination(destination *v1.ResourcePath) error {
	//is destination unique
	if destination.Workflow != nil || destination.Pipeline != nil {
		return fmt.Errorf(v1.DestinationNotUnique, "task")
	}
	//does destination exist
	return nil
}

func validatePrePostDeploymentVersion(task *v1.Task) error {
	if task.ApiVersion == "" || !util.ContainsString(validPrePostDeploymentVersions, task.ApiVersion) {
		return fmt.Errorf(v1.UnsupportedVersion, task.ApiVersion, "task")
	}
	return nil
}

func validatePrePostDeploymentClone(task *v1.Task) error {
	if task.GetOperation() != v1.Clone {
		return nil
	}
	errs := make([]string, 0)

	//source should be uniquely identifiable
	errs = util.AppendErrorString(errs, validatePrePostDeploymentSourceClone(task.Source))
	//validate destination match
	return util.GetErrorOrNil(errs)
}

func validatePrePostDeploymentSourceClone(source *v1.ResourcePath) error {
	//is source unique
	if source.Workflow != nil || source.Pipeline != nil {
		return fmt.Errorf(v1.SourceNotUnique, "task")
	}
	//does source exist
	return nil
}

func validatePrePostDeploymentCreate(task *v1.Task) error {
	if task.GetOperation() != v1.Clone {
		return nil
	}
	errs := make([]string, 0)
	//should have stages
	if len(task.Stages) == 0 {
		util.AppendErrorString(errs, fmt.Errorf(v1.StagesMissing))
	}
	return util.GetErrorOrNil(errs)
}

func validatePrePostDeploymentEdit(task *v1.Task) error {
	if task.GetOperation() != v1.Clone {
		return nil
	}
	errs := make([]string, 0)
	//should have stages
	if len(task.Stages) == 0 {
		util.AppendErrorString(errs, fmt.Errorf(v1.StagesMissing))
	}
	return util.GetErrorOrNil(errs)
}

func validatePrePostDeploymentDelete(task *v1.Task) error {
	if task.GetOperation() != v1.Clone {
		return nil
	}
	errs := make([]string, 0)

	return util.GetErrorOrNil(errs)
}

func validatePrePostDeploymentAppend(task *v1.Task) error {
	if task.GetOperation() != v1.Clone {
		return nil
	}
	errs := make([]string, 0)
	if len(task.Stages) == 0 {
		errs = util.AppendErrorString(errs, fmt.Errorf(v1.StagesMissing))
	}
	for i, _ := range task.Stages {
		errs = util.AppendErrorString(errs, validateStageAppend(&task.Stages[i]))
	}
	return util.GetErrorOrNil(errs)
}

//validations for stage

func validateStageCreate(stage *v1.Stage) error {
	if stage.Operation != v1.Create {
		return nil
	}
	if stage.Script == nil {
		return fmt.Errorf(v1.ScriptMissing)
	}
	return nil
}

func validateStageEdit(stage *v1.Stage) error {
	if stage.Operation != v1.Edit {
		return nil
	}
	if stage.Position == nil && stage.Name == "" {
		return fmt.Errorf("name and position cannot be null in stage edit")
	}
	if stage.Script == nil {
		return fmt.Errorf(v1.ScriptMissing)
	}
	return nil
}

func validateStageDelete(stage *v1.Stage) error {
	if stage.Operation != v1.Delete {
		return nil
	}
	if stage.Position == nil && stage.Name == "" {
		return fmt.Errorf("name and position cannot be null in stage delete")
	}
	return nil
}

func validateStageAppend(stage *v1.Stage) error {
	if stage.Operation != v1.Append {
		return nil
	}
	if stage.Script == nil {
		return fmt.Errorf(v1.ScriptMissing)
	}
	return nil
}
