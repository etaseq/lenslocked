package templates

import "embed"

// NOTE: This embeds ALL files in the current folder into the binary.

//go:embed *
var FS embed.FS
