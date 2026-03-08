package cost

import (
	"github.com/MuhammadHananAsghar/probe/internal/store"
)

const tokensPerMillion = 1_000_000.0

// Calculate sets req.InputCost, req.OutputCost, req.TotalCost, and
// req.PricingKnown based on the request's token counts and the pricing
// database. If the model is unknown, PricingKnown is set to false and all
// cost fields are left at zero.
func Calculate(db *DB, req *store.Request) {
	pricing, ok := db.Lookup(req.Model)
	if !ok {
		req.PricingKnown = false
		return
	}

	req.PricingKnown = true
	req.InputCost = float64(req.InputTokens) / tokensPerMillion * pricing.InputPer1M
	req.OutputCost = float64(req.OutputTokens) / tokensPerMillion * pricing.OutputPer1M
	req.TotalCost = req.InputCost + req.OutputCost
}
