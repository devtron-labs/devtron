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

import (
	"testing"
)

func Test_updatePath(t *testing.T) {

	toString := func(st string) *string { return &st }
	compare := func(first, second *ResourcePath) bool {
		appSame := (first.App == nil && second.App == nil) || *first.App == *second.App
		configMapSame := (first.Configmap == nil && second.Configmap == nil) || *first.Configmap == *second.Configmap
		environmentSame := (first.Environment == nil && second.Environment == nil) || *first.Environment == *second.Environment
		pipelineSame := (first.Pipeline == nil && second.Pipeline == nil) || *first.Pipeline == *second.Pipeline
		secretSame := (first.Secret == nil && second.Secret == nil) || *first.Secret == *second.Secret
		uidSame := (first.Uid == nil && second.Uid == nil) || *first.Uid == *second.Uid
		workflowSame := (first.Workflow == nil && second.Workflow == nil) || *first.Workflow == *second.Workflow
		return appSame && configMapSame && environmentSame && pipelineSame && secretSame && uidSame && workflowSame
	}
	type args struct {
		to   *ResourcePath
		from *ResourcePath
		out  *ResourcePath
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Update All",
			args: args{
				to: &ResourcePath{
					App:         nil,
					Configmap:   nil,
					Environment: nil,
					Pipeline:    nil,
					Secret:      nil,
					Uid:         nil,
					Workflow:    nil,
				},
				from: &ResourcePath{
					App:         toString("App-f"),
					Configmap:   toString("cf-f"),
					Environment: toString("env-f"),
					Pipeline:    toString("pip-f"),
					Secret:      toString("sec-f"),
					Uid:         toString("uid-f"),
					Workflow:    toString("wf-f"),
				},
				out: &ResourcePath{
					App:         toString("App-g"),
					Configmap:   toString("cf-f"),
					Environment: toString("env-f"),
					Pipeline:    toString("pip-f"),
					Secret:      toString("sec-f"),
					Uid:         toString("uid-f"),
					Workflow:    toString("wf-f"),
				},
			},
		}, {
			name: "One not to be overriden",
			args: args{
				to: &ResourcePath{
					App:         nil,
					Configmap:   nil,
					Environment: nil,
					Pipeline:    toString("pip-t"),
					Secret:      nil,
					Uid:         nil,
					Workflow:    nil,
				},
				from: &ResourcePath{
					App:         toString("App-f"),
					Configmap:   toString("cf-f"),
					Environment: toString("env-f"),
					Pipeline:    toString("pip-f"),
					Secret:      toString("sec-f"),
					Uid:         toString("uid-f"),
					Workflow:    toString("wf-f"),
				},
				out: &ResourcePath{
					App:         toString("App-g"),
					Configmap:   toString("cf-f"),
					Environment: toString("env-f"),
					Pipeline:    toString("pip-t"),
					Secret:      toString("sec-f"),
					Uid:         toString("uid-f"),
					Workflow:    toString("wf-f"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatePath(tt.args.to, tt.args.from)
			if compare(tt.args.to, tt.args.out) {
				t.Errorf("not match expected :%+v found :%+v\n", *tt.args.out, *tt.args.to)
			}
		})
	}
}
