-- Migration: 002
-- Description: Add SSH username and password hash fields for registration system
-- Version: 0.5

-- Add ssh_username field to players table
ALTER TABLE players ADD COLUMN ssh_username TEXT;

-- Add password_hash field to players table
ALTER TABLE players ADD COLUMN password_hash TEXT;

-- Create unique index on ssh_username for fast lookups
CREATE UNIQUE INDEX idx_players_ssh_username ON players (ssh_username);
