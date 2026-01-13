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
	catppuccinSurface0 = lipgloss.Color("#313244") // Surface color for borders
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(catppuccinMauve).
			MarginBottom(1)

	clipboardStyle = lipgloss.NewStyle().
			Foreground(catppuccinText).
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

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(catppuccinMauve).
			Padding(1, 2)
)

// View renders the TUI (Bubble Tea interface)
func (m Model) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Clipboard Router"))
	b.WriteString("\n")

	// Clipboard excerpt
	clipText := fmt.Sprintf("Clipboard: \"%s\"", m.clipboardExcerpt)
	b.WriteString(clipboardStyle.Render(clipText))
	b.WriteString("\n")

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

	content := b.String()

	// Apply border with 2-character padding on all sides
	borderedContent := borderStyle.Render(content)

	// Center the content in the available terminal space
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			borderedContent,
		)
	}

	return borderedContent
}
