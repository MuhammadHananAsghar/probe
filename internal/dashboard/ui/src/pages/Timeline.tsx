import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import type { Request } from '../hooks/useWebSocket'
import { formatDuration } from '../utils'

export function Timeline({ requests }: { requests: Request[] }) {
  const [zoom, setZoom] = useState(1)

  const sorted = useMemo(() =>
    [...requests].sort((a, b) => new Date(a.started_at).getTime() - new Date(b.started_at).getTime()),
    [requests]
  )

  if (sorted.length === 0) {
    return <div className="p-8 text-center text-gray-500 text-sm">No requests yet.</div>
  }

  const sessionStart = new Date(sorted[0].started_at).getTime()
  const sessionEnd = sorted.reduce((max, r) => {
    const latencyMs = (r.latency || 0) / 1_000_000
    const t = new Date(r.started_at).getTime() + latencyMs
    return Math.max(max, t)
  }, sessionStart + 1)
  const totalDuration = sessionEnd - sessionStart
  const barWidth = Math.min(800 * zoom, 2400)

  return (
    <div className="p-4">
      <div className="flex items-center gap-3 mb-4">
        <h1 className="text-lg font-bold">Timeline</h1>
        <span className="text-xs text-gray-500">total: {formatDuration(totalDuration * 1_000_000)}</span>
        <div className="flex gap-1 ml-auto items-center">
          <button onClick={() => setZoom(z => Math.max(0.25, z - 0.25))}
            className="px-2 py-1 bg-gray-800 rounded text-xs hover:bg-gray-700 border-0 cursor-pointer text-gray-300">−</button>
          <span className="px-2 py-1 text-xs text-gray-400">{Math.round(zoom * 100)}%</span>
          <button onClick={() => setZoom(z => Math.min(4, z + 0.25))}
            className="px-2 py-1 bg-gray-800 rounded text-xs hover:bg-gray-700 border-0 cursor-pointer text-gray-300">+</button>
        </div>
      </div>

      <div className="overflow-x-auto">
        <div style={{ minWidth: barWidth + 320, paddingBottom: 16 }}>
          {sorted.map(req => {
            const startMs = new Date(req.started_at).getTime() - sessionStart
            const latencyMs = (req.latency || 0) / 1_000_000
            const ttftMs = (req.ttft || 0) / 1_000_000

            const leftPct = totalDuration > 0 ? (startMs / totalDuration) * 100 : 0
            const widthPct = totalDuration > 0 ? Math.max((latencyMs / totalDuration) * 100, 0.3) : 0.3
            const ttftPct = latencyMs > 0 && ttftMs > 0 ? Math.min((ttftMs / latencyMs) * 100, 99) : 0

            const isError = req.status_code >= 400 || req.status === 'error'
            const barColor = isError ? 'bg-red-800/80' : req.stream ? 'bg-blue-800/80' : 'bg-emerald-800/80'

            return (
              <div key={req.id} className="flex items-center gap-2 mb-1 group">
                <div className="w-52 flex-shrink-0 text-right pr-2">
                  <Link to={`/request/${req.id}`} className="text-xs text-gray-500 hover:text-purple-400 transition-colors truncate block">
                    #{req.seq} {req.model || req.provider}
                  </Link>
                </div>
                <div className="relative h-5" style={{ width: barWidth }}>
                  <div
                    className={`absolute inset-y-0 rounded ${barColor} group-hover:brightness-125 transition-all`}
                    style={{ left: `${leftPct}%`, width: `${widthPct}%` }}
                  >
                    {ttftPct > 0 && (
                      <div className="absolute top-0 bottom-0 w-px bg-yellow-400/80"
                        style={{ left: `${ttftPct}%` }} />
                    )}
                  </div>
                </div>
                <div className="w-20 flex-shrink-0 text-xs text-gray-500">
                  {formatDuration(req.latency)}
                </div>
              </div>
            )
          })}
        </div>

        {/* Legend */}
        <div className="mt-2 flex gap-4 text-xs text-gray-500 flex-wrap">
          <span className="flex items-center gap-1"><span className="inline-block w-3 h-3 bg-blue-800/80 rounded" />Streaming</span>
          <span className="flex items-center gap-1"><span className="inline-block w-3 h-3 bg-emerald-800/80 rounded" />Non-streaming</span>
          <span className="flex items-center gap-1"><span className="inline-block w-3 h-3 bg-red-800/80 rounded" />Error</span>
          <span className="flex items-center gap-1"><span className="inline-block w-px h-3 bg-yellow-400" />TTFT</span>
        </div>
      </div>
    </div>
  )
}
