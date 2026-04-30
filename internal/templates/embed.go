package templates

import (
	"embed"
)

//go:embed project/*.tmpl web/*.tmpl
var FS embed.FS
