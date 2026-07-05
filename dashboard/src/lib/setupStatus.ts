/** GET /api/setup/status — single-path readiness for dashboard gates and banners. */
export type SetupStatus = {
  inference_mode?: string
  orchestrator_configured?: boolean
  orchestrator_reachable?: boolean
  tenant_ready?: boolean
  providers_connected?: number
  gaiol_keys_count?: number
  setup_complete?: boolean
}

export async function fetchSetupStatus(get: (path: string) => Promise<unknown>): Promise<SetupStatus | null> {
  try {
    return (await get('/api/setup/status')) as SetupStatus
  } catch {
    return null
  }
}
