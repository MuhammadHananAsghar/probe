import { useState, useCallback } from 'react'

interface ViewerProps {
  data: unknown
  depth?: number
  collapsed?: boolean
}

export function JsonViewer({ data, depth = 0, collapsed = depth > 2 }: ViewerProps) {
  if (data === null) return <span className="text-gray-500">null</span>
  if (data === undefined) return <span className="text-gray-500">undefined</span>
  if (typeof data === 'boolean') return <span className="text-orange-400">{String(data)}</span>
  if (typeof data === 'number') return <span className="text-blue-400">{data}</span>
  if (typeof data === 'string') return <span className="text-green-400">"{escapeStr(data)}"</span>
  if (Array.isArray(data)) return <CollapsibleNode data={data} depth={depth} type="array" initialCollapsed={collapsed} />
  if (typeof data === 'object') return <CollapsibleNode data={data as Record<string, unknown>} depth={depth} type="object" initialCollapsed={collapsed} />
  return <span>{String(data)}</span>
}

function escapeStr(s: string): string {
  return s.replace(/\\/g, '\\\\').replace(/"/g, '\\"').replace(/\n/g, '\\n').replace(/\t/g, '\\t')
}

function CollapsibleNode({ data, depth, type, initialCollapsed }: {
  data: Record<string, unknown> | unknown[]
  depth: number
  type: 'object' | 'array'
  initialCollapsed: boolean
}) {
  const [collapsed, setCollapsed] = useState(initialCollapsed)
  const entries: [string | number, unknown][] = Array.isArray(data)
    ? data.map((v, i) => [i, v])
    : Object.entries(data)
  const open = type === 'object' ? '{' : '['
  const close = type === 'object' ? '}' : ']'

  if (entries.length === 0) return <span className="text-gray-400">{open}{close}</span>

  return (
    <span>
      <button onClick={() => setCollapsed(!collapsed)} className="text-gray-500 hover:text-gray-300 select-none w-4">
        {collapsed ? '▶' : '▼'}
      </button>
      <span className="text-gray-400">{open}</span>
      {collapsed ? (
        <span className="text-gray-500 cursor-pointer hover:text-gray-300" onClick={() => setCollapsed(false)}>
          {' '}…{entries.length}{' '}
        </span>
      ) : (
        <>
          {entries.map(([k, v], i) => (
            <div key={String(k)} style={{ paddingLeft: `${(depth + 1) * 14}px` }}>
              {type === 'object' && <span className="text-purple-300">"{k}"</span>}
              {type === 'object' && <span className="text-gray-400">: </span>}
              <JsonViewer data={v} depth={depth + 1} />
              {i < entries.length - 1 && <span className="text-gray-600">,</span>}
            </div>
          ))}
          <div style={{ paddingLeft: `${depth * 14}px` }}>
            <span className="text-gray-400">{close}</span>
          </div>
        </>
      )}
      {collapsed && <span className="text-gray-400">{close}</span>}
    </span>
  )
}

export function JsonBlock({ raw }: { raw: string }) {
  const [copied, setCopied] = useState(false)
  let parsed: unknown
  let parseError = false
  try {
    parsed = JSON.parse(raw)
  } catch {
    parseError = true
  }

  const copy = useCallback(() => {
    navigator.clipboard.writeText(raw)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }, [raw])

  return (
    <div className="relative bg-gray-900 rounded p-3 text-xs font-mono overflow-auto max-h-96">
      <button onClick={copy} className="absolute top-2 right-2 text-gray-500 hover:text-gray-300 text-xs border-0 p-0 bg-transparent">
        {copied ? '✓ copied' : 'copy'}
      </button>
      {parseError ? (
        <pre className="text-gray-300 whitespace-pre-wrap break-all">{raw}</pre>
      ) : (
        <JsonViewer data={parsed} depth={0} collapsed={false} />
      )}
    </div>
  )
}
