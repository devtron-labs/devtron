package chart

import "testing"

func Test_checkCompatibility(t *testing.T) {
	type args struct {
		oldChartId int
		newChartId int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "when charts are compatible, it should return true",
			args: args{
				oldChartId: 32,
				newChartId: 33,
			},
			want: true,
		},
		{
			name: "when oldChart is not found, it should return false",
			args: args{
				oldChartId: -32,
				newChartId: 33,
			},
			want: false,
		},
		{
			name: "when newChart is not found, it should return false",
			args: args{
				oldChartId: 32,
				newChartId: -33,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckCompatibility(tt.args.oldChartId, tt.args.newChartId); got != tt.want {
				t.Errorf("checkCompatibility() = %v, want %v", got, tt.want)
			}
		})
	}
}
