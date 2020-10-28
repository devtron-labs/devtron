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
	"github.com/devtron-labs/devtron/pkg/apis/devtron/v1"
	"github.com/devtron-labs/devtron/util"
)

var validateHolderFunc = []func(holder *v1.DataHolder, holderType string) error{validateHolderVersion, validateHolderClone, validateHolderCreate, validateHolderAppend, validateHolderEdit, validateHolderDelete}
var validSecretVersions = []string{"app/v1"}
var validConfigMapVersions = []string{"app/v1"}

// Below code is for secret and configmap validation

func validateConfigMap(holder *v1.DataHolder, props v1.InheritedProps) error {
	errs := make([]string, 0)
	holder.UpdateMissingProps(props)

	for _, f := range validateHolderFunc {
		errs = util.AppendErrorString(errs, f(holder, "configMap"))
	}
	return util.GetErrorOrNil(errs)
}

func validateSecret(holder *v1.DataHolder, props v1.InheritedProps) error {
	errs := make([]string, 0)
	holder.UpdateMissingProps(props)

	for _, f := range validateHolderFunc {
		errs = util.AppendErrorString(errs, f(holder, "secret"))
	}
	return util.GetErrorOrNil(errs)
}

func validateHolderVersion(holder *v1.DataHolder, holderType string) error {
	if holderType == "secret" && (len(holder.ApiVersion) == 0 || !util.ContainsString(validSecretVersions, holder.ApiVersion)) {
		return fmt.Errorf(v1.UnsupportedVersion, holder.ApiVersion, "secret")
	} else if len(holder.ApiVersion) == 0 || !util.ContainsString(validConfigMapVersions, holder.ApiVersion) {
		return fmt.Errorf(v1.UnsupportedVersion, holder.ApiVersion, "configMap")
	}
	return nil
}

func validateHolderClone(holder *v1.DataHolder, holderType string) error {
	if holder.GetOperation() != v1.Clone {
		return nil
	}
	errs := make([]string, 0)
	//source and destination cannot be same
	if v1.CompareResourcePath(holder.Source, holder.Destination) {
		errs = util.AppendErrorString(errs, fmt.Errorf(v1.SourceDestinationSame, holderType))
	}

	return util.GetErrorOrNil(errs)
}

func validateHolderCreate(holder *v1.DataHolder, holderType string) error {
	if holder.GetOperation() != v1.Create {
		return nil
	}
	errs := make([]string, 0)

	//if len(holder.Data) == 0 && (holder.External == nil || !*holder.External) {
	//	errs = util.AppendErrorString(errs, fmt.Errorf(v1.DataEmpty, holderType, v1.Create))
	//}
	return util.GetErrorOrNil(errs)
}

func validateHolderAppend(holder *v1.DataHolder, holderType string) error {
	if holder.GetOperation() != v1.Append {
		return nil
	}
	errs := make([]string, 0)

	//if holder.External != nil || holder.External {
	//	errs = util.AppendErrorString(errs, fmt.Errorf("external %s cannot be %s", holderType, v1.Append))
	//}

	if len(holder.Data) == 0 {
		errs = util.AppendErrorString(errs, fmt.Errorf(v1.DataEmpty, holderType, v1.Append))
	}
	return util.GetErrorOrNil(errs)
}

func validateHolderEdit(holder *v1.DataHolder, holderType string) error {
	if holder.GetOperation() != v1.Edit {
		return nil
	}
	errs := make([]string, 0)

	//if holder.External != nil || *holder.External {
	//	errs = util.AppendErrorString(errs, fmt.Errorf("external %s cannot be %s", holderType, v1.Edit))
	//}
	if len(holder.Data) == 0 {
		errs = util.AppendErrorString(errs, fmt.Errorf(v1.DataEmpty, holderType, v1.Edit))
	}
	return util.GetErrorOrNil(errs)
}

func validateHolderDelete(holder *v1.DataHolder, holderType string) error {
	if holder.GetOperation() != v1.Delete {
		return nil
	}
	errs := make([]string, 0)
	//if holder.Environment != nil {
	//	errs = util.AppendErrorString(errs, fmt.Errorf("cannot delete %s specifically for environment", holderType))
	//}
	return util.GetErrorOrNil(errs)
}
