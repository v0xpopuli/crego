package templates

import (
	"embed"
)

//go:embed project/*.tmpl
var FS embed.FS
