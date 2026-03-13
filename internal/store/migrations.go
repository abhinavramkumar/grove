package store

import (
	"database/sql"
	"fmt"
)

const currentVersion = 3

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
				id              TEXT PRIMARY KEY,
				name            TEXT NOT NULL,
				tool            TEXT NOT NULL,
				worktree        TEXT,
				directory       TEXT NOT NULL,
				repo_root       TEXT,
				prompt          TEXT,
				plan_file       TEXT,
				tmux_session    TEXT NOT NULL,
				tool_session_id TEXT,
				status          TEXT DEFAULT 'running',
				created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
				stopped_at      DATETIME
			);
		`); err != nil {
			return fmt.Errorf("creating sessions table: %w", err)
		}
	}

	if version >= 1 && version < 2 {
		// Existing v1 databases: add the tool_session_id column.
		if _, err := db.Exec(`ALTER TABLE sessions ADD COLUMN tool_session_id TEXT`); err != nil {
			return fmt.Errorf("adding tool_session_id column: %w", err)
		}
	}

	if version >= 2 && version < 3 {
		if _, err := db.Exec(`ALTER TABLE sessions ADD COLUMN repo_root TEXT`); err != nil {
			return fmt.Errorf("adding repo_root column: %w", err)
		}
	}

	if _, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", currentVersion)); err != nil {
		return fmt.Errorf("setting user_version: %w", err)
	}

	return nil
}
