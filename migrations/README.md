# Database Migrations

This directory contains SQL migration files for the BlackSector database schema.

## Migration System

The migration system automatically applies database schema changes in order when the server starts or when tests run. It tracks which migrations have been applied using a `schema_migrations` table.

### How It Works

1. On database initialization, the system creates a `schema_migrations` table if it doesn't exist
2. It reads all migration files from this directory in order
3. For each migration, it checks if it has already been applied
4. If not applied, it executes the migration SQL within a transaction
5. On success, it records the migration in `schema_migrations` with a timestamp

### Migration Files

Migrations are numbered sequentially and should follow this naming convention:
```
NNN_description.sql
```

Where:
- `NNN` is a zero-padded 3-digit number (001, 002, 003, etc.)
- `description` is a brief description using underscores

**Current Migrations:**
- `001_initial_schema.sql` - Base schema with all core tables
- `002_add_auth_columns.sql` - SSH authentication columns for registration
- `003_add_combat_instances.sql` - Combat tracking table
- `004_add_mission_current_objective.sql` - Mission objective tracking
- `005_add_performance_indexes.sql` - Performance optimization indexes

### Adding New Migrations

1. Create a new file with the next sequential number
2. Write idempotent SQL (use `IF NOT EXISTS`, `IF EXISTS`, etc.)
3. Add the filename to the `migrations` array in `internal/db/db.go`
4. Test the migration on a fresh database and an existing database

### Idempotency

All migrations should be idempotent (safe to run multiple times) using:
- `CREATE TABLE IF NOT EXISTS`
- `CREATE INDEX IF NOT EXISTS`
- `ALTER TABLE ... ADD COLUMN` (SQLite will error if column exists, handle gracefully)

### Schema Migrations Table

```sql
CREATE TABLE schema_migrations (
  version TEXT PRIMARY KEY,      -- Migration filename (e.g., "001_initial_schema.sql")
  applied_at INTEGER NOT NULL    -- Unix timestamp when migration was applied
)
```

### Testing

The migration system is tested in `internal/db/migrations_test.go`:
- New database applies all migrations
- Existing database skips already-applied migrations
- Migrations are applied in correct order
- Idempotent migrations can be re-run safely

### Notes

- Migration 002 has two variants in the directory (`002_add_auth_columns.sql` and `002_add_registration_fields.sql`). The system uses `002_add_auth_columns.sql` (non-unique index) for flexibility.
- All migrations are applied within transactions for atomicity
- Migration failures will rollback and prevent database initialization
- The system tries multiple paths to find migration files (supports both production and test environments)
