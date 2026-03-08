package cost

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	liteLLMURL      = "https://raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json"
	pricingCacheFile = "pricing_cache.json"
)

// liteLLMEntry mirrors a single model entry in LiteLLM's pricing JSON.
// LiteLLM stores costs per token; we convert to per-1M for internal use.
type liteLLMEntry struct {
	InputCostPerToken  float64 `json:"input_cost_per_token"`
	OutputCostPerToken float64 `json:"output_cost_per_token"`
	MaxInputTokens     int     `json:"max_input_tokens"`
	MaxOutputTokens    int     `json:"max_output_tokens"`
	Provider           string  `json:"litellm_provider"`
	Mode               string  `json:"mode"`
}

// FetchLiteLLM fetches the latest model pricing from LiteLLM's GitHub repo.
// On success it saves a local cache at ~/.probe/pricing_cache.json.
// If the network request fails, it falls back to the cache if available.
// Returns a nil map (no error) when both fetch and cache are unavailable —
// the caller should then fall back to the embedded pricing data.
func FetchLiteLLM(ctx context.Context) (map[string]ModelPricing, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	data, fetchErr := fetchRaw(ctx)
	if fetchErr != nil {
		// Try local cache before giving up.
		cached, cacheErr := readCache()
		if cacheErr != nil {
			// Both failed — caller will use embedded data.
			return nil, nil //nolint:nilerr
		}
		data = cached
	}

	parsed, err := parseLiteLLM(data)
	if err != nil {
		return nil, fmt.Errorf("cost: parsing LiteLLM pricing: %w", err)
	}

	// Persist fresh data so future runs can use it offline.
	if fetchErr == nil {
		_ = writeCache(data) // best-effort; ignore write errors
	}

	return parsed, nil
}

// fetchRaw performs the HTTP GET from LiteLLM's GitHub URL.
func fetchRaw(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, liteLLMURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "probe-llm-debugger")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from LiteLLM pricing URL", resp.StatusCode)
	}

	return io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10 MB max
}

// parseLiteLLM converts raw LiteLLM JSON into our internal ModelPricing map.
// Models with zero input/output cost are skipped (e.g. free tier or incomplete entries).
func parseLiteLLM(data []byte) (map[string]ModelPricing, error) {
	var raw map[string]liteLLMEntry
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	out := make(map[string]ModelPricing, len(raw))
	for name, e := range raw {
		if e.InputCostPerToken == 0 && e.OutputCostPerToken == 0 {
			continue
		}
		// Skip non-chat modes that aren't useful for cost display.
		if e.Mode != "" && e.Mode != "chat" && e.Mode != "completion" && e.Mode != "text-completion-openai" {
			continue
		}
		out[strings.ToLower(name)] = ModelPricing{
			Provider:      e.Provider,
			InputPer1M:    e.InputCostPerToken * 1_000_000,
			OutputPer1M:   e.OutputCostPerToken * 1_000_000,
			ContextWindow: e.MaxInputTokens,
		}
	}
	return out, nil
}

// cacheDir returns the ~/.probe directory, creating it if needed.
func cacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".probe")
	return dir, os.MkdirAll(dir, 0o700)
}

func cachePath() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, pricingCacheFile), nil
}

func readCache() ([]byte, error) {
	p, err := cachePath()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(p)
}

func writeCache(data []byte) error {
	p, err := cachePath()
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}
