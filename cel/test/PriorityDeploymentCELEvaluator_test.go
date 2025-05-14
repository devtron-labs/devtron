/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package test

import (
	"fmt"
	"github.com/devtron-labs/devtron/cel"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvaluatorServiceImpl_EvaluateCELRequest(t *testing.T) {
	type args struct {
		request cel.Request
	}
	log, err := util.NewSugardLogger()
	if err != nil {
		log.Panic(err)
	}
	impl := cel.NewCELServiceImpl(log)
	testParams := getTestParams()
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "TEST 1: Priority Deployment Default Case",
			args: args{
				request: cel.Request{
					Expression: "isProdEnv == true",
					ExpressionMetadata: cel.ExpressionMetadata{
						Params: testParams,
					},
				},
			},
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name: "TEST 2: Priority Deployment Custom Case",
			args: args{
				request: cel.Request{
					Expression: "appName.startsWith('test') && cdPipelineName.startsWith('test')",
					ExpressionMetadata: cel.ExpressionMetadata{
						Params: testParams,
					},
				},
			},
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name: "TEST 3: Priority Deployment Custom Case",
			args: args{
				request: cel.Request{
					Expression: "chartRefId in [15, 16]",
					ExpressionMetadata: cel.ExpressionMetadata{
						Params: testParams,
					},
				},
			},
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name: "TEST 4: Invalid CEL Expression Case",
			args: args{
				request: cel.Request{
					Expression: "IsProdEnv == (test)",
					ExpressionMetadata: cel.ExpressionMetadata{
						Params: testParams,
					},
				},
			},
			want:    false,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := impl.EvaluateCELRequest(tt.args.request)
			if !tt.wantErr(t, err, fmt.Sprintf("EvaluateCELRequest(Expression: %v, ExpressionMetadata: %v)", tt.args.request.Expression, tt.args.request.ExpressionMetadata)) {
				return
			}
			assert.Equalf(t, tt.want, got, "EvaluateCELRequest(Expression: %v, ExpressionMetadata: %v)", tt.args.request.Expression, tt.args.request.ExpressionMetadata)
		})
	}
}

func getTestParams() []cel.ExpressionParam {
	return []cel.ExpressionParam{
		{
			ParamName: cel.AppName,
			Value:     "test-app",
			Type:      cel.ParamTypeString,
		},
		{
			ParamName: cel.CdPipelineName,
			Value:     "test-pipeline",
			Type:      cel.ParamTypeString,
		},
		{
			ParamName: cel.IsProdEnv,
			Value:     true,
			Type:      cel.ParamTypeBool,
		},
		{
			ParamName: cel.ChartRefId,
			Value:     15,
			Type:      cel.ParamTypeInteger,
		},
	}
}
