package missions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Database interface defines required database operations for the mission system
type Database interface {
	// Mission instance operations
	CreateMissionInstance(instance *MissionInstance) error
	GetMissionInstance(instanceID string) (*MissionInstance, error)
	GetActiveMissionByPlayer(playerID string) (*MissionInstance, error)
	GetAllInProgressMissions() ([]*MissionInstance, error)
	GetCompletedMissionsByPlayer(playerID string) ([]*MissionInstance, error)
	UpdateMissionStatus(instanceID string, status string, tick int64) error
	UpdateMissionObjectiveIndex(instanceID string, objectiveIndex int) error
	DeleteMissionInstance(instanceID string) error

	// Objective progress operations
	CreateObjectiveProgress(progress *ObjectiveProgress) error
	GetObjectiveProgress(instanceID string, objectiveIndex int) (*ObjectiveProgress, error)
	GetAllObjectiveProgress(instanceID string) ([]*ObjectiveProgress, error)
	UpdateObjectiveProgress(instanceID string, objectiveIndex int, status string, currentValue int) error
	DeleteObjectiveProgress(instanceID string) error

	// Player and ship operations
	GetPlayerByID(playerID string) (*Player, error)
	GetShipByPlayerID(playerID string) (*Ship, error)
	UpdatePlayerCredits(playerID string, credits int) error
	GetCargoByShipID(shipID string) ([]*CargoSlot, error)

	// World queries
	GetPortByID(portID int) (*Port, error)
	GetSystemSecurityLevel(systemID int) (float64, error)
}

// Player represents a player (simplified for interface)
type Player struct {
	PlayerID string
	Credits  int64
}

// Ship represents a ship (simplified for interface)
type Ship struct {
	ShipID          string
	PlayerID        string
	CurrentSystemID int
	Status          string
	DockedAtPortID  *int
}

// CargoSlot represents cargo (simplified for interface)
type CargoSlot struct {
	ShipID      string
	CommodityID string
	Quantity    int
}

// Port represents a port (simplified for interface)
type Port struct {
	PortID   int
	SystemID int
}

// Config holds mission system configuration
type Config struct {
	MissionConfigPath string // Path to mission JSON files (default: "config/missions/")
	EnableHotReload   bool   // Enable hot-reload of mission files (Phase 2)
}

// DefaultConfig returns the default mission configuration
func DefaultConfig() *Config {
	return &Config{
		MissionConfigPath: "config/missions/",
		EnableHotReload:   false,
	}
}

// MissionManager manages mission lifecycle and objective tracking
type MissionManager struct {
	cfg       *Config
	db        Database
	registry  map[string]*MissionDefinition // mission_id -> definition
	mu        sync.RWMutex                  // Protects registry for concurrent reads
	logger    zerolog.Logger
}

// NewMissionManager creates a new mission manager instance
func NewMissionManager(cfg *Config, db Database, logger zerolog.Logger) *MissionManager {
	return &MissionManager{
		cfg:      cfg,
		db:       db,
		registry: make(map[string]*MissionDefinition),
		logger:   logger,
	}
}

// LoadMissions loads mission definitions from config/missions/ directory
func (m *MissionManager) LoadMissions(configPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear existing registry
	m.registry = make(map[string]*MissionDefinition)

	// Read mission directory
	files, err := os.ReadDir(configPath)
	if err != nil {
		return fmt.Errorf("read mission directory: %w", err)
	}

	loadedCount := 0
	errorCount := 0

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(configPath, file.Name())
		if err := m.loadMissionFile(filePath); err != nil {
			m.logger.Error().
				Err(err).
				Str("file", file.Name()).
				Msg("failed to load mission file")
			errorCount++
			continue
		}

		loadedCount++
	}

	m.logger.Info().
		Int("loaded", loadedCount).
		Int("errors", errorCount).
		Int("total_missions", len(m.registry)).
		Msg("mission definitions loaded")

	if loadedCount == 0 && errorCount > 0 {
		return fmt.Errorf("no mission files loaded successfully")
	}

	return nil
}

// loadMissionFile loads and validates a single mission JSON file
func (m *MissionManager) loadMissionFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var missionFile MissionFile
	if err := json.Unmarshal(data, &missionFile); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	// Validate mission file
	if err := ValidateMissionFile(&missionFile); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Register missions
	for _, mission := range missionFile.Missions {
		m.registry[mission.MissionID] = mission
		m.logger.Debug().
			Str("mission_id", mission.MissionID).
			Str("name", mission.Name).
			Bool("enabled", mission.Enabled).
			Msg("mission registered")
	}

	return nil
}

// GetAvailableMissions returns missions available at a port for a player
func (m *MissionManager) GetAvailableMissions(portID int, playerID string) ([]*MissionListing, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if player has active mission
	activeMission, err := m.db.GetActiveMissionByPlayer(playerID)
	if err != nil {
		return nil, fmt.Errorf("get available missions: %w", err)
	}
	if activeMission != nil {
		// Player already has active mission - return empty list
		return []*MissionListing{}, nil
	}

	// Get port to determine security zone
	port, err := m.db.GetPortByID(portID)
	if err != nil {
		return nil, fmt.Errorf("get available missions: %w", err)
	}
	if port == nil {
		return nil, fmt.Errorf("get available missions: port not found")
	}

	// TODO: Extract security level from port - for now, accept all zones
	// This will be refined when we have proper port data structures

	listings := make([]*MissionListing, 0)

	for _, mission := range m.registry {
		// Skip disabled missions
		if !mission.Enabled {
			continue
		}

		// Check if mission is available at this port
		if len(mission.AvailableAtPorts) > 0 {
			available := false
			for _, pid := range mission.AvailableAtPorts {
				if pid == portID {
					available = true
					break
				}
			}
			if !available {
				continue
			}
		}

		// TODO: Check security zone filtering
		// TODO: Check repeat cooldown for repeatable missions

		listings = append(listings, &MissionListing{
			MissionID:     mission.MissionID,
			Name:          mission.Name,
			Description:   mission.Description,
			Objectives:    mission.Objectives,
			Rewards:       mission.Rewards,
			ExpiryTicks:   mission.ExpiryTicks,
			Repeatable:    mission.Repeatable,
			SecurityZones: mission.SecurityZones,
		})
	}

	return listings, nil
}

// AcceptMission creates a mission instance for a player
func (m *MissionManager) AcceptMission(missionID string, playerID string, tick int64) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get mission definition
	mission, exists := m.registry[missionID]
	if !exists {
		return fmt.Errorf("accept mission: %w", ErrMissionNotFound)
	}

	if !mission.Enabled {
		return fmt.Errorf("accept mission: %w", ErrMissionNotEnabled)
	}

	// Check if player already has active mission
	activeMission, err := m.db.GetActiveMissionByPlayer(playerID)
	if err != nil {
		return fmt.Errorf("accept mission: %w", err)
	}
	if activeMission != nil {
		return fmt.Errorf("accept mission: %w", ErrPlayerHasActiveMission)
	}

	// Create mission instance
	instanceID := uuid.New().String()
	var expiresAt *int64
	if mission.ExpiryTicks != nil {
		expiry := tick + int64(*mission.ExpiryTicks)
		expiresAt = &expiry
	}

	instance := &MissionInstance{
		InstanceID:    instanceID,
		MissionID:     missionID,
		PlayerID:      playerID,
		Status:        MissionInProgress,
		AcceptedTick:  tick,
		StartedTick:   nil,
		CompletedTick: nil,
		FailedReason:  nil,
		ExpiresAtTick: expiresAt,
	}

	if err := m.db.CreateMissionInstance(instance); err != nil {
		return fmt.Errorf("accept mission: %w", err)
	}

	// Initialize objective progress for all objectives
	for i, obj := range mission.Objectives {
		status := string(ObjectivePending)
		if i == 0 {
			status = string(ObjectiveActive) // First objective is active
		}

		requiredValue := m.extractRequiredValue(obj)

		progress := &ObjectiveProgress{
			InstanceID:     instanceID,
			ObjectiveIndex: i,
			Status:         status,
			CurrentValue:   0,
			RequiredValue:  requiredValue,
		}

		if err := m.db.CreateObjectiveProgress(progress); err != nil {
			// Cleanup mission instance
			m.db.DeleteMissionInstance(instanceID)
			return fmt.Errorf("accept mission: %w", err)
		}
	}

	m.logger.Info().
		Str("instance_id", instanceID).
		Str("mission_id", missionID).
		Str("player_id", playerID).
		Int64("tick", tick).
		Msg("mission accepted")

	return nil
}

// extractRequiredValue extracts the required value from objective parameters
func (m *MissionManager) extractRequiredValue(obj *ObjectiveDefinition) int {
	switch obj.Type {
	case "deliver_commodity":
		if quantity, ok := obj.Parameters["quantity"].(float64); ok {
			return int(quantity)
		}
	case "navigate_to":
		return 1 // Just need to reach the destination
	}
	return 1
}

// GetActiveMission returns the player's current active mission
func (m *MissionManager) GetActiveMission(playerID string) (*MissionInstance, error) {
	instance, err := m.db.GetActiveMissionByPlayer(playerID)
	if err != nil {
		return nil, fmt.Errorf("get active mission: %w", err)
	}
	return instance, nil
}

// AbandonMission cancels the active mission
func (m *MissionManager) AbandonMission(playerID string, tick int64) error {
	// Get active mission
	instance, err := m.db.GetActiveMissionByPlayer(playerID)
	if err != nil {
		return fmt.Errorf("abandon mission: %w", err)
	}
	if instance == nil {
		return fmt.Errorf("abandon mission: %w", ErrNoActiveMission)
	}

	// Update status to abandoned
	if err := m.db.UpdateMissionStatus(instance.InstanceID, string(MissionAbandoned), tick); err != nil {
		return fmt.Errorf("abandon mission: %w", err)
	}

	m.logger.Info().
		Str("instance_id", instance.InstanceID).
		Str("mission_id", instance.MissionID).
		Str("player_id", playerID).
		Int64("tick", tick).
		Msg("mission abandoned")

	return nil
}

// EvaluateObjectives checks mission progress each tick
func (m *MissionManager) EvaluateObjectives(tick int64) ([]*MissionEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := make([]*MissionEvent, 0)

	// Get all IN_PROGRESS missions
	missions, err := m.db.GetAllInProgressMissions()
	if err != nil {
		return nil, fmt.Errorf("evaluate objectives: %w", err)
	}

	for _, instance := range missions {
		// Check for expiry first
		if instance.ExpiresAtTick != nil && tick >= *instance.ExpiresAtTick {
			if err := m.expireMission(instance, tick); err != nil {
				m.logger.Error().
					Err(err).
					Str("instance_id", instance.InstanceID).
					Msg("failed to expire mission")
				continue
			}

			events = append(events, &MissionEvent{
				Type:       "expired",
				InstanceID: instance.InstanceID,
				PlayerID:   instance.PlayerID,
				MissionID:  instance.MissionID,
				Tick:       tick,
			})
			continue
		}

		// Get mission definition
		missionDef, exists := m.registry[instance.MissionID]
		if !exists {
			m.logger.Warn().
				Str("mission_id", instance.MissionID).
				Str("instance_id", instance.InstanceID).
				Msg("mission definition not found for instance")
			continue
		}

		// Get all objective progress
		progressList, err := m.db.GetAllObjectiveProgress(instance.InstanceID)
		if err != nil {
			m.logger.Error().
				Err(err).
				Str("instance_id", instance.InstanceID).
				Msg("failed to get objective progress")
			continue
		}

		// Find current active objective
		var currentObjectiveIndex int = -1
		for _, progress := range progressList {
			if progress.Status == string(ObjectiveActive) {
				currentObjectiveIndex = progress.ObjectiveIndex
				break
			}
		}

		if currentObjectiveIndex == -1 {
			// No active objective - this shouldn't happen
			m.logger.Warn().
				Str("instance_id", instance.InstanceID).
				Msg("no active objective found")
			continue
		}

		// Get the objective definition
		if currentObjectiveIndex >= len(missionDef.Objectives) {
			m.logger.Error().
				Str("instance_id", instance.InstanceID).
				Int("objective_index", currentObjectiveIndex).
				Msg("objective index out of bounds")
			continue
		}

		objectiveDef := missionDef.Objectives[currentObjectiveIndex]
		currentProgress := progressList[currentObjectiveIndex]

		// Evaluate the objective based on type
		completed, err := m.evaluateObjective(instance, objectiveDef, currentProgress, tick)
		if err != nil {
			m.logger.Error().
				Err(err).
				Str("instance_id", instance.InstanceID).
				Int("objective_index", currentObjectiveIndex).
				Msg("failed to evaluate objective")
			continue
		}

		if completed {
			// Mark current objective as completed
			if err := m.db.UpdateObjectiveProgress(instance.InstanceID, currentObjectiveIndex, string(ObjectiveCompleted), currentProgress.RequiredValue); err != nil {
				m.logger.Error().
					Err(err).
					Str("instance_id", instance.InstanceID).
					Msg("failed to update objective progress")
				continue
			}

			// Check if this was the final objective
			if currentObjectiveIndex == len(missionDef.Objectives)-1 {
				// Mission complete!
				if err := m.completeMission(instance, missionDef, tick); err != nil {
					m.logger.Error().
						Err(err).
						Str("instance_id", instance.InstanceID).
						Msg("failed to complete mission")
					continue
				}

				events = append(events, &MissionEvent{
					Type:       "completed",
					InstanceID: instance.InstanceID,
					PlayerID:   instance.PlayerID,
					MissionID:  instance.MissionID,
					Tick:       tick,
					Details: map[string]interface{}{
						"reward_credits": missionDef.Rewards.Credits,
					},
				})
			} else {
				// Activate next objective
				nextObjectiveIndex := currentObjectiveIndex + 1
				if err := m.db.UpdateObjectiveProgress(instance.InstanceID, nextObjectiveIndex, string(ObjectiveActive), 0); err != nil {
					m.logger.Error().
						Err(err).
						Str("instance_id", instance.InstanceID).
						Msg("failed to activate next objective")
					continue
				}

				m.logger.Info().
					Str("instance_id", instance.InstanceID).
					Int("objective_index", nextObjectiveIndex).
					Msg("objective completed, next objective activated")
			}
		}
	}

	return events, nil
}

// evaluateObjective checks if an objective is complete
func (m *MissionManager) evaluateObjective(instance *MissionInstance, objectiveDef *ObjectiveDefinition, progress *ObjectiveProgress, tick int64) (bool, error) {
	switch objectiveDef.Type {
	case "deliver_commodity":
		return m.evaluateDeliverCommodity(instance, objectiveDef, progress)
	case "navigate_to":
		return m.evaluateNavigateTo(instance, objectiveDef, progress)
	default:
		m.logger.Warn().
			Str("objective_type", objectiveDef.Type).
			Msg("unsupported objective type")
		return false, nil
	}
}

// evaluateDeliverCommodity checks if player is at destination with required commodity
func (m *MissionManager) evaluateDeliverCommodity(instance *MissionInstance, objectiveDef *ObjectiveDefinition, progress *ObjectiveProgress) (bool, error) {
	// Get player's ship
	ship, err := m.db.GetShipByPlayerID(instance.PlayerID)
	if err != nil {
		return false, fmt.Errorf("get ship: %w", err)
	}
	if ship == nil {
		return false, fmt.Errorf("ship not found for player")
	}

	// Check if ship is docked
	if ship.Status != "DOCKED" || ship.DockedAtPortID == nil {
		return false, nil
	}

	// Get destination port from parameters
	destinationPortID, ok := objectiveDef.Parameters["destination_port_id"].(float64)
	if !ok {
		return false, fmt.Errorf("destination_port_id parameter missing or invalid")
	}

	// Check if docked at correct port
	if *ship.DockedAtPortID != int(destinationPortID) {
		return false, nil
	}

	// Get required commodity and quantity
	commodityID, ok := objectiveDef.Parameters["commodity_id"].(string)
	if !ok {
		return false, fmt.Errorf("commodity_id parameter missing or invalid")
	}

	quantity, ok := objectiveDef.Parameters["quantity"].(float64)
	if !ok {
		return false, fmt.Errorf("quantity parameter missing or invalid")
	}

	// Check player's cargo
	cargo, err := m.db.GetCargoByShipID(ship.ShipID)
	if err != nil {
		return false, fmt.Errorf("get cargo: %w", err)
	}

	// Find the commodity in cargo
	var cargoQuantity int
	for _, slot := range cargo {
		if slot.CommodityID == commodityID {
			cargoQuantity += slot.Quantity
		}
	}

	// Check if player has enough
	if cargoQuantity >= int(quantity) {
		return true, nil
	}

	return false, nil
}

// evaluateNavigateTo checks if player is at destination
func (m *MissionManager) evaluateNavigateTo(instance *MissionInstance, objectiveDef *ObjectiveDefinition, progress *ObjectiveProgress) (bool, error) {
	// Get player's ship
	ship, err := m.db.GetShipByPlayerID(instance.PlayerID)
	if err != nil {
		return false, fmt.Errorf("get ship: %w", err)
	}
	if ship == nil {
		return false, fmt.Errorf("ship not found for player")
	}

	// Get destination system from parameters
	destinationSystemID, ok := objectiveDef.Parameters["destination_system_id"].(float64)
	if !ok {
		return false, fmt.Errorf("destination_system_id parameter missing or invalid")
	}

	// Check if ship is in the destination system
	if ship.CurrentSystemID == int(destinationSystemID) {
		return true, nil
	}

	return false, nil
}

// completeMission marks mission as complete and distributes rewards
func (m *MissionManager) completeMission(instance *MissionInstance, missionDef *MissionDefinition, tick int64) error {
	// Update mission status
	if err := m.db.UpdateMissionStatus(instance.InstanceID, string(MissionCompleted), tick); err != nil {
		return fmt.Errorf("update mission status: %w", err)
	}

	// Distribute credit rewards
	if missionDef.Rewards != nil && missionDef.Rewards.Credits > 0 {
		player, err := m.db.GetPlayerByID(instance.PlayerID)
		if err != nil {
			return fmt.Errorf("get player: %w", err)
		}
		if player == nil {
			return fmt.Errorf("player not found")
		}

		newCredits := int(player.Credits) + missionDef.Rewards.Credits
		if err := m.db.UpdatePlayerCredits(instance.PlayerID, newCredits); err != nil {
			return fmt.Errorf("update player credits: %w", err)
		}

		m.logger.Info().
			Str("instance_id", instance.InstanceID).
			Str("player_id", instance.PlayerID).
			Int("reward_credits", missionDef.Rewards.Credits).
			Int("new_credits", newCredits).
			Msg("mission rewards distributed")
	}

	// TODO: Distribute item rewards (Phase 2)

	m.logger.Info().
		Str("instance_id", instance.InstanceID).
		Str("mission_id", instance.MissionID).
		Str("player_id", instance.PlayerID).
		Int64("tick", tick).
		Msg("mission completed")

	return nil
}

// expireMission marks mission as expired
func (m *MissionManager) expireMission(instance *MissionInstance, tick int64) error {
	if err := m.db.UpdateMissionStatus(instance.InstanceID, string(MissionExpired), tick); err != nil {
		return fmt.Errorf("expire mission: %w", err)
	}

	m.logger.Info().
		Str("instance_id", instance.InstanceID).
		Str("mission_id", instance.MissionID).
		Str("player_id", instance.PlayerID).
		Int64("tick", tick).
		Msg("mission expired")

	return nil
}

// GetMissionDefinition retrieves a mission definition by ID
func (m *MissionManager) GetMissionDefinition(missionID string) (*MissionDefinition, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mission, exists := m.registry[missionID]
	if !exists {
		return nil, fmt.Errorf("get mission definition: %w", ErrMissionNotFound)
	}

	return mission, nil
}

// GetMissionCount returns the number of loaded missions
func (m *MissionManager) GetMissionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.registry)
}
