package ui

import (
	"cliprouter/internal/config"

	"github.com/sahilm/fuzzy"
)

// FilterScripts filters scripts based on fuzzy search query
func FilterScripts(scripts []config.Script, query string) []config.Script {
	if query == "" {
		return scripts
	}

	// Create list of script names for fuzzy matching
	names := make([]string, len(scripts))
	for i, script := range scripts {
		names[i] = script.Name
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, names)

	// Return filtered scripts in match order
	filtered := make([]config.Script, len(matches))
	for i, match := range matches {
		filtered[i] = scripts[match.Index]
	}

	return filtered
}
