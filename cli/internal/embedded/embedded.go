// Package embedded ships a copy of the MobiAI catalog (plugins/) inside the
// binary so the CLI works out of the box without --catalog-root or env vars.
//
// The plugins/ directory is populated at build time by `go generate ./...`
// (see gen/main.go). It is gitignored.
package embedded

import "embed"

//go:generate go run ./gen/main.go

//go:embed all:plugins
var Plugins embed.FS
