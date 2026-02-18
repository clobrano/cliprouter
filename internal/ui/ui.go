package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Catppuccin Mocha color palette
var (
	catppuccinMauve    = lipgloss.Color("#cba6f7") // Purple - for titles
	catppuccinPink     = lipgloss.Color("#f5c2e7") // Pink - for selected items
	catppuccinBlue     = lipgloss.Color("#89b4fa") // Blue - for accents
	catppuccinText     = lipgloss.Color("#cdd6f4") // Main text
	catppuccinSubtext0 = lipgloss.Color("#a6adc8") // Subtle text
)

var (
	catppuccinYellow = lipgloss.Color("#f9e2af") // Yellow - for warnings
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(catppuccinMauve).
			MarginBottom(1)

	clipboardStyle = lipgloss.NewStyle().
			Foreground(catppuccinText).
			MarginBottom(1)

	warningStyle = lipgloss.NewStyle().
			Foreground(catppuccinYellow).
			Bold(true).
			MarginBottom(1)

	instructionsStyle = lipgloss.NewStyle().
				Foreground(catppuccinSubtext0).
				MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(catppuccinPink).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(catppuccinText)

	footerStyle = lipgloss.NewStyle().
			Foreground(catppuccinSubtext0).
			MarginTop(1)

	searchStyle = lipgloss.NewStyle().
			Foreground(catppuccinBlue).
			MarginBottom(1)
)

// View renders the TUI (Bubble Tea interface)
func (m Model) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Clipboard Router"))
	b.WriteString("\n")

	// Clipboard excerpt or empty warning
	if m.clipboardEmpty {
		b.WriteString(warningStyle.Render("Warning: Clipboard is empty"))
		b.WriteString("\n")
	} else {
		clipText := fmt.Sprintf("Clipboard: \"%s\"", m.clipboardExcerpt)
		b.WriteString(clipboardStyle.Render(clipText))
		b.WriteString("\n")
	}

	// Instructions
	instructions := "Select a script (↑↓ to navigate, type to search):"
	b.WriteString(instructionsStyle.Render(instructions))
	b.WriteString("\n")

	// Search query (if active)
	if m.searchQuery != "" {
		searchText := fmt.Sprintf("Search: %s", m.searchQuery)
		b.WriteString(searchStyle.Render(searchText))
		b.WriteString("\n")
	}

	// Script list
	if len(m.filteredScripts) == 0 {
		b.WriteString(normalStyle.Render("No matching scripts found."))
		b.WriteString("\n")
	} else {
		for i, script := range m.filteredScripts {
			prefix := "  "
			if i == m.selectedIndex {
				prefix = "> "
				b.WriteString(selectedStyle.Render(prefix + script.Name))
			} else {
				b.WriteString(normalStyle.Render(prefix + script.Name))
			}
			b.WriteString("\n")
		}
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("[Press ESC to quit]"))

	return b.String()
}
