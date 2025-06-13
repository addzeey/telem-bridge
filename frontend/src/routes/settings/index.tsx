import { createFileRoute, Link } from '@tanstack/react-router'
import { useEffect, useState } from 'react'

export const Route = createFileRoute('/settings/')({
  component: RouteComponent,
})

function RouteComponent() {
  // AppConfig shape from backend
  type AppConfig = {
    udp_addr: string
    udp_port: number
    osc_addr: string
    osc_port: number
    enable_osc: boolean
    broadcast_rate_hz?: number
    debug_output?: boolean
  }
  const [config, setConfig] = useState<AppConfig | null>(null)
  const [saving, setSaving] = useState(false)
  const [saveStatus, setSaveStatus] = useState<null | 'saving' | 'saved'>(null)

  useEffect(() => {
    fetch('/api/config')
      .then(r => r.json())
      .then(setConfig)
  }, [])

  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    if (!config) return
    const { name, value, type, checked } = e.target
    setConfig({
      ...config,
      [name]: type === 'checkbox' ? checked : (type === 'number' ? Number(value) : value),
    })
  }

  function handleSave() {
    setSaving(true)
    setSaveStatus('saving')
    fetch('/api/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(config),
    }).then(() => {
      setSaving(false)
      setSaveStatus('saved')
      setTimeout(() => setSaveStatus(null), 1500)
    })
  }

  return (
    <div className="container py-4 bg-secondary-subtle">
      <h1 className="h3 mb-4">Settings</h1>
      {saveStatus === 'saving' && (
        <div className="alert alert-info py-2">Saving...</div>
      )}
      {saveStatus === 'saved' && (
        <div className="alert alert-success py-2">Saved! Services will reconnect.</div>
      )}
      <form onSubmit={e => { e.preventDefault(); handleSave() }} className="mb-4">
        {!config ? (
          <div className="alert alert-info">Loading...</div>
        ) : (
          <>
            <div className="mb-3">
              <label className="form-label">Enable OSC Forwarding</label>
              <div className="form-check">
                <input
                  className="form-check-input"
                  type="checkbox"
                  name="enable_osc"
                  checked={config.enable_osc}
                  onChange={handleChange}
                  id="enable_osc"
                />
                <label className="form-check-label" htmlFor="enable_osc">
                  Enable OSC
                </label>
              </div>
            </div>
            <div className="mb-3">
              <label className="form-label">UDP Address</label>
              <input
                className="form-control"
                type="text"
                name="udp_addr"
                value={config.udp_addr}
                onChange={handleChange}
              />
            </div>
            <div className="mb-3">
              <label className="form-label">UDP Port</label>
              <input
                className="form-control"
                type="number"
                name="udp_port"
                value={config.udp_port}
                onChange={handleChange}
              />
            </div>
            <div className="mb-3">
              <label className="form-label">OSC Address</label>
              <input
                className="form-control"
                type="text"
                name="osc_addr"
                value={config.osc_addr}
                onChange={handleChange}
              />
            </div>
            <div className="mb-3">
              <label className="form-label">OSC Port</label>
              <input
                className="form-control"
                type="number"
                name="osc_port"
                value={config.osc_port}
                onChange={handleChange}
              />
            </div>
            <div className="mb-3">
              <label className="form-label">Broadcast Rate (Hz)</label>
              <input
                className="form-control"
                type="number"
                name="broadcast_rate_hz"
                min={1}
                max={60}
                title="How many times per second to send data to WebSocket/OSC"
                value={config.broadcast_rate_hz ?? 2}
                onChange={handleChange}
              />
              <div className="form-text">How many times per second to send data to WebSocket/OSC (default: 2)</div>
            </div>
            <div className="mb-3">
              <label className="form-label">Debug Output</label>
              <div className="form-check">
                <input
                  className="form-check-input"
                  type="checkbox"
                  name="debug_output"
                  checked={!!config.debug_output}
                  onChange={handleChange}
                  id="debug_output"
                />
                <label className="form-check-label" htmlFor="debug_output">
                  Enable debug info logs (show extra info in backend log)
                </label>
              </div>
            </div>
            <button type="submit" className="btn btn-primary" disabled={saving}>
              {saving ? 'Saving...' : 'Save'}
            </button>
          </>
        )}
      </form>
      <div className="list-group">
        <Link
          to="/settings/telemetry"
          className="list-group-item list-group-item-action bg-secondary text-white"
        >
          Telemetry Settings
        </Link>
      </div>
    </div>
  )
}
