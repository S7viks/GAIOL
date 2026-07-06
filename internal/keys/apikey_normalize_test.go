package keys

import "testing"

func TestNormalizeAPIKey(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"  sk-or-v1-abc  ", "sk-or-v1-abc"},
		{"Bearer sk-or-v1-abc", "sk-or-v1-abc"},
		{"bearer sk-or-v1-abc", "sk-or-v1-abc"},
		{"Bearer Bearer sk-or-v1-abc", "sk-or-v1-abc"},
		{"Bearer", ""},
		{"Bearer ", ""},
	}
	for _, tc := range tests {
		if got := NormalizeAPIKey(tc.in); got != tc.want {
			t.Errorf("NormalizeAPIKey(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
