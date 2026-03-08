// Package cost provides LLM API pricing data and cost calculation utilities.
package cost

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MuhammadHananAsghar/probe/pkg/config"
)

//go:embed pricing_data.json
var pricingDataJSON []byte

// ModelPricing holds per-1M-token costs for a model.
type ModelPricing struct {
	Provider      string  `json:"provider"`
	InputPer1M    float64 `json:"input_cost_per_1m"`
	OutputPer1M   float64 `json:"output_cost_per_1m"`
	ContextWindow int     `json:"context_window"`
}

// pricingFile mirrors the top-level shape of pricing_data.json.
type pricingFile struct {
	UpdatedAt string                  `json:"updated_at"`
	Models    map[string]ModelPricing `json:"models"`
}

// DB is the pricing database. It holds built-in pricing data merged with any
// user-supplied custom overrides.
type DB struct {
	models map[string]ModelPricing // key: model name (lower-case)
	custom map[string]ModelPricing // user overrides (lower-case key)
}

// NewDB loads pricing from the embedded JSON file and merges custom overrides
// from the provided config map. If livePricing is non-nil (fetched from LiteLLM
// at startup), it replaces the embedded data as the base layer. Custom config
// entries always take the highest precedence.
func NewDB(customPricing map[string]config.CustomPricing, livePricing map[string]ModelPricing) (*DB, error) {
	var pf pricingFile
	if err := json.Unmarshal(pricingDataJSON, &pf); err != nil {
		return nil, fmt.Errorf("cost: parsing embedded pricing data: %w", err)
	}

	db := &DB{
		models: make(map[string]ModelPricing, len(pf.Models)),
		custom: make(map[string]ModelPricing, len(customPricing)),
	}

	// Start with embedded data as the base.
	for name, p := range pf.Models {
		db.models[strings.ToLower(name)] = p
	}

	// Overlay with live LiteLLM data if available — covers far more models
	// and stays up to date with provider price changes.
	for name, p := range livePricing {
		db.models[strings.ToLower(name)] = p
	}

	// User-defined custom overrides always win.
	for name, cp := range customPricing {
		db.custom[strings.ToLower(name)] = ModelPricing{
			InputPer1M:  cp.InputPer1M,
			OutputPer1M: cp.OutputPer1M,
		}
	}

	return db, nil
}

// Lookup returns pricing for a model name using the following strategy:
//  1. Exact match in custom overrides.
//  2. Exact match in built-in data.
//  3. Prefix match: any key that is a prefix of the model name
//     (e.g. "claude-3-5-sonnet" matches "claude-3-5-sonnet-20241022").
//  4. Suffix match: the model name is a prefix of a built-in key.
//
// Returns (pricing, true) on success, or (zero, false) when unknown.
func (db *DB) Lookup(model string) (ModelPricing, bool) {
	key := strings.ToLower(model)

	// 1. Exact custom override.
	if p, ok := db.custom[key]; ok {
		return p, true
	}

	// 2. Exact built-in match.
	if p, ok := db.models[key]; ok {
		return p, true
	}

	// 3. A known key is a prefix of the requested model name
	//    (e.g. "claude-3-5-sonnet" -> "claude-3-5-sonnet-20241022").
	for k, p := range db.models {
		if strings.HasPrefix(key, k) {
			return p, true
		}
	}

	// 4. The requested model name is a prefix of a known key.
	for k, p := range db.models {
		if strings.HasPrefix(k, key) {
			return p, true
		}
	}

	return ModelPricing{}, false
}
