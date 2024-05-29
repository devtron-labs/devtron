/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package security

import (
	securityBean "github.com/devtron-labs/devtron/internal/sql/repository/security/bean"
	"github.com/go-pg/pg"
	"reflect"
	"testing"
)

func TestCvePolicyRepositoryImpl_enforceCvePolicy(t *testing.T) {
	type fields struct {
		dbConnection *pg.DB
	}
	type args struct {
		cves           []*CveStore
		cvePolicy      map[string]*CvePolicy
		severityPolicy map[securityBean.Severity]*CvePolicy
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantBlockedCVE []*CveStore
	}{
		// TODO: Add test cases.
		{
			name: "Test 1",
			args: args{
				cves: []*CveStore{
					{
						Name: "abc",
					},
					{
						Severity: securityBean.Low,
					},
				},
				cvePolicy: map[string]*CvePolicy{
					"abc": {
						Action: securityBean.Allow,
					},
				},
				severityPolicy: map[securityBean.Severity]*CvePolicy{
					securityBean.Low: {
						Action: securityBean.Allow,
					},
				},
			},
			wantBlockedCVE: nil,
		},
		{
			name: "Test 2",
			args: args{
				cves: []*CveStore{
					{
						Name: "abc",
					},
				},
				cvePolicy: map[string]*CvePolicy{
					"abc": {
						Action: securityBean.Block,
					},
				},
				severityPolicy: map[securityBean.Severity]*CvePolicy{},
			},
			wantBlockedCVE: []*CveStore{
				{
					Name: "abc",
				},
			},
		},
		{
			name: "Test 3",
			args: args{
				cves: []*CveStore{
					{
						Severity: securityBean.High,
					},
				},
				cvePolicy: map[string]*CvePolicy{},
				severityPolicy: map[securityBean.Severity]*CvePolicy{
					securityBean.High: {
						Action: securityBean.Block,
					},
				},
			},
			wantBlockedCVE: []*CveStore{
				{
					Severity: securityBean.High,
				},
			},
		},
		{
			name: "Test 4",
			args: args{
				cves: []*CveStore{
					{
						Name:         "abc",
						FixedVersion: "1.0.0",
					},
				},
				cvePolicy: map[string]*CvePolicy{
					"abc": {
						Action: securityBean.Blockiffixed,
					},
				},
				severityPolicy: map[securityBean.Severity]*CvePolicy{},
			},
			wantBlockedCVE: []*CveStore{
				{
					Name:         "abc",
					FixedVersion: "1.0.0",
				},
			},
		},
		{
			name: "Test 5",
			args: args{
				cves: []*CveStore{
					{
						Name: "abc",
					},
				},
				cvePolicy: map[string]*CvePolicy{
					"abc": {
						Action: securityBean.Blockiffixed,
					},
				},
				severityPolicy: map[securityBean.Severity]*CvePolicy{},
			},
			wantBlockedCVE: nil,
		},
		{
			name: "Test 6",
			args: args{
				cves: []*CveStore{
					{
						Severity:     securityBean.High,
						FixedVersion: "1.0.0",
					},
				},
				cvePolicy: map[string]*CvePolicy{},
				severityPolicy: map[securityBean.Severity]*CvePolicy{
					securityBean.High: {
						Action: securityBean.Blockiffixed,
					},
				},
			},
			wantBlockedCVE: []*CveStore{
				{
					Severity:     securityBean.High,
					FixedVersion: "1.0.0",
				},
			},
		},
		{
			name: "Test 7",
			args: args{
				cves: []*CveStore{
					{
						Severity: securityBean.High,
					},
				},
				cvePolicy: map[string]*CvePolicy{},
				severityPolicy: map[securityBean.Severity]*CvePolicy{
					securityBean.High: {
						Action: securityBean.Blockiffixed,
					},
				},
			},
			wantBlockedCVE: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotBlockedCVE := EnforceCvePolicy(tt.args.cves, tt.args.cvePolicy, tt.args.severityPolicy); !reflect.DeepEqual(gotBlockedCVE, tt.wantBlockedCVE) {
				t.Errorf("EnforceCvePolicy() = %v, want %v", gotBlockedCVE, tt.wantBlockedCVE)
			}
		})
	}
}
