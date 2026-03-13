package store

import (
	"database/sql"
	"fmt"
)

const currentVersion = 1

func migrate(db *sql.DB) error {
	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("reading user_version: %w", err)
	}

	if version >= currentVersion {
		return nil
	}

	if version < 1 {
		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS sessions (
				id           TEXT PRIMARY KEY,
				name         TEXT NOT NULL,
				tool         TEXT NOT NULL,
				worktree     TEXT,
				directory    TEXT NOT NULL,
				prompt       TEXT,
				plan_file    TEXT,
				tmux_session TEXT NOT NULL,
				status       TEXT DEFAULT 'running',
				created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
				stopped_at   DATETIME
			);
		`); err != nil {
			return fmt.Errorf("creating sessions table: %w", err)
		}
	}

	if _, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", currentVersion)); err != nil {
		return fmt.Errorf("setting user_version: %w", err)
	}

	return nil
}
