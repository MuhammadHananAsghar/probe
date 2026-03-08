import { useMemo } from 'react'
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts'
import type { Request } from '../hooks/useWebSocket'
import { formatCost } from '../utils'

export function CostDashboard({ requests }: { requests: Request[] }) {
  const sorted = useMemo(() =>
    [...requests].sort((a, b) => new Date(a.started_at).getTime() - new Date(b.started_at).getTime()),
    [requests]
  )

  const timeData = useMemo(() => {
    let cumulative = 0
    return sorted.map(r => {
      cumulative += r.total_cost || 0
      return { name: `#${r.seq}`, cost: r.total_cost || 0, cumulative }
    })
  }, [sorted])

  const byModel = useMemo(() => {
    const m: Record<string, number> = {}
    for (const r of requests) {
      const key = r.model || 'unknown'
      m[key] = (m[key] || 0) + (r.total_cost || 0)
    }
    return Object.entries(m).sort((a, b) => b[1] - a[1]).slice(0, 5).map(([name, cost]) => ({ name, cost }))
  }, [requests])

  const byProvider = useMemo(() => {
    const m: Record<string, number> = {}
    for (const r of requests) {
      const key = r.provider || 'unknown'
      m[key] = (m[key] || 0) + (r.total_cost || 0)
    }
    return Object.entries(m).map(([name, cost]) => ({ name, cost }))
  }, [requests])

  const totalCost = requests.reduce((s, r) => s + (r.total_cost || 0), 0)
  const mostExpensive = requests.length ? requests.reduce((max, r) => (r.total_cost || 0) > (max.total_cost || 0) ? r : max, requests[0]) : null
  const sessionDurationMs = sorted.length > 1
    ? new Date(sorted[sorted.length - 1].ended_at || sorted[sorted.length - 1].started_at).getTime() - new Date(sorted[0].started_at).getTime()
    : 0
  const hourlyRate = sessionDurationMs > 60_000 ? (totalCost / sessionDurationMs) * 3_600_000 : 0

  const tooltipStyle = { backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: 4, color: '#d1d5db', fontSize: 11 }
  const axisStyle = { fontSize: 10, fill: '#6b7280' }

  if (requests.length === 0) {
    return <div className="p-8 text-center text-gray-500 text-sm">No data yet.</div>
  }

  return (
    <div className="p-4 max-w-5xl mx-auto">
      <h1 className="text-lg font-bold mb-4">Cost Analytics</h1>

      {/* Summary cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6">
        <Card label="Session Total" value={formatCost(totalCost, true)} color="text-yellow-400" />
        <Card label="Most Expensive" value={mostExpensive ? formatCost(mostExpensive.total_cost, mostExpensive.pricing_known) : '—'} color="text-orange-400" />
        <Card label="Top Model" value={byModel[0]?.name || '—'} color="text-purple-400" />
        <Card label="~/hour rate" value={hourlyRate > 0 ? formatCost(hourlyRate, true) : '—'} color="text-green-400" />
      </div>

      {/* Cost over time */}
      <div className="mb-6">
        <h2 className="text-xs text-gray-500 uppercase tracking-wider mb-2">Cost Over Time</h2>
        <div className="bg-gray-900 rounded p-3">
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={timeData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#1f2937" />
              <XAxis dataKey="name" tick={axisStyle} />
              <YAxis tick={axisStyle} tickFormatter={v => `$${Number(v).toFixed(4)}`} width={80} />
              <Tooltip contentStyle={tooltipStyle} formatter={(v: unknown) => [formatCost(Number(v), true)]} />
              <Line type="monotone" dataKey="cost" stroke="#a78bfa" dot={false} name="Per request" strokeWidth={1.5} />
              <Line type="monotone" dataKey="cumulative" stroke="#fbbf24" dot={false} name="Cumulative" strokeWidth={1.5} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <h2 className="text-xs text-gray-500 uppercase tracking-wider mb-2">Top Models by Cost</h2>
          <div className="bg-gray-900 rounded p-3">
            <ResponsiveContainer width="100%" height={Math.max(120, byModel.length * 32)}>
              <BarChart data={byModel} layout="vertical">
                <XAxis type="number" tick={axisStyle} tickFormatter={v => `$${Number(v).toFixed(4)}`} />
                <YAxis type="category" dataKey="name" tick={axisStyle} width={130} />
                <Tooltip contentStyle={tooltipStyle} formatter={(v: unknown) => [formatCost(Number(v), true)]} />
                <Bar dataKey="cost" fill="#a78bfa" radius={[0, 3, 3, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div>
          <h2 className="text-xs text-gray-500 uppercase tracking-wider mb-2">Cost by Provider</h2>
          <div className="bg-gray-900 rounded p-3">
            <ResponsiveContainer width="100%" height={Math.max(120, byProvider.length * 32)}>
              <BarChart data={byProvider} layout="vertical">
                <XAxis type="number" tick={axisStyle} tickFormatter={v => `$${Number(v).toFixed(4)}`} />
                <YAxis type="category" dataKey="name" tick={axisStyle} width={100} />
                <Tooltip contentStyle={tooltipStyle} formatter={(v: unknown) => [formatCost(Number(v), true)]} />
                <Bar dataKey="cost" fill="#34d399" radius={[0, 3, 3, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>
    </div>
  )
}

function Card({ label, value, color }: { label: string; value: string; color: string }) {
  return (
    <div className="bg-gray-900 rounded p-3">
      <div className="text-xs text-gray-500 mb-1">{label}</div>
      <div className={`text-sm font-bold ${color} truncate`}>{value}</div>
    </div>
  )
}
