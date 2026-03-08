# Probe — Universal LLM API Debugger
## Project Plan & Architecture

---

## Philosophy

```bash
probe listen
```

That's it. One command. Every LLM API call your app makes — visible, measured, debuggable.

No config. No SDK integration. No code changes. Just start probe, point your app at it, and see everything.

**Probe is Wireshark for the LLM era.**

---

## The Core Insight

Every developer building on LLM APIs is flying blind:

- "Why did that request cost $0.47?" — No visibility into token usage across requests
- "Why is streaming stuttering?" — No way to see SSE chunk timing
- "The tool call failed but I don't know why" — JSON blobs 500 lines deep
- "We're over budget this month" — No cost tracking until the invoice arrives
- "It worked yesterday, what changed?" — No request history or diff

Current debugging workflow: `curl` → squint at JSON → add `print()` statements → check provider dashboard 30 minutes later. This is absurd for a protocol that costs real money per request.

**Probe intercepts LLM API traffic and gives it structure, cost, timing, and a beautiful UI.**

```
┌─────────────────────────────────────────────────┐
│              Who Needs Probe                     │
│                                                  │
│  DEFAULT (80% of users)                          │
│  ┌────────────────────────────────────────────┐  │
│  │  Solo dev / small team                     │  │
│  │  • Building an app on LLM APIs             │  │
│  │  • Wants to see what's happening           │  │
│  │  • Wants to control costs                  │  │
│  │  • probe listen → done                     │  │
│  └────────────────────────────────────────────┘  │
│                                                  │
│  POWER USER (15% of users)                       │
│  ┌────────────────────────────────────────────┐  │
│  │  Team / production debugging               │  │
│  │  • Shared sessions                         │  │
│  │  • CI integration                          │  │
│  │  • Cost alerting                           │  │
│  │  • HAR/JSON export for bug reports         │  │
│  └────────────────────────────────────────────┘  │
│                                                  │
│  PLATFORM (5% of users)                          │
│  ┌────────────────────────────────────────────┐  │
│  │  AI-native companies                       │  │
│  │  • Multi-service LLM observability         │  │
│  │  • Cost allocation per feature/team        │  │
│  │  • Model comparison & regression testing   │  │
│  └────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

---

## How It Works

```
YOUR APP                         PROBE (localhost:8080)                    LLM PROVIDERS
┌──────────────┐                ┌──────────────────────────┐             ┌──────────────┐
│              │  HTTP/HTTPS    │                          │  HTTPS      │              │
│  Next.js     │───────────────►│  1. Intercept request    │────────────►│  OpenAI      │
│  Python      │                │  2. Detect provider      │             │  Anthropic   │
│  Go app      │◄───────────────│  3. Parse: model, tokens │◄────────────│  Google      │
│  Any app     │   Response     │  4. Measure: latency,    │  Response   │  Mistral     │
│              │                │     TTFT, chunk timing   │             │  Cohere      │
└──────────────┘                │  5. Calculate cost       │             │  Ollama      │
                                │  6. Store in ring buffer │             │  Any OpenAI- │
                                │  7. Stream to terminal   │             │  compatible  │
                                │  8. Stream to dashboard  │             └──────────────┘
                                │                          │
                                │  Terminal UI (live)      │
                                │  Dashboard (localhost:4041)│
                                └──────────────────────────┘
```

**Two ways to connect your app:**

```bash
# Option A: Proxy mode (zero code changes)
# Set environment variable, probe intercepts automatically
HTTPS_PROXY=http://localhost:8080 node app.js
HTTPS_PROXY=http://localhost:8080 python app.py

# Option B: Base URL mode (one-line code change)
# Point your SDK at probe instead of the provider
# Python (OpenAI SDK)
client = OpenAI(base_url="http://localhost:8080/v1")

# Python (Anthropic SDK)
client = Anthropic(base_url="http://localhost:8080")

# Node.js
const openai = new OpenAI({ baseURL: "http://localhost:8080/v1" });
```

Both modes work. Proxy mode requires zero code changes. Base URL mode avoids TLS certificate setup.

---

## User Experience

### The Default (Zero Config)
```bash
$ probe listen

  🔍 Probe v1.0.0 — LLM API Debugger

  Mode:         Proxy (intercept all HTTPS to LLM providers)
  Proxy:        http://localhost:8080
  Dashboard:    http://localhost:4041
  Providers:    auto-detect (OpenAI, Anthropic, Google, Mistral, Cohere, Ollama)

  Tip: Run your app with HTTPS_PROXY=http://localhost:8080

  Waiting for requests...
```

### Requests Streaming In
```
  ┌──────────────────────────────────────────────────────────────────┐
  │ #1  POST anthropic /v1/messages            claude-sonnet-4-20250514│
  │     Tokens: 1,247 in → 583 out            Cost: $0.0089         │
  │     Latency: 1.2s (TTFT: 180ms)           Tools: 2 calls        │
  │     Status: 200 ✓                                                │
  │                                                                  │
  │ #2  POST openai /v1/chat/completions       gpt-4o               │
  │     Tokens: 892 in → 1,204 out            Cost: $0.0134         │
  │     Latency: 2.8s (TTFT: 420ms)           Stream: 47 chunks     │
  │     Status: 200 ✓                                                │
  │                                                                  │
  │ #3  POST anthropic /v1/messages            claude-sonnet-4-20250514│
  │     Tokens: 3,891 in → 12 out             Cost: $0.0121         │
  │     Latency: 8.1s                         ⚠ TIMEOUT (retry #2)  │
  │     Status: 529 OVERLOADED                                       │
  │                                                                  │
  │ #4  POST openai /v1/chat/completions       gpt-4o-mini          │
  │     Tokens: 456 in → 234 out              Cost: $0.0003         │
  │     Latency: 0.8s (TTFT: 95ms)            Stream: 12 chunks     │
  │     Status: 200 ✓                                                │
  └──────────────────────────────────────────────────────────────────┘

  Session: 4 requests | $0.0347 total | Avg TTFT: 231ms | 1 error
```

### Detailed Request View
```bash
$ probe inspect 3
# → Show full detail of request #3

  ┌─ Request #3 ─────────────────────────────────────────────────────┐
  │                                                                  │
  │ Provider:   Anthropic                                            │
  │ Endpoint:   POST /v1/messages                                    │
  │ Model:      claude-sonnet-4-20250514                                │
  │ Status:     529 OVERLOADED (retried 2x, then failed)             │
  │                                                                  │
  │ ── Timing ──────────────────────────────────────────────────────│
  │ Total:      8.1s                                                 │
  │ Attempt 1:  2.1s → 529 (retry after 1s)                         │
  │ Attempt 2:  2.4s → 529 (retry after 2s)                         │
  │ Attempt 3:  2.6s → 529 (gave up)                                │
  │                                                                  │
  │ ── Tokens & Cost ───────────────────────────────────────────────│
  │ Input:      3,891 tokens ($0.0117)                               │
  │ Output:     12 tokens ($0.0004)                                  │
  │ Total:      $0.0121 (charged even on failure)                    │
  │                                                                  │
  │ ── Messages (3) ────────────────────────────────────────────────│
  │ [system]  You are a helpful assistant that...  (892 tokens)      │
  │ [user]    Analyze the following code and...    (2,847 tokens)    │
  │ [asst]    I'll analyze...                      (12 tokens)       │
  │                                                                  │
  │ ── Tool Calls ──────────────────────────────────────────────────│
  │ (none — request failed before tool execution)                    │
  │                                                                  │
  │ ── Headers ─────────────────────────────────────────────────────│
  │ anthropic-version: 2023-06-01                                    │
  │ x-api-key: sk-ant-...••••                                       │
  │ anthropic-ratelimit-remaining: 0  ← THIS IS WHY                 │
  │                                                                  │
  │ Actions: [r]eplay  [d]iff with another  [e]xport  [c]opy curl   │
  └──────────────────────────────────────────────────────────────────┘
```

---

## Installation

```bash
# macOS / Linux
curl -fsSL https://probe.dev/install.sh | sh

# macOS (Homebrew)
brew install probe-dev/tap/probe

# Windows
scoop install probe

# Go devs
go install github.com/probe-dev/probe@latest
```

Single binary. No dependencies. No runtime. Works in 3 seconds.

---

## Tech Stack

| Component | Technology | Why |
|---|---|---|
| Language | Go 1.22+ | Single binary, native TLS, fast proxy |
| Proxy Engine | net/http + httputil.ReverseProxy | Battle-tested, handles streaming |
| TLS Interception | Custom CA + dynamic cert generation | Intercept HTTPS without code changes |
| SSE Parser | Custom (line-by-line scanner) | Parse streaming chunks from all providers |
| Terminal UI | bubbletea + lipgloss | Beautiful, interactive terminal dashboard |
| Web Dashboard | Embedded React SPA (go:embed) | Full inspector in browser |
| Live Updates | WebSocket (gorilla/websocket) | Terminal ↔ Dashboard real-time sync |
| Storage | In-memory ring buffer + optional SQLite | Fast, bounded memory, opt-in persistence |
| Cost Database | Embedded JSON (go:embed) | Model pricing auto-updated via GitHub |
| Config | ~/.probe/config.yaml | Auto-generated, never required |

---

## Provider Detection & Parsing

Probe auto-detects which LLM provider a request targets based on the hostname or URL path:

```
┌───────────────────────────────────────────────────────────────────┐
│ Provider        Host / Pattern              API Format            │
├───────────────────────────────────────────────────────────────────┤
│ OpenAI          api.openai.com              /v1/chat/completions  │
│ Anthropic       api.anthropic.com           /v1/messages          │
│ Google          generativelanguage.          /v1/models/*/        │
│                 googleapis.com              generateContent       │
│ Mistral         api.mistral.ai              /v1/chat/completions  │
│ Cohere          api.cohere.com              /v2/chat              │
│ Groq            api.groq.com                /v1/chat/completions  │
│ Together        api.together.xyz            /v1/chat/completions  │
│ Fireworks       api.fireworks.ai            /v1/chat/completions  │
│ Ollama          localhost:11434             /api/chat             │
│ OpenRouter      openrouter.ai               /api/v1/chat/...     │
│ Azure OpenAI    *.openai.azure.com          /openai/deployments/* │
│ AWS Bedrock     bedrock-runtime.*.          /model/*/invoke       │
│                 amazonaws.com                                     │
│ OpenAI-compat   (any)                       /v1/chat/completions  │
│                                             (auto-detected)       │
└───────────────────────────────────────────────────────────────────┘
```

For each provider, Probe extracts:

```
REQUEST:                          RESPONSE:
├── model name                    ├── output tokens (from usage or counted)
├── input tokens (from usage      ├── finish reason (stop, tool_call, length)
│   or estimated via tiktoken)    ├── tool call results
├── messages array                ├── streaming chunks + timing
├── tools/functions defined       ├── rate limit headers
├── temperature, max_tokens       ├── error details
├── stream: true/false            └── response headers
└── request headers
```

---

## Project Structure

```
probe/
├── cmd/
│   └── probe/
│       └── main.go                 # Single binary entrypoint
│
├── internal/
│   ├── proxy/                      # ── THE CORE: HTTP(S) PROXY ──
│   │   ├── proxy.go                # Main proxy server
│   │   ├── tls.go                  # Dynamic TLS cert generation
│   │   ├── ca.go                   # Local CA creation & trust
│   │   └── passthrough.go          # Non-LLM traffic passthrough
│   │
│   ├── intercept/                  # ── REQUEST INTERCEPTION ──
│   │   ├── interceptor.go          # Middleware: capture req/res pairs
│   │   ├── body.go                 # Body reading without consuming (tee reader)
│   │   └── streaming.go            # SSE stream interception + chunk timing
│   │
│   ├── provider/                   # ── PROVIDER DETECTION & PARSING ──
│   │   ├── detect.go               # Auto-detect provider from host/path
│   │   ├── provider.go             # Provider interface
│   │   ├── openai.go               # OpenAI request/response parser
│   │   ├── anthropic.go            # Anthropic parser
│   │   ├── google.go               # Google Gemini parser
│   │   ├── mistral.go              # Mistral parser
│   │   ├── cohere.go               # Cohere parser
│   │   ├── ollama.go               # Ollama parser
│   │   ├── bedrock.go              # AWS Bedrock parser
│   │   ├── azure.go                # Azure OpenAI parser
│   │   └── generic.go              # OpenAI-compatible fallback
│   │
│   ├── cost/                       # ── COST CALCULATION ──
│   │   ├── calculator.go           # Cost per request
│   │   ├── pricing.go              # Model pricing database
│   │   ├── pricing_data.json       # Embedded pricing (go:embed)
│   │   └── tokens.go               # Token counting (tiktoken-go)
│   │
│   ├── store/                      # ── REQUEST STORAGE ──
│   │   ├── store.go                # Storage interface
│   │   ├── memory.go               # In-memory ring buffer (default)
│   │   ├── sqlite.go               # SQLite persistence (opt-in)
│   │   └── models.go               # Request/Response data models
│   │
│   ├── analyze/                    # ── ANALYSIS ENGINE ──
│   │   ├── session.go              # Session-level aggregates (cost, latency)
│   │   ├── anomaly.go              # Detect: high cost, slow TTFT, errors
│   │   ├── diff.go                 # Diff two requests (prompt changes)
│   │   └── timeline.go             # Request timeline visualization data
│   │
│   ├── replay/                     # ── REQUEST REPLAY ──
│   │   ├── replay.go               # Re-send captured request
│   │   ├── modify.go               # Modify before replay (change model, params)
│   │   └── compare.go              # Compare original vs replayed response
│   │
│   ├── export/                     # ── EXPORT ──
│   │   ├── har.go                  # HAR format export
│   │   ├── json.go                 # JSON export
│   │   ├── curl.go                 # Generate curl command
│   │   └── markdown.go             # Markdown report
│   │
│   ├── tui/                        # ── TERMINAL UI ──
│   │   ├── app.go                  # Bubbletea main app
│   │   ├── list.go                 # Request list view
│   │   ├── detail.go               # Request detail view
│   │   ├── stream.go               # Streaming chunk visualizer
│   │   ├── tools.go                # Tool call tree view
│   │   ├── stats.go                # Session stats bar
│   │   └── styles.go               # Lipgloss styles
│   │
│   └── dashboard/                  # ── WEB DASHBOARD ──
│       ├── embed.go                # go:embed built SPA
│       ├── server.go               # Dashboard HTTP server
│       ├── ws.go                   # WebSocket: live request stream
│       └── ui/                     # React SPA
│           ├── src/
│           │   ├── App.tsx
│           │   ├── pages/
│           │   │   ├── RequestList.tsx
│           │   │   ├── RequestDetail.tsx
│           │   │   ├── StreamView.tsx
│           │   │   ├── ToolCallTree.tsx
│           │   │   ├── CostDashboard.tsx
│           │   │   ├── Timeline.tsx
│           │   │   └── Settings.tsx
│           │   └── components/
│           │       ├── JsonViewer.tsx
│           │       ├── TokenCounter.tsx
│           │       ├── CostBadge.tsx
│           │       ├── LatencyBar.tsx
│           │       ├── StreamChunks.tsx
│           │       └── DiffView.tsx
│           └── package.json
│
├── pkg/
│   ├── config/                     # ~/.probe/config.yaml
│   │   └── config.go
│   └── logger/
│       └── logger.go
│
├── pricing/                        # Model pricing data (auto-updated)
│   ├── update.go                   # Script to fetch latest pricing
│   └── models.json                 # Pricing database
│
├── deployments/
│   └── install.sh                  # Universal installer
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Build Phases

### Phase 1 — "probe listen" Shows LLM Traffic (Week 1-2)
**Goal:** Start probe, run your app, see every LLM API call with tokens, cost, and latency.

**Proxy core:**
- [ ] HTTP forward proxy (`HTTPS_PROXY` mode)
- [ ] Reverse proxy / base URL mode (`localhost:8080/v1/...`)
- [ ] TLS interception with auto-generated local CA
- [ ] First-run: generate CA cert, prompt to trust it (or skip for base URL mode)
- [ ] Non-LLM traffic passthrough (don't break the app)
- [ ] Graceful shutdown

**Provider detection:**
- [ ] Auto-detect provider from hostname (openai, anthropic, google, mistral)
- [ ] OpenAI request/response parser (messages, model, tokens, tools)
- [ ] Anthropic request/response parser
- [ ] Generic OpenAI-compatible fallback parser
- [ ] Extract: model name, input/output tokens, finish reason, error details

**Cost calculation:**
- [ ] Embedded pricing database (JSON, updated at build time)
- [ ] Cost per request: input_tokens × input_price + output_tokens × output_price
- [ ] Session running totals

**Terminal UI:**
- [ ] Live request list (method, provider, model, tokens, cost, latency, status)
- [ ] Color-coded status (green 200, yellow retries, red errors)
- [ ] Session stats bar (total cost, request count, avg latency)
- [ ] Request detail view (press Enter on a request)

**What the user does:**
```bash
curl -fsSL https://probe.dev/install.sh | sh
probe listen
# In another terminal:
HTTPS_PROXY=http://localhost:8080 python my_llm_app.py
# → See every LLM call in the probe terminal
```

### Phase 2 — Streaming Debugger (Week 2-3)
**Goal:** See SSE streaming in real-time — chunk timing, TTFT, stalls.

- [ ] SSE stream interception without buffering (tee reader pattern)
- [ ] Parse streaming chunks per provider (OpenAI `data: {...}`, Anthropic `event: content_block_delta`)
- [ ] Time to First Token (TTFT) measurement
- [ ] Inter-chunk timing (detect stalls: >500ms between chunks)
- [ ] Chunk count and token-per-chunk statistics
- [ ] Terminal UI: stream visualization (dots appearing in real-time like a progress bar)
- [ ] Stream replay: replay a streaming response at original timing
- [ ] Handle edge cases: interrupted streams, empty chunks, error mid-stream

```
  ┌─ Stream View (#2) ─────────────────────────────────────────────┐
  │                                                                │
  │  TTFT: 420ms  │  Chunks: 47  │  Duration: 2.8s               │
  │                                                                │
  │  Timeline:                                                     │
  │  ├── 0ms     [connect]                                        │
  │  ├── 420ms   chunk 1: "I" (TTFT)                              │
  │  ├── 445ms   chunk 2: "'ll"                                   │
  │  ├── 468ms   chunk 3: " analyze"                              │
  │  ├── 491ms   chunk 4: " the"                                  │
  │  │   ... (43 more chunks, avg 52ms apart)                     │
  │  ├── 2,780ms chunk 47: "." [stop]                             │
  │  └── 2,800ms [done]                                           │
  │                                                                │
  │  Throughput: 16.8 tokens/sec                                  │
  │  Stalls: 0 (no gaps > 500ms)                                  │
  └────────────────────────────────────────────────────────────────┘
```

### Phase 3 — Tool Call Inspector (Week 3-4)
**Goal:** Visualize function calling / tool use chains clearly.

- [ ] Parse tool definitions from request (functions/tools array)
- [ ] Parse tool call decisions from response
- [ ] Parse tool results from follow-up messages
- [ ] Build tool call chain: request → model decides → function called → result → model continues
- [ ] Terminal UI: tree view of tool call chains
- [ ] Show: function name, arguments (formatted JSON), result, timing per step
- [ ] Detect issues: tool call with no result, malformed arguments, schema mismatches
- [ ] Multi-turn tool call tracking (correlate across requests in a session)

```
  ┌─ Tool Calls (#5) ──────────────────────────────────────────────┐
  │                                                                │
  │  Chain: 3 steps                                                │
  │                                                                │
  │  ① Model decided to call: get_weather                         │
  │     Args: { "city": "London", "units": "celsius" }            │
  │     ↓                                                          │
  │  ② Tool result:                                                │
  │     { "temp": 12, "condition": "cloudy", "wind": "15km/h" }  │
  │     Latency: 340ms (your function execution time)              │
  │     ↓                                                          │
  │  ③ Model decided to call: get_forecast                         │
  │     Args: { "city": "London", "days": 3 }                    │
  │     ↓                                                          │
  │  ④ Tool result:                                                │
  │     { "forecast": [...] }                                      │
  │     Latency: 520ms                                             │
  │     ↓                                                          │
  │  ⑤ Model final response:                                      │
  │     "The weather in London is currently 12°C and cloudy..."   │
  │                                                                │
  │  Total: 2 tool calls | 860ms in tool execution                │
  └────────────────────────────────────────────────────────────────┘
```

### Phase 4 — Web Dashboard (Week 4-5)
**Goal:** Full-featured browser-based inspector at localhost:4041.

- [ ] React SPA embedded via go:embed
- [ ] WebSocket connection for live request streaming
- [ ] Request list with filtering (provider, model, status, cost range)
- [ ] Request detail with tabs: Overview, Messages, Tools, Stream, Headers, Raw
- [ ] JSON viewer with syntax highlighting and collapsible nodes
- [ ] Cost dashboard: spending over time, cost per model, cost per hour
- [ ] Timeline view: waterfall chart of all requests (like Chrome DevTools Network tab)
- [ ] Search: full-text search across request/response bodies
- [ ] Diff view: compare two requests side-by-side (prompt changes, response changes)
- [ ] Dark/light theme
- [ ] Dashboard auto-opens on first run (`--no-browser` to disable)

### Phase 5 — Replay & Compare (Week 5-6)
**Goal:** Re-send any request, optionally to a different model. Compare results.

- [ ] Replay a captured request to the same provider
- [ ] Replay with modifications: change model, temperature, max_tokens, system prompt
- [ ] Replay to a different provider (translate OpenAI request → Anthropic format)
- [ ] Compare: original vs replayed (side-by-side diff of responses)
- [ ] Compare: same prompt across multiple models simultaneously
- [ ] CLI: `probe replay 3` or `probe replay 3 --model gpt-4o`
- [ ] Dashboard: replay button on every request with modification UI
- [ ] Cost comparison: "same prompt, claude-sonnet costs $0.008, gpt-4o costs $0.012"
- [ ] Export comparison as markdown (for sharing with team)

```bash
$ probe replay 3 --model gpt-4o-mini

  ┌─ Replay Comparison ────────────────────────────────────────────┐
  │                                                                │
  │  Original: claude-sonnet-4-20250514    Replay: gpt-4o-mini      │
  │  Tokens:   3,891 in → 583 out     Tokens: 3,891 in → 612 out│
  │  Cost:     $0.0089                 Cost:   $0.0004           │
  │  Latency:  1.2s (TTFT: 180ms)     Latency: 0.6s (TTFT: 90ms)│
  │  Quality:  (your judgment)         Quality: (your judgment)   │
  │                                                                │
  │  Cost savings: 95.5% with gpt-4o-mini                        │
  │  Speed improvement: 2x faster                                 │
  │                                                                │
  │  Response diff: 23 lines changed (see dashboard for full diff)│
  └────────────────────────────────────────────────────────────────┘
```

### Phase 6 — Export, Persistence & Advanced Features (Week 6-7)
**Goal:** Save sessions, export data, advanced analysis.

**Export:**
- [ ] HAR format (compatible with Chrome DevTools, Charles Proxy)
- [ ] JSON (machine-readable, for custom analysis)
- [ ] curl command generation (reproduce any request)
- [ ] Markdown report (shareable summary with cost breakdown)
- [ ] `probe export --format har --output session.har`
- [ ] `probe export --last 1h` (export last hour of requests)

**Persistence:**
- [ ] `probe listen --persist` — Save all requests to SQLite (survives restart)
- [ ] `probe history` — Browse past sessions
- [ ] `probe history --cost --last 7d` — Cost summary for last 7 days
- [ ] Auto-cleanup: configurable retention (default: 7 days)

**Alerting:**
- [ ] `--alert-cost 1.00` — Terminal notification when session cost exceeds $1
- [ ] `--alert-latency 5s` — Flag requests slower than 5 seconds
- [ ] `--alert-error` — Sound/notification on errors (rate limits, timeouts)

**Analysis:**
- [ ] Token waste detector: "This request sends 2,000 tokens of system prompt that never changes — consider caching"
- [ ] Duplicate request detector: "Requests #4 and #7 are identical — add caching"
- [ ] Cost optimizer: "Switching model X to Y would save 60% with similar quality"
- [ ] Rate limit tracker: parse rate limit headers, show remaining quota

### Phase 7 — Multi-Provider Intelligence & Polish (Week 7-8)
**Goal:** Ship a production-ready tool.

**Provider support:**
- [ ] Complete Google Gemini parser (different API format)
- [ ] Cohere V2 parser
- [ ] Groq, Together, Fireworks parsers
- [ ] Ollama parser (local models, no cost)
- [ ] Azure OpenAI parser
- [ ] AWS Bedrock parser
- [ ] OpenRouter parser
- [ ] Generic OpenAI-compatible fallback (catches everything else)

**Pricing:**
- [ ] `probe update-pricing` — Fetch latest model pricing from providers
- [ ] GitHub Action to auto-update pricing.json weekly
- [ ] User overrides: custom pricing in config for private/fine-tuned models

**Distribution:**
- [ ] Cross-compile: Linux, macOS, Windows × amd64, arm64
- [ ] GitHub Actions CI/CD + goreleaser
- [ ] Homebrew tap
- [ ] Scoop bucket
- [ ] `probe update` self-update command
- [ ] Install script with OS/arch detection

**Polish:**
- [ ] `probe listen --port 9090` (custom port)
- [ ] `probe listen --filter anthropic` (only intercept Anthropic calls)
- [ ] `probe listen --no-tls` (skip CA cert for base URL mode only)
- [ ] First-run wizard: detect OS, install CA cert, explain proxy setup
- [ ] Clear error messages: "No requests seen. Make sure your app uses HTTPS_PROXY=..."
- [ ] README with terminal GIFs
- [ ] Benchmarks: prove probe adds <2ms latency overhead

---

## CLI Design

```bash
# ── THE BASICS (90% of usage) ──────────────────────────
probe listen                                # Start intercepting (port 8080)
probe listen --port 9090                    # Custom port
probe listen --no-browser                   # Don't auto-open dashboard

# ── INSPECT ────────────────────────────────────────────
probe inspect 3                             # View request #3 in detail
probe inspect --last                        # View most recent request
probe inspect 3 --messages                  # Show just the messages array
probe inspect 3 --tools                     # Show just tool calls
probe inspect 3 --stream                    # Show stream chunk timeline
probe inspect 3 --curl                      # Print as curl command
probe inspect 3 --raw                       # Print raw request/response

# ── REPLAY ─────────────────────────────────────────────
probe replay 3                              # Re-send request #3
probe replay 3 --model gpt-4o-mini          # Replay with different model
probe replay 3 --provider openai            # Translate to OpenAI format
probe replay 3 --temperature 0              # Override parameters

# ── COMPARE ────────────────────────────────────────────
probe compare 3 7                           # Diff two requests
probe compare 3 --models claude-sonnet,gpt-4o,gemini-pro
                                            # Same prompt, 3 models

# ── EXPORT ─────────────────────────────────────────────
probe export                                # Export session as JSON
probe export --format har                   # HAR format
probe export --format markdown              # Cost report
probe export --last 1h                      # Last hour only

# ── HISTORY (with --persist) ───────────────────────────
probe history                               # Browse past sessions
probe history --cost --last 7d              # 7-day cost summary
probe history --errors                      # Show only errors

# ── CONFIG ─────────────────────────────────────────────
probe config set pricing.custom.my-model 0.001/0.003
                                            # Custom model pricing
probe update-pricing                        # Fetch latest pricing
probe update                                # Self-update

# ── META ───────────────────────────────────────────────
probe version
probe help
```

---

## TLS Interception: How It Actually Works

The biggest technical challenge is intercepting HTTPS traffic without breaking the app. Here's the approach:

```
┌─────────────────────────────────────────────────────────────────┐
│                  TLS Interception Flow                           │
│                                                                 │
│  1. First run: probe generates a local CA certificate           │
│     ~/.probe/ca-cert.pem + ~/.probe/ca-key.pem                 │
│                                                                 │
│  2. User trusts the CA (probe prompts with OS-specific cmd):    │
│     macOS:  sudo security add-trusted-cert -d ...               │
│     Linux:  sudo cp ca-cert.pem /usr/local/share/ca-certs/     │
│     Windows: certutil -addstore -f "ROOT" ca-cert.pem           │
│                                                                 │
│  3. App makes HTTPS request via proxy:                          │
│     App → CONNECT api.openai.com:443 → Probe                   │
│                                                                 │
│  4. Probe dynamically generates a cert FOR api.openai.com       │
│     signed by the local CA. App trusts it (CA is trusted).      │
│                                                                 │
│  5. Probe decrypts, inspects, re-encrypts, forwards to real     │
│     api.openai.com                                              │
│                                                                 │
│  6. Only LLM provider hostnames are intercepted.                │
│     All other HTTPS traffic passes through untouched.           │
└─────────────────────────────────────────────────────────────────┘
```

**If TLS setup is too complex, use base URL mode instead:**
```python
# No CA cert needed. Just point your SDK at probe.
client = Anthropic(base_url="http://localhost:8080")
```

Probe detects which mode is being used and adapts automatically.

---

## Cost Database (Embedded)

Probe ships with a pricing database that covers all major models. Updated weekly via GitHub Actions.

```json
{
  "models": {
    "claude-sonnet-4-20250514": {
      "provider": "anthropic",
      "input_cost_per_1m": 3.00,
      "output_cost_per_1m": 15.00,
      "context_window": 200000
    },
    "gpt-4o": {
      "provider": "openai",
      "input_cost_per_1m": 2.50,
      "output_cost_per_1m": 10.00,
      "context_window": 128000
    },
    "gpt-4o-mini": {
      "provider": "openai",
      "input_cost_per_1m": 0.15,
      "output_cost_per_1m": 0.60,
      "context_window": 128000
    },
    "gemini-2.0-flash": {
      "provider": "google",
      "input_cost_per_1m": 0.10,
      "output_cost_per_1m": 0.40,
      "context_window": 1000000
    }
  },
  "updated_at": "2026-03-07"
}
```

Users can override pricing for fine-tuned or private models:
```yaml
# ~/.probe/config.yaml
pricing:
  custom:
    my-fine-tuned-gpt4: { input: 6.00, output: 12.00 }
```

---

## Key Design Decisions

### Why a proxy, not an SDK?
SDKs require code changes in every project, in every language. A proxy works with **any language, any framework, any SDK** — Python, Node, Go, Rust, curl. Set one environment variable and everything flows through probe. Same philosophy as Wormhole: zero friction.

### Why both proxy mode and base URL mode?
TLS interception requires trusting a local CA certificate, which can be intimidating for new users or blocked in some environments. Base URL mode (`base_url="http://localhost:8080"`) requires one line of code but no certificate setup. Give users the choice — most will start with base URL mode and graduate to proxy mode.

### Why embedded pricing instead of API lookup?
Pricing lookups would add latency to every request and require internet access. Embedded JSON is instant, works offline, and updates weekly. Users with custom/fine-tuned models can override in config.

### Why ring buffer, not append-only log?
Developers debug recent requests, not requests from last Tuesday. A ring buffer (default: last 1,000 requests) keeps memory bounded and fast. `--persist` enables SQLite for users who want history.

### Why Go?
Same reason as Wormhole: single binary, native TLS handling, goroutine-per-connection for high concurrency, cross-compiles everywhere. The proxy architecture maps perfectly to Go's stdlib.

### Why bubbletea for terminal UI?
It's the gold standard for Go terminal apps. Rich, interactive, and the same library Wormhole uses — code reuse. The terminal UI is not just a log stream — it's a navigable interface where you can select requests, view details, and trigger replays.

---

## Performance Targets

| Metric | Target | How |
|---|---|---|
| Proxy latency overhead | < 2ms (non-streaming) | Direct forwarding, no body copy |
| Streaming overhead | < 1ms per chunk | Tee reader, no buffering |
| Memory (idle) | < 15 MB | Minimal footprint |
| Memory (1000 requests) | < 80 MB | Ring buffer with bounded size |
| Startup time | < 50ms | No heavy init |
| Binary size | < 12 MB | Stripped, gzipped dashboard |
| Dashboard load | < 200ms | Embedded, no CDN fetch |

**Critical constraint:** Probe must NEVER slow down the app's LLM calls. The proxy is on the hot path. Every microsecond counts. Body inspection happens via tee readers (zero-copy), not buffering.

---

## Dependencies

```
# Core
net/http (stdlib)                    # Proxy server
crypto/tls (stdlib)                  # TLS interception
crypto/x509 (stdlib)                 # CA cert generation
httputil (stdlib)                    # Reverse proxy

# Provider parsing
github.com/tidwall/gjson             # Fast JSON path extraction (no full parse)
github.com/pkoukk/tiktoken-go       # Token counting (OpenAI tokenizer)

# Terminal UI
github.com/charmbracelet/bubbletea  # Interactive TUI framework
github.com/charmbracelet/lipgloss   # Styling
github.com/charmbracelet/bubbles    # Pre-built TUI components

# Dashboard
github.com/gorilla/websocket        # Live updates to browser

# Storage (optional)
modernc.org/sqlite                  # Persistent history (pure Go)

# CLI
github.com/spf13/cobra              # Command framework

# Logging
github.com/rs/zerolog               # Structured logging
```

---

## What Makes Probe Different

| Feature | Charles Proxy | Fiddler | Postman | Probe |
|---|---|---|---|---|
| Understands LLM APIs | ❌ Raw JSON | ❌ Raw JSON | ❌ Manual | ✅ Semantic parsing |
| Token counting | ❌ | ❌ | ❌ | ✅ Per-request, per-model |
| Cost tracking | ❌ | ❌ | ❌ | ✅ Real-time $/request |
| Streaming debugger | ❌ | ❌ | Partial | ✅ Chunk timing, TTFT |
| Tool call visualizer | ❌ | ❌ | ❌ | ✅ Tree view + chain |
| Model comparison | ❌ | ❌ | Manual | ✅ Replay to any model |
| CLI-first | ❌ GUI only | ❌ GUI only | ❌ GUI first | ✅ Terminal + dashboard |
| Zero config | ❌ Complex | ❌ Complex | ❌ Setup | ✅ `probe listen` |
| Free | ❌ $50-$250 | ❌ $12/mo | Partial | ✅ Open source |
| Single binary | ❌ | ❌ | ❌ Electron | ✅ ~12MB Go binary |

---

## Monetization Path (Future, Optional)

**Free forever (open source):**
- Local proxy + terminal UI + dashboard
- All provider support
- Export (HAR, JSON, curl)
- Request replay

**Probe Pro ($9/mo, if we want to go there):**
- Cloud dashboard (share sessions with team via URL)
- Persistent history with search
- Cost alerts (Slack/email when daily spend exceeds threshold)
- CI mode: `probe ci --budget 5.00` fails the build if LLM costs exceed budget
- Team cost allocation: tag requests by feature/service, split costs

This is entirely optional. The open-source version should be complete and useful on its own. Monetization is a bonus, not a requirement.
