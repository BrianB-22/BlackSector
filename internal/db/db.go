package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	_ "modernc.org/sqlite"
)

// Database wraps the SQLite connection with proper configuration
type Database struct {
	conn   *sql.DB
	logger zerolog.Logger
}

// InitDatabase opens a SQLite connection and applies required PRAGMAs
func InitDatabase(dbPath string, logger zerolog.Logger) (*Database, error) {
	logger.Info().Str("db_path", dbPath).Msg("Initializing database")

	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open SQLite connection
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Apply required PRAGMAs for WAL mode and performance
	pragmas := []struct {
		name  string
		value string
	}{
		{"journal_mode", "WAL"},
		{"synchronous", "NORMAL"},
		{"foreign_keys", "ON"},
		{"busy_timeout", "5000"},
	}

	for _, pragma := range pragmas {
		query := fmt.Sprintf("PRAGMA %s = %s", pragma.name, pragma.value)
		if _, err := conn.Exec(query); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to set PRAGMA %s: %w", pragma.name, err)
		}
		logger.Debug().Str("pragma", pragma.name).Str("value", pragma.value).Msg("Applied PRAGMA")
	}

	db := &Database{
		conn:   conn,
		logger: logger,
	}

	// Apply migrations (both for new and existing databases)
	logger.Info().Msg("Checking for pending migrations")
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}
	logger.Info().Msg("Database migrations complete")

	return db, nil
}

// initSchema creates the schema_migrations table and applies all pending migrations
func (db *Database) initSchema() error {
	// Create schema_migrations table if it doesn't exist
	if err := db.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of applied migrations
	appliedMigrations, err := db.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Define migration files in order
	// Note: We skip 002_add_registration_fields.sql as it's a duplicate of 002_add_auth_columns.sql
	// with a UNIQUE constraint that would conflict. We use the non-unique version for flexibility.
	migrations := []string{
		"001_initial_schema.sql",
		"002_add_auth_columns.sql",
		"003_add_combat_instances.sql",
		"004_add_mission_current_objective.sql",
		"005_add_performance_indexes.sql",
	}

	// Apply each migration that hasn't been applied yet
	for _, migrationFile := range migrations {
		if appliedMigrations[migrationFile] {
			db.logger.Debug().Str("migration", migrationFile).Msg("Migration already applied, skipping")
			continue
		}

		if err := db.applyMigration(migrationFile); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migrationFile, err)
		}

		db.logger.Info().Str("migration", migrationFile).Msg("Migration applied successfully")
	}

	return nil
}

// createMigrationsTable creates the schema_migrations table for tracking applied migrations
func (db *Database) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at INTEGER NOT NULL
		)
	`
	if _, err := db.conn.Exec(query); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}
	return nil
}

// getAppliedMigrations returns a map of migration filenames that have been applied
func (db *Database) getAppliedMigrations() (map[string]bool, error) {
	rows, err := db.conn.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration rows: %w", err)
	}

	return applied, nil
}

// applyMigration reads and executes a migration file, then records it in schema_migrations
func (db *Database) applyMigration(migrationFile string) error {
	// Try multiple possible paths for the migration file
	possiblePaths := []string{
		filepath.Join("migrations", migrationFile),
		filepath.Join("../..", "migrations", migrationFile), // For tests running from internal/db
		filepath.Join("..", "..", "..", "migrations", migrationFile), // For tests running from internal/db subdirectories
	}

	var migrationSQL []byte
	var err error
	var usedPath string

	for _, migrationPath := range possiblePaths {
		migrationSQL, err = os.ReadFile(migrationPath)
		if err == nil {
			usedPath = migrationPath
			break
		}
	}

	if err != nil {
		return fmt.Errorf("failed to read migration file (tried %v): %w", possiblePaths, err)
	}

	// Begin transaction for migration
	tx, err := db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer db.RollbackTx(tx) // Rollback if we don't commit

	// Execute the migration SQL
	if _, err := tx.Exec(string(migrationSQL)); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record the migration as applied
	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
		migrationFile,
		db.getCurrentTimestamp(),
	)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit the transaction
	if err := db.CommitTx(tx); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	db.logger.Debug().Str("migration", usedPath).Msg("Migration file executed")
	return nil
}

// getCurrentTimestamp returns the current Unix timestamp
func (db *Database) getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// Close closes the database connection
func (db *Database) Close() error {
	db.logger.Info().Msg("Closing database connection")
	if err := db.conn.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}

// Conn returns the underlying sql.DB connection for query execution
func (db *Database) Conn() *sql.DB {
	return db.conn
}

// BeginTx starts a new database transaction
func (db *Database) BeginTx() (*sql.Tx, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	db.logger.Debug().Msg("Transaction started")
	return tx, nil
}

// CommitTx commits a database transaction
func (db *Database) CommitTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("cannot commit nil transaction")
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	db.logger.Debug().Msg("Transaction committed")
	return nil
}

// RollbackTx rolls back a database transaction
func (db *Database) RollbackTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("cannot rollback nil transaction")
	}
	if err := tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	db.logger.Debug().Msg("Transaction rolled back")
	return nil
}
