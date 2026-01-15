package static

import "embed"

//go:embed ../../cmd/server/static
var WebFS embed.FS
