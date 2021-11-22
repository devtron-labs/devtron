package util

import "testing"

func TestCompareLimitsRequests(t *testing.T) {
	requests := "requests"
	limits := "limits"
	resources := "resources"
	cpu := "cpu"
	memory := "memory"
	type args struct {
		dat map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "empty base object",
			args: args{dat: nil},
			want: true,
			wantErr: false,
		},
		{
			name: "empty resources object",
			args: args{dat: map[string]interface{}{resources: nil}},
			want: true,
			wantErr: false,
		},
		{
			name: "empty resources requests limits object",
			args: args{dat: map[string]interface{}{resources: map[string]interface{}{limits: nil, requests: nil}}},
			want: true,
			wantErr: false,
		},
		{
			name: "non-empty resources limits object",
			args: args{dat: map[string]interface{}{resources: map[string]interface{}{limits: map[string]interface{}{cpu: "10Gi", memory: ".1" }, requests: nil}}},
			want: true,
			wantErr: false,
		},
		{
			name: "non-empty resources requests object",
			args: args{dat: map[string]interface{}{resources: map[string]interface{}{limits: nil, requests: map[string]interface{}{cpu: "10Gi", memory: ".1" }}}},
			want: true,
			wantErr: false,
		},
		{
			name: "non-empty  and equal resources limits and requests object",
			args: args{dat: map[string]interface{}{resources: map[string]interface{}{limits: map[string]interface{}{cpu: "10Gi", memory: ".1" }, requests: map[string]interface{}{cpu: "10Gi", memory: ".1" }}}},
			want: true,
			wantErr: false,
		},
		{
			name: "negative: non-empty  and not equal resources limits and requests object",
			args: args{dat: map[string]interface{}{resources: map[string]interface{}{limits: map[string]interface{}{cpu: "10Gi", memory: ".1" }, requests: map[string]interface{}{cpu: "11Gi", memory: ".15" }}}},
			want: false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareLimitsRequests(tt.args.dat)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompareLimitsRequests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CompareLimitsRequests() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAutoScale(t *testing.T) {
	autoScaling := "autoscaling"
	minReplicas := "MinReplicas"
	maxReplicas := "MaxReplicas"
	enabled := "enabled"
	type args struct {
		dat map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "empty base object",
			args: args{dat: nil},
			want: true,
			wantErr: false,
		},
		{
			name: "empty autoscaling object",
			args: args{dat: map[string]interface{}{autoScaling: nil}},
			want: true,
			wantErr: false,
		},
		{
			name: "negative : non-empty autoscaling empty enabled minReplicas maxReplicas object",
			args: args{dat: map[string]interface{}{autoScaling: map[string]interface{}{}}},
			want: true,
			wantErr: false,
		},
		{
			name: "non-empty autoscaling enabled minReplicas maxReplicas object",
			args: args{dat: map[string]interface{}{autoScaling: map[string]interface{}{enabled: false,minReplicas: float64(10), maxReplicas: float64(11)}}},
			want: true,
			wantErr: false,
		},
		{
			name: "non-empty autoscaling enabled minReplicas empty maxReplicas object",
			args: args{dat: map[string]interface{}{autoScaling: map[string]interface{}{enabled: false,minReplicas: float64(11)}}},
			want: true,
			wantErr: false,
		},
		{
			name: "non-empty autoscaling minReplicas maxReplicas object empty enabled",
			args: args{dat: map[string]interface{}{autoScaling: map[string]interface{}{minReplicas: float64(10), maxReplicas: float64(11)}}},
			want: true,
			wantErr: false,
		},
		{
			name: "negative: non-empty and greater minReplicas than maxReplicas object",
			args: args{dat: map[string]interface{}{autoScaling: map[string]interface{}{enabled: true, minReplicas: float64(10), maxReplicas: float64(9)}}},
			want: false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AutoScale(tt.args.dat)
			if (err != nil) != tt.wantErr {
				t.Errorf("AutoScale() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AutoScale() got = %v, want %v", got, tt.want)
			}
		})
	}
}