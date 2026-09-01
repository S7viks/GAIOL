package auth

import (
	"context"
	"testing"
)

// Skeleton tests for internal/auth. Replace calls with real package API.
func TestVerifyToken_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		wantError bool
	}{
		{"valid token", "test-valid-jwt", false},
		{"expired token", "test-expired-jwt", true},
		{"bad signature", "test-bad-jwt", true},
	}

	// TODO: initialize your verifier with test keys or mock dependency.
	// v := NewVerifier(WithKey("test-key"))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = context.Background()
			// err := v.VerifyToken(ctx, tt.token)
			// if (err != nil) != tt.wantError {
			//     t.Fatalf("VerifyToken() error = %v, wantErr %v", err, tt.wantError)
			// }
		})
	}
}
