package puzzles

import "strings"

func normalizeTheme(theme string) string {
	return strings.ToLower(strings.TrimSpace(theme))
}
