import { createFileRoute } from '@tanstack/react-router'
import { useEffect, useState } from 'react'

// Known packet labels for display
const KNOWN_PACKET_LABELS: Record<string, string> = {
  0: 'Motion',
  1: 'Session',
  2: 'Lap Data',
  3: 'Event',
  4: 'Participants',
  5: 'Car Setups',
  6: 'Car Telemetry',
  7: 'Car Status',
  8: 'Final Classification',
  9: 'Lobby Info',
  10: 'Car Damage',
  11: 'Session History',
  12: 'Tyre Sets',
  13: 'MotionEx',
  14: 'Time Trial',
  15: 'Lap Positions',
}

export const Route = createFileRoute('/settings/telemetry')({
  component: PacketForwardingPage,
})

function PacketForwardingPage() {
  const [config, setConfig] = useState<Record<string, boolean> | null>(null)
  const [saving, setSaving] = useState(false)
  const [saveStatus, setSaveStatus] = useState<null | 'saving' | 'saved'>(null)

  useEffect(() => {
    fetch('/api/packet-forwarding')
      .then(r => r.json())
      .then(setConfig)
  }, [])

  function handleChange(packetId: string) {
    if (!config) return
    setConfig({ ...config, [packetId]: !config[packetId] })
  }

  function handleSave() {
    setSaving(true)
    setSaveStatus('saving')
    fetch('/api/packet-forwarding', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(config),
    }).then(() => {
      setSaving(false)
      setSaveStatus('saved')
      setTimeout(() => setSaveStatus(null), 1500)
    })
  }

  // Get all packet IDs from config, sorted numerically
  const packetIds = config ? Object.keys(config).sort((a, b) => Number(a) - Number(b)) : []

  return (
    <div className="container py-4">
      <h1 className="h3 mb-4">Packet Forwarding Settings</h1>
      {saveStatus === 'saving' && (
        <div className="alert alert-info py-2">Saving...</div>
      )}
      {saveStatus === 'saved' && (
        <div className="alert alert-success py-2">Saved!</div>
      )}
      {!config ? (
        <div className="alert alert-info">Loading...</div>
      ) : (
        <form onSubmit={e => { e.preventDefault(); handleSave() }}>
          <div className="mb-4">
            {packetIds.map(id => (
              <div className="form-check mb-2" key={id}>
                <input
                  className="form-check-input"
                  type="checkbox"
                  id={`packet-${id}`}
                  checked={!!config[id]}
                  onChange={() => handleChange(id)}
                />
                <label className="form-check-label" htmlFor={`packet-${id}`}>
                  {KNOWN_PACKET_LABELS[id] || `Packet ${id}`}
                </label>
              </div>
            ))}
          </div>
          <button
            type="submit"
            className="btn btn-primary"
            disabled={saving}
          >
            {saving ? 'Saving...' : 'Save'}
          </button>
        </form>
      )}
    </div>
  )
}
