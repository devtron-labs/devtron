package util

import (
	"fmt"
	"testing"
)

func TestMatchRegex(t *testing.T) {
	type args struct {
		exp  string
		text string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Invalid regex",
			args: args{
				exp:  "**",
				text: "v1.23+",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Valid regex,text not matching with regex",
			args: args{
				exp:  EphemeralServerVersionRegex,
				text: "v1.03+",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Valid regex,text not matching with regex",
			args: args{
				exp:  EphemeralServerVersionRegex,
				text: "v1.22+",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Valid regex, text match with regex",
			args: args{
				exp:  EphemeralServerVersionRegex,
				text: "v1.3",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Valid regex, text match with regex",
			args: args{
				exp:  EphemeralServerVersionRegex,
				text: "v1.23+",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid regex, text match with regex",
			args: args{
				exp:  EphemeralServerVersionRegex,
				text: "v1.26.6",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid regex, text match with regex",
			args: args{
				exp:  EphemeralServerVersionRegex,
				text: "v1.26",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid regex, text match with regex",
			args: args{
				exp:  EphemeralServerVersionRegex,
				text: "v1.30",
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchRegex(tt.args.exp, tt.args.text)
			fmt.Println(err)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MatchRegex() got = %v, want %v", got, tt.want)
			}
		})
	}
}
