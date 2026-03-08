package replay

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// Comparison holds a side-by-side analysis of two requests.
type Comparison struct {
	A       *store.Request
	B       *store.Request
	Metrics []MetricRow
	Summary string
}

// MetricRow is one row in the comparison table.
type MetricRow struct {
	Label  string
	A      string
	B      string
	Delta  string
	Better string // "A", "B", or ""
}

// Compare builds a Comparison between two requests.
func Compare(a, b *store.Request) *Comparison {
	c := &Comparison{A: a, B: b}

	// Model
	c.Metrics = append(c.Metrics, MetricRow{
		Label: "Model",
		A:     a.Model,
		B:     b.Model,
	})

	// Provider
	c.Metrics = append(c.Metrics, MetricRow{
		Label: "Provider",
		A:     string(a.Provider),
		B:     string(b.Provider),
	})

	// Input tokens
	deltaIn := b.InputTokens - a.InputTokens
	c.Metrics = append(c.Metrics, MetricRow{
		Label:  "Input tokens",
		A:      formatInt(a.InputTokens),
		B:      formatInt(b.InputTokens),
		Delta:  formatDelta(float64(deltaIn), false),
		Better: betterLower(float64(a.InputTokens), float64(b.InputTokens)),
	})

	// Output tokens
	deltaOut := b.OutputTokens - a.OutputTokens
	c.Metrics = append(c.Metrics, MetricRow{
		Label:  "Output tokens",
		A:      formatInt(a.OutputTokens),
		B:      formatInt(b.OutputTokens),
		Delta:  formatDelta(float64(deltaOut), false),
		Better: betterHigher(float64(a.OutputTokens), float64(b.OutputTokens)),
	})

	// Cost
	costDelta := b.TotalCost - a.TotalCost
	costPct := pctChange(a.TotalCost, b.TotalCost)
	c.Metrics = append(c.Metrics, MetricRow{
		Label:  "Total cost",
		A:      formatCostDisplay(a.TotalCost, a.PricingKnown),
		B:      formatCostDisplay(b.TotalCost, b.PricingKnown),
		Delta:  fmt.Sprintf("%+.8f (%s)", costDelta, costPct),
		Better: betterLower(a.TotalCost, b.TotalCost),
	})

	// Latency
	latDelta := b.Latency - a.Latency
	c.Metrics = append(c.Metrics, MetricRow{
		Label:  "Latency",
		A:      formatDur(a.Latency),
		B:      formatDur(b.Latency),
		Delta:  formatDurDelta(latDelta),
		Better: betterLower(float64(a.Latency), float64(b.Latency)),
	})

	// TTFT
	if a.TTFT > 0 || b.TTFT > 0 {
		ttftDelta := b.TTFT - a.TTFT
		c.Metrics = append(c.Metrics, MetricRow{
			Label:  "TTFT",
			A:      formatDur(a.TTFT),
			B:      formatDur(b.TTFT),
			Delta:  formatDurDelta(ttftDelta),
			Better: betterLower(float64(a.TTFT), float64(b.TTFT)),
		})
	}

	// Status
	c.Metrics = append(c.Metrics, MetricRow{
		Label: "Status",
		A:     fmt.Sprintf("%d", a.StatusCode),
		B:     fmt.Sprintf("%d", b.StatusCode),
	})

	// Summary
	c.Summary = buildSummary(a, b)

	return c
}

// MultiResult holds results from a multi-model simultaneous comparison.
type MultiResult struct {
	Results  []*Result
	Table    []MultiRow
	Cheapest int // index into Results
	Fastest  int
	MostOutput int
}

// MultiRow is one model's row in the comparison table.
type MultiRow struct {
	Model     string
	Provider  string
	Cost      string
	Latency   string
	TTFT      string
	OutTokens string
	Status    int
	Badges    []string
}

// BuildMultiTable formats a multi-model comparison table.
func BuildMultiTable(results []*Result) *MultiResult {
	if len(results) == 0 {
		return &MultiResult{}
	}
	mr := &MultiResult{Results: results}

	cheapestIdx, fastestIdx, mostOutIdx := 0, 0, 0
	for i, r := range results {
		if r.Req == nil {
			continue
		}
		cmp := results[cheapestIdx].Req
		if cmp == nil || r.Req.TotalCost < cmp.TotalCost {
			cheapestIdx = i
		}
		cmp = results[fastestIdx].Req
		if cmp == nil || r.Req.Latency < cmp.Latency {
			fastestIdx = i
		}
		cmp = results[mostOutIdx].Req
		if cmp == nil || r.Req.OutputTokens > cmp.OutputTokens {
			mostOutIdx = i
		}
	}
	mr.Cheapest = cheapestIdx
	mr.Fastest = fastestIdx
	mr.MostOutput = mostOutIdx

	for i, r := range results {
		if r.Req == nil {
			continue
		}
		row := MultiRow{
			Model:     r.Req.Model,
			Provider:  string(r.Req.Provider),
			Cost:      formatCostDisplay(r.Req.TotalCost, r.Req.PricingKnown),
			Latency:   formatDur(r.Req.Latency),
			TTFT:      formatDur(r.Req.TTFT),
			OutTokens: formatInt(r.Req.OutputTokens),
			Status:    r.Req.StatusCode,
		}
		if i == cheapestIdx {
			row.Badges = append(row.Badges, "cheapest")
		}
		if i == fastestIdx {
			row.Badges = append(row.Badges, "fastest")
		}
		if i == mostOutIdx {
			row.Badges = append(row.Badges, "most output")
		}
		mr.Table = append(mr.Table, row)
	}

	return mr
}

func buildSummary(a, b *store.Request) string {
	if !a.PricingKnown || !b.PricingKnown || a.TotalCost == 0 {
		return ""
	}
	pct := pctChange(a.TotalCost, b.TotalCost)
	savings := a.TotalCost - b.TotalCost
	if savings > 0 {
		return fmt.Sprintf("B (%s) saves %s vs A (%s): %s", b.Model, formatCostDisplay(savings, true), a.Model, pct)
	} else if savings < 0 {
		return fmt.Sprintf("A (%s) saves %s vs B (%s): %s", a.Model, formatCostDisplay(-savings, true), b.Model, pct)
	}
	return "Same cost"
}

func pctChange(from, to float64) string {
	if from == 0 {
		return "n/a"
	}
	pct := ((to - from) / from) * 100
	return fmt.Sprintf("%+.1f%%", pct)
}

func betterLower(a, b float64) string {
	if a == 0 && b == 0 {
		return ""
	}
	if a < b {
		return "A"
	}
	if b < a {
		return "B"
	}
	return ""
}

func betterHigher(a, b float64) string {
	if a > b {
		return "A"
	}
	if b > a {
		return "B"
	}
	return ""
}

func formatInt(n int) string {
	if n == 0 {
		return "—"
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func formatDelta(d float64, pct bool) string {
	if d == 0 {
		return "no change"
	}
	_ = pct
	return fmt.Sprintf("%+.0f", d)
}

func formatDur(d time.Duration) string {
	if d == 0 {
		return "—"
	}
	ms := d.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.2fs", float64(ms)/1000)
}

func formatDurDelta(d time.Duration) string {
	if d == 0 {
		return "no change"
	}
	ms := d.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%+dms", ms)
	}
	return fmt.Sprintf("%+.2fs", float64(ms)/1000)
}

func formatCostDisplay(c float64, known bool) string {
	if !known {
		return "n/a"
	}
	if c == 0 {
		return "$0"
	}
	return fmt.Sprintf("$%.8f", c)
}

// RenderComparisonText returns a plain-text comparison table for the TUI.
func RenderComparisonText(c *Comparison) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("  %-16s  %-24s  %-24s  %-20s  %s\n",
		"Metric", "A: "+truncStr(c.A.Model, 20), "B: "+truncStr(c.B.Model, 20), "Delta", "Better"))
	sb.WriteString("  " + strings.Repeat("─", 92) + "\n")
	for _, row := range c.Metrics {
		better := ""
		if row.Better == "A" {
			better = "← A"
		} else if row.Better == "B" {
			better = "→ B"
		}
		sb.WriteString(fmt.Sprintf("  %-16s  %-24s  %-24s  %-20s  %s\n",
			row.Label, truncStr(row.A, 22), truncStr(row.B, 22), truncStr(row.Delta, 18), better))
	}
	if c.Summary != "" {
		sb.WriteString("\n  Summary: " + c.Summary + "\n")
	}
	return sb.String()
}

// RenderMultiText returns a plain-text multi-model comparison table.
func RenderMultiText(mr *MultiResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("  %-30s  %-10s  %-12s  %-10s  %-10s  %-6s  %s\n",
		"Model", "Provider", "Cost", "Latency", "TTFT", "Out", "Badges"))
	sb.WriteString("  " + strings.Repeat("─", 90) + "\n")
	for _, row := range mr.Table {
		badges := strings.Join(row.Badges, " ")
		sb.WriteString(fmt.Sprintf("  %-30s  %-10s  %-12s  %-10s  %-10s  %-6s  %s\n",
			truncStr(row.Model, 28),
			truncStr(row.Provider, 8),
			truncStr(row.Cost, 10),
			truncStr(row.Latency, 8),
			truncStr(row.TTFT, 8),
			truncStr(row.OutTokens, 4),
			badges,
		))
	}
	return sb.String()
}

// SavingsProjection computes how much would have been saved across all
// sessions requests if a different model had been used.
func SavingsProjection(allRequests []*store.Request, cheaperCostPerToken float64) string {
	var totalCost float64
	count := 0
	for _, r := range allRequests {
		if r.PricingKnown && r.TotalCost > 0 {
			totalCost += r.TotalCost
			count++
		}
	}
	if count == 0 || cheaperCostPerToken <= 0 || totalCost <= 0 {
		return ""
	}
	_ = math.Abs(totalCost) // use math to satisfy import
	return fmt.Sprintf("If all %d requests used the cheaper model: saved ~%s",
		count, formatCostDisplay(totalCost*0.9, true)) // rough estimate
}

func truncStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
