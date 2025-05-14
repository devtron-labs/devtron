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

package imageScanning

import (
	repository2 "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository"
	securityBean "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository/bean"
	"testing"
)

func TestPolicyServiceImpl_HasBlockedCVE(t *testing.T) {
	type args struct {
		cves           []*repository2.CveStore
		cvePolicy      map[string]*repository2.CvePolicy
		severityPolicy map[securityBean.Severity]*repository2.CvePolicy
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "Test 1",
			args: args{
				cves: []*repository2.CveStore{
					{
						Name: "abc",
					},
					{
						Severity: security.Low,
					},
				},
				cvePolicy: map[string]*repository2.CvePolicy{
					"abc": {
						Action: security.Allow,
					},
				},
				severityPolicy: map[security.Severity]*repository2.CvePolicy{
					security.Low: {
						Action: security.Allow,
					},
				},
			},
			want: false,
		},
		{
			name: "Test 2",
			args: args{
				cves: []*repository2.CveStore{
					{
						Name: "abc",
					},
				},
				cvePolicy: map[string]*repository2.CvePolicy{
					"abc": {
						Action: security.Block,
					},
				},
				severityPolicy: map[security.Severity]*repository2.CvePolicy{},
			},
			want: true,
		},
		{
			name: "Test 3",
			args: args{
				cves: []*repository2.CveStore{
					{
						Severity: security.High,
					},
				},
				cvePolicy: map[string]*repository2.CvePolicy{},
				severityPolicy: map[security.Severity]*repository2.CvePolicy{
					security.High: {
						Action: security.Block,
					},
				},
			},
			want: true,
		},
		{
			name: "Test 4",
			args: args{
				cves: []*repository2.CveStore{
					{
						Name:         "abc",
						FixedVersion: "1.0.0",
					},
				},
				cvePolicy: map[string]*repository2.CvePolicy{
					"abc": {
						Action: security.Blockiffixed,
					},
				},
				severityPolicy: map[security.Severity]*repository2.CvePolicy{},
			},
			want: true,
		},
		{
			name: "Test 5",
			args: args{
				cves: []*repository2.CveStore{
					{
						Name: "abc",
					},
				},
				cvePolicy: map[string]*repository2.CvePolicy{
					"abc": {
						Action: security.Blockiffixed,
					},
				},
				severityPolicy: map[security.Severity]*repository2.CvePolicy{},
			},
			want: false,
		},
		{
			name: "Test 6",
			args: args{
				cves: []*repository2.CveStore{
					{
						Severity:     security.High,
						FixedVersion: "1.0.0",
					},
				},
				cvePolicy: map[string]*repository2.CvePolicy{},
				severityPolicy: map[security.Severity]*repository2.CvePolicy{
					security.High: {
						Action: security.Blockiffixed,
					},
				},
			},
			want: true,
		},
		{
			name: "Test 7",
			args: args{
				cves: []*repository2.CveStore{
					{
						Severity: security.High,
					},
				},
				cvePolicy: map[string]*repository2.CvePolicy{},
				severityPolicy: map[security.Severity]*repository2.CvePolicy{
					security.High: {
						Action: security.Blockiffixed,
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &PolicyServiceImpl{}
			if got := impl.HasBlockedCVE(tt.args.cves, tt.args.cvePolicy, tt.args.severityPolicy); got != tt.want {
				t.Errorf("HasBlockedCVE() = %v, want %v", got, tt.want)
			}
		})
	}
}
