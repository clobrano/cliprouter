package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginBottom(1)

	clipboardStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginBottom(1)

	instructionsStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			MarginBottom(1)
)

// View renders the TUI (Bubble Tea interface)
func (m Model) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("┌─ Clipboard Router ─────────────────────────────┐"))
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

	return b.String()
}
