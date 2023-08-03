package chart

import (
	"reflect"
	"testing"
)

func Test_checkCompatibility(t *testing.T) {
	type args struct {
		oldChartType string
		newChartType string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "when charts are compatible, it should return true",
			args: args{
				oldChartType: DeploymentChartType,
				newChartType: RolloutChartType,
			},
			want: true,
		},
		{
			name: "when oldChart is not found, it should return false",
			args: args{
				oldChartType: "Sdfasdf",
				newChartType: RolloutChartType,
			},
			want: false,
		},
		{
			name: "when newChart is not found, it should return false",
			args: args{
				oldChartType: DeploymentChartType,
				newChartType: "random sdfasdf saldfj",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckCompatibility(tt.args.oldChartType, tt.args.newChartType); got != tt.want {
				t.Errorf("checkCompatibility() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompatibleChartsWith(t *testing.T) {
	type args struct {
		chartType string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "when chart is not found, it should return empty slice",
			args: args{
				chartType: "sdfsdf",
			},
			want: []string{},
		},
		{
			name: "when chart is found, it should return a slice of all cpmpatible chartIds",
			args: args{
				chartType: DeploymentChartType,
			},
			want: []string{RolloutChartType},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompatibleChartsWith(tt.args.chartType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CompatibleChartsWith() = %v, want %v", got, tt.want)
			}
		})
	}
}
