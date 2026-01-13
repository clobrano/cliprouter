package ui

import (
	"github.com/clobrano/cliprouter/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the Bubble Tea application model
type Model struct {
	scripts          []config.Script // All available scripts
	filteredScripts  []config.Script // Filtered scripts based on search
	selectedIndex    int             // Currently selected script index
	searchQuery      string          // Current search query
	clipboardExcerpt string          // Clipboard content excerpt
	selectedScript   *config.Script  // Script chosen by user (when Enter pressed)
	quitting         bool            // Whether user is quitting
	width            int             // Terminal width
	height           int             // Terminal height
}

// NewModel creates a new TUI model
func NewModel(scripts []config.Script, clipboardExcerpt string) Model {
	return Model{
		scripts:          scripts,
		filteredScripts:  scripts,
		selectedIndex:    0,
		searchQuery:      "",
		clipboardExcerpt: clipboardExcerpt,
		selectedScript:   nil,
		quitting:         false,
	}
}

// Init initializes the model (Bubble Tea interface)
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model (Bubble Tea interface)
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Capture terminal dimensions for centering
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			// User wants to quit
			m.quitting = true
			return m, tea.Quit

		case "enter":
			// User selected a script
			if len(m.filteredScripts) > 0 && m.selectedIndex < len(m.filteredScripts) {
				m.selectedScript = &m.filteredScripts[m.selectedIndex]
				return m, tea.Quit
			}

		case "up", "k":
			// Move selection up
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case "down", "j":
			// Move selection down
			if m.selectedIndex < len(m.filteredScripts)-1 {
				m.selectedIndex++
			}

		case "backspace":
			// Remove last character from search
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.updateFilter()
			}

		default:
			// Handle regular character input for search
			if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
				m.searchQuery += msg.String()
				m.updateFilter()
			}
		}
	}

	return m, nil
}

// updateFilter applies the search query to filter scripts
func (m *Model) updateFilter() {
	m.filteredScripts = FilterScripts(m.scripts, m.searchQuery)
	// Reset selection to first item
	if m.selectedIndex >= len(m.filteredScripts) {
		m.selectedIndex = 0
	}
	if len(m.filteredScripts) > 0 && m.selectedIndex >= len(m.filteredScripts) {
		m.selectedIndex = len(m.filteredScripts) - 1
	}
}

// GetSelectedScript returns the script selected by the user (if any)
func (m Model) GetSelectedScript() *config.Script {
	return m.selectedScript
}

// IsQuitting returns true if the user quit without selecting
func (m Model) IsQuitting() bool {
	return m.quitting
}
