package bean

import "testing"

func TestValidateAttributeKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"plain key", HostUrlKey, false},
		{"key with dash and dot", "panel-template.v1", false},
		{"empty key", "", false},
		{"trailing nul", HostUrlKey + "\x00", true},
		{"nul in middle", "u\x00rl", true},
		{"newline", "url\n", true},
		{"tab", "url\t", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateAttributeKey(tt.key); (err != nil) != tt.wantErr {
				t.Errorf("ValidateAttributeKey(%q) err = %v, wantErr %v", tt.key, err, tt.wantErr)
			}
		})
	}
}

func TestIsInternalOnlyKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{API_SECRET_KEY, true},
		{ENCRYPTION_KEY, true},
		{" " + API_SECRET_KEY + " ", true},
		{HostUrlKey, false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsInternalOnlyKey(tt.key); got != tt.want {
			t.Errorf("IsInternalOnlyKey(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}
