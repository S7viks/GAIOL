import { apiUrl, configuredApiOrigin, isCrossOriginApi } from '../../lib/apiBase'

export function ApiUnreachableHelp() {
  const healthUrl = apiUrl('/health')
  const crossOrigin = isCrossOriginApi()
  const pageOrigin = typeof window !== 'undefined' ? window.location.origin : ''

  if (crossOrigin) {
    return (
      <p className="error-message" role="alert">
        Cannot reach the Go API at <code>{healthUrl}</code> from <code>{pageOrigin}</code>. On your API host
        (Render/Fly/etc.), set <code>ALLOWED_ORIGINS</code> to include <code>{pageOrigin}</code> (comma-separated,
        no trailing slash), then redeploy the API. Free Render instances can take up to a minute to wake — refresh after
        waiting. <code>VITE_API_BASE</code> on the dashboard build is{' '}
        <code>{configuredApiOrigin() || '(not set)'}</code>.
      </p>
    )
  }

  return (
    <p className="error-message" role="alert">
      Cannot reach the Go API at <code>{healthUrl}</code>. Start it with <code>go run cmd/web-server/main.go</code> or{' '}
      <code>.\scripts\dev\start-relay.ps1 -Dashboard</code>, then refresh. If you use a hosted API, open its URL directly (same host serves API + dashboard) or fix{' '}
      <code>ALLOWED_ORIGINS</code> for split hosting.
    </p>
  )
}
