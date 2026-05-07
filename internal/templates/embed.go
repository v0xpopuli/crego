package templates

import (
	"embed"
)

//go:embed cli/*.tmpl project/*.tmpl web/*.tmpl
var FS embed.FS
