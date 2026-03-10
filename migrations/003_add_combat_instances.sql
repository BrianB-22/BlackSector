-- Migration: 003
-- Description: Add combat_instances table for active combat tracking
-- Version: 0.6
-- Related: Milestone 2 - Vertical Slice (Phase 1) - Combat System

-- ============================================================================
-- Combat Domain
-- ============================================================================

-- Combat instances track active encounters between players and pirates
-- Pirates are ephemeral entities (not persisted after combat ends)
CREATE TABLE combat_instances (
  combat_id         TEXT PRIMARY KEY,
  player_ship_id    TEXT NOT NULL REFERENCES ships(ship_id),
  pirate_ship_id    TEXT NOT NULL,
  system_id         INTEGER NOT NULL REFERENCES systems(system_id),
  start_tick        INTEGER NOT NULL,
  status            TEXT NOT NULL CHECK (status IN ('ACTIVE', 'ENDED', 'FLED')),
  turn_number       INTEGER NOT NULL DEFAULT 0
);

-- Index on player_ship_id for fast lookups of active combat by player
CREATE INDEX idx_combat_player_ship ON combat_instances (player_ship_id);

-- Index on status for querying active combats during tick processing
CREATE INDEX idx_combat_status ON combat_instances (status);

-- Index on system_id for system-based combat queries
CREATE INDEX idx_combat_system ON combat_instances (system_id);
