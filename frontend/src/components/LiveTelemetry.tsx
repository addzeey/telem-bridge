import { useTelemetry } from "../hooks/useTelemetry.ts"

export default function LiveTelemetry() {
  const data = useTelemetry()

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Live Telemetry</h1>
      {data ? (
        <pre className="bg-gray-800 text-green-400 p-4 rounded">
          {JSON.stringify(data, null, 2)}
        </pre>
      ) : (
        <p>Waiting for data...</p>
      )}
    </div>
  )
}
