package frontend

import "embed"

//go:embed all:templates all:static
var Assets embed.FS
