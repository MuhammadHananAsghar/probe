// Package tui provides the terminal user interface for probe.
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	colorGreen  = lipgloss.Color("#00D26A")
	colorYellow = lipgloss.Color("#FFB300")
	colorRed    = lipgloss.Color("#FF4444")
	colorBlue   = lipgloss.Color("#4A9EFF")
	colorGray   = lipgloss.Color("#6C7A8A")
	colorWhite  = lipgloss.Color("#E8EBF0")
	colorDim    = lipgloss.Color("#4A5568")

	// Provider badge colors
	providerColors = map[string]lipgloss.Color{
		"openai":    lipgloss.Color("#74AA9C"),
		"anthropic": lipgloss.Color("#D4A27F"),
		"google":    lipgloss.Color("#4285F4"),
		"mistral":   lipgloss.Color("#FF6B35"),
		"groq":      lipgloss.Color("#F55036"),
		"ollama":    lipgloss.Color("#8B5CF6"),
	}

	// Styles
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(colorWhite)
	dimStyle      = lipgloss.NewStyle().Foreground(colorGray)
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#1E3A5F")).Foreground(colorWhite)
	statsBarStyle = lipgloss.NewStyle().Background(lipgloss.Color("#0D1117")).Foreground(colorGray).Padding(0, 1)
	errorStyle    = lipgloss.NewStyle().Foreground(colorRed)
	successStyle  = lipgloss.NewStyle().Foreground(colorGreen)
	warningStyle  = lipgloss.NewStyle().Foreground(colorYellow)
	costStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	modelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#A78BFA"))
	borderStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorDim).Padding(0, 1)
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Padding(0, 1)
)

// StatusColor returns the appropriate lipgloss color for an HTTP status code.
func StatusColor(code int) lipgloss.Color {
	switch {
	case code == 0:
		return colorGray
	case code >= 200 && code < 300:
		return colorGreen
	case code >= 400 && code < 500:
		return colorYellow
	case code >= 500:
		return colorRed
	default:
		return colorGray
	}
}

// ProviderColor returns the display color associated with a provider name.
// Falls back to colorGray for unknown providers.
func ProviderColor(p string) lipgloss.Color {
	if c, ok := providerColors[strings.ToLower(p)]; ok {
		return c
	}
	return colorGray
}

// FormatCost formats a float64 cost value for display.
// Examples: "$0.0089", "<$0.0001", "$1.23"
func FormatCost(cost float64, known bool) string {
	if !known {
		return dimStyle.Render("n/a")
	}
	if cost == 0 {
		return costStyle.Render("$0.00")
	}
	if cost < 0.0001 {
		return costStyle.Render("<$0.0001")
	}
	if cost < 0.01 {
		return costStyle.Render(fmt.Sprintf("$%.4f", cost))
	}
	return costStyle.Render(fmt.Sprintf("$%.4f", cost))
}

// FormatDuration formats a time.Duration for concise display.
// Examples: "1.2s", "420ms", "12µs"
func FormatDuration(d time.Duration) string {
	switch {
	case d == 0:
		return dimStyle.Render("—")
	case d >= time.Second:
		return fmt.Sprintf("%.1fs", d.Seconds())
	case d >= time.Millisecond:
		return fmt.Sprintf("%dms", d.Milliseconds())
	default:
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
}

// FormatTokens formats an integer token count with comma separators.
// Example: 1247 → "1,247"
func FormatTokens(n int) string {
	if n == 0 {
		return "0"
	}
	s := fmt.Sprintf("%d", n)
	// Insert commas every three digits from the right.
	var result []byte
	for i, c := range s {
		pos := len(s) - i
		if i > 0 && pos%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

// ProviderBadge renders a colored provider name badge using lipgloss styles.
func ProviderBadge(p string) string {
	color := ProviderColor(p)
	return lipgloss.NewStyle().Foreground(color).Bold(true).Render(p)
}
