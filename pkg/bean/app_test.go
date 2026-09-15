package bean

import "testing"

func boolPtr(b bool) *bool { return &b }

func TestCDPatchRequest_GetForegroundDelete(t *testing.T) {
	tests := []struct {
		name       string
		field      *bool
		envDefault bool
		want       bool
	}{
		{"nil payload, env false", nil, false, false},
		{"nil payload, env true", nil, true, true},
		{"payload true overrides env false", boolPtr(true), false, true},
		{"payload false overrides env true", boolPtr(false), true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CDPatchRequest{ForegroundDelete: tt.field}
			if got := r.GetForegroundDelete(tt.envDefault); got != tt.want {
				t.Errorf("GetForegroundDelete(%v) with field=%v = %v, want %v",
					tt.envDefault, tt.field, got, tt.want)
			}
		})
	}
}
