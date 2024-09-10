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
		severityPolicy map[Severity]*CvePolicy
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
						Severity: Low,
					},
				},
				cvePolicy: map[string]*CvePolicy{
					"abc": {
						Action: Allow,
					},
				},
				severityPolicy: map[Severity]*CvePolicy{
					Low: {
						Action: Allow,
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
						Action: Block,
					},
				},
				severityPolicy: map[Severity]*CvePolicy{},
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
						Severity: High,
					},
				},
				cvePolicy: map[string]*CvePolicy{},
				severityPolicy: map[Severity]*CvePolicy{
					High: {
						Action: Block,
					},
				},
			},
			wantBlockedCVE: []*CveStore{
				{
					Severity: High,
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
						Action: Blockiffixed,
					},
				},
				severityPolicy: map[Severity]*CvePolicy{},
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
						Action: Blockiffixed,
					},
				},
				severityPolicy: map[Severity]*CvePolicy{},
			},
			wantBlockedCVE: nil,
		},
		{
			name: "Test 6",
			args: args{
				cves: []*CveStore{
					{
						Severity:     High,
						FixedVersion: "1.0.0",
					},
				},
				cvePolicy: map[string]*CvePolicy{},
				severityPolicy: map[Severity]*CvePolicy{
					High: {
						Action: Blockiffixed,
					},
				},
			},
			wantBlockedCVE: []*CveStore{
				{
					Severity:     High,
					FixedVersion: "1.0.0",
				},
			},
		},
		{
			name: "Test 7",
			args: args{
				cves: []*CveStore{
					{
						Severity: High,
					},
				},
				cvePolicy: map[string]*CvePolicy{},
				severityPolicy: map[Severity]*CvePolicy{
					High: {
						Action: Blockiffixed,
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
