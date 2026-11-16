// Package dbmigrations embeds the SQL migration files so the server can apply them at startup.
package dbmigrations

import "embed"

// FS holds db/migrations/*.sql.
//
//go:embed migrations/*.sql
var FS embed.FS
