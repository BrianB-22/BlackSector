-- Migration: Add authentication columns for registration system
-- Date: 2024-03-06

-- Add ssh_username column (nullable for backward compatibility)
ALTER TABLE players ADD COLUMN ssh_username TEXT;

-- Add password_hash column (nullable for backward compatibility)
ALTER TABLE players ADD COLUMN password_hash TEXT;

-- Create index on ssh_username for fast lookups during registration
CREATE INDEX IF NOT EXISTS idx_players_ssh_username ON players(ssh_username);
