import { useEffect, useRef, useState } from 'react'

export interface Message {
  role: string
  content: string
}

export interface ToolCall {
  id: string
  name: string
  arguments_json: string
  parse_error?: boolean
}

export interface ToolDefinition {
  name: string
  description: string
  schema: string
}

export interface ToolResult {
  tool_call_id: string
  content: string
  is_error?: boolean
}

export interface Anomaly {
  kind: string
  message: string
}

export interface StreamStats {
  chunk_count: number
  stream_duration: number
  throughput_tps: number
}

export interface Request {
  id: string
  seq: number
  started_at: string
  ended_at: string
  method: string
  url: string
  path: string
  provider: string
  provider_host: string
  model: string
  status: string
  status_code: number
  latency: number
  ttft: number
  input_tokens: number
  output_tokens: number
  input_cost: number
  output_cost: number
  total_cost: number
  pricing_known: boolean
  stream: boolean
  stream_stats?: StreamStats
  system_prompt: string
  messages: Message[]
  tools: ToolDefinition[]
  tool_calls: ToolCall[]
  tool_results: ToolResult[]
  anomalies: Anomaly[]
  finish_reason: string
  response_content: string
  error_message: string
  conversation_id: string
  replay_of?: string
  many_tools: boolean
  request_headers: Record<string, string>
  response_headers: Record<string, string>
  request_body: string
  response_body: string
}

type WSMessage =
  | { type: 'snapshot'; data: Request[] }
  | { type: 'request'; data: Request }
  | { type: 'update'; id: string; data: Request }

export function useWebSocket() {
  const [requests, setRequests] = useState<Request[]>([])
  const [connected, setConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    function connect() {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const ws = new WebSocket(`${protocol}//${window.location.host}/api/ws`)
      wsRef.current = ws

      ws.onopen = () => setConnected(true)
      ws.onclose = () => {
        setConnected(false)
        reconnectTimer.current = setTimeout(connect, 2000)
      }
      ws.onerror = () => ws.close()

      ws.onmessage = (e) => {
        const msg: WSMessage = JSON.parse(e.data)
        if (msg.type === 'snapshot') {
          setRequests((msg.data || []).slice().reverse())
        } else if (msg.type === 'request') {
          setRequests(prev => [msg.data, ...prev])
        } else if (msg.type === 'update') {
          setRequests(prev => prev.map(r => r.id === msg.id ? msg.data : r))
        }
      }
    }

    connect()
    return () => {
      wsRef.current?.close()
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current)
    }
  }, [])

  return { requests, connected }
}
