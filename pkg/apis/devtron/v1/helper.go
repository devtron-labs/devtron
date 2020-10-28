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

package v1

import "fmt"

type Component interface {
	GetOperation() Operation
	GetProps() InheritedProps
	UpdateMissingProps(props InheritedProps) error
	CompareSource(path *ResourcePath) error
	CompareDestination(path *ResourcePath) error
}

func (o *Build) GetOperation() Operation {
	return o.Operation
}

func (o *Build) GetProps() InheritedProps {
	return InheritedProps{
		Destination: o.Destination,
		Operation:   o.Operation,
		Source:      o.Source,
	}
}

func (o *Build) UpdateMissingProps(props InheritedProps) error {
	if len(o.Operation) == 0 {
		o.Operation = props.Operation
	}
	if o.Destination == nil {
		o.Destination = props.Destination
	} else {
		updatePath(o.Destination, props.Destination)
	}
	if o.Source == nil {
		o.Source = props.Source
	} else {
		updatePath(o.Source, props.Source)
	}
	return nil
}

func (o *Deployment) GetOperation() Operation {
	return o.Operation
}

func (o *Deployment) GetProps() InheritedProps {
	return InheritedProps{
		Destination: o.Destination,
		Operation:   o.Operation,
		Source:      o.Source,
	}
}

func (o *Deployment) UpdateMissingProps(props InheritedProps) error {
	if len(o.Operation) == 0 {
		o.Operation = props.Operation
	}
	if o.Destination == nil {
		o.Destination = props.Destination
	} else {
		updatePath(o.Destination, props.Destination)
	}
	if o.Source == nil {
		o.Source = props.Source
	} else {
		updatePath(o.Source, props.Source)
	}
	return nil
}

func (o *Deployment) CompareSource(with *ResourcePath) error {
	same := compareDeployment(o.Source, with)
	if !same {
		err := fmt.Errorf(SourceNotSame, "deployment", "parent")
		return err
	}
	return nil
}

func (o *Deployment) CompareDestination(with *ResourcePath) error {
	same := compareDeployment(o.Destination, with)
	if !same {
		err := fmt.Errorf(DestinationNotSame, "deployment", "parent")
		return err
	}
	return nil
}

func compareDeployment(resource *ResourcePath, with *ResourcePath) bool {
	if with == nil {
		return true
	}
	appSame := (resource.App == nil && with.App == nil) || *resource.App == *with.App
	pipelineSame := (resource.Pipeline == nil && with.Pipeline == nil) || *resource.Pipeline == *with.Pipeline
	uidSame := (resource.Uid == nil && with.Uid == nil) || *resource.Uid == *with.Uid
	workflowSame := (resource.Workflow == nil && with.Workflow == nil) || *resource.Workflow == *with.Workflow
	if !(appSame && pipelineSame && uidSame && workflowSame) {
		return false
	}
	return true
}

func (o *DataHolder) GetOperation() Operation {
	return o.Operation
}

func (o *DataHolder) GetProps() InheritedProps {
	return InheritedProps{
		Destination: o.Destination,
		Operation:   o.Operation,
		Source:      o.Source,
	}
}

func (o *DataHolder) UpdateMissingProps(props InheritedProps) error {
	if len(o.Operation) == 0 {
		o.Operation = props.Operation
	}
	if o.Destination == nil {
		o.Destination = props.Destination
	} else {
		updatePath(o.Destination, props.Destination)
	}
	if o.Source == nil {
		o.Source = props.Source
	} else {
		updatePath(o.Source, props.Source)
	}
	return nil
}

func (o *Task) GetOperation() Operation {
	return o.Operation
}

func (o *Task) GetProps() InheritedProps {
	return InheritedProps{
		Destination: o.Destination,
		Operation:   o.Operation,
		Source:      o.Source,
	}
}

func (o *Task) UpdateMissingProps(props InheritedProps) error {
	if len(o.Operation) == 0 {
		o.Operation = props.Operation
	}
	if o.Destination == nil {
		o.Destination = props.Destination
	} else {
		updatePath(o.Destination, props.Destination)
	}
	if o.Source == nil {
		o.Source = props.Source
	} else {
		updatePath(o.Source, props.Source)
	}
	return nil
}

func (o *Task) CompareSource(with *ResourcePath) error {
	return compareTask(o.Source, with)
}

func (o *Task) CompareDestination(with *ResourcePath) error {
	return compareTask(o.Destination, with)
}

func (o *DeploymentTemplate) GetOperation() Operation {
	return o.Operation
}

func (o *DeploymentTemplate) GetProps() InheritedProps {
	return InheritedProps{
		Destination: o.Destination,
		Operation:   o.Operation,
		Source:      o.Source,
	}
}

func (o *DeploymentTemplate) UpdateMissingProps(props InheritedProps) error {
	if len(o.Operation) == 0 {
		o.Operation = props.Operation
	}
	if o.Destination == nil {
		o.Destination = props.Destination
	} else {
		updatePath(o.Destination, props.Destination)
	}
	if o.Source == nil {
		o.Source = props.Source
	} else {
		updatePath(o.Source, props.Source)
	}
	return nil
}

func (o *DeploymentTemplate) CompareSource(with *ResourcePath) error {
	return compareTask(o.Source, with)
}

func (o *DeploymentTemplate) CompareDestination(with *ResourcePath) error {
	return compareTask(o.Destination, with)
}

func (o *Workflow) GetOperation() Operation {
	return o.Operation
}

func (o *Workflow) GetProps() InheritedProps {
	return InheritedProps{
		Destination: o.Destination,
		Operation:   o.Operation,
		Source:      o.Source,
	}
}

func (o *Workflow) UpdateMissingProps(props InheritedProps) error {
	if len(o.Operation) == 0 {
		o.Operation = props.Operation
	}
	if o.Destination == nil {
		o.Destination = props.Destination
	} else {
		updatePath(o.Destination, props.Destination)
	}
	if o.Source == nil {
		o.Source = props.Source
	} else {
		updatePath(o.Source, props.Source)
	}
	return nil
}

func (o *Workflow) CompareSource(with *ResourcePath) error {
	return compareTask(o.Source, with)
}

func (o *Workflow) CompareDestination(with *ResourcePath) error {
	return compareTask(o.Destination, with)
}

func compareTask(resource *ResourcePath, with *ResourcePath) error {
	if with == nil {
		return nil
	}
	appSame := (resource.App == nil && with.App == nil) || *resource.App == *with.App
	pipelineSame := (resource.Pipeline == nil && with.Pipeline == nil) || *resource.Pipeline == *with.Pipeline
	uidSame := (resource.Uid == nil && with.Uid == nil) || *resource.Uid == *with.Uid
	workflowSame := (resource.Workflow == nil && with.Workflow == nil) || *resource.Workflow == *with.Workflow
	if !(appSame && pipelineSame && uidSame && workflowSame) {
		return fmt.Errorf(DestinationNotSame, "task", "parent")
	}
	return nil
}

//from cannot be null
func updatePath(to *ResourcePath, from *ResourcePath) {
	if to == nil {
		to = from
		return
	}
	if from == nil {
		return
	}
	if to.App == nil {
		to.App = from.App
	}
	if to.Workflow == nil {
		to.Workflow = from.Workflow
	}
	if to.Pipeline == nil {
		to.Pipeline = from.Pipeline
	}
	if to.ConfigMap == nil {
		to.ConfigMap = from.ConfigMap
	}
	if to.Secret == nil {
		to.Secret = from.Secret
	}
	if to.Environment == nil {
		to.Environment = from.Environment
	}
	if to.Uid == nil {
		to.Uid = from.Uid
	}
}

func CompareResourcePath(first, second *ResourcePath) bool {
	if first == nil && second == nil {
		return true
	}
	if (first == nil && second != nil) || (first != nil && second == nil) {
		return false
	}
	appSame := (first.App == nil && second.App == nil) || *first.App == *second.App
	configMapSame := (first.ConfigMap == nil && second.ConfigMap == nil) || *first.ConfigMap == *second.ConfigMap
	environmentSame := (first.Environment == nil && second.Environment == nil) || *first.Environment == *second.Environment
	pipelineSame := (first.Pipeline == nil && second.Pipeline == nil) || *first.Pipeline == *second.Pipeline
	secretSame := (first.Secret == nil && second.Secret == nil) || *first.Secret == *second.Secret
	uidSame := (first.Uid == nil && second.Uid == nil) || *first.Uid == *second.Uid
	workflowSame := (first.Workflow == nil && second.Workflow == nil) || *first.Workflow == *second.Workflow
	return appSame && configMapSame && environmentSame && pipelineSame && secretSame && uidSame && workflowSame
}
