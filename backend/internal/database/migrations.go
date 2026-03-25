package database

import "embed"

//go:embed migrations/*.sql
var MigrationFS embed.FS

const MigrationDir = "migrations"
