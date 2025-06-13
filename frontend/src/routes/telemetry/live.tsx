import { createFileRoute } from '@tanstack/react-router'
import { useTelemetry } from '../../hooks/useTelemetry'
import { ServiceRestartButton } from '../../components/ServiceRestartButton'
export const Route = createFileRoute('/telemetry/live')({
  component: LiveComponent,
})

export function LiveComponent () {
  console.log('Live telemetry component mounted')
  const data = useTelemetry()

  return (
    <div className="container py-4">
      <h1 className="h3 mb-4">Live Telemetry</h1>
      <div className="mb-3">
        <ServiceRestartButton service="udp" />
      </div>
      {data ? (
        <pre className="bg-dark text-success p-4 rounded">
          {JSON.stringify(data, null, 2)}
        </pre>
      ) : (
        <div className="alert alert-info">Waiting for telemetry data...</div>
      )}
    </div>
  )
}