// Package migrate applies the numbered SQL files of the migrations directory
// in order, recording each applied version in schema_migrations.
package migrate

import (
	"context"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Migration is one SQL file.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

var fileName = regexp.MustCompile(`^(\d{4})_([a-z0-9_]+)\.sql$`)

// Load reads and sorts the migrations of a directory. Every .sql file must be
// named NNNN_description.sql and versions must be unique.
func Load(fsys fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}
	seen := map[int]string{}
	var out []Migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		m := fileName.FindStringSubmatch(e.Name())
		if m == nil {
			return nil, fmt.Errorf("invalid migration file name %q", e.Name())
		}
		version, _ := strconv.Atoi(m[1])
		if other, dup := seen[version]; dup {
			return nil, fmt.Errorf("duplicate migration version %d: %s and %s", version, other, e.Name())
		}
		seen[version] = e.Name()
		body, err := fs.ReadFile(fsys, e.Name())
		if err != nil {
			return nil, err
		}
		out = append(out, Migration{Version: version, Name: m[2], SQL: string(body)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}

// DB is the subset of pgx used by Up; *pgx.Conn and *pgxpool.Pool satisfy it.
type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Up applies the pending migrations and returns the names of those applied.
func Up(ctx context.Context, db DB, migrations []Migration) ([]string, error) {
	if _, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		name       TEXT NOT NULL,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		return nil, fmt.Errorf("create schema_migrations: %w", err)
	}
	rows, err := db.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	applied, err := pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil {
		return nil, err
	}
	done := map[int]bool{}
	for _, v := range applied {
		done[v] = true
	}

	var names []string
	for _, m := range Pending(migrations, done) {
		tx, err := db.Begin(ctx)
		if err != nil {
			return names, err
		}
		if _, err := tx.Exec(ctx, m.SQL); err != nil {
			_ = tx.Rollback(ctx)
			return names, fmt.Errorf("migration %04d_%s: %w", m.Version, m.Name, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, m.Version, m.Name); err != nil {
			_ = tx.Rollback(ctx)
			return names, err
		}
		if err := tx.Commit(ctx); err != nil {
			return names, err
		}
		names = append(names, fmt.Sprintf("%04d_%s", m.Version, m.Name))
	}
	return names, nil
}

// Pending returns the migrations whose version is not in applied, in order.
func Pending(migrations []Migration, applied map[int]bool) []Migration {
	var out []Migration
	for _, m := range migrations {
		if !applied[m.Version] {
			out = append(out, m)
		}
	}
	return out
}
