package ui

import "github.com/japacx/aisand/internal/embed"

// getOpenCodeScript returns the embedded opencode.sh script bytes.
func getOpenCodeScript() []byte {
	return embed.OpenCodeScript
}
