import { useParams, Link } from 'react-router-dom'
import { useState } from 'react'
import type { Request } from '../hooks/useWebSocket'
import { JsonBlock } from '../components/JsonViewer'
import { formatDuration, formatCost, formatTokens, statusColor, providerColor } from '../utils'

type Tab = 'overview' | 'messages' | 'tools' | 'stream' | 'headers' | 'raw' | 'compare'

export function RequestDetail({ requests }: { requests: Request[] }) {
  const { id } = useParams<{ id: string }>()
  const req = requests.find(r => r.id === id)
  const [tab, setTab] = useState<Tab>('overview')
  const [showReplay, setShowReplay] = useState(false)

  if (!req) return (
    <div className="p-8 text-center text-gray-500 text-sm">
      Request not found. <Link to="/" className="text-purple-400 hover:underline">Back to list</Link>
    </div>
  )

  const tabs: { key: Tab; label: string }[] = [
    { key: 'overview', label: 'Overview' },
    { key: 'messages', label: `Messages (${(req.messages?.length || 0) + (req.system_prompt ? 1 : 0)})` },
    { key: 'tools', label: `Tools (${req.tool_calls?.length || 0})` },
    { key: 'stream', label: 'Stream' },
    { key: 'headers', label: 'Headers' },
    { key: 'raw', label: 'Raw' },
    { key: 'compare', label: 'Compare' },
  ]

  // Find related replays
  const relatedReplays = requests.filter(r => r.replay_of === req.id || r.id === req.replay_of)

  return (
    <div className="p-4 max-w-5xl mx-auto">
      <div className="mb-4">
        <Link to="/" className="text-gray-500 hover:text-gray-300 text-xs">← All requests</Link>
        <div className="flex items-start justify-between mt-1">
          <div>
            <h1 className="text-lg font-bold">Request #{req.seq}</h1>
            {req.replay_of && (
              <div className="text-xs text-purple-400 mt-0.5">
                ↩ Replay of <Link to={`/request/${req.replay_of}`} className="underline hover:text-purple-300">{req.replay_of.slice(0, 8)}</Link>
              </div>
            )}
          </div>
          <button
            onClick={() => setShowReplay(true)}
            className="text-xs bg-purple-700 hover:bg-purple-600 text-white px-3 py-1.5 rounded border-0 cursor-pointer transition-colors"
          >
            ↩ Replay
          </button>
        </div>
        <div className="flex gap-2 mt-1 text-xs text-gray-500 flex-wrap items-center">
          <span className={`px-1.5 py-0.5 rounded ${providerColor(req.provider)}`}>{req.provider}</span>
          <span className="text-purple-300">{req.model}</span>
          <span>{req.method} {req.path}</span>
          <span className={statusColor(req.status_code)}>{req.status_code || req.status}</span>
        </div>
      </div>

      {showReplay && (
        <ReplayModal req={req} onClose={() => setShowReplay(false)} />
      )}

      {req.anomalies?.length > 0 && (
        <div className="mb-4 bg-yellow-950/50 border border-yellow-800/50 rounded p-3">
          {req.anomalies.map((a, i) => (
            <div key={i} className={`text-xs py-0.5 ${a.kind === 'malformed_args' || a.kind === 'tool_loop' ? 'text-red-400' : 'text-yellow-400'}`}>
              ⚠ {a.message}
            </div>
          ))}
        </div>
      )}

      {/* Tabs */}
      <div className="flex gap-0.5 border-b border-gray-800 mb-4 flex-wrap">
        {tabs.map(t => (
          <button key={t.key} onClick={() => setTab(t.key)}
            className={`px-3 py-1.5 text-xs rounded-t-sm transition-colors border-0 cursor-pointer ${tab === t.key ? 'bg-gray-800 text-white border-b-2 border-purple-500' : 'text-gray-500 hover:text-gray-300 bg-transparent'}`}>
            {t.label}
          </button>
        ))}
      </div>

      {tab === 'overview' && <OverviewTab req={req} />}
      {tab === 'messages' && <MessagesTab req={req} />}
      {tab === 'tools' && <ToolsTab req={req} />}
      {tab === 'stream' && <StreamTab req={req} />}
      {tab === 'headers' && <HeadersTab req={req} />}
      {tab === 'raw' && <RawTab req={req} />}
      {tab === 'compare' && <CompareTab req={req} requests={requests} relatedReplays={relatedReplays} />}
    </div>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="mb-6">
      <h2 className="text-xs text-gray-500 uppercase tracking-wider mb-2 border-b border-gray-800 pb-1">{title}</h2>
      {children}
    </div>
  )
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex gap-4 py-0.5 text-xs">
      <span className="text-gray-500 w-32 flex-shrink-0">{label}</span>
      <span className="text-gray-200">{value}</span>
    </div>
  )
}

function OverviewTab({ req }: { req: Request }) {
  return (
    <div>
      <Section title="Timing">
        <Row label="Latency" value={formatDuration(req.latency)} />
        <Row label="TTFT" value={req.ttft > 0 ? formatDuration(req.ttft) : '—'} />
        <Row label="Started" value={new Date(req.started_at).toLocaleTimeString()} />
        {req.ended_at && <Row label="Ended" value={new Date(req.ended_at).toLocaleTimeString()} />}
      </Section>
      <Section title="Tokens & Cost">
        <Row label="Input tokens" value={`${formatTokens(req.input_tokens)}  ·  ${formatCost(req.input_cost, req.pricing_known)}`} />
        <Row label="Output tokens" value={`${formatTokens(req.output_tokens)}  ·  ${formatCost(req.output_cost, req.pricing_known)}`} />
        <Row label="Total cost" value={<span className="text-yellow-400 font-bold">{formatCost(req.total_cost, req.pricing_known)}</span>} />
      </Section>
      <Section title="Details">
        <Row label="Finish reason" value={req.finish_reason || '—'} />
        <Row label="Streaming" value={req.stream ? 'Yes' : 'No'} />
        <Row label="Conversation" value={<span className="text-gray-400 font-mono">{req.conversation_id || '—'}</span>} />
        {req.many_tools && <Row label="Tool count" value={<span className="text-yellow-400">⚠ &gt;20 tools may degrade model behavior</span>} />}
        {req.error_message && <Row label="Error" value={<span className="text-red-400">{req.error_message}</span>} />}
      </Section>
    </div>
  )
}

function MessagesTab({ req }: { req: Request }) {
  const msgs: { role: string; content: string }[] = []
  if (req.system_prompt) msgs.push({ role: 'system', content: req.system_prompt })
  for (const m of (req.messages || [])) msgs.push(m)

  return (
    <div className="space-y-3">
      {msgs.length === 0 && <p className="text-gray-500 text-xs">(none)</p>}
      {msgs.map((m, i) => (
        <div key={i} className="bg-gray-900 rounded p-3">
          <div className="text-xs text-gray-500 mb-1.5 font-bold uppercase tracking-wider">{m.role}</div>
          <pre className="text-xs text-gray-200 whitespace-pre-wrap break-words font-mono leading-relaxed">{m.content}</pre>
        </div>
      ))}
      {req.response_content && (
        <div className="bg-gray-900 rounded p-3 border-l-2 border-purple-700">
          <div className="text-xs text-purple-400 mb-1.5 font-bold uppercase tracking-wider">assistant (response)</div>
          <pre className="text-xs text-gray-200 whitespace-pre-wrap break-words font-mono leading-relaxed">{req.response_content}</pre>
        </div>
      )}
    </div>
  )
}

function ToolsTab({ req }: { req: Request }) {
  const resultMap = new Map((req.tool_results || []).map(tr => [tr.tool_call_id, tr]))

  return (
    <div>
      <Section title={`Tool Calls (${req.tool_calls?.length || 0})`}>
        {(!req.tool_calls || req.tool_calls.length === 0) && (
          <p className="text-gray-500 text-xs">(none)</p>
        )}
        {(req.tool_calls || []).map((tc, i) => {
          const result = resultMap.get(tc.id)
          return (
            <div key={i} className="mb-4 bg-gray-900 rounded p-3">
              <div className="flex items-center gap-2 mb-2">
                <span className="text-xs text-gray-500">Step {i + 1}</span>
                <span className="text-purple-300 font-bold">{tc.name}</span>
                {tc.parse_error && <span className="text-red-400 text-xs bg-red-900/30 px-1 rounded">[malformed JSON]</span>}
              </div>
              <div className="mb-2">
                <div className="text-xs text-gray-500 mb-1">arguments:</div>
                <JsonBlock raw={tc.arguments_json || '{}'} />
              </div>
              {result ? (
                <div>
                  <div className={`text-xs mb-1 ${result.is_error ? 'text-red-400' : 'text-green-400'}`}>
                    result{result.is_error ? ' (error)' : ''}:
                  </div>
                  <pre className="text-xs text-gray-300 whitespace-pre-wrap break-words bg-gray-800 p-2 rounded font-mono">{result.content}</pre>
                </div>
              ) : (
                <div className="text-xs text-yellow-400">⏳ awaiting result...</div>
              )}
            </div>
          )
        })}
      </Section>
    </div>
  )
}

function StreamTab({ req }: { req: Request }) {
  if (!req.stream) return <p className="text-gray-500 text-xs">(not a streaming request)</p>
  const ss = req.stream_stats
  if (!ss && req.status === 'streaming') return <p className="text-yellow-400 text-xs">⏳ Streaming in progress...</p>
  if (!ss) return <p className="text-gray-500 text-xs">(no stream stats)</p>

  return (
    <Section title="Stream Statistics">
      <Row label="TTFT" value={req.ttft > 0 ? formatDuration(req.ttft) : '—'} />
      <Row label="Chunks" value={ss.chunk_count} />
      <Row label="Duration" value={formatDuration(ss.stream_duration)} />
      <Row label="Throughput" value={ss.throughput_tps > 0 ? `${ss.throughput_tps.toFixed(1)} tok/s` : '—'} />
    </Section>
  )
}

function HeadersTab({ req }: { req: Request }) {
  const mask = (k: string, v: string) => {
    if (/auth|key|secret|token/i.test(k)) {
      const visible = v.slice(0, 12)
      return visible + '...'
    }
    return v
  }
  return (
    <div>
      <Section title="Request Headers">
        {Object.keys(req.request_headers || {}).length === 0
          ? <p className="text-gray-500 text-xs">(none captured)</p>
          : Object.entries(req.request_headers || {}).map(([k, v]) => (
            <div key={k} className="flex gap-2 text-xs py-0.5 border-b border-gray-800/30">
              <span className="text-gray-500 w-48 flex-shrink-0">{k}</span>
              <span className="text-gray-300 break-all">{mask(k, v)}</span>
            </div>
          ))}
      </Section>
      <Section title="Response Headers">
        {Object.keys(req.response_headers || {}).length === 0
          ? <p className="text-gray-500 text-xs">(none captured)</p>
          : Object.entries(req.response_headers || {}).map(([k, v]) => (
            <div key={k} className="flex gap-2 text-xs py-0.5 border-b border-gray-800/30">
              <span className="text-gray-500 w-48 flex-shrink-0">{k}</span>
              <span className="text-gray-300 break-all">{v}</span>
            </div>
          ))}
      </Section>
    </div>
  )
}

function RawTab({ req }: { req: Request }) {
  let reqBody = ''
  let resBody = ''
  try { reqBody = req.request_body ? atob(req.request_body) : '' } catch { reqBody = req.request_body || '' }
  try { resBody = req.response_body ? atob(req.response_body) : '' } catch { resBody = req.response_body || '' }

  return (
    <div>
      <Section title="Request Body">
        <JsonBlock raw={reqBody || '(empty)'} />
      </Section>
      <Section title="Response Body">
        <JsonBlock raw={resBody || '(empty)'} />
      </Section>
    </div>
  )
}

// ── Compare tab ───────────────────────────────────────────────────────────────

function CompareTab({ req, requests, relatedReplays }: { req: Request; requests: Request[]; relatedReplays: Request[] }) {
  const [compareId, setCompareId] = useState('')
  const [comparing, setComparing] = useState(false)
  const [result, setResult] = useState<CompareResult | null>(null)

  const doCompare = async () => {
    if (!compareId) return
    setComparing(true)
    try {
      const res = await fetch('/api/compare', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id_a: req.id, id_b: compareId }),
      })
      if (res.ok) setResult(await res.json())
    } finally {
      setComparing(false)
    }
  }

  const others = requests.filter(r => r.id !== req.id)

  return (
    <div>
      <Section title="Compare with...">
        <div className="flex gap-2 mb-4">
          <select value={compareId} onChange={e => setCompareId(e.target.value)}
            className="flex-1 bg-gray-800 border border-gray-700 rounded px-2 py-1.5 text-xs focus:outline-none focus:border-purple-500">
            <option value="">Select a request to compare</option>
            {relatedReplays.length > 0 && (
              <optgroup label="Replays">
                {relatedReplays.map(r => (
                  <option key={r.id} value={r.id}>#{r.seq} {r.model} ({r.provider})</option>
                ))}
              </optgroup>
            )}
            <optgroup label="All requests">
              {others.map(r => (
                <option key={r.id} value={r.id}>#{r.seq} {r.model} ({r.provider})</option>
              ))}
            </optgroup>
          </select>
          <button onClick={doCompare} disabled={!compareId || comparing}
            className="px-3 py-1.5 bg-purple-700 hover:bg-purple-600 disabled:opacity-50 text-white text-xs rounded border-0 cursor-pointer transition-colors">
            {comparing ? '...' : 'Compare'}
          </button>
        </div>
      </Section>

      {result && (
        <Section title="Comparison Results">
          <div className="overflow-x-auto">
            <table className="w-full text-xs border-collapse">
              <thead>
                <tr className="text-gray-500 border-b border-gray-800">
                  <th className="pb-1 pr-3 text-left font-normal">Metric</th>
                  <th className="pb-1 pr-3 text-left font-normal">A (#{req.seq})</th>
                  <th className="pb-1 pr-3 text-left font-normal">B</th>
                  <th className="pb-1 pr-3 text-left font-normal">Delta</th>
                  <th className="pb-1 text-left font-normal">Better</th>
                </tr>
              </thead>
              <tbody>
                {(result.Metrics || []).map((row: MetricRow, i: number) => (
                  <tr key={i} className="border-b border-gray-800/30">
                    <td className="py-1 pr-3 text-gray-400">{row.Label}</td>
                    <td className="py-1 pr-3 text-gray-200">{row.A}</td>
                    <td className="py-1 pr-3 text-gray-200">{row.B}</td>
                    <td className="py-1 pr-3 text-gray-400">{row.Delta}</td>
                    <td className={`py-1 text-xs font-bold ${row.Better === 'A' ? 'text-blue-400' : row.Better === 'B' ? 'text-green-400' : 'text-gray-600'}`}>
                      {row.Better || '—'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {result.Summary && (
            <div className="mt-3 text-xs text-yellow-400 bg-yellow-950/30 border border-yellow-800/30 rounded p-2">
              {result.Summary}
            </div>
          )}
        </Section>
      )}
    </div>
  )
}

interface MetricRow { Label: string; A: string; B: string; Delta: string; Better: string }
interface CompareResult { Metrics: MetricRow[]; Summary: string }

// ── Replay modal ──────────────────────────────────────────────────────────────

function ReplayModal({ req, onClose }: { req: Request; onClose: () => void }) {
  const [model, setModel] = useState(req.model || '')
  const [provider, setProvider] = useState(req.provider || '')
  const [temp, setTemp] = useState('')
  const [maxTok, setMaxTok] = useState('')
  const [system, setSystem] = useState(req.system_prompt || '')
  const [replaying, setReplaying] = useState(false)
  const [diffs, setDiffs] = useState<string[]>([])
  const [error, setError] = useState('')
  const [done, setDone] = useState(false)

  const doReplay = async () => {
    setReplaying(true)
    setError('')
    try {
      const body: Record<string, unknown> = { model, provider }
      if (system) body.system_prompt = system
      if (temp !== '') { body.temperature = parseFloat(temp); body.has_temperature = true }
      if (maxTok !== '') { body.max_tokens = parseInt(maxTok); body.has_max_tokens = true }

      const res = await fetch(`/api/replay/${req.id}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!res.ok) {
        setError(await res.text())
        return
      }
      const data = await res.json()
      setDiffs(data.parameter_diffs || [])
      setDone(true)
    } catch (e) {
      setError(String(e))
    } finally {
      setReplaying(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50" onClick={e => e.target === e.currentTarget && onClose()}>
      <div className="bg-gray-900 border border-gray-700 rounded-lg p-5 w-full max-w-md">
        <h2 className="font-bold mb-4 text-sm">Replay Request #{req.seq}</h2>
        {done ? (
          <div>
            <div className="text-green-400 text-sm mb-2">✓ Replay dispatched</div>
            {diffs.length > 0 && <div className="text-xs text-gray-400 mb-3">Changes: {diffs.join(', ')}</div>}
            <div className="text-xs text-gray-500">Check the Requests list for the new replay entry.</div>
            <button onClick={onClose} className="mt-4 px-3 py-1.5 bg-gray-700 text-white text-xs rounded border-0 cursor-pointer">Close</button>
          </div>
        ) : (
          <div className="space-y-3">
            <Field label="Model" value={model} onChange={setModel} placeholder={req.model} />
            <Field label="Provider" value={provider} onChange={setProvider} placeholder={req.provider} />
            <Field label="Temperature" value={temp} onChange={setTemp} placeholder="(unchanged)" />
            <Field label="Max tokens" value={maxTok} onChange={setMaxTok} placeholder="(unchanged)" />
            <div>
              <label className="text-xs text-gray-500 block mb-1">System prompt</label>
              <textarea value={system} onChange={e => setSystem(e.target.value)} rows={3}
                className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1.5 text-xs font-mono focus:outline-none focus:border-purple-500 resize-none" />
            </div>
            {error && <div className="text-red-400 text-xs">{error}</div>}
            <div className="flex gap-2 justify-end pt-1">
              <button onClick={onClose} className="px-3 py-1.5 bg-gray-700 text-white text-xs rounded border-0 cursor-pointer">Cancel</button>
              <button onClick={doReplay} disabled={replaying}
                className="px-3 py-1.5 bg-purple-700 hover:bg-purple-600 disabled:opacity-50 text-white text-xs rounded border-0 cursor-pointer transition-colors">
                {replaying ? 'Replaying...' : 'Replay'}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function Field({ label, value, onChange, placeholder }: { label: string; value: string; onChange: (v: string) => void; placeholder?: string }) {
  return (
    <div>
      <label className="text-xs text-gray-500 block mb-1">{label}</label>
      <input value={value} onChange={e => onChange(e.target.value)} placeholder={placeholder}
        className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1.5 text-xs focus:outline-none focus:border-purple-500 placeholder-gray-600" />
    </div>
  )
}
