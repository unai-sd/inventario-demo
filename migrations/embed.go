// Package migrations embeds the SQL migration files so the binaries carry them.
package migrations

import "embed"

// FS contains every NNNN_name.sql file of this directory.
//
//go:embed *.sql
var FS embed.FS
