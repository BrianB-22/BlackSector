package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationSystem(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	t.Run("new database applies all migrations", func(t *testing.T) {
		// Create a temporary database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		// Initialize database
		db, err := InitDatabase(dbPath, logger)
		require.NoError(t, err)
		defer db.Close()

		// Verify schema_migrations table exists
		var count int
		err = db.conn.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 5, count, "Should have 5 migrations applied")

		// Verify all expected migrations are recorded
		expectedMigrations := []string{
			"001_initial_schema.sql",
			"002_add_auth_columns.sql",
			"003_add_combat_instances.sql",
			"004_add_mission_current_objective.sql",
			"005_add_performance_indexes.sql",
		}

		for _, migration := range expectedMigrations {
			var exists int
			err = db.conn.QueryRow(
				"SELECT COUNT(*) FROM schema_migrations WHERE version = ?",
				migration,
			).Scan(&exists)
			require.NoError(t, err)
			assert.Equal(t, 1, exists, "Migration %s should be recorded", migration)
		}

		// Verify key tables exist
		tables := []string{
			"players",
			"ships",
			"combat_instances",
			"mission_instances",
		}

		for _, table := range tables {
			var exists int
			err = db.conn.QueryRow(
				"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
				table,
			).Scan(&exists)
			require.NoError(t, err)
			assert.Equal(t, 1, exists, "Table %s should exist", table)
		}

		// Verify auth columns exist in players table
		var hasSSHUsername, hasPasswordHash int
		err = db.conn.QueryRow(`
			SELECT 
				COUNT(CASE WHEN name = 'ssh_username' THEN 1 END),
				COUNT(CASE WHEN name = 'password_hash' THEN 1 END)
			FROM pragma_table_info('players')
		`).Scan(&hasSSHUsername, &hasPasswordHash)
		require.NoError(t, err)
		assert.Equal(t, 1, hasSSHUsername, "ssh_username column should exist")
		assert.Equal(t, 1, hasPasswordHash, "password_hash column should exist")

		// Verify current_objective_index exists in mission_instances
		var hasCurrentObjective int
		err = db.conn.QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('mission_instances')
			WHERE name = 'current_objective_index'
		`).Scan(&hasCurrentObjective)
		require.NoError(t, err)
		assert.Equal(t, 1, hasCurrentObjective, "current_objective_index column should exist")
	})

	t.Run("existing database skips applied migrations", func(t *testing.T) {
		// Create a temporary database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		// Initialize database first time
		db1, err := InitDatabase(dbPath, logger)
		require.NoError(t, err)
		
		// Get initial migration count
		var initialCount int
		err = db1.conn.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&initialCount)
		require.NoError(t, err)
		db1.Close()

		// Open database again
		db2, err := InitDatabase(dbPath, logger)
		require.NoError(t, err)
		defer db2.Close()

		// Verify migration count hasn't changed
		var finalCount int
		err = db2.conn.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&finalCount)
		require.NoError(t, err)
		assert.Equal(t, initialCount, finalCount, "Should not re-apply migrations")
	})

	t.Run("idempotent migrations can be re-run safely", func(t *testing.T) {
		// Create a temporary database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		// Initialize database
		db, err := InitDatabase(dbPath, logger)
		require.NoError(t, err)
		defer db.Close()

		// Manually delete a migration record to simulate partial migration
		_, err = db.conn.Exec("DELETE FROM schema_migrations WHERE version = ?", "005_add_performance_indexes.sql")
		require.NoError(t, err)

		// Re-run migrations
		err = db.initSchema()
		require.NoError(t, err)

		// Verify the migration was re-applied
		var exists int
		err = db.conn.QueryRow(
			"SELECT COUNT(*) FROM schema_migrations WHERE version = ?",
			"005_add_performance_indexes.sql",
		).Scan(&exists)
		require.NoError(t, err)
		assert.Equal(t, 1, exists, "Migration should be re-applied")
	})
}

func TestMigrationOrder(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	t.Run("migrations are applied in correct order", func(t *testing.T) {
		// Create a temporary database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		// Initialize database
		db, err := InitDatabase(dbPath, logger)
		require.NoError(t, err)
		defer db.Close()

		// Query migrations in order they were applied
		rows, err := db.conn.Query("SELECT version FROM schema_migrations ORDER BY applied_at")
		require.NoError(t, err)
		defer rows.Close()

		var appliedMigrations []string
		for rows.Next() {
			var version string
			err := rows.Scan(&version)
			require.NoError(t, err)
			appliedMigrations = append(appliedMigrations, version)
		}

		// Verify order
		expectedOrder := []string{
			"001_initial_schema.sql",
			"002_add_auth_columns.sql",
			"003_add_combat_instances.sql",
			"004_add_mission_current_objective.sql",
			"005_add_performance_indexes.sql",
		}

		assert.Equal(t, expectedOrder, appliedMigrations, "Migrations should be applied in order")
	})
}
