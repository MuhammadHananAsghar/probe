<p align="center">
  <pre align="center">
  █▀▀█ █▀▀█ █▀▀█ █▀▀▄ █▀▀
  █  █ █▄▄▀ █  █ █▀▀▄ █▀▀
  █▀▀▀ ▀ ▀▀ ▀▀▀▀ ▀▀▀  ▀▀▀
  </pre>
  <br>
  <strong>Universal LLM API Debugger — intercept, inspect, replay every AI call your app makes.</strong>
  <br><br>
  <a href="https://github.com/MuhammadHananAsghar/probe/releases"><img src="https://img.shields.io/github/v/release/MuhammadHananAsghar/probe?style=flat-square" alt="Release"></a>
  <a href="https://github.com/MuhammadHananAsghar/probe/blob/main/LICENSE"><img src="https://img.shields.io/github/license/MuhammadHananAsghar/probe?style=flat-square" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/MuhammadHananAsghar/probe"><img src="https://goreportcard.com/badge/github.com/MuhammadHananAsghar/probe?style=flat-square" alt="Go Report"></a>
  <a href="https://github.com/MuhammadHananAsghar/probe/actions"><img src="https://img.shields.io/github/actions/workflow/status/MuhammadHananAsghar/probe/ci.yml?style=flat-square&label=CI" alt="CI"></a>
</p>

<p align="center">
  <img src="assets/demo.gif" alt="probe demo" width="800" />
</p>

---

**Probe** is an open-source LLM API debugger that sits between your application and any AI provider. Point your SDK at `http://localhost:9000` and get a full live view of every request, response, token count, cost, and latency — with zero code changes.

```bash
probe listen
# Proxy:     http://localhost:9000
# Dashboard: http://localhost:9001
```

```
  █▀▀█ █▀▀█ █▀▀█ █▀▀▄ █▀▀
  █  █ █▄▄▀ █  █ █▀▀▄ █▀▀
  █▀▀▀ ▀ ▀▀ ▀▀▀▀ ▀▀▀  ▀▀▀
  probe v0.1.1  ·  listening on :9000  ·  dashboard http://localhost:9001

  #   Provider    Model                          Tokens     Cost      Latency  Status
  ──────────────────────────────────────────────────────────────────────────────────────
  1   openai      gpt-4.1-mini                   1,842      $0.0028   312ms    ✓ 200
  2   anthropic   claude-haiku-4-5-20251001       2,105      $0.0053   481ms    ✓ 200
  3   openai      gpt-4.1-mini (stream)           938        $0.0014   128ms    ✓ 200
  4   groq        llama-4-maverick               4,210      $0.0021   89ms     ✓ 200
  5   openai      gpt-4.1-mini                   —          —         —        ✗ 429
```

## Features

- **Zero code changes** — just redirect your `base_url` to probe, nothing else
- **All major providers** — OpenAI, Anthropic, Google Gemini, Ollama, Azure OpenAI, OpenRouter, Mistral, Cohere, Groq, Together AI, Fireworks AI, AWS Bedrock
- **Streaming debugger** — chunk timeline, TTFT (time-to-first-token), full SSE inspection
- **Tool call inspector** — visualise every function definition, invocation, and result
- **Token & cost tracking** — per-request and session totals with real pricing data
- **Web dashboard** — live React SPA at `localhost:9001` with WebSocket updates
- **Request replay** — re-send any captured request, optionally with a different model or provider
- **Cross-provider compare** — run the same prompt against two models side-by-side
- **Anomaly alerts** — configurable cost/latency/error banners in the TUI
- **Export everything** — HAR, JSON, NDJSON, Markdown, curl/python/node snippets
- **SQLite persistence** — optional `--persist` mode with history queries and cleanup
- **Token waste detection** — duplicate requests, unchanged system prompts, high I/O ratio
- **Rate limit visibility** — parses provider rate limit headers into readable warnings
- **Single binary** — no runtime dependencies, no Docker, no setup

## Install

### Quick install (macOS / Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/MuhammadHananAsghar/probe/main/install.sh | sh
```

### Homebrew (macOS)

```bash
brew install MuhammadHananAsghar/tap/probe
```

### Go install

```bash
go install github.com/MuhammadHananAsghar/probe/cmd/probe@latest
```

### Build from source

```bash
git clone https://github.com/MuhammadHananAsghar/probe.git
cd probe
# Build the React dashboard
cd internal/dashboard/ui && npm ci && npm run build && cd ../../..
# Build the binary
go build -o probe ./cmd/probe
```

### Download binary

Pre-built binaries for macOS, Linux, and Windows are available on the [releases page](https://github.com/MuhammadHananAsghar/probe/releases).

---

## Quick Start

### 1. Start probe

```bash
probe listen
# Proxy:     http://localhost:9000
# Dashboard: http://localhost:9001
```

### 2. Point your SDK at probe

One line change in any SDK — no code changes anywhere else.

### 3. Watch traffic live

Every request appears in the TUI instantly. Press `Enter` on any row to inspect the full request/response, `s` for the streaming chunk timeline, `t` for tool calls. Open `http://localhost:9001` for the live web dashboard.

---

## SDK Integration

Probe works with every major LLM SDK. Change one line — the `base_url` — and all traffic flows through probe automatically.

### OpenAI

```python
from openai import OpenAI

client = OpenAI(
    api_key="your-key",
    base_url="http://localhost:9000/v1",  # ← probe
)

response = client.chat.completions.create(
    model="gpt-4.1-mini",
    messages=[{"role": "user", "content": "Hello!"}],
)
```

```bash
# Or via environment variable — zero code changes
export OPENAI_BASE_URL=http://localhost:9000/v1
python your_app.py
```

### Anthropic

```python
import anthropic

client = anthropic.Anthropic(
    api_key="your-key",
    base_url="http://localhost:9000",  # ← probe (no /v1 suffix)
)

response = client.messages.create(
    model="claude-haiku-4-5-20251001",
    max_tokens=256,
    messages=[{"role": "user", "content": "Hello!"}],
)
```

```bash
# Or via environment variable
export ANTHROPIC_BASE_URL=http://localhost:9000
python your_app.py
```

### Google Gemini

```python
from google import genai
from google.genai import types

client = genai.Client(
    api_key="your-key",
    http_options=types.HttpOptions(base_url="http://localhost:9000"),  # ← probe
)

response = client.models.generate_content(
    model="gemini-2.5-flash",
    contents="Hello!",
)
```

### Groq

```python
from groq import Groq

client = Groq(
    api_key="your-key",
    base_url="http://localhost:9000",  # ← probe
)

response = client.chat.completions.create(
    model="meta-llama/llama-4-maverick-17b-128e-instruct",
    messages=[{"role": "user", "content": "Hello!"}],
)
```

### Mistral

```python
from mistralai import Mistral

client = Mistral(
    api_key="your-key",
    server_url="http://localhost:9000",  # ← probe
)

response = client.chat.complete(
    model="mistral-small-2503",
    messages=[{"role": "user", "content": "Hello!"}],
)
```

### Cohere

```python
import cohere

client = cohere.ClientV2(
    api_key="your-key",
    base_url="http://localhost:9000",  # ← probe
)

response = client.chat(
    model="command-a-03-2025",
    messages=[{"role": "user", "content": "Hello!"}],
)
```

### Together AI

```python
from together import Together

client = Together(
    api_key="your-key",
    base_url="http://localhost:9000/v1",  # ← probe
)

response = client.chat.completions.create(
    model="meta-llama/Llama-4-Maverick-17B-128E-Instruct-FP8",
    messages=[{"role": "user", "content": "Hello!"}],
)
```

### Fireworks AI

```python
from fireworks.client import Fireworks

client = Fireworks(
    api_key="your-key",
    base_url="http://localhost:9000/v1",  # ← probe
)

response = client.chat.completions.create(
    model="accounts/fireworks/models/llama4-maverick-instruct-basic",
    messages=[{"role": "user", "content": "Hello!"}],
)
```

### OpenRouter

```python
from openai import OpenAI  # OpenRouter is OpenAI-compatible

client = OpenAI(
    api_key="your-openrouter-key",
    base_url="http://localhost:9000/v1",  # ← probe
    default_headers={"HTTP-Referer": "https://yourapp.com"},
)

response = client.chat.completions.create(
    model="openai/gpt-4.1-mini",  # or any OpenRouter model
    messages=[{"role": "user", "content": "Hello!"}],
)
```

### Ollama (local)

```python
from openai import OpenAI  # Ollama is OpenAI-compatible

client = OpenAI(
    api_key="ollama",
    base_url="http://localhost:9000/v1",  # ← probe (forwards to Ollama at :11434)
)

response = client.chat.completions.create(
    model="llama3.2",  # must be pulled: ollama pull llama3.2
    messages=[{"role": "user", "content": "Hello!"}],
)
```

### Azure OpenAI

```python
import httpx
from openai import AzureOpenAI

client = AzureOpenAI(
    api_key="your-key",
    azure_endpoint="https://your-resource.openai.azure.com",
    api_version="2024-10-21",
    http_client=httpx.Client(
        transport=httpx.HTTPTransport(proxy="http://localhost:9000"),  # ← probe
    ),
)
```

### AWS Bedrock

```python
import boto3
from botocore.config import Config

client = boto3.client(
    "bedrock-runtime",
    region_name="us-east-1",
    config=Config(proxies={"https": "http://localhost:9000"}),  # ← probe
)
```

### `base_url` Reference

| Provider | `base_url` / config |
|---|---|
| OpenAI | `http://localhost:9000/v1` |
| Anthropic | `http://localhost:9000` |
| Google Gemini | `http_options=HttpOptions(base_url="http://localhost:9000")` |
| Groq | `http://localhost:9000` |
| Mistral | `server_url="http://localhost:9000"` |
| Cohere | `http://localhost:9000` |
| Together AI | `http://localhost:9000/v1` |
| Fireworks AI | `http://localhost:9000/v1` |
| OpenRouter | `http://localhost:9000/v1` |
| Ollama | `http://localhost:9000/v1` |
| Azure OpenAI | `httpx.HTTPTransport(proxy="http://localhost:9000")` |
| AWS Bedrock | `Config(proxies={"https": "http://localhost:9000"})` |

> **Why `/v1` for some but not others?**
> OpenAI-compatible SDKs (OpenAI, Groq, Together, Fireworks, OpenRouter, Ollama) append paths like `/chat/completions` to the base URL themselves, so probe needs `/v1` in the base.
> Anthropic, Mistral, and Cohere SDKs include `/v1` in their path already, so the base should be just `http://localhost:9000`.

---

## CLI Reference

```bash
# Start the proxy + TUI
probe listen
probe listen --port 9000              # custom proxy port
probe listen --persist                # save all requests to SQLite
probe listen --quiet                  # no TUI, plain log output
probe listen --alert-cost 0.10        # banner when a single request costs >$0.10
probe listen --alert-latency 5s       # banner on requests slower than 5s
probe listen --alert-error            # banner on any non-2xx response

# Inspect a captured request
probe inspect 3                       # show full detail for request #3
probe inspect 3 --curl                # print a curl snippet
probe inspect 3 --lang python         # print a Python snippet
probe inspect 3 --lang node           # print a Node.js snippet

# Replay a request
probe replay 3                        # re-send as-is
probe replay 3 --model gpt-4.1-mini  # swap the model
probe replay 3 --provider anthropic   # translate to a different provider
probe replay 3 --temperature 0.2      # override parameters
probe replay 3 --export replay.md     # save comparison report

# Compare two responses
probe compare 3 5                     # diff requests #3 and #5
probe compare 3 --models gpt-4.1-mini,gpt-4.1  # run same prompt on two models

# Export captured traffic
probe export                          # JSON to stdout
probe export --format har -o out.har  # HAR file (Chrome DevTools compatible)
probe export --format ndjson          # streaming NDJSON
probe export --format markdown        # human-readable session report
probe export --filter provider=openai # filter by field

# Query history (requires --persist)
probe history                         # list recent requests
probe history --cost                  # sort by cost
probe history --errors                # errors only
probe history --model gpt-4.1-mini    # filter by model
probe history --limit 50              # last 50 requests
probe history --cleanup               # delete requests older than retention_days

# Detect waste and duplicates
probe analyze
probe analyze --waste                 # token waste analysis
probe analyze --duplicates            # find identical requests

# Version
probe version

# Update to latest version (auto-detects install method)
probe update

# Uninstall (removes binary, config, CA certs, history — all install methods)
probe uninstall
probe uninstall --yes         # skip confirmation
```

## TUI Keybindings

| View | Key | Action |
|---|---|---|
| List | `↑↓` / `jk` | Navigate requests |
| List | `Enter` | Inspect request |
| List | `q` / `Ctrl+C` | Quit |
| Detail | `↑↓` / `jk` | Scroll |
| Detail | `s` | Open stream view |
| Detail | `t` | Open tool call view |
| Detail | `c` | Copy curl to clipboard |
| Detail | `Esc` / `q` | Back to list |
| Stream / Tools | `↑↓` / `jk` | Scroll |
| Stream / Tools | `Esc` / `q` | Back |
| Global | `d` | Dismiss alert banner |

## Web Dashboard

Open `http://localhost:9001` in your browser for the live dashboard:

- Real-time request stream via WebSocket
- Full request/response detail with syntax highlighting
- Cost and token charts across the session
- Tool call explorer — definitions, invocations, and results
- Works alongside the TUI — both update simultaneously

## Supported Providers

| Provider | Auto-detected | Streaming | Tool Calls | Cost Tracking |
|---|---|---|---|---|
| OpenAI | ✓ | ✓ | ✓ | ✓ |
| Anthropic | ✓ | ✓ | ✓ | ✓ |
| Google Gemini | ✓ | ✓ | ✓ | ✓ |
| Ollama (local) | ✓ | ✓ | — | free |
| Azure OpenAI | ✓ | ✓ | ✓ | ✓ |
| OpenRouter | ✓ | ✓ | ✓ | ✓ |
| Mistral | ✓ | ✓ | ✓ | ✓ |
| Cohere | ✓ | ✓ | — | ✓ |
| Groq | ✓ | ✓ | ✓ | ✓ |
| Together AI | ✓ | ✓ | — | ✓ |
| Fireworks AI | ✓ | ✓ | — | ✓ |
| AWS Bedrock | ✓ | — | — | ✓ |
| OpenAI-compatible | ✓ (path) | ✓ | ✓ | — |

## Configuration

Probe reads `~/.probe/config.yaml` on startup. All values can be overridden with CLI flags.

```yaml
proxy:
  port: 9000
  dashboard_port: 9001

storage:
  persist: false          # enable SQLite persistence
  retention_days: 7       # cleanup threshold for probe history --cleanup
  ring_buffer_size: 1000  # in-memory request history size

alerts:
  cost_threshold: 0.10    # alert when a single request costs more than $0.10
  latency_threshold: 5s   # alert on requests slower than 5s
  alert_on_error: true    # alert on any non-2xx response
```

## How It Works

```
YOUR APP                        PROBE                        LLM PROVIDER
┌─────────────┐   HTTP/HTTPS   ┌──────────────────────┐   HTTPS   ┌──────────────┐
│             │───────────────►│                      │──────────►│              │
│  SDK call   │                │  1. Intercept req    │           │  OpenAI /    │
│             │◄───────────────│  2. Forward upstream │◄──────────│  Anthropic / │
│             │                │  3. Parse response   │           │  Gemini ...  │
└─────────────┘                │  4. Store + display  │           └──────────────┘
                               └──────────┬───────────┘
                                          │
                               ┌──────────▼───────────┐
                               │   TUI + Dashboard    │
                               │  (bubbletea + React) │
                               └──────────────────────┘
```

1. Your app sends an API request to `http://localhost:9000` instead of the real provider
2. Probe forwards it upstream over HTTPS with your real API key untouched
3. The response (including SSE streams) is intercepted, parsed, and stored
4. The TUI and web dashboard update live via the in-memory ring buffer

Your API key is **never logged, stored, or displayed** — it passes through transparently.

## Architecture

| Component | Technology |
|---|---|
| CLI / Proxy | Go, Cobra, `net/http` reverse proxy |
| TUI | Bubbletea, Lipgloss |
| Web dashboard | React, TypeScript, Tailwind, Vite |
| Live updates | WebSocket (`coder/websocket`) |
| Persistence | SQLite (`modernc.org/sqlite`, pure Go) |
| Pricing data | LiteLLM pricing DB (embedded JSON) |
| Logging | zerolog |

## Project Structure

```
probe/
├── cmd/probe/             # CLI entrypoint (cobra commands)
├── internal/
│   ├── proxy/             # HTTP reverse proxy + SSE interceptor
│   ├── provider/          # Per-provider request/response parsers
│   ├── store/             # In-memory ring buffer + SQLite persistence
│   ├── tui/               # Bubbletea TUI (list, detail, stream, tools views)
│   ├── dashboard/         # REST API + WebSocket hub + go:embed React SPA
│   │   └── ui/            # React/TS/Tailwind frontend
│   ├── replay/            # Request replay + cross-provider translation
│   ├── export/            # HAR, JSON, NDJSON, Markdown, curl generation
│   ├── analyze/           # Anomaly detection, waste, rate limit parsing
│   └── cost/              # Pricing DB + token cost calculator
├── pkg/
│   ├── config/            # YAML config loader (~/.probe/config.yaml)
│   └── logger/            # zerolog wrapper
├── examples/              # Python SDK examples for every provider
├── pricing/               # Embedded LiteLLM pricing data
├── .github/workflows/     # CI + release (goreleaser)
└── .goreleaser.yml        # Cross-platform release config
```

## License

MIT — see [LICENSE](LICENSE).
