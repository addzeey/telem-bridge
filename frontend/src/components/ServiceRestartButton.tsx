import { useState } from 'react'

export type ServiceType = 'osc' | 'udp' | 'all'

interface ServiceRestartButtonProps {
  service: ServiceType
  label?: string
  className?: string
}

export function ServiceRestartButton({ service, label, className }: ServiceRestartButtonProps) {
  const [status, setStatus] = useState<'idle' | 'loading' | 'success' | 'error'>('idle')
  const endpoint = `/api/restart/${service}`.replace(/\/api\/+/g, '/api/');
  const buttonLabel = label || `Restart ${service.toUpperCase()} Service`

  async function handleRestart() {
    setStatus('loading')
    try {
      const res = await fetch(endpoint, { method: 'POST' })
      if (res.ok) {
        setStatus('success')
        setTimeout(() => setStatus('idle'), 1200)
      } else {
        setStatus('error')
        setTimeout(() => setStatus('idle'), 2000)
      }
    } catch {
      setStatus('error')
      setTimeout(() => setStatus('idle'), 2000)
    }
  }

  return (
    <div className={className}>
      <button
        className={`btn btn-secondary me-2`}
        onClick={handleRestart}
        disabled={status === 'loading'}
        type="button"
      >
        {status === 'loading' ? 'Restarting...' : buttonLabel}
      </button>
      {status === 'success' && (
        <div className="alert alert-success py-1 px-2 d-inline-block ms-2 mb-0" role="alert">
          Service restarted!
        </div>
      )}
      {status === 'error' && (
        <div className="alert alert-danger py-1 px-2 d-inline-block ms-2 mb-0" role="alert">
          Restart failed
        </div>
      )}
    </div>
  )
}