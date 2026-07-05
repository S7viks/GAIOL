package keys

import (
	"context"
	"fmt"
	"strings"

	"gaiol/internal/database"
)

// EnsureDefaultModelsForProvider registers the first suggested model for a provider when the
// tenant has no active models for that provider yet. Routing uses all connected keys/models;
// no tenant-wide default model id is set.
func EnsureDefaultModelsForProvider(ctx context.Context, db *database.Client, tenantID string, provider string) error {
	if db == nil || db.Client == nil || strings.TrimSpace(tenantID) == "" {
		return nil
	}
	p := normalizeProvider(provider)
	if p == "" {
		return nil
	}

	existing, err := LoadTenantModelsForTenant(ctx, db, tenantID)
	if err != nil {
		return err
	}
	for _, m := range existing {
		if strings.EqualFold(strings.TrimSpace(m.ProviderKey), p) && m.IsActive {
			return nil
		}
	}

	suggestions := DefaultModelsForProvider(p)
	if len(suggestions) == 0 {
		return nil
	}
	first := suggestions[0]
	qs := 0.75
	if err := UpsertTenantModel(ctx, db, tenantID, p, first.ModelID, first.DisplayName, &qs, nil, nil, nil, []string{"auto-default"}); err != nil {
		return fmt.Errorf("auto-register default model: %w", err)
	}
	return nil
}
