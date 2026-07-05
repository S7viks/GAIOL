package httpserver

import (
	"testing"

	"gaiol/internal/database"
)

func TestCanManageKeys(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		tc   database.TenantContext
		want bool
	}{
		{"admin", database.TenantContext{UserID: "u1", TenantID: "t1", Role: "admin"}, true},
		{"owner", database.TenantContext{UserID: "u1", TenantID: "t1", Role: "owner"}, true},
		{"empty role legacy", database.TenantContext{UserID: "u1", TenantID: "t1", Role: ""}, true},
		{"personal tenant user", database.TenantContext{UserID: "u1", TenantID: "u1", Role: "user"}, true},
		{"org member user", database.TenantContext{UserID: "u2", TenantID: "t1", Role: "user"}, false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := canManageKeys(tc.tc); got != tc.want {
				t.Fatalf("canManageKeys(%+v) = %v, want %v", tc.tc, got, tc.want)
			}
		})
	}
}
