package engine

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/BrianB-22/BlackSector/internal/session"
	"github.com/BrianB-22/BlackSector/internal/snapshot"
)

// Command represents a player action to be processed by the tick engine
type Command struct {
	SessionID   string
	PlayerID    string
	CommandType string
	Payload     []byte // JSON payload
	EnqueuedAt  int64  // Unix timestamp
}
// StateUpdate represents a game state snapshot broadcast to sessions
// Contains all information a session needs to update its TUI
type StateUpdate struct {
	TickNumber int64
	Timestamp  int64
	PlayerState *PlayerState
	Events     []GameEvent
}

// PlayerState contains the current state for a specific player
type PlayerState struct {
	PlayerID string
	Credits  int64
	Ship     *db.Ship
	Cargo    []db.CargoSlot
	// Future: combat state, mission state, etc.
}

// GameEvent represents a notable event that occurred this tick
type GameEvent struct {
	EventType string // "jump_complete", "trade_complete", "combat_start", etc.
	Message   string
	Data      map[string]interface{}
}

// ErrorEvent represents an error that occurred during command processing
// This is sent to the affected player's session to display the error
type ErrorEvent struct {
	CommandType string
	ErrorMsg    string
	Tick        int64
}

// Navigation command payloads
type JumpPayload struct {
	TargetSystemID int `json:"target_system_id"`
}

type DockPayload struct {
	PortID int `json:"port_id"`
}

// UndockPayload is empty - undock doesn't need parameters
type UndockPayload struct{}

// Economy command payloads
type BuyPayload struct {
	PortID      int    `json:"port_id"`
	CommodityID string `json:"commodity_id"`
	Quantity    int    `json:"quantity"`
}

type SellPayload struct {
	PortID      int    `json:"port_id"`
	CommodityID string `json:"commodity_id"`
	Quantity    int    `json:"quantity"`
}

// Combat command payloads
type AttackPayload struct {
	CombatID string `json:"combat_id"`
}

type FleePayload struct {
	CombatID string `json:"combat_id"`
}

type SurrenderPayload struct {
	CombatID string `json:"combat_id"`
}

// Mission command payloads
type MissionAcceptPayload struct {
	MissionID string `json:"mission_id"`
}

type MissionAbandonPayload struct{}


// Navigator defines the interface for ship navigation operations
type Navigator interface {
	Jump(shipID string, targetSystemID int, currentTick int64) error
	Dock(shipID string, portID int, currentTick int64) error
	Undock(shipID string, currentTick int64) error
}

// Trader defines the interface for commodity trading operations
type Trader interface {
	BuyCommodity(shipID string, portID int, commodityID string, quantity int, tick int64) error
	SellCommodity(shipID string, portID int, commodityID string, quantity int, tick int64) error
}

// CombatResolver defines the interface for combat operations
type CombatResolver interface {
	ProcessAttack(combatID string, attackerID string, tick int64) (interface{}, error)
	ProcessFlee(combatID string, playerID string, tick int64) (interface{}, error)
	ProcessSurrender(combatID string, playerID string, tick int64) error
	ResolveCombatTick(tick int64) ([]interface{}, error)
	CheckPirateSpawns(tick int64) ([]interface{}, error)
}

// MissionController defines the interface for mission operations
type MissionController interface {
	AcceptMission(missionID string, playerID string, tick int64) error
	AbandonMission(playerID string, tick int64) error
	EvaluateObjectives(tick int64) ([]interface{}, error)
}

// TickEngine is the authoritative simulation loop that owns all game state mutations.
// It runs in a single goroutine and processes commands from player sessions.
type TickEngine struct {
	tickNumber            int64
	tickInterval          time.Duration
	running               bool
	commandQueue          chan Command
	shutdownChan          chan struct{}
	db                    *db.Database
	sessionMgr            *session.SessionManager
	navigation            Navigator        // Navigation subsystem
	trader                Trader           // Economy subsystem
	combat                CombatResolver   // Combat subsystem
	missions              MissionController // Mission subsystem
	logger                zerolog.Logger
	snapshotIntervalTicks int
	serverName            string
}

// Config contains configuration parameters for the TickEngine
type Config struct {
	TickIntervalMs        int
	SnapshotIntervalTicks int
	ServerName            string
	InitialTickNumber     int64 // Tick number to start from (0 for fresh, snapshot.Tick+1 for recovery)
}

// NewTickEngine creates a new TickEngine instance.
// The tick number is initialized from the provided config (either 0 for a fresh server
// or snapshot.Tick+1 for recovery from a snapshot).
func NewTickEngine(cfg Config, database *db.Database, sessionMgr *session.SessionManager, nav Navigator, trader Trader, combat CombatResolver, missions MissionController, logger zerolog.Logger) *TickEngine {
	return &TickEngine{
		tickNumber:            cfg.InitialTickNumber,
		tickInterval:          time.Duration(cfg.TickIntervalMs) * time.Millisecond,
		running:               false,
		commandQueue:          make(chan Command, 1000), // Buffered channel for commands
		shutdownChan:          make(chan struct{}),
		db:                    database,
		sessionMgr:            sessionMgr,
		navigation:            nav,
		trader:                trader,
		combat:                combat,
		missions:              missions,
		logger:                logger,
		snapshotIntervalTicks: cfg.SnapshotIntervalTicks,
		serverName:            cfg.ServerName,
	}
}

// Start sets the running flag to true, allowing the tick loop to begin
func (te *TickEngine) Start() {
	te.running = true
	te.logger.Info().
		Int64("initial_tick", te.tickNumber).
		Int("tick_interval_ms", int(te.tickInterval.Milliseconds())).
		Int("snapshot_interval_ticks", te.snapshotIntervalTicks).
		Msg("Tick engine started")
}

// Stop sets the running flag to false, signaling the tick loop to stop
func (te *TickEngine) Stop() {
	te.running = false
	te.logger.Info().Msg("Tick engine stop requested")
}

// IsRunning returns whether the tick engine is currently running
func (te *TickEngine) IsRunning() bool {
	return te.running
}

// TickNumber returns the current tick number
func (te *TickEngine) TickNumber() int64 {
	return te.tickNumber
}

// EnqueueCommand adds a command to the command queue for processing in the next tick
func (te *TickEngine) EnqueueCommand(cmd Command) {
	select {
	case te.commandQueue <- cmd:
		te.logger.Debug().
			Str("session_id", cmd.SessionID).
			Str("player_id", cmd.PlayerID).
			Str("command_type", cmd.CommandType).
			Int64("tick", te.tickNumber).
			Msg("Command enqueued")
	default:
		te.logger.Warn().
			Str("session_id", cmd.SessionID).
			Str("command_type", cmd.CommandType).
			Msg("Command queue full, command dropped")
	}
}

// ShutdownChan returns the shutdown channel for graceful shutdown coordination
func (te *TickEngine) ShutdownChan() <-chan struct{} {
	return te.shutdownChan
}
// RunTickLoop executes the main game simulation loop.
// It runs continuously while the running flag is true, incrementing the tick number
// each iteration and maintaining consistent timing based on the configured tick interval.
// Each tick measures its duration and logs a warning if it exceeds 500ms.
// RunTickLoop executes the main game simulation loop.
// It runs continuously while the running flag is true, incrementing the tick number
// each iteration and maintaining consistent timing based on the configured tick interval.
// Each tick measures its duration and logs a warning if it exceeds 500ms.
// Panics are recovered with stack traces logged, and graceful shutdown is attempted.
func (te *TickEngine) RunTickLoop() {
	// Recover from panics in the tick loop
	// Requirement 15.7: Recover from panics, log with stack trace, attempt graceful shutdown
	defer func() {
		if r := recover(); r != nil {
			te.logger.Error().
				Interface("panic", r).
				Stack().
				Int64("tick", te.tickNumber).
				Msg("PANIC in tick loop - attempting graceful shutdown")

			// Set running to false to signal shutdown
			te.running = false

			// Attempt to create a final snapshot before exiting
			if err := te.CreateFinalSnapshot(); err != nil {
				te.logger.Error().
					Err(err).
					Msg("Failed to create snapshot after panic")
			}

			// Signal shutdown channel
			close(te.shutdownChan)
		}
	}()

	te.logger.Info().Msg("Tick loop starting")

	for te.running {
		tickStart := time.Now()

		// Log current tick at DEBUG level
		te.logger.Debug().
			Int64("tick", te.tickNumber).
			Msg("Tick executing")

		// Phase 1: Drain command queue (placeholder for M1)
		// In M1, we just drain commands but don't process them
		te.drainCommandQueue()

		// Phase 2: Resolve combat turns for all active combats
		if te.combat != nil {
			if events, err := te.combat.ResolveCombatTick(te.tickNumber); err != nil {
				te.logger.Error().
					Err(err).
					Int64("tick", te.tickNumber).
					Msg("Failed to resolve combat tick")
			} else if len(events) > 0 {
				te.logger.Debug().
					Int("event_count", len(events)).
					Int64("tick", te.tickNumber).
					Msg("Combat events generated")
			}
		}

		// Phase 3: Check for pirate spawns (every 5 ticks in low security systems)
		if te.combat != nil && te.tickNumber%5 == 0 {
			if events, err := te.combat.CheckPirateSpawns(te.tickNumber); err != nil {
				te.logger.Error().
					Err(err).
					Int64("tick", te.tickNumber).
					Msg("Failed to check pirate spawns")
			} else if len(events) > 0 {
				te.logger.Info().
					Int("spawn_count", len(events)).
					Int64("tick", te.tickNumber).
					Msg("Pirates spawned")
			}
		}

		// Phase 4: Evaluate mission objectives
		if te.missions != nil {
			if events, err := te.missions.EvaluateObjectives(te.tickNumber); err != nil {
				te.logger.Error().
					Err(err).
					Int64("tick", te.tickNumber).
					Msg("Failed to evaluate mission objectives")
			} else if len(events) > 0 {
				te.logger.Debug().
					Int("event_count", len(events)).
					Int64("tick", te.tickNumber).
					Msg("Mission events generated")
			}
		}

		// Phase 4.5: Broadcast state updates to sessions
		te.broadcastStateUpdate()

		// Phase 5: Snapshot trigger check
		if te.snapshotIntervalTicks > 0 && te.tickNumber%int64(te.snapshotIntervalTicks) == 0 {
			te.logger.Info().
				Int64("tick", te.tickNumber).
				Int("snapshot_interval_ticks", te.snapshotIntervalTicks).
				Msg("Snapshot triggered")

			// Create snapshot with current game state
			snap, err := te.createSnapshot()
			if err != nil {
				te.logger.Error().
					Err(err).
					Int64("tick", te.tickNumber).
					Msg("Failed to create snapshot")
			} else {
				// Write snapshot asynchronously to avoid blocking tick loop
				go func(s *snapshot.Snapshot) {
					if err := snapshot.SaveSnapshot(s, "snapshots", te.logger); err != nil {
						te.logger.Error().
							Err(err).
							Int64("tick", s.Tick).
							Msg("Failed to save snapshot")
					}
				}(snap)
			}
		}

		// Calculate tick duration
		tickDuration := time.Since(tickStart)

		// Log warning if tick took too long (>100ms threshold)
		if tickDuration > 100*time.Millisecond {
			te.logger.Warn().
				Int64("tick", te.tickNumber).
				Int64("duration_ms", tickDuration.Milliseconds()).
				Msg("Slow tick detected")
		}

		// Increment tick number for next iteration
		te.tickNumber++

		// Sleep for remaining interval time
		sleepDuration := te.tickInterval - tickDuration
		if sleepDuration > 0 {
			time.Sleep(sleepDuration)
		}
		// If tick took longer than interval, proceed immediately to next tick
	}

	te.logger.Info().
		Int64("final_tick", te.tickNumber-1).
		Msg("Tick loop stopped")
}

// drainCommandQueue removes all pending commands from the queue.
// Commands are drained in FIFO order, validated, and processed.
// Navigation commands (jump, dock, undock) are routed to the navigation subsystem.
// Invalid commands are rejected and logged.
// Errors from subsystem operations are logged and sent to the affected player's session.
func (te *TickEngine) drainCommandQueue() {
	drained := 0
	valid := 0
	invalid := 0
	processed := 0
	
	for {
		select {
		case cmd := <-te.commandQueue:
			drained++
			
			// Validate command structure
			if !te.validateCommand(cmd) {
				invalid++
				// Log rejection for invalid command
				te.logger.Debug().
					Str("session_id", cmd.SessionID).
					Str("player_id", cmd.PlayerID).
					Str("command_type", cmd.CommandType).
					Int64("tick", te.tickNumber).
					Msg("Command rejected: invalid structure")
				
				// Send error to player session
				te.sendErrorToSession(cmd.SessionID, cmd.CommandType, "Invalid command structure")
				continue
			}
			
			valid++
			
			// Process command by type
			if err := te.processCommand(cmd); err != nil {
				// Log error with structured fields
				te.logger.Warn().
					Err(err).
					Str("session_id", cmd.SessionID).
					Str("player_id", cmd.PlayerID).
					Str("command_type", cmd.CommandType).
					Int64("tick", te.tickNumber).
					Msg("Command processing failed")
				
				// Send error message to affected player's session
				te.sendErrorToSession(cmd.SessionID, cmd.CommandType, err.Error())
			} else {
				processed++
				te.logger.Debug().
					Str("session_id", cmd.SessionID).
					Str("player_id", cmd.PlayerID).
					Str("command_type", cmd.CommandType).
					Int64("tick", te.tickNumber).
					Msg("Command processed successfully")
			}
			
		default:
			// Queue is empty
			if drained > 0 {
				te.logger.Debug().
					Int("commands_drained", drained).
					Int("valid", valid).
					Int("invalid", invalid).
					Int("processed", processed).
					Int64("tick", te.tickNumber).
					Msg("Command queue drained")
			}
			return
		}
	}
}

// validateCommand performs basic structure validation on a command.
// Returns true if the command is valid, false otherwise.
func (te *TickEngine) validateCommand(cmd Command) bool {
	// Basic structure validation:
	// 1. SessionID must be non-empty
	if cmd.SessionID == "" {
		return false
	}
	
	// 2. PlayerID must be non-empty
	if cmd.PlayerID == "" {
		return false
	}
	
	// 3. CommandType must be non-empty
	if cmd.CommandType == "" {
		return false
	}
	
	// 4. EnqueuedAt must be positive (valid Unix timestamp)
	if cmd.EnqueuedAt <= 0 {
		return false
	}
	
	// Payload can be empty or nil for some commands, so we don't validate it here
	
	return true
}


// createSnapshot creates a snapshot of the current game state
// It retrieves all players, sessions, combat instances, and mission data from the database
// and packages them into a Snapshot struct with metadata
func (te *TickEngine) createSnapshot() (*snapshot.Snapshot, error) {
	// Get all players from database
	players, err := te.db.GetAllPlayers()
	if err != nil {
		return nil, fmt.Errorf("failed to get players for snapshot: %w", err)
	}

	// Get all sessions from database
	sessions, err := te.db.GetAllSessions()
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions for snapshot: %w", err)
	}

	// Get all combat instances from database
	combatInstances, err := te.db.GetAllCombatInstances()
	if err != nil {
		return nil, fmt.Errorf("failed to get combat instances for snapshot: %w", err)
	}

	// Get all mission instances from database
	missionInstances, err := te.db.GetAllMissionInstances()
	if err != nil {
		return nil, fmt.Errorf("failed to get mission instances for snapshot: %w", err)
	}

	// Get all objective progress from database
	objectiveProgress, err := te.db.GetAllObjectiveProgressForSnapshot()
	if err != nil {
		return nil, fmt.Errorf("failed to get objective progress for snapshot: %w", err)
	}

	// Create snapshot with current state
	snap := &snapshot.Snapshot{
		SnapshotVersion: "1.0",
		Tick:            te.tickNumber,
		Timestamp:       time.Now().Unix(),
		ServerName:      te.serverName,
		ProtocolVersion: "1.0",
		State: snapshot.SnapshotState{
			Players:           players,
			Sessions:          sessions,
			CombatInstances:   combatInstances,
			MissionInstances:  missionInstances,
			ObjectiveProgress: objectiveProgress,
		},
	}

	return snap, nil
}


// CreateFinalSnapshot creates and saves a final snapshot synchronously during shutdown.
// Unlike the async snapshots during normal operation, this blocks until the snapshot is written.
func (te *TickEngine) CreateFinalSnapshot() error {
	te.logger.Info().
		Int64("tick", te.tickNumber).
		Msg("Creating final snapshot")

	snap, err := te.createSnapshot()
	if err != nil {
		return fmt.Errorf("failed to create final snapshot: %w", err)
	}

	// Write synchronously (blocking)
	if err := snapshot.SaveSnapshot(snap, "snapshots", te.logger); err != nil {
		return fmt.Errorf("failed to save final snapshot: %w", err)
	}

	te.logger.Info().
		Int64("tick", snap.Tick).
		Msg("Final snapshot saved successfully")

	return nil
}

// processCommand routes a command to the appropriate subsystem for processing
func (te *TickEngine) processCommand(cmd Command) error {
	switch cmd.CommandType {
	case "jump":
		return te.processJumpCommand(cmd)
	case "dock":
		return te.processDockCommand(cmd)
	case "undock":
		return te.processUndockCommand(cmd)
	case "buy":
		return te.processBuyCommand(cmd)
	case "sell":
		return te.processSellCommand(cmd)
	case "attack":
		return te.processAttackCommand(cmd)
	case "flee":
		return te.processFleeCommand(cmd)
	case "surrender":
		return te.processSurrenderCommand(cmd)
	case "mission_accept":
		return te.processMissionAcceptCommand(cmd)
	case "mission_abandon":
		return te.processMissionAbandonCommand(cmd)
	default:
		return fmt.Errorf("unknown command type: %s", cmd.CommandType)
	}
}

// processJumpCommand handles jump navigation commands
func (te *TickEngine) processJumpCommand(cmd Command) error {
	// Parse payload
	var payload JumpPayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		return fmt.Errorf("parse jump payload: %w", err)
	}
	
	// Get ship ID from player
	// For now, we assume one ship per player - we'll need to query the DB for the ship
	// This is a simplified implementation for M2 integration
	ship, err := te.db.GetShipByPlayerID(cmd.PlayerID)
	if err != nil {
		return fmt.Errorf("get ship for player %s: %w", cmd.PlayerID, err)
	}
	if ship == nil {
		return fmt.Errorf("no ship found for player %s", cmd.PlayerID)
	}
	
	// Call navigation subsystem
	if err := te.navigation.Jump(ship.ShipID, payload.TargetSystemID, te.tickNumber); err != nil {
		return fmt.Errorf("jump: %w", err)
	}
	
	te.logger.Info().
		Str("player_id", cmd.PlayerID).
		Str("ship_id", ship.ShipID).
		Int("target_system", payload.TargetSystemID).
		Int64("tick", te.tickNumber).
		Msg("Jump command processed")
	
	return nil
}

// processDockCommand handles dock commands
func (te *TickEngine) processDockCommand(cmd Command) error {
	// Parse payload
	var payload DockPayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		return fmt.Errorf("parse dock payload: %w", err)
	}
	
	// Get ship ID from player
	ship, err := te.db.GetShipByPlayerID(cmd.PlayerID)
	if err != nil {
		return fmt.Errorf("get ship for player %s: %w", cmd.PlayerID, err)
	}
	if ship == nil {
		return fmt.Errorf("no ship found for player %s", cmd.PlayerID)
	}
	
	// Call navigation subsystem
	if err := te.navigation.Dock(ship.ShipID, payload.PortID, te.tickNumber); err != nil {
		return fmt.Errorf("dock: %w", err)
	}
	
	te.logger.Info().
		Str("player_id", cmd.PlayerID).
		Str("ship_id", ship.ShipID).
		Int("port_id", payload.PortID).
		Int64("tick", te.tickNumber).
		Msg("Dock command processed")
	
	return nil
}

// processUndockCommand handles undock commands
func (te *TickEngine) processUndockCommand(cmd Command) error {
	// Get ship ID from player
	ship, err := te.db.GetShipByPlayerID(cmd.PlayerID)
	if err != nil {
		return fmt.Errorf("get ship for player %s: %w", cmd.PlayerID, err)
	}
	if ship == nil {
		return fmt.Errorf("no ship found for player %s", cmd.PlayerID)
	}
	
	// Call navigation subsystem
	if err := te.navigation.Undock(ship.ShipID, te.tickNumber); err != nil {
		return fmt.Errorf("undock: %w", err)
	}
	
	te.logger.Info().
		Str("player_id", cmd.PlayerID).
		Str("ship_id", ship.ShipID).
		Int64("tick", te.tickNumber).
		Msg("Undock command processed")
	
	return nil
}

// processBuyCommand handles commodity purchase commands
func (te *TickEngine) processBuyCommand(cmd Command) error {
	// Parse payload
	var payload BuyPayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		return fmt.Errorf("parse buy payload: %w", err)
	}
	
	// Get ship ID from player
	ship, err := te.db.GetShipByPlayerID(cmd.PlayerID)
	if err != nil {
		return fmt.Errorf("get ship for player %s: %w", cmd.PlayerID, err)
	}
	if ship == nil {
		return fmt.Errorf("no ship found for player %s", cmd.PlayerID)
	}
	
	// Call economy subsystem
	if err := te.trader.BuyCommodity(ship.ShipID, payload.PortID, payload.CommodityID, payload.Quantity, te.tickNumber); err != nil {
		return fmt.Errorf("buy commodity: %w", err)
	}
	
	te.logger.Info().
		Str("player_id", cmd.PlayerID).
		Str("ship_id", ship.ShipID).
		Int("port_id", payload.PortID).
		Str("commodity_id", payload.CommodityID).
		Int("quantity", payload.Quantity).
		Int64("tick", te.tickNumber).
		Msg("Buy command processed")
	
	return nil
}

// processSellCommand handles commodity sale commands
func (te *TickEngine) processSellCommand(cmd Command) error {
	// Parse payload
	var payload SellPayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		return fmt.Errorf("parse sell payload: %w", err)
	}
	
	// Get ship ID from player
	ship, err := te.db.GetShipByPlayerID(cmd.PlayerID)
	if err != nil {
		return fmt.Errorf("get ship for player %s: %w", cmd.PlayerID, err)
	}
	if ship == nil {
		return fmt.Errorf("no ship found for player %s", cmd.PlayerID)
	}
	
	// Call economy subsystem
	if err := te.trader.SellCommodity(ship.ShipID, payload.PortID, payload.CommodityID, payload.Quantity, te.tickNumber); err != nil {
		return fmt.Errorf("sell commodity: %w", err)
	}
	
	te.logger.Info().
		Str("player_id", cmd.PlayerID).
		Str("ship_id", ship.ShipID).
		Int("port_id", payload.PortID).
		Str("commodity_id", payload.CommodityID).
		Int("quantity", payload.Quantity).
		Int64("tick", te.tickNumber).
		Msg("Sell command processed")
	
	return nil
}

// processAttackCommand handles combat attack commands
func (te *TickEngine) processAttackCommand(cmd Command) error {
	if te.combat == nil {
		return fmt.Errorf("combat subsystem not initialized")
	}

	// Parse payload
	var payload AttackPayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		return fmt.Errorf("parse attack payload: %w", err)
	}
	
	// Get ship ID from player
	ship, err := te.db.GetShipByPlayerID(cmd.PlayerID)
	if err != nil {
		return fmt.Errorf("get ship for player %s: %w", cmd.PlayerID, err)
	}
	if ship == nil {
		return fmt.Errorf("no ship found for player %s", cmd.PlayerID)
	}
	
	// Call combat subsystem
	if _, err := te.combat.ProcessAttack(payload.CombatID, ship.ShipID, te.tickNumber); err != nil {
		return fmt.Errorf("attack: %w", err)
	}
	
	te.logger.Info().
		Str("player_id", cmd.PlayerID).
		Str("ship_id", ship.ShipID).
		Str("combat_id", payload.CombatID).
		Int64("tick", te.tickNumber).
		Msg("Attack command processed")
	
	return nil
}

// processFleeCommand handles combat flee commands
func (te *TickEngine) processFleeCommand(cmd Command) error {
	if te.combat == nil {
		return fmt.Errorf("combat subsystem not initialized")
	}

	// Parse payload
	var payload FleePayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		return fmt.Errorf("parse flee payload: %w", err)
	}
	
	// Get ship ID from player
	ship, err := te.db.GetShipByPlayerID(cmd.PlayerID)
	if err != nil {
		return fmt.Errorf("get ship for player %s: %w", cmd.PlayerID, err)
	}
	if ship == nil {
		return fmt.Errorf("no ship found for player %s", cmd.PlayerID)
	}
	
	// Call combat subsystem
	if _, err := te.combat.ProcessFlee(payload.CombatID, cmd.PlayerID, te.tickNumber); err != nil {
		return fmt.Errorf("flee: %w", err)
	}
	
	te.logger.Info().
		Str("player_id", cmd.PlayerID).
		Str("ship_id", ship.ShipID).
		Str("combat_id", payload.CombatID).
		Int64("tick", te.tickNumber).
		Msg("Flee command processed")
	
	return nil
}

// processSurrenderCommand handles combat surrender commands
func (te *TickEngine) processSurrenderCommand(cmd Command) error {
	if te.combat == nil {
		return fmt.Errorf("combat subsystem not initialized")
	}

	// Parse payload
	var payload SurrenderPayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		return fmt.Errorf("parse surrender payload: %w", err)
	}
	
	// Call combat subsystem
	if err := te.combat.ProcessSurrender(payload.CombatID, cmd.PlayerID, te.tickNumber); err != nil {
		return fmt.Errorf("surrender: %w", err)
	}
	
	te.logger.Info().
		Str("player_id", cmd.PlayerID).
		Str("combat_id", payload.CombatID).
		Int64("tick", te.tickNumber).
		Msg("Surrender command processed")
	
	return nil
}

// processMissionAcceptCommand handles mission acceptance commands
func (te *TickEngine) processMissionAcceptCommand(cmd Command) error {
	if te.missions == nil {
		return fmt.Errorf("mission subsystem not initialized")
	}

	// Parse payload
	var payload MissionAcceptPayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		return fmt.Errorf("parse mission accept payload: %w", err)
	}
	
	// Call mission subsystem
	if err := te.missions.AcceptMission(payload.MissionID, cmd.PlayerID, te.tickNumber); err != nil {
		return fmt.Errorf("accept mission: %w", err)
	}
	
	te.logger.Info().
		Str("player_id", cmd.PlayerID).
		Str("mission_id", payload.MissionID).
		Int64("tick", te.tickNumber).
		Msg("Mission accept command processed")
	
	return nil
}

// processMissionAbandonCommand handles mission abandonment commands
func (te *TickEngine) processMissionAbandonCommand(cmd Command) error {
	if te.missions == nil {
		return fmt.Errorf("mission subsystem not initialized")
	}

	// Call mission subsystem
	if err := te.missions.AbandonMission(cmd.PlayerID, te.tickNumber); err != nil {
		return fmt.Errorf("abandon mission: %w", err)
	}
	
	te.logger.Info().
		Str("player_id", cmd.PlayerID).
		Int64("tick", te.tickNumber).
		Msg("Mission abandon command processed")
	
	return nil
}

// sendErrorToSession sends an error message to a specific player's session
// This is called when a command fails to process, allowing the player to see what went wrong
func (te *TickEngine) sendErrorToSession(sessionID string, commandType string, errorMsg string) {
	errorEvent := ErrorEvent{
		CommandType: commandType,
		ErrorMsg:    errorMsg,
		Tick:        te.tickNumber,
	}
	
	if err := te.sessionMgr.BroadcastToSession(sessionID, errorEvent); err != nil {
		te.logger.Warn().
			Err(err).
			Str("session_id", sessionID).
			Str("command_type", commandType).
			Msg("Failed to send error message to session")
	}
}

// broadcastStateUpdate sends state updates to all active sessions
// This is called after command processing in each tick
func (te *TickEngine) broadcastStateUpdate() {
	// Get all active sessions
	sessions := te.sessionMgr.ActiveSessions()
	
	if len(sessions) == 0 {
		return // No active sessions to update
	}
	
	// Build state updates for each session
	updates := make(map[string]interface{})
	
	for _, session := range sessions {
		// Get player state
		player, err := te.db.GetPlayerByID(session.PlayerID)
		if err != nil {
			te.logger.Warn().
				Err(err).
				Str("session_id", session.SessionID).
				Str("player_id", session.PlayerID).
				Msg("Failed to get player for state update")
			continue
		}
		
		// Get ship state
		ship, err := te.db.GetShipByPlayerID(session.PlayerID)
		if err != nil {
			te.logger.Warn().
				Err(err).
				Str("session_id", session.SessionID).
				Str("player_id", session.PlayerID).
				Msg("Failed to get ship for state update")
			continue
		}
		
		// Get cargo state
		var cargo []db.CargoSlot
		if ship != nil {
			cargo, err = te.db.GetShipCargo(ship.ShipID)
			if err != nil {
				te.logger.Warn().
					Err(err).
					Str("session_id", session.SessionID).
					Str("ship_id", ship.ShipID).
					Msg("Failed to get cargo for state update")
				// Continue with empty cargo rather than skipping the update
				cargo = []db.CargoSlot{}
			}
		}
		
		// Build player state
		playerState := &PlayerState{
			PlayerID: session.PlayerID,
			Credits:  player.Credits,
			Ship:     ship,
			Cargo:    cargo,
		}
		
		// Create state update
		update := StateUpdate{
			TickNumber:  te.tickNumber,
			Timestamp:   time.Now().Unix(),
			PlayerState: playerState,
			Events:      []GameEvent{}, // Future: populate with events from this tick
		}
		
		updates[session.SessionID] = update
	}
	
	// Broadcast to all sessions
	successCount, errors := te.sessionMgr.BroadcastToAllSessions(updates)
	
	if len(errors) > 0 {
		te.logger.Warn().
			Int("success_count", successCount).
			Int("error_count", len(errors)).
			Int64("tick", te.tickNumber).
			Msg("State broadcast completed with errors")
	} else {
		te.logger.Debug().
			Int("session_count", successCount).
			Int64("tick", te.tickNumber).
			Msg("State broadcast completed successfully")
	}
}
