package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// renderStatsBar renders the bottom stats bar given current session statistics.
// width is the terminal width used to right-align the address info.
func renderStatsBar(stats store.SessionStats, proxyAddr, dashAddr string, width int) string {
	// Left side: session stats
	reqPart := fmt.Sprintf("Session: %d requests", stats.RequestCount)

	costPart := costStyle.Render(fmt.Sprintf("$%.4f total", stats.TotalCost))

	var ttftPart string
	if avg := stats.AvgTTFT(); avg > 0 {
		ttftPart = fmt.Sprintf("Avg TTFT: %s", FormatDuration(avg))
	} else {
		ttftPart = dimStyle.Render("Avg TTFT: —")
	}

	var errPart string
	if stats.ErrorCount > 0 {
		errPart = errorStyle.Render(fmt.Sprintf("%d error", stats.ErrorCount))
		if stats.ErrorCount > 1 {
			errPart = errorStyle.Render(fmt.Sprintf("%d errors", stats.ErrorCount))
		}
	}

	leftParts := []string{reqPart, costPart, ttftPart}
	if errPart != "" {
		leftParts = append(leftParts, errPart)
	}
	left := strings.Join(leftParts, dimStyle.Render(" | "))

	// Right side: addresses
	right := ""
	if proxyAddr != "" || dashAddr != "" {
		parts := []string{}
		if proxyAddr != "" {
			parts = append(parts, fmt.Sprintf("Proxy: %s", proxyAddr))
		}
		if dashAddr != "" {
			parts = append(parts, fmt.Sprintf("Dashboard: %s", dashAddr))
		}
		right = dimStyle.Render(strings.Join(parts, "  "))
	}

	// Compute padding to right-align
	leftRaw := lipgloss.Width(left)
	rightRaw := lipgloss.Width(right)
	gap := width - leftRaw - rightRaw - 2 // -2 for the statsBarStyle padding(0,1)
	if gap < 1 {
		gap = 1
	}
	line := left + strings.Repeat(" ", gap) + right

	return statsBarStyle.Width(width).Render(line)
}
