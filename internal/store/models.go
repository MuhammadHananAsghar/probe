// Package store defines the core data models used throughout probe.
package store

import "time"

// ProviderName identifies an LLM API provider.
type ProviderName string

const (
	ProviderOpenAI      ProviderName = "openai"
	ProviderAnthropic   ProviderName = "anthropic"
	ProviderGoogle      ProviderName = "google"
	ProviderMistral     ProviderName = "mistral"
	ProviderCohere      ProviderName = "cohere"
	ProviderGroq        ProviderName = "groq"
	ProviderTogether    ProviderName = "together"
	ProviderFireworks   ProviderName = "fireworks"
	ProviderOllama      ProviderName = "ollama"
	ProviderOpenRouter  ProviderName = "openrouter"
	ProviderAzureOpenAI ProviderName = "azure-openai"
	ProviderBedrock     ProviderName = "bedrock"
	ProviderCompatible  ProviderName = "openai-compatible"
	ProviderUnknown     ProviderName = "unknown"
)

// FinishReason is why the model stopped generating.
type FinishReason string

const (
	FinishStop      FinishReason = "stop"
	FinishLength    FinishReason = "length"
	FinishToolCall  FinishReason = "tool_call"
	FinishError     FinishReason = "error"
	FinishUnknown   FinishReason = ""
)

// Message is a single conversation turn.
type Message struct {
	Role       string `json:"role"`
	Content    string `json:"content"`
	TokenCount int    `json:"token_count,omitempty"`
}

// ToolDefinition is a tool the model can call.
type ToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      string `json:"schema"` // raw JSON schema
}

// ToolCall is a tool invocation decided by the model.
type ToolCall struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ArgumentsJSON string `json:"arguments_json"`
	ParseError    bool   `json:"parse_error,omitempty"`
}

// StreamChunk is a single SSE chunk with timing.
type StreamChunk struct {
	Index     int           `json:"index"`
	Content   string        `json:"content"`
	ArrivedAt time.Time     `json:"arrived_at"`
	Gap       time.Duration `json:"gap,omitempty"` // gap from previous chunk
	IsStall   bool          `json:"is_stall,omitempty"`
}

// StreamStats holds aggregate streaming metrics.
type StreamStats struct {
	ChunkCount       int           `json:"chunk_count"`
	TTFT             time.Duration `json:"ttft"`              // time to first token
	StreamDuration   time.Duration `json:"stream_duration"`
	ThroughputTPS    float64       `json:"throughput_tps"`    // tokens/sec
	AvgTokensPerChunk float64      `json:"avg_tokens_per_chunk"`
	StallCount       int           `json:"stall_count"`
	StallThreshold   time.Duration `json:"stall_threshold"`
	Interrupted      bool          `json:"interrupted,omitempty"`
}

// RequestStatus represents the lifecycle state of a captured request.
type RequestStatus string

const (
	StatusPending   RequestStatus = "pending"
	StatusStreaming  RequestStatus = "streaming"
	StatusDone      RequestStatus = "done"
	StatusError     RequestStatus = "error"
)

// Request is the core model for a captured LLM API call.
type Request struct {
	ID        string    `json:"id"`
	Seq       int       `json:"seq"` // display number (#1, #2, ...)
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at,omitempty"`

	// Connection
	Method string `json:"method"`
	URL    string `json:"url"`
	Path   string `json:"path"`

	// Provider
	Provider   ProviderName `json:"provider"`
	ProviderHost string     `json:"provider_host"`

	// Model
	Model string `json:"model"`

	// Parsed request fields
	Messages        []Message        `json:"messages,omitempty"`
	SystemPrompt    string           `json:"system_prompt,omitempty"`
	Tools           []ToolDefinition `json:"tools,omitempty"`
	Temperature     *float64         `json:"temperature,omitempty"`
	MaxTokens       *int             `json:"max_tokens,omitempty"`
	Stream          bool             `json:"stream"`

	// Parsed response fields
	ResponseContent string       `json:"response_content,omitempty"`
	ToolCalls       []ToolCall   `json:"tool_calls,omitempty"`
	FinishReason    FinishReason `json:"finish_reason,omitempty"`

	// Tokens
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`

	// Cost
	InputCost  float64 `json:"input_cost"`
	OutputCost float64 `json:"output_cost"`
	TotalCost  float64 `json:"total_cost"`
	PricingKnown bool  `json:"pricing_known"`

	// Timing
	Latency time.Duration `json:"latency"`
	TTFT    time.Duration `json:"ttft,omitempty"`

	// Streaming
	Chunks      []StreamChunk `json:"chunks,omitempty"`
	StreamStats *StreamStats  `json:"stream_stats,omitempty"`

	// HTTP
	RequestHeaders  map[string]string `json:"request_headers,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
	StatusCode      int               `json:"status_code"`

	// Raw bodies (for Raw tab / export)
	RequestBody  []byte `json:"-"`
	ResponseBody []byte `json:"-"`

	// Error
	ErrorMessage string `json:"error_message,omitempty"`

	// Status
	Status   RequestStatus `json:"status"`
	ReplayOf string        `json:"replay_of,omitempty"` // ID of original if this is a replay

	// Rate limit info
	RateLimitRemaining int       `json:"rate_limit_remaining,omitempty"`
	RateLimitReset     time.Time `json:"rate_limit_reset,omitempty"`

	// Conversation grouping
	ConversationID string `json:"conversation_id,omitempty"`
}

// SessionStats holds aggregate stats for the current probe session.
type SessionStats struct {
	RequestCount int
	ErrorCount   int
	TotalCost    float64
	TotalLatency time.Duration
	TotalTTFT    time.Duration
	TTFTCount    int
}

// AvgLatency returns the mean latency across all completed requests.
func (s SessionStats) AvgLatency() time.Duration {
	if s.RequestCount == 0 {
		return 0
	}
	return s.TotalLatency / time.Duration(s.RequestCount)
}

// AvgTTFT returns the mean TTFT across all streaming requests.
func (s SessionStats) AvgTTFT() time.Duration {
	if s.TTFTCount == 0 {
		return 0
	}
	return s.TotalTTFT / time.Duration(s.TTFTCount)
}
