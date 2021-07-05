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

package restHandler

import (
	"github.com/devtron-labs/devtron/pkg/apis/devtron/v1"
	"testing"
)

func Test_validatePipeline(t *testing.T) {
	type args struct {
		pipeline *v1.Pipeline
		props    v1.InheritedProps
		err      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Missing Operation",
			args: args{
				pipeline: &v1.Pipeline{
					Build: &v1.Build{
						ApiVersion:       "app/v1",
						Destination:      nil,
						DockerArguments:  nil,
						NextPipeline:     nil,
						Operation:        v1.Clone,
						PostBuild:        nil,
						PreBuild:         nil,
						PreviousPipeline: nil,
						Repo:             nil,
						Source:           nil,
						Trigger:          nil,
					},
					Deployment: &v1.Deployment{
						ApiVersion:       "app/v1",
						ConfigMaps:       nil,
						Destination:      nil,
						Environment:      nil,
						NextPipeline:     nil,
						Operation:        v1.Clone,
						PostDeployment:   nil,
						PreDeployment:    nil,
						PreviousPipeline: nil,
						Secrets:          nil,
						Source:           nil,
						Trigger:          nil,
					},
				},
				props: v1.InheritedProps{
					Destination: nil,
					Operation:   "",
					Source:      nil,
				},
				err: "no operation defined for build",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePipeline(tt.args.pipeline, tt.args.props); (err != nil) != tt.wantErr || (err != nil && err.Error() != tt.args.err) {
				t.Errorf("validatePipeline() error = %v, wantErr %v, error string %s", err, tt.wantErr, tt.args.err)
			}
		})
	}

