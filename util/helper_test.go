package util

import "testing"

func TestAutoscale(t *testing.T) {
	type args struct {
		dat map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Autoscale(tt.args.dat)
			if (err != nil) != tt.wantErr {
				t.Errorf("Autoscale() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Autoscale() got = %v, want %v", got, tt.want)
			}
		})
	}
}
