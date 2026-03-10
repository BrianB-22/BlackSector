-- Migration: 004
-- Description: Add current_objective_index to mission_instances table
-- Version: 0.7
-- Related: Milestone 2 - Vertical Slice (Phase 1) - Mission System (Task 9.2)

-- ============================================================================
-- Mission Domain Enhancement
-- ============================================================================

-- Add current_objective_index to track which objective the player is currently working on
-- This allows the mission system to evaluate objectives sequentially
-- Default to 0 (first objective) for any existing missions
ALTER TABLE mission_instances ADD COLUMN current_objective_index INTEGER NOT NULL DEFAULT 0;

-- Add index on player_id and status for fast lookups of active missions
-- This supports the "one active mission per player" constraint (REQ-MISSION-007)
CREATE INDEX IF NOT EXISTS idx_missions_player_status ON mission_instances (player_id, status);

-- Add index on status for querying missions by state during tick processing
CREATE INDEX IF NOT EXISTS idx_missions_status ON mission_instances (status);

-- Add index on expires_at_tick for efficient expiry checking during tick processing
CREATE INDEX IF NOT EXISTS idx_missions_expiry ON mission_instances (expires_at_tick) WHERE expires_at_tick IS NOT NULL;
