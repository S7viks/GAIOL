export interface ModelRow {
  id: string
  provider?: string
  model_name?: string
  display_name?: string
  cost_per_token?: number
  capabilities?: Record<string, boolean>
  quality_score?: number
  context_window?: number
  max_tokens?: number
  tags?: string[]
}

export interface ModelsListResponse {
  models?: ModelRow[]
  count?: number
}

export interface SmartQueryResponse {
  response?: string
  result?: { data?: string }
  metadata?: {
    session_id?: string
    trace_id?: string
    engine?: string
    steps_executed?: number
    cost_info?: { total_cost?: number }
  }
  orchestration?: {
    trace_id?: string
    schema_version?: string
    trust_updates_count?: number
    consensus_mode?: string
    beam_width?: number
  }
  cost?: number
  latency_ms?: number
  strategy?: string
}

export interface TrustRecord {
  model_id: string
  domain: string
  distribution: { alpha: number; beta: number }
  updated_at: string
}

export interface TrustListResponse {
  records: TrustRecord[]
  count: number
  domain: string | null
}

export interface TraceIdsResponse {
  trace_ids: string[]
  count: number
}

export interface TraceBundle {
  trace?: unknown
  timeline_rebuilt?: unknown
  metrics_summary?: Record<string, unknown>
}

export interface ActivityEntry {
  id?: string
  action?: string
  created_at?: string
  metadata?: Record<string, unknown>
}

export interface ActivityResponse {
  activity?: ActivityEntry[]
}

export interface PreferencesResponse {
  budget_limit?: number | null
  strategy?: string
  beam_width?: number
  consensus_mode?: string
  domain?: string
  explore_paths?: boolean
}

/** Matches keys.ProviderKeyRow from GET /api/settings/provider-keys */
export interface ProviderKeyRow {
  id?: string
  provider?: string
  key_hint?: string
  is_active?: boolean
  created_at?: string
  updated_at?: string
}

/** GET /api/gaiol-keys — tenant Relay/GAIOL API keys (secret not returned). */
export interface GaiolKeyRow {
  id?: string
  name?: string
  created_at?: string
  last_used_at?: string | null
}

/** GET /api/settings/models — tenant-routable models (production DB). */
export interface TenantModelRow {
  id?: string
  provider_key?: string
  model_id?: string
  display_name?: string
  quality_score?: number
  cost_per_token?: number
  context_window?: number
  max_tokens?: number
  tags?: string[]
  is_active?: boolean
}

export interface TenantModelsResponse {
  models?: TenantModelRow[]
}

/** GET /api/usage */
export interface UsageSummary {
  requests?: number
  tokens?: number
  cost?: number
}

export interface UsageBreakdownRow {
  date?: string
  provider?: string
  key_id?: string
  key_name?: string
  requests?: number
  tokens?: number
  cost?: number
}

export interface UsageResponse {
  summary?: UsageSummary
  by_day?: UsageBreakdownRow[]
  by_provider?: UsageBreakdownRow[]
  by_key?: UsageBreakdownRow[]
}

/** GET /api/billing/summary */
export interface BillingSummaryResponse {
  period?: string
  total_cost?: number
  by_provider?: { provider?: string; cost?: number }[]
}

/** GET /api/billing/history */
export interface BillingHistoryEntry {
  month?: string
  total_cost?: number
}

export interface BillingHistoryResponse {
  history?: BillingHistoryEntry[]
}
