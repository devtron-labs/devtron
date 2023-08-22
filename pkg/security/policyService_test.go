package security

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"reflect"
	"testing"
)

func TestPolicyServiceImpl_enforceCvePolicy(t *testing.T) {
	type args struct {
		cves           []*security.CveStore
		cvePolicy      map[string]*security.CvePolicy
		severityPolicy map[security.Severity]*security.CvePolicy
	}
	tests := []struct {
		name           string
		args           args
		wantBlockedCVE []*security.CveStore
	}{
		// TODO: Add test cases.
		{
			name: "Test 1",
			args: args{
				cves: []*security.CveStore{
					{
						Name: "abc",
					},
					{
						Severity: security.Low,
					},
				},
				cvePolicy: map[string]*security.CvePolicy{
					"abc": {
						Action: security.Allow,
					},
				},
				severityPolicy: map[security.Severity]*security.CvePolicy{
					security.Low: {
						Action: security.Allow,
					},
				},
			},
			wantBlockedCVE: nil,
		},
		{
			name: "Test 2",
			args: args{
				cves: []*security.CveStore{
					{
						Name: "abc",
					},
				},
				cvePolicy: map[string]*security.CvePolicy{
					"abc": {
						Action: security.Block,
					},
				},
				severityPolicy: map[security.Severity]*security.CvePolicy{},
			},
			wantBlockedCVE: []*security.CveStore{
				{
					Name: "abc",
				},
			},
		},
		{
			name: "Test 3",
			args: args{
				cves: []*security.CveStore{
					{
						Severity: security.High,
					},
				},
				cvePolicy: map[string]*security.CvePolicy{},
				severityPolicy: map[security.Severity]*security.CvePolicy{
					security.High: {
						Action: security.Block,
					},
				},
			},
			wantBlockedCVE: []*security.CveStore{
				{
					Severity: security.High,
				},
			},
		},
		{
			name: "Test 4",
			args: args{
				cves: []*security.CveStore{
					{
						Name:         "abc",
						FixedVersion: "1.0.0",
					},
				},
				cvePolicy: map[string]*security.CvePolicy{
					"abc": {
						Action: security.Blockiffixed,
					},
				},
				severityPolicy: map[security.Severity]*security.CvePolicy{},
			},
			wantBlockedCVE: []*security.CveStore{
				{
					Name:         "abc",
					FixedVersion: "1.0.0",
				},
			},
		},
		{
			name: "Test 5",
			args: args{
				cves: []*security.CveStore{
					{
						Name: "abc",
					},
				},
				cvePolicy: map[string]*security.CvePolicy{
					"abc": {
						Action: security.Blockiffixed,
					},
				},
				severityPolicy: map[security.Severity]*security.CvePolicy{},
			},
			wantBlockedCVE: nil,
		},
		{
			name: "Test 6",
			args: args{
				cves: []*security.CveStore{
					{
						Severity:     security.High,
						FixedVersion: "1.0.0",
					},
				},
				cvePolicy: map[string]*security.CvePolicy{},
				severityPolicy: map[security.Severity]*security.CvePolicy{
					security.High: {
						Action: security.Blockiffixed,
					},
				},
			},
			wantBlockedCVE: []*security.CveStore{
				{
					Severity:     security.High,
					FixedVersion: "1.0.0",
				},
			},
		},
		{
			name: "Test 7",
			args: args{
				cves: []*security.CveStore{
					{
						Severity: security.High,
					},
				},
				cvePolicy: map[string]*security.CvePolicy{},
				severityPolicy: map[security.Severity]*security.CvePolicy{
					security.High: {
						Action: security.Blockiffixed,
					},
				},
			},
			wantBlockedCVE: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &PolicyServiceImpl{}
			if gotBlockedCVE := impl.enforceCvePolicy(tt.args.cves, tt.args.cvePolicy, tt.args.severityPolicy); !reflect.DeepEqual(gotBlockedCVE, tt.wantBlockedCVE) {
				t.Errorf("enforceCvePolicy() = %v, want %v", gotBlockedCVE, tt.wantBlockedCVE)
			}
		})
	}
}

func TestPolicyServiceImpl_HasBlockedCVE(t *testing.T) {
	type args struct {
		cves           []*security.CveStore
		cvePolicy      map[string]*security.CvePolicy
		severityPolicy map[security.Severity]*security.CvePolicy
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
				cves: []*security.CveStore{
					{
						Name: "abc",
					},
					{
						Severity: security.Low,
					},
				},
				cvePolicy: map[string]*security.CvePolicy{
					"abc": {
						Action: security.Allow,
					},
				},
				severityPolicy: map[security.Severity]*security.CvePolicy{
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
				cves: []*security.CveStore{
					{
						Name: "abc",
					},
				},
				cvePolicy: map[string]*security.CvePolicy{
					"abc": {
						Action: security.Block,
					},
				},
				severityPolicy: map[security.Severity]*security.CvePolicy{},
			},
			want: true,
		},
		{
			name: "Test 3",
			args: args{
				cves: []*security.CveStore{
					{
						Severity: security.High,
					},
				},
				cvePolicy: map[string]*security.CvePolicy{},
				severityPolicy: map[security.Severity]*security.CvePolicy{
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
				cves: []*security.CveStore{
					{
						Name:         "abc",
						FixedVersion: "1.0.0",
					},
				},
				cvePolicy: map[string]*security.CvePolicy{
					"abc": {
						Action: security.Blockiffixed,
					},
				},
				severityPolicy: map[security.Severity]*security.CvePolicy{},
			},
			want: true,
		},
		{
			name: "Test 5",
			args: args{
				cves: []*security.CveStore{
					{
						Name: "abc",
					},
				},
				cvePolicy: map[string]*security.CvePolicy{
					"abc": {
						Action: security.Blockiffixed,
					},
				},
				severityPolicy: map[security.Severity]*security.CvePolicy{},
			},
			want: false,
		},
		{
			name: "Test 6",
			args: args{
				cves: []*security.CveStore{
					{
						Severity:     security.High,
						FixedVersion: "1.0.0",
					},
				},
				cvePolicy: map[string]*security.CvePolicy{},
				severityPolicy: map[security.Severity]*security.CvePolicy{
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
				cves: []*security.CveStore{
					{
						Severity: security.High,
					},
				},
				cvePolicy: map[string]*security.CvePolicy{},
				severityPolicy: map[security.Severity]*security.CvePolicy{
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
