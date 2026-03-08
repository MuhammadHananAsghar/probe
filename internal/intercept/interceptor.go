package intercept

import (
	"net/http"
	"strings"

	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// sensitiveHeaders is the set of header names (lowercased) that contain credentials.
var sensitiveHeaders = map[string]bool{
	"authorization":   true,
	"x-api-key":       true,
	"anthropic-api-key": true,
}

// ParsedRequest holds extracted metadata from an HTTP request before forwarding.
type ParsedRequest struct {
	// Provider identifies which LLM provider this request targets.
	Provider store.ProviderName
	// Host is the target provider hostname.
	Host string
	// Path is the request URL path.
	Path string
	// Body is the raw request body bytes.
	Body []byte
	// Headers is a flat map of header name → value, with sensitive values masked.
	Headers map[string]string
	// IsStreaming indicates the request asks for a streaming response.
	IsStreaming bool
}

// IsSensitiveHeader returns true for headers that contain API credentials.
func IsSensitiveHeader(key string) bool {
	return sensitiveHeaders[strings.ToLower(key)]
}

// MaskHeaderValue masks an API key value for safe display.
// It keeps the first 8 characters and the last 4 characters, joining them with "...".
// If the value is too short to mask, the whole value is replaced with "***".
func MaskHeaderValue(key, value string) string {
	if !IsSensitiveHeader(key) {
		return value
	}

	// For Bearer tokens, operate on the token part only.
	prefix := ""
	v := value
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		prefix = value[:7] // preserve "Bearer "
		v = value[7:]
	}

	const minLen = 13 // 8 head + 3 dots + at least some tail
	if len(v) < minLen {
		return prefix + "***"
	}

	return prefix + v[:8] + "..." + v[len(v)-4:]
}

// ExtractHeaders copies relevant HTTP headers into a flat map, masking
// sensitive keys (Authorization, x-api-key, anthropic-api-key).
// Each header is stored as a single string (multiple values joined with ", ").
func ExtractHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for name, values := range h {
		joined := strings.Join(values, ", ")
		if IsSensitiveHeader(name) {
			joined = MaskHeaderValue(name, joined)
		}
		out[name] = joined
	}
	return out
}
