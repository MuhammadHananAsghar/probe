export function formatDuration(ns: number): string {
  if (!ns || ns === 0) return '—'
  const ms = ns / 1_000_000
  if (ms < 1000) return `${ms.toFixed(0)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

export function formatCost(cost: number, known: boolean): string {
  if (!known) return 'n/a'
  if (!cost || cost === 0) return '$0'
  return `$${cost.toFixed(8)}`
}

export function formatTokens(n: number): string {
  if (!n || n === 0) return '—'
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`
  return String(n)
}

export function statusColor(code: number): string {
  if (!code || code === 0) return 'text-gray-500'
  if (code < 300) return 'text-green-400'
  if (code < 400) return 'text-yellow-400'
  if (code < 500) return 'text-orange-400'
  return 'text-red-400'
}

export function providerColor(provider: string): string {
  switch ((provider || '').toLowerCase()) {
    case 'openai': return 'bg-emerald-900 text-emerald-300'
    case 'anthropic': return 'bg-orange-900 text-orange-300'
    case 'google': return 'bg-blue-900 text-blue-300'
    default: return 'bg-gray-800 text-gray-300'
  }
}
