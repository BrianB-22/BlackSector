package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rs/zerolog"
	"github.com/BrianB-22/BlackSector/internal/db"
)

const (
	// DefaultRetentionCount is the number of snapshots to keep
	DefaultRetentionCount = 10
	// SnapshotLatestSymlink is the name of the symlink pointing to the latest snapshot
	SnapshotLatestSymlink = "snapshot_latest.json"
)

// Snapshot represents the complete server state at a specific tick
type Snapshot struct {
	SnapshotVersion string        `json:"snapshot_version"`
	Tick            int64         `json:"tick"`
	Timestamp       int64         `json:"timestamp"`
	ServerName      string        `json:"server_name"`
	ProtocolVersion string        `json:"protocol_version"`
	State           SnapshotState `json:"state"`
}

// SnapshotState contains all game state data to be persisted
type SnapshotState struct {
	Players            []db.Player             `json:"players"`
	Sessions           []db.Session            `json:"sessions"`
	CombatInstances    []db.CombatInstance     `json:"combat_instances"`
	MissionInstances   []db.MissionInstance    `json:"mission_instances"`
	ObjectiveProgress  []db.ObjectiveProgress  `json:"objective_progress"`
	// Ships, traders, etc. will be added in future milestones
}

// SaveSnapshot writes a snapshot to disk atomically and manages retention
// It writes to snapshots/snapshot_{tick}_{timestamp}.json, updates the
// snapshot_latest.json symlink, and deletes old snapshots beyond retention count.
// Errors are logged but do not crash the server.
func SaveSnapshot(snapshot *Snapshot, dir string, logger zerolog.Logger) error {
	if snapshot == nil {
		return fmt.Errorf("snapshot cannot be nil")
	}

	// Ensure snapshots directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Error().Err(err).Str("dir", dir).Msg("Failed to create snapshots directory")
		return fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	// Generate filename: snapshot_{tick}_{timestamp}.json
	filename := fmt.Sprintf("snapshot_%d_%d.json", snapshot.Tick, snapshot.Timestamp)
	finalPath := filepath.Join(dir, filename)
	tempPath := finalPath + ".tmp"

	// Marshal snapshot to JSON
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal snapshot to JSON")
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	// Write to temporary file (atomic write step 1)
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		logger.Error().Err(err).Str("path", tempPath).Msg("Failed to write temporary snapshot file")
		return fmt.Errorf("failed to write temporary snapshot: %w", err)
	}

	// Rename temp file to final name (atomic write step 2)
	if err := os.Rename(tempPath, finalPath); err != nil {
		logger.Error().Err(err).Str("from", tempPath).Str("to", finalPath).Msg("Failed to rename snapshot file")
		// Clean up temp file
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename snapshot: %w", err)
	}

	logger.Info().
		Int64("tick", snapshot.Tick).
		Int64("timestamp", snapshot.Timestamp).
		Str("file", filename).
		Msg("Snapshot saved successfully")

	// Update snapshot_latest.json symlink
	if err := updateLatestSymlink(dir, filename, logger); err != nil {
		// Log error but don't fail - this is not critical
		logger.Warn().Err(err).Msg("Failed to update snapshot_latest.json symlink")
	}

	// Delete old snapshots beyond retention count
	if err := cleanOldSnapshots(dir, DefaultRetentionCount, logger); err != nil {
		// Log error but don't fail - this is not critical
		logger.Warn().Err(err).Msg("Failed to clean old snapshots")
	}

	return nil
}

// updateLatestSymlink updates the snapshot_latest.json symlink to point to the latest snapshot
func updateLatestSymlink(dir, filename string, logger zerolog.Logger) error {
	symlinkPath := filepath.Join(dir, SnapshotLatestSymlink)
	
	// Remove existing symlink if it exists
	if _, err := os.Lstat(symlinkPath); err == nil {
		if err := os.Remove(symlinkPath); err != nil {
			return fmt.Errorf("failed to remove old symlink: %w", err)
		}
	}

	// Create new symlink
	if err := os.Symlink(filename, symlinkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	logger.Debug().Str("symlink", SnapshotLatestSymlink).Str("target", filename).Msg("Updated snapshot symlink")
	return nil
}

// cleanOldSnapshots deletes snapshots beyond the retention count, keeping the most recent ones
func cleanOldSnapshots(dir string, retentionCount int, logger zerolog.Logger) error {
	// List all snapshot files
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	// Filter for snapshot files (snapshot_*.json, excluding symlink)
	var snapshots []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "snapshot_") && strings.HasSuffix(name, ".json") && name != SnapshotLatestSymlink {
			snapshots = append(snapshots, name)
		}
	}

	// If we have fewer snapshots than retention count, nothing to delete
	if len(snapshots) <= retentionCount {
		return nil
	}

	// Sort snapshots by name (which includes tick and timestamp, so chronological)
	sort.Strings(snapshots)

	// Delete oldest snapshots (keep the last retentionCount)
	toDelete := snapshots[:len(snapshots)-retentionCount]
	for _, filename := range toDelete {
		path := filepath.Join(dir, filename)
		if err := os.Remove(path); err != nil {
			logger.Warn().Err(err).Str("file", filename).Msg("Failed to delete old snapshot")
			continue
		}
		logger.Debug().Str("file", filename).Msg("Deleted old snapshot")
	}

	if len(toDelete) > 0 {
		logger.Info().Int("count", len(toDelete)).Msg("Cleaned old snapshots")
	}

	return nil
}

// LoadSnapshot loads the most recent snapshot from the snapshots directory.
// It first attempts to use the snapshot_latest.json symlink, falling back to
// finding the most recent snapshot file if the symlink doesn't exist.
// Returns nil, nil if no snapshots exist (not an error - allows clean first-time startup).
// Returns error if snapshot is corrupted or validation fails.
func LoadSnapshot(dir string, logger zerolog.Logger) (*Snapshot, error) {
	// Check if snapshots directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Info().Str("dir", dir).Msg("Snapshots directory does not exist, starting with empty state")
		return nil, nil
	}

	// Try to use snapshot_latest.json symlink first
	symlinkPath := filepath.Join(dir, SnapshotLatestSymlink)
	var snapshotPath string

	if info, err := os.Lstat(symlinkPath); err == nil && info.Mode()&os.ModeSymlink != 0 {
		// Symlink exists, resolve it
		target, err := os.Readlink(symlinkPath)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to read snapshot_latest.json symlink, falling back to directory scan")
		} else {
			snapshotPath = filepath.Join(dir, target)
			// Verify the target file exists
			if _, err := os.Stat(snapshotPath); err != nil {
				logger.Warn().Err(err).Str("target", target).Msg("Symlink target does not exist, falling back to directory scan")
				snapshotPath = ""
			}
		}
	}

	// If symlink didn't work, find the most recent snapshot file
	if snapshotPath == "" {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
		}

		// Filter for snapshot files (snapshot_*.json, excluding symlink)
		var snapshots []string
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasPrefix(name, "snapshot_") && strings.HasSuffix(name, ".json") && name != SnapshotLatestSymlink {
				snapshots = append(snapshots, name)
			}
		}

		// If no snapshots found, return nil (not an error)
		if len(snapshots) == 0 {
			logger.Info().Msg("No snapshots found, starting with empty state")
			return nil, nil
		}

		// Sort snapshots by name (chronological due to tick_timestamp format)
		sort.Strings(snapshots)

		// Get the most recent snapshot (last in sorted list)
		snapshotPath = filepath.Join(dir, snapshots[len(snapshots)-1])
	}

	// Read the snapshot file
	data, err := os.ReadFile(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot file %s: %w", snapshotPath, err)
	}

	// Unmarshal JSON
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot (corrupted): %w", err)
	}

	// Validate snapshot_version
	if snapshot.SnapshotVersion == "" {
		return nil, fmt.Errorf("snapshot missing snapshot_version field")
	}

	// Validate protocol_version
	if snapshot.ProtocolVersion == "" {
		return nil, fmt.Errorf("snapshot missing protocol_version field")
	}

	// Log successful load
	logger.Info().
		Int64("tick", snapshot.Tick).
		Int64("timestamp", snapshot.Timestamp).
		Str("snapshot_version", snapshot.SnapshotVersion).
		Str("protocol_version", snapshot.ProtocolVersion).
		Str("file", filepath.Base(snapshotPath)).
		Int("players", len(snapshot.State.Players)).
		Int("sessions", len(snapshot.State.Sessions)).
		Int("combat_instances", len(snapshot.State.CombatInstances)).
		Int("mission_instances", len(snapshot.State.MissionInstances)).
		Int("objective_progress", len(snapshot.State.ObjectiveProgress)).
		Msg("Snapshot loaded successfully")

	return &snapshot, nil
}
