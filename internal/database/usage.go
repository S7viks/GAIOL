package database

import (
	"context"
	"fmt"
	"time"
)

// UsageRow represents one row from api_queries for usage aggregation.
type UsageRow struct {
	ID               string    `json:"id"`
	ModelID          string    `json:"model_id"`
	TokensUsed       int       `json:"tokens_used"`
	Cost             float64   `json:"cost"`
	ProcessingTimeMs int       `json:"processing_time_ms"`
	Success          bool      `json:"success"`
	CreatedAt        time.Time `json:"created_at"`
	GAIOLKeyID       *string   `json:"gaiol_api_key_id"`
}

// GetUsageForTenant queries api_queries for the tenant and date range, returns rows for client-side aggregation.
// from and to are optional (empty = no filter on that side).
func (c *Client) GetUsageForTenant(ctx context.Context, tenantID string, from, to *time.Time) ([]UsageRow, error) {
	if c == nil || c.Client == nil {
		return nil, nil
	}
	q := c.From("api_queries").
		Select("id,model_id,tokens_used,cost,processing_time_ms,success,created_at,gaiol_api_key_id", "", false).
		Filter("tenant_id", "eq", tenantID)
	if from != nil {
		q = q.Filter("created_at", "gte", from.Format(time.RFC3339))
	}
	if to != nil {
		q = q.Filter("created_at", "lte", to.Format(time.RFC3339))
	}
	var rows []UsageRow
	_, err := q.ExecuteTo(&rows)
	if err != nil {
		return nil, fmt.Errorf("get usage: %w", err)
	}
	return rows, nil
}
