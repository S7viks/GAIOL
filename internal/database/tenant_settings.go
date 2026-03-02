package database

import (
	"context"
	"fmt"
	"time"
)

// TenantSettings holds per-tenant preferences and budget.
type TenantSettings struct {
	TenantID         string     `json:"tenant_id"`
	BudgetLimit      *float64   `json:"budget_limit"`
	BudgetAlertSentAt *time.Time `json:"budget_alert_sent_at"`
	DefaultModelID   string     `json:"default_model_id"`
	Strategy         string     `json:"strategy"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// GetTenantSettings returns settings for the tenant, or nil if none.
func (c *Client) GetTenantSettings(ctx context.Context, tenantID string) (*TenantSettings, error) {
	if c == nil || c.Client == nil {
		return nil, nil
	}
	var rows []struct {
		TenantID          string     `json:"tenant_id"`
		BudgetLimit       *float64   `json:"budget_limit"`
		BudgetAlertSentAt *time.Time `json:"budget_alert_sent_at"`
		DefaultModelID    *string    `json:"default_model_id"`
		Strategy         *string    `json:"strategy"`
		UpdatedAt        time.Time  `json:"updated_at"`
	}
	_, err := c.From("tenant_settings").
		Select("tenant_id,budget_limit,budget_alert_sent_at,default_model_id,strategy,updated_at", "", false).
		Filter("tenant_id", "eq", tenantID).
		ExecuteTo(&rows)
	if err != nil || len(rows) == 0 {
		return nil, nil
	}
	r := &TenantSettings{
		TenantID:   rows[0].TenantID,
		BudgetLimit: rows[0].BudgetLimit,
		BudgetAlertSentAt: rows[0].BudgetAlertSentAt,
		UpdatedAt: rows[0].UpdatedAt,
	}
	if rows[0].DefaultModelID != nil {
		r.DefaultModelID = *rows[0].DefaultModelID
	}
	if rows[0].Strategy != nil {
		r.Strategy = *rows[0].Strategy
	} else {
		r.Strategy = "balanced"
	}
	return r, nil
}

// UpsertTenantSettings inserts or updates tenant settings. Pass nil for fields to leave unchanged (for partial update we'd need PATCH; here we upsert full row).
func (c *Client) UpsertTenantSettings(ctx context.Context, s *TenantSettings) error {
	if c == nil || c.Client == nil {
		return fmt.Errorf("database client is required")
	}
	if s == nil || s.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	row := map[string]interface{}{
		"tenant_id":           s.TenantID,
		"budget_limit":        s.BudgetLimit,
		"budget_alert_sent_at": s.BudgetAlertSentAt,
		"default_model_id":    nullIfEmpty(s.DefaultModelID),
		"strategy":            nullIfEmpty(s.Strategy),
		"updated_at":         time.Now().UTC().Format(time.RFC3339),
	}
	_, _, err := c.From("tenant_settings").Insert(row, true, "tenant_id", "", "").Execute()
	if err != nil {
		return fmt.Errorf("upsert tenant_settings: %w", err)
	}
	return nil
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
