import { useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import type { Request } from '../hooks/useWebSocket'
import { formatDuration, formatCost, formatTokens, statusColor, providerColor } from '../utils'

type SortKey = 'time' | 'cost' | 'latency' | 'tokens'

interface Props {
  requests: Request[]
  search: string
}

export function RequestList({ requests, search }: Props) {
  const [provider, setProvider] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [sort, setSort] = useState<SortKey>('time')

  const filtered = useMemo(() => {
    let list = [...requests]
    if (search) {
      const q = search.toLowerCase()
      list = list.filter(r =>
        r.model?.toLowerCase().includes(q) ||
        r.provider?.toLowerCase().includes(q) ||
        r.path?.toLowerCase().includes(q) ||
        r.response_content?.toLowerCase().includes(q) ||
        r.error_message?.toLowerCase().includes(q) ||
        r.tool_calls?.some(tc => tc.name.toLowerCase().includes(q))
      )
    }
    if (provider) list = list.filter(r => r.provider === provider)
    if (statusFilter) {
      const code = parseInt(statusFilter)
      if (!isNaN(code)) list = list.filter(r => Math.floor(r.status_code / 100) === Math.floor(code / 100))
    }
    if (sort !== 'time') {
      list = [...list]
      switch (sort) {
        case 'cost': list.sort((a, b) => (b.total_cost || 0) - (a.total_cost || 0)); break
        case 'latency': list.sort((a, b) => (b.latency || 0) - (a.latency || 0)); break
        case 'tokens': list.sort((a, b) => ((b.input_tokens || 0) + (b.output_tokens || 0)) - ((a.input_tokens || 0) + (a.output_tokens || 0))); break
      }
    }
    return list
  }, [requests, search, provider, statusFilter, sort])

  const totalCost = requests.reduce((s, r) => s + (r.total_cost || 0), 0)
  const avgLatency = requests.length ? requests.reduce((s, r) => s + (r.latency || 0), 0) / requests.length : 0
  const providers = [...new Set(requests.map(r => r.provider).filter(Boolean))]

  return (
    <div className="p-4">
      {/* Session summary */}
      <div className="flex gap-6 mb-4 text-sm text-gray-400 flex-wrap">
        <span><span className="text-white font-bold">{requests.length}</span> requests</span>
        <span>Total cost: <span className="text-yellow-400 font-bold">{formatCost(totalCost, true)}</span></span>
        <span>Avg latency: <span className="text-blue-400">{formatDuration(avgLatency)}</span></span>
        {search && <span className="text-purple-400">{filtered.length} results for "{search}"</span>}
      </div>

      {/* Filters */}
      <div className="flex gap-2 mb-4 flex-wrap items-center">
        <select value={provider} onChange={e => setProvider(e.target.value)}
          className="bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs focus:outline-none focus:border-purple-500">
          <option value="">All providers</option>
          {providers.map(p => <option key={p} value={p}>{p}</option>)}
        </select>
        <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)}
          className="bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs focus:outline-none focus:border-purple-500">
          <option value="">All statuses</option>
          <option value="200">2xx OK</option>
          <option value="400">4xx Client Error</option>
          <option value="500">5xx Server Error</option>
        </select>
        <div className="flex gap-1 ml-auto">
          {(['time', 'cost', 'latency', 'tokens'] as SortKey[]).map(k => (
            <button key={k} onClick={() => setSort(k)}
              className={`px-2 py-1 rounded text-xs border-0 cursor-pointer ${sort === k ? 'bg-purple-700 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'}`}>
              {k}
            </button>
          ))}
        </div>
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full text-xs border-collapse">
          <thead>
            <tr className="text-gray-500 border-b border-gray-800 text-left">
              <th className="pb-2 pr-3 w-8 font-normal">#</th>
              <th className="pb-2 pr-3 font-normal">Provider / Model</th>
              <th className="pb-2 pr-3 font-normal">Path</th>
              <th className="pb-2 pr-3 text-right font-normal">In</th>
              <th className="pb-2 pr-3 text-right font-normal">Out</th>
              <th className="pb-2 pr-3 text-right font-normal">Cost</th>
              <th className="pb-2 pr-3 text-right font-normal">Latency</th>
              <th className="pb-2 pr-3 text-right font-normal">TTFT</th>
              <th className="pb-2 text-right font-normal">Status</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 && (
              <tr>
                <td colSpan={9} className="py-12 text-center text-gray-600">
                  {requests.length === 0 ? 'No requests yet. Make an LLM API call through probe.' : 'No results match your filters.'}
                </td>
              </tr>
            )}
            {filtered.map(r => (
              <tr key={r.id} className="border-b border-gray-800/50 hover:bg-gray-800/40 transition-colors">
                <td className="py-1.5 pr-3 text-gray-600">{r.seq}</td>
                <td className="py-1.5 pr-3">
                  <Link to={`/request/${r.id}`} className="flex items-center gap-1.5 hover:text-purple-400 transition-colors">
                    <span className={`px-1 py-0.5 rounded text-xs ${providerColor(r.provider)}`}>{r.provider}</span>
                    <span className="text-purple-300 truncate max-w-[180px]">{r.model || '—'}</span>
                  </Link>
                </td>
                <td className="py-1.5 pr-3 text-gray-400 max-w-[200px] truncate">
                  <Link to={`/request/${r.id}`} className="hover:text-gray-200 transition-colors">
                    {r.method} {r.path}
                  </Link>
                </td>
                <td className="py-1.5 pr-3 text-right text-gray-500">{formatTokens(r.input_tokens)}</td>
                <td className="py-1.5 pr-3 text-right text-gray-500">{formatTokens(r.output_tokens)}</td>
                <td className="py-1.5 pr-3 text-right text-yellow-400">{formatCost(r.total_cost, r.pricing_known)}</td>
                <td className="py-1.5 pr-3 text-right text-blue-400">{formatDuration(r.latency)}</td>
                <td className="py-1.5 pr-3 text-right text-cyan-400">{r.ttft > 0 ? formatDuration(r.ttft) : '—'}</td>
                <td className="py-1.5 text-right">
                  <StatusBadge req={r} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function StatusBadge({ req }: { req: Request }) {
  if (req.status === 'pending' || req.status === 'streaming') {
    return <span className="text-yellow-400">⏳ {req.status}</span>
  }
  if (req.status === 'error' && !req.status_code) {
    return <span className="text-red-400">error</span>
  }
  const hasAnomaly = req.anomalies?.length > 0
  return (
    <span className={statusColor(req.status_code)}>
      {req.status_code || '—'}
      {hasAnomaly && <span className="text-yellow-400 ml-1">⚠</span>}
    </span>
  )
}
