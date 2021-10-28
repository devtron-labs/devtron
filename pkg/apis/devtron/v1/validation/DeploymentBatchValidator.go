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

var validateDeploymentFunc = []func(deployment *v1.Deployment) error{validateDeploymentVersion, validateDeploymentClone, validateDeploymentCreate}
var validDeploymentVersions = []string{"app/v1"}

func ValidateDeployment(deployment *v1.Deployment, props v1.InheritedProps) error {
	if len(deployment.GetOperation()) == 0 {
		return fmt.Errorf(v1.OperationUndefinedError, "deployment")
	}
	errs := make([]string, 0)

	deployment.UpdateMissingProps(props)
	errs = util.AppendErrorString(errs, deployment.CompareSource(props.Source))

	errs = util.AppendErrorString(errs, deployment.CompareDestination(props.Destination))

	for i := range deployment.ConfigMaps {
		errs = util.AppendErrorString(errs, validateConfigMap(&deployment.ConfigMaps[i], deployment.GetProps()))
	}

	for i := range deployment.Secrets {
		errs = util.AppendErrorString(errs, validateSecret(&deployment.Secrets[i], deployment.GetProps()))
	}

	errs = util.AppendErrorString(errs, validatePrePostDeployment(deployment.PreDeployment, deployment.GetProps()))

	errs = util.AppendErrorString(errs, validatePrePostDeployment(deployment.PostDeployment, deployment.GetProps()))

	//is destination uniquely identifiable
	errs = util.AppendErrorString(errs, validateDeploymentDestination(deployment.Destination))

	for _, f := range validateDeploymentFunc {
		errs = util.AppendErrorString(errs, f(deployment))
	}

	return util.GetErrorOrNil(errs)
}

func validateDeploymentDestination(destination *v1.ResourcePath) error {
	//is destination unique
	if destination.Workflow != nil || destination.Pipeline != nil {
		return fmt.Errorf(v1.DestinationNotUnique, "deployment")
	}
	//does destination exist

	return nil
}

func validateDeploymentVersion(deployment *v1.Deployment) error {
	if deployment.ApiVersion == "" || !util.ContainsString(validDeploymentVersions, deployment.ApiVersion) {
		return fmt.Errorf(v1.UnsupportedVersion, deployment.ApiVersion, "deployment")
	}
	return nil
}

func validateDeploymentClone(deployment *v1.Deployment) error {
	if deployment.GetOperation() != v1.Clone {
		return nil
	}

	errs := make([]string, 0)
	//is source uniquely identifiable
	errs = util.AppendErrorString(errs, validateDeploymentSourceClone(deployment.Source))

	//check that environment doesnt already exist in this application

	//Source and destination cannot be same
	if v1.CompareResourcePath(deployment.Source, deployment.Destination) {
		errs = util.AppendErrorString(errs, fmt.Errorf(v1.SourceDestinationSame, "deployment"))
	}

	return util.GetErrorOrNil(errs)
}

func validateDeploymentSourceClone(source *v1.ResourcePath) error {
	//is source unique
	if source.Workflow != nil || source.Pipeline != nil {
		return fmt.Errorf(v1.SourceNotUnique, "deployment")
	}
	//does source exist
	return nil
}

func validateDeploymentCreate(deployment *v1.Deployment) error {
	if deployment.GetOperation() != v1.Clone {
		return nil
	}

	errs := make([]string, 0)
	//is source uniquely identifiable
	errs = util.AppendErrorString(errs, validateDeploymentSourceClone(deployment.Source))

	//stages cannot be nil
	if deployment.Destination.Environment == nil {
		errs = append(errs, fmt.Errorf(v1.EnvironmentEmpty, "deployment", "create").Error())
	}

	if deployment.Strategy.BlueGreen == nil && deployment.Strategy.Canary == nil && deployment.Strategy.Rolling == nil && deployment.Strategy.Recreate == nil {
		errs = append(errs, fmt.Errorf("strategy cannot be empty for deployment create").Error())
	}

	return util.GetErrorOrNil(errs)
}
