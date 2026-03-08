package analyze

import (
	"strconv"
	"strings"
	"time"

	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// ParseRateLimitHeaders extracts rate limit info from response headers and
// populates req.RateLimitRemaining and req.RateLimitReset.
// Supports OpenAI and Anthropic header conventions.
func ParseRateLimitHeaders(req *store.Request) {
	if len(req.ResponseHeaders) == 0 {
		return
	}

	// Normalise header names to lowercase for matching.
	lower := make(map[string]string, len(req.ResponseHeaders))
	for k, v := range req.ResponseHeaders {
		lower[strings.ToLower(k)] = v
	}

	// OpenAI: x-ratelimit-remaining-requests, x-ratelimit-reset-requests
	if v, ok := lower["x-ratelimit-remaining-requests"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			req.RateLimitRemaining = n
		}
	}
	if v, ok := lower["x-ratelimit-reset-requests"]; ok {
		if t := parseResetTime(v); !t.IsZero() {
			req.RateLimitReset = t
		}
	}

	// Anthropic: anthropic-ratelimit-requests-remaining, anthropic-ratelimit-requests-reset
	if v, ok := lower["anthropic-ratelimit-requests-remaining"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			req.RateLimitRemaining = n
		}
	}
	if v, ok := lower["anthropic-ratelimit-requests-reset"]; ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			req.RateLimitReset = t
		}
	}

	// Token-level limits (use when request limit is not set).
	if req.RateLimitRemaining == 0 {
		if v, ok := lower["x-ratelimit-remaining-tokens"]; ok {
			if n, err := strconv.Atoi(v); err == nil {
				req.RateLimitRemaining = n
			}
		}
		if v, ok := lower["anthropic-ratelimit-tokens-remaining"]; ok {
			if n, err := strconv.Atoi(v); err == nil {
				req.RateLimitRemaining = n
			}
		}
	}
}

// parseResetTime parses OpenAI's reset time which is a relative duration like
// "1m30s" or an RFC3339 timestamp.
func parseResetTime(v string) time.Time {
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t
	}
	// OpenAI uses relative durations like "1m30s".
	if d, err := time.ParseDuration(v); err == nil {
		return time.Now().Add(d)
	}
	return time.Time{}
}
