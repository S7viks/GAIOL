package database

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// AuditEntry represents one row in audit_log.
type AuditEntry struct {
	ID        string                 `json:"id"`
	TenantID  string                 `json:"tenant_id"`
	UserID    string                 `json:"user_id"`
	Action    string                 `json:"action"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// InsertAuditLog inserts an audit log entry. UserID may be empty for key-based actions.
func (c *Client) InsertAuditLog(ctx context.Context, tenantID, userID, action string, metadata map[string]interface{}) error {
	if c == nil || c.Client == nil {
		return nil
	}
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	row := map[string]interface{}{
		"tenant_id": tenantID,
		"user_id":   userID,
		"action":    action,
		"metadata":  metadata,
	}
	_, _, err := c.From("audit_log").Insert(row, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("insert audit_log: %w", err)
	}
	return nil
}

// GetAuditLogForTenant returns recent audit entries for the tenant, newest first.
func (c *Client) GetAuditLogForTenant(ctx context.Context, tenantID string, limit int) ([]AuditEntry, error) {
	if c == nil || c.Client == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 50
	}
	var rows []struct {
		ID        string                 `json:"id"`
		TenantID  string                 `json:"tenant_id"`
		UserID    *string                `json:"user_id"`
		Action    string                 `json:"action"`
		Metadata  map[string]interface{} `json:"metadata"`
		CreatedAt time.Time              `json:"created_at"`
	}
	_, err := c.From("audit_log").
		Select("id,tenant_id,user_id,action,metadata,created_at", "", false).
		Filter("tenant_id", "eq", tenantID).
		ExecuteTo(&rows)
	if err != nil {
		return nil, fmt.Errorf("get audit_log: %w", err)
	}
	// Sort newest first and take up to limit
	sort.Slice(rows, func(i, j int) bool { return rows[i].CreatedAt.After(rows[j].CreatedAt) })
	if len(rows) > limit {
		rows = rows[:limit]
	}
	out := make([]AuditEntry, len(rows))
	for i := range rows {
		uid := ""
		if rows[i].UserID != nil {
			uid = *rows[i].UserID
		}
		out[i] = AuditEntry{
			ID:        rows[i].ID,
			TenantID:  rows[i].TenantID,
			UserID:    uid,
			Action:    rows[i].Action,
			Metadata:  rows[i].Metadata,
			CreatedAt: rows[i].CreatedAt,
		}
	}
	return out, nil
}
