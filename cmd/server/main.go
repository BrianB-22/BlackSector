package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BrianB-22/BlackSector/internal/admin"
	"github.com/BrianB-22/BlackSector/internal/combat"
	"github.com/BrianB-22/BlackSector/internal/config"
	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/BrianB-22/BlackSector/internal/economy"
	"github.com/BrianB-22/BlackSector/internal/engine"
	"github.com/BrianB-22/BlackSector/internal/missions"
	"github.com/BrianB-22/BlackSector/internal/navigation"
	"github.com/BrianB-22/BlackSector/internal/registration"
	"github.com/BrianB-22/BlackSector/internal/session"
	"github.com/BrianB-22/BlackSector/internal/snapshot"
	"github.com/BrianB-22/BlackSector/internal/sshserver"
	"github.com/BrianB-22/BlackSector/internal/world"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Version information (injected at build time via -ldflags)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
	GitBranch = "unknown"
)

// dbAdapter adapts db.Database to implement navigation.ShipRepository interface
type dbAdapter struct {
	db *db.Database
}

func (a *dbAdapter) GetShipByID(shipID string) (*navigation.Ship, error) {
	dbShip, err := a.db.GetShipByID(shipID)
	if err != nil {
		return nil, err
	}
	if dbShip == nil {
		return nil, nil
	}
	
	// Convert db.Ship to navigation.Ship
	return &navigation.Ship{
		ShipID:          dbShip.ShipID,
		PlayerID:        dbShip.PlayerID,
		ShipClass:       dbShip.ShipClass,
		HullPoints:      dbShip.HullPoints,
		MaxHullPoints:   dbShip.MaxHullPoints,
		ShieldPoints:    dbShip.ShieldPoints,
		MaxShieldPoints: dbShip.MaxShieldPoints,
		EnergyPoints:    dbShip.EnergyPoints,
		MaxEnergyPoints: dbShip.MaxEnergyPoints,
		CargoCapacity:   dbShip.CargoCapacity,
		MissilesCurrent: dbShip.MissilesCurrent,
		CurrentSystemID: dbShip.CurrentSystemID,
		PositionX:       dbShip.PositionX,
		PositionY:       dbShip.PositionY,
		Status:          navigation.ShipStatus(dbShip.Status),
		DockedAtPortID:  dbShip.DockedAtPortID,
		LastUpdatedTick: dbShip.LastUpdatedTick,
	}, nil
}

func (a *dbAdapter) UpdateShipPosition(shipID string, systemID int, tick int64) error {
	return a.db.UpdateShipPosition(shipID, systemID, tick)
}

func (a *dbAdapter) UpdateShipDockStatus(shipID string, status navigation.ShipStatus, dockedAtPortID *int, tick int64) error {
	return a.db.UpdateShipDockStatus(shipID, string(status), dockedAtPortID, tick)
}

// combatDBAdapter adapts db.Database to implement combat.Database interface
type combatDBAdapter struct {
	db *db.Database
}

func (a *combatDBAdapter) GetShipByID(shipID string) (*combat.Ship, error) {
	dbShip, err := a.db.GetShipByID(shipID)
	if err != nil {
		return nil, err
	}
	if dbShip == nil {
		return nil, nil
	}
	
	// Convert db.Ship to combat.Ship
	return &combat.Ship{
		ShipID:          dbShip.ShipID,
		PlayerID:        dbShip.PlayerID,
		ShipClass:       dbShip.ShipClass,
		HullPoints:      dbShip.HullPoints,
		MaxHullPoints:   dbShip.MaxHullPoints,
		ShieldPoints:    dbShip.ShieldPoints,
		MaxShieldPoints: dbShip.MaxShieldPoints,
		CurrentSystemID: dbShip.CurrentSystemID,
		Status:          dbShip.Status,
		WeaponDamage:    20, // TODO: Get from ship class config
	}, nil
}

func (a *combatDBAdapter) UpdateShipStatus(shipID string, status string, tick int64) error {
	return a.db.UpdateShipDockStatus(shipID, status, nil, tick)
}

func (a *combatDBAdapter) UpdateShipDamage(shipID string, hull, shield int, tick int64) error {
	// TODO: Implement UpdateShipDamage in db.Database
	return fmt.Errorf("UpdateShipDamage not yet implemented")
}

func (a *combatDBAdapter) ClearShipCargo(shipID string) error {
	// TODO: Implement ClearShipCargo in db.Database
	return fmt.Errorf("ClearShipCargo not yet implemented")
}

func (a *combatDBAdapter) RespawnShip(shipID string, systemID int, portID int, tick int64) error {
	// TODO: Implement RespawnShip in db.Database
	return fmt.Errorf("RespawnShip not yet implemented")
}

func (a *combatDBAdapter) GetPlayerByShipID(shipID string) (string, error) {
	ship, err := a.db.GetShipByID(shipID)
	if err != nil {
		return "", err
	}
	if ship == nil {
		return "", fmt.Errorf("ship not found")
	}
	return ship.PlayerID, nil
}

func (a *combatDBAdapter) GetPlayerCredits(playerID string) (int64, error) {
	player, err := a.db.GetPlayerByID(playerID)
	if err != nil {
		return 0, err
	}
	if player == nil {
		return 0, fmt.Errorf("player not found")
	}
	return player.Credits, nil
}

func (a *combatDBAdapter) UpdatePlayerCredits(playerID string, credits int) error {
	return a.db.UpdatePlayerCredits(playerID, credits)
}

func (a *combatDBAdapter) CreateCombatInstance(combat *combat.CombatInstance) error {
	return a.db.CreateCombatInstance(combat.CombatID, combat.PlayerShipID, combat.PirateShipID, combat.SystemID, combat.StartTick)
}

func (a *combatDBAdapter) GetCombatInstance(combatID string) (*combat.CombatInstance, error) {
	// TODO: Implement GetCombatInstance in db.Database
	return nil, fmt.Errorf("GetCombatInstance not yet implemented")
}

func (a *combatDBAdapter) GetActiveCombatByShip(shipID string) (*combat.CombatInstance, error) {
	dbCombat, err := a.db.GetActiveCombatByPlayerShip(shipID)
	if err != nil {
		return nil, err
	}
	if dbCombat == nil {
		return nil, nil
	}
	
	// Convert db.CombatInstance to combat.CombatInstance
	return &combat.CombatInstance{
		CombatID:     dbCombat.CombatID,
		PlayerShipID: dbCombat.PlayerShipID,
		PirateShipID: dbCombat.PirateShipID,
		SystemID:     dbCombat.SystemID,
		StartTick:    dbCombat.StartTick,
		Status:       combat.CombatStatus(dbCombat.Status),
		TurnNumber:   dbCombat.TurnNumber,
	}, nil
}

func (a *combatDBAdapter) UpdateCombatStatus(combatID string, status combat.CombatStatus, tick int64) error {
	return a.db.UpdateCombatStatus(combatID, string(status))
}

func (a *combatDBAdapter) UpdateCombatTurn(combatID string, turnNumber int) error {
	return a.db.UpdateCombatTurn(combatID, turnNumber)
}

func (a *combatDBAdapter) DeleteCombatInstance(combatID string) error {
	return a.db.DeleteCombat(combatID)
}

func (a *combatDBAdapter) CreatePirateShip(pirate *combat.PirateShip) error {
	// Pirates are ephemeral - store in memory only
	return nil
}

func (a *combatDBAdapter) GetPirateShip(pirateShipID string) (*combat.PirateShip, error) {
	// Pirates are ephemeral - not persisted
	return nil, fmt.Errorf("pirate ship not found (ephemeral)")
}

func (a *combatDBAdapter) UpdatePirateShip(pirate *combat.PirateShip) error {
	// Pirates are ephemeral - no persistence needed
	return nil
}

func (a *combatDBAdapter) DeletePirateShip(pirateShipID string) error {
	// Pirates are ephemeral - no persistence needed
	return nil
}

func (a *combatDBAdapter) GetSystemSecurityLevel(systemID int) (float64, error) {
	// TODO: Implement GetSystemSecurityLevel in db.Database
	return 0.5, nil // Default to medium security
}

func (a *combatDBAdapter) GetShipsInSpace() ([]*combat.Ship, error) {
	// TODO: Implement GetShipsInSpace in db.Database
	return nil, nil
}

func (a *combatDBAdapter) FindNearestPort(systemID int) (int, error) {
	// TODO: Implement FindNearestPort in db.Database
	return 1, nil // Default to port 1
}

func (a *combatDBAdapter) BeginTx() (*sql.Tx, error) {
	return a.db.BeginTx()
}

func (a *combatDBAdapter) CommitTx(tx *sql.Tx) error {
	return a.db.CommitTx(tx)
}

func (a *combatDBAdapter) RollbackTx(tx *sql.Tx) error {
	return a.db.RollbackTx(tx)
}

// missionDBAdapter adapts db.Database to implement missions.Database interface
type missionDBAdapter struct {
	db *db.Database
}

func (a *missionDBAdapter) CreateMissionInstance(instance *missions.MissionInstance) error {
	// TODO: Implement CreateMissionInstance in db.Database
	return fmt.Errorf("CreateMissionInstance not yet implemented")
}

func (a *missionDBAdapter) GetMissionInstance(instanceID string) (*missions.MissionInstance, error) {
	// TODO: Implement GetMissionInstance in db.Database
	return nil, fmt.Errorf("GetMissionInstance not yet implemented")
}

func (a *missionDBAdapter) GetActiveMissionByPlayer(playerID string) (*missions.MissionInstance, error) {
	// TODO: Implement GetActiveMissionByPlayer in db.Database
	return nil, nil
}

func (a *missionDBAdapter) GetAllInProgressMissions() ([]*missions.MissionInstance, error) {
	// TODO: Implement GetAllInProgressMissions in db.Database
	return nil, nil
}

func (a *missionDBAdapter) GetCompletedMissionsByPlayer(playerID string) ([]*missions.MissionInstance, error) {
	// TODO: Implement GetCompletedMissionsByPlayer in db.Database
	return nil, nil
}

func (a *missionDBAdapter) UpdateMissionStatus(instanceID string, status string, tick int64) error {
	// TODO: Implement UpdateMissionStatus in db.Database
	return fmt.Errorf("UpdateMissionStatus not yet implemented")
}

func (a *missionDBAdapter) UpdateMissionObjectiveIndex(instanceID string, objectiveIndex int) error {
	// TODO: Implement UpdateMissionObjectiveIndex in db.Database
	return fmt.Errorf("UpdateMissionObjectiveIndex not yet implemented")
}

func (a *missionDBAdapter) DeleteMissionInstance(instanceID string) error {
	// TODO: Implement DeleteMissionInstance in db.Database
	return fmt.Errorf("DeleteMissionInstance not yet implemented")
}

func (a *missionDBAdapter) CreateObjectiveProgress(progress *missions.ObjectiveProgress) error {
	// TODO: Implement CreateObjectiveProgress in db.Database
	return fmt.Errorf("CreateObjectiveProgress not yet implemented")
}

func (a *missionDBAdapter) GetObjectiveProgress(instanceID string, objectiveIndex int) (*missions.ObjectiveProgress, error) {
	// TODO: Implement GetObjectiveProgress in db.Database
	return nil, fmt.Errorf("GetObjectiveProgress not yet implemented")
}

func (a *missionDBAdapter) GetAllObjectiveProgress(instanceID string) ([]*missions.ObjectiveProgress, error) {
	// TODO: Implement GetAllObjectiveProgress in db.Database
	return nil, nil
}

func (a *missionDBAdapter) UpdateObjectiveProgress(instanceID string, objectiveIndex int, status string, currentValue int) error {
	// TODO: Implement UpdateObjectiveProgress in db.Database
	return fmt.Errorf("UpdateObjectiveProgress not yet implemented")
}

func (a *missionDBAdapter) DeleteObjectiveProgress(instanceID string) error {
	// TODO: Implement DeleteObjectiveProgress in db.Database
	return fmt.Errorf("DeleteObjectiveProgress not yet implemented")
}

func (a *missionDBAdapter) GetPlayerByID(playerID string) (*missions.Player, error) {
	dbPlayer, err := a.db.GetPlayerByID(playerID)
	if err != nil {
		return nil, err
	}
	if dbPlayer == nil {
		return nil, nil
	}
	
	return &missions.Player{
		PlayerID: dbPlayer.PlayerID,
		Credits:  dbPlayer.Credits,
	}, nil
}

func (a *missionDBAdapter) GetShipByPlayerID(playerID string) (*missions.Ship, error) {
	dbShip, err := a.db.GetShipByPlayerID(playerID)
	if err != nil {
		return nil, err
	}
	if dbShip == nil {
		return nil, nil
	}
	
	return &missions.Ship{
		ShipID:          dbShip.ShipID,
		PlayerID:        dbShip.PlayerID,
		CurrentSystemID: dbShip.CurrentSystemID,
		Status:          dbShip.Status,
		DockedAtPortID:  dbShip.DockedAtPortID,
	}, nil
}

func (a *missionDBAdapter) UpdatePlayerCredits(playerID string, credits int) error {
	return a.db.UpdatePlayerCredits(playerID, credits)
}

func (a *missionDBAdapter) GetCargoByShipID(shipID string) ([]*missions.CargoSlot, error) {
	dbCargo, err := a.db.GetShipCargo(shipID)
	if err != nil {
		return nil, err
	}
	
	cargo := make([]*missions.CargoSlot, len(dbCargo))
	for i, slot := range dbCargo {
		cargo[i] = &missions.CargoSlot{
			ShipID:      slot.ShipID,
			CommodityID: slot.CommodityID,
			Quantity:    slot.Quantity,
		}
	}
	
	return cargo, nil
}

func (a *missionDBAdapter) GetPortByID(portID int) (*missions.Port, error) {
	// TODO: Implement GetPortByID in db.Database
	return &missions.Port{
		PortID:   portID,
		SystemID: 1, // Default
	}, nil
}

func (a *missionDBAdapter) GetSystemSecurityLevel(systemID int) (float64, error) {
	// TODO: Implement GetSystemSecurityLevel in db.Database
	return 0.5, nil // Default to medium security
}

// combatSystemAdapter wraps combat.CombatSystem to match engine.CombatResolver interface
type combatSystemAdapter struct {
	*combat.CombatSystem
}

func (a *combatSystemAdapter) ProcessAttack(combatID string, attackerID string, tick int64) (interface{}, error) {
	return a.CombatSystem.ProcessAttack(combatID, attackerID, tick)
}

func (a *combatSystemAdapter) ProcessFlee(combatID string, playerID string, tick int64) (interface{}, error) {
	return a.CombatSystem.ProcessFlee(combatID, playerID, tick)
}

func (a *combatSystemAdapter) ResolveCombatTick(tick int64) ([]interface{}, error) {
	events, err := a.CombatSystem.ResolveCombatTick(tick)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(events))
	for i, e := range events {
		result[i] = e
	}
	return result, nil
}

func (a *combatSystemAdapter) CheckPirateSpawns(tick int64) ([]interface{}, error) {
	events, err := a.CombatSystem.CheckPirateSpawns(tick)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(events))
	for i, e := range events {
		result[i] = e
	}
	return result, nil
}

// missionSystemAdapter wraps missions.MissionManager to match engine.MissionController interface
type missionSystemAdapter struct {
	*missions.MissionManager
}

func (a *missionSystemAdapter) EvaluateObjectives(tick int64) ([]interface{}, error) {
	events, err := a.MissionManager.EvaluateObjectives(tick)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(events))
	for i, e := range events {
		result[i] = e
	}
	return result, nil
}

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "config/server.json", "Path to server configuration file")
	showVersion := flag.Bool("version", false, "Show version information and exit")
	flag.Parse()

	// Show version and exit if requested
	if *showVersion {
		fmt.Printf("BlackSector Server\n")
		fmt.Printf("  Version:    %s\n", Version)
		fmt.Printf("  Build Time: %s\n", BuildTime)
		fmt.Printf("  Git Commit: %s\n", GitCommit)
		fmt.Printf("  Git Branch: %s\n", GitBranch)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	initLogger(cfg.Logging)

	log.Info().
		Str("version", Version).
		Str("git_commit", GitCommit).
		Int("ssh_port", cfg.Server.SSHPort).
		Int("tick_interval_ms", cfg.Server.TickIntervalMs).
		Msg("BlackSector server starting")

	// Initialize database
	database, err := initDatabase(cfg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize database")
		os.Exit(1)
	}
	defer database.Close()

	// Load snapshot or initialize empty state
	tickNumber, err := loadSnapshotOrInitialize(cfg, database)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load snapshot")
		os.Exit(1)
	}

	log.Info().
		Int64("starting_tick", tickNumber).
		Msg("Server initialization complete (foundation stub)")

	// Initialize session manager
	sessionMgr := initSessionManager(database)

	// Initialize tick engine
	tickEngine := initTickEngine(cfg, database, sessionMgr, tickNumber)

	// Initialize registrar for new player registration
	registrar := initRegistrar(cfg, database, tickEngine)

	// Start SSH server (needs tick engine reference for command forwarding)
	sshServer, err := startSSHServer(cfg, sessionMgr, tickEngine, registrar)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start SSH server")
		os.Exit(1)
	}

	// Initialize and start admin CLI
	initAdminCLI(tickEngine, sessionMgr)

	// Start tick loop in goroutine
	log.Info().Msg("Tick loop starting")
	go tickEngine.RunTickLoop()

	log.Info().Msg("Server running - press Ctrl+C to shutdown or type 'shutdown' to stop")

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal or admin shutdown command
	// The admin CLI handler will call tickEngine.Stop() when shutdown is requested
	// We need to wait for either a signal or for the tick engine to stop
	shutdownReason := ""
	
	go func() {
		sig := <-sigChan
		shutdownReason = sig.String()
		tickEngine.Stop()
	}()
	
	// Wait for tick engine to stop (either from signal or admin command)
	for tickEngine.IsRunning() {
		time.Sleep(100 * time.Millisecond)
	}
	
	if shutdownReason == "" {
		shutdownReason = "admin command"
	}

	log.Info().
		Str("reason", shutdownReason).
		Msg("Shutdown signal received")

	// Perform graceful shutdown
	performGracefulShutdown(tickEngine, sshServer, sessionMgr, database)

	log.Info().Msg("Server stopped")
}

// initLogger configures zerolog based on the logging configuration
func initLogger(cfg config.LoggingConfig) {
	// Set log level
	var level zerolog.Level
	switch cfg.Level {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	default:
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output
	if cfg.LogFile != "" {
		logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
			os.Exit(1)
		}
		log.Logger = zerolog.New(logFile).With().Timestamp().Logger()
	} else {
		// Use console output with pretty formatting for development
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Enable debug logging if configured
	if cfg.DebugEnabled {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

// initDatabase initializes the database connection with proper PRAGMAs
func initDatabase(cfg *config.Config) (*db.Database, error) {
	log.Info().Str("db_path", cfg.Server.DBPath).Msg("Initializing database")
	
	database, err := db.InitDatabase(cfg.Server.DBPath, log.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	
	log.Info().Msg("Database initialized successfully")
	return database, nil
}

// loadSnapshotOrInitialize loads the most recent snapshot or initializes with empty state
// Returns the starting tick number for the server
func loadSnapshotOrInitialize(cfg *config.Config, database *db.Database) (int64, error) {
	snapshotsDir := "snapshots"
	
	// Attempt to load the most recent snapshot
	snap, err := snapshot.LoadSnapshot(snapshotsDir, log.Logger)
	if err != nil {
		return 0, fmt.Errorf("failed to load snapshot: %w", err)
	}
	
	// If no snapshot exists, start with empty state at tick 0
	if snap == nil {
		log.Info().
			Int64("tick", 0).
			Msg("No snapshot found, starting with empty state")
		return 0, nil
	}
	
	// Snapshot found - restore state
	tickNumber := snap.Tick + 1
	
	log.Info().
		Int64("snapshot_tick", snap.Tick).
		Int64("starting_tick", tickNumber).
		Str("snapshot_version", snap.SnapshotVersion).
		Str("protocol_version", snap.ProtocolVersion).
		Int("players", len(snap.State.Players)).
		Int("sessions", len(snap.State.Sessions)).
		Msg("Restored from snapshot")
	
	// In Milestone 1, we don't have a tick engine yet, so we just log the recovery
	// In future milestones, this is where we would:
	// 1. Restore player data to the tick engine state
	// 2. Restore session data to the session manager
	// 3. Restore any other game state (ships, traders, etc.)
	
	// For now, we verify the snapshot data is accessible
	log.Debug().
		Int("player_count", len(snap.State.Players)).
		Int("session_count", len(snap.State.Sessions)).
		Msg("Snapshot state verified")
	
	return tickNumber, nil
}

// initSessionManager creates and initializes the session manager
func initSessionManager(database *db.Database) *session.SessionManager {
	log.Info().Msg("Initializing session manager")
	return session.NewSessionManager(database, log.Logger)
}

// startSSHServer creates and starts the SSH server
func startSSHServer(cfg *config.Config, sessionMgr *session.SessionManager, tickEngine *engine.TickEngine, registrar *registration.Registrar) (*sshserver.Server, error) {
	handshakeConfig := session.HandshakeConfig{
		ServerName:     "Black Sector",
		MOTD:           "Welcome. Watch your back out there.",
		TickIntervalMs: cfg.Server.TickIntervalMs,
	}

	sshCfg := sshserver.Config{
		Port:                 cfg.Server.SSHPort,
		MaxConcurrentPlayers: cfg.Server.MaxConcurrentPlayers,
		SessionManager:       sessionMgr,
		TickEngine:           tickEngine,
		Registrar:            registrar,
		Logger:               log.Logger,
		HandshakeConfig:      handshakeConfig,
	}

	server, err := sshserver.NewServer(sshCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH server: %w", err)
	}

	if err := server.Start(); err != nil {
		return nil, fmt.Errorf("failed to start SSH server: %w", err)
	}

	log.Info().
		Int("port", cfg.Server.SSHPort).
		Msg("SSH listener on port")

	return server, nil
}

// initTickEngine creates and initializes the tick engine
func initTickEngine(cfg *config.Config, database *db.Database, sessionMgr *session.SessionManager, tickNumber int64) *engine.TickEngine {
	log.Info().Msg("Initializing tick engine")

	// Load world configuration
	log.Info().Msg("Loading world configuration")
	worldGen := world.NewWorldGenerator(log.Logger)
	universe, err := worldGen.LoadWorld(cfg.Server.WorldConfigPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load world configuration")
		os.Exit(1)
	}
	
	// Validate world topology
	if err := worldGen.ValidateTopology(universe); err != nil {
		log.Error().Err(err).Msg("World topology validation failed")
		os.Exit(1)
	}
	log.Info().
		Int("systems", len(universe.Systems)).
		Int("ports", len(universe.Ports)).
		Int("connections", len(universe.JumpConnections)).
		Msg("World configuration loaded")

	// Initialize navigation subsystem
	log.Info().Msg("Initializing navigation subsystem")
	dbAdapter := &dbAdapter{db: database}
	navSystem := navigation.NewNavigationSystem(universe, dbAdapter, log.Logger)
	log.Info().Msg("Navigation subsystem initialized")

	// Initialize economy subsystem
	log.Info().Msg("Initializing economy subsystem")
	economyCfg := &economy.Config{
		LowSecPriceMultiplier: cfg.Economy.LowSecPriceMultiplier,
		BuyMarkup:             cfg.Economy.BuyMarkup,
		SellMarkdown:          cfg.Economy.SellMarkdown,
	}
	econSystem := economy.NewEconomySystem(economyCfg, database, log.Logger)
	
	// Load commodity definitions
	if err := econSystem.LoadCommodities(cfg.Server.WorldConfigPath); err != nil {
		log.Error().Err(err).Msg("Failed to load commodities for economy")
		os.Exit(1)
	}
	log.Info().Msg("Economy subsystem initialized")

	// Initialize combat subsystem
	log.Info().Msg("Initializing combat subsystem")
	combatCfg := &combat.Config{
		PirateActivityBase:   cfg.Universe.PirateActivityBase,
		SpawnCheckInterval:   cfg.Universe.PirateSpawnIntervalTicks,
		InsurancePayout:      cfg.Combat.InsurancePayout,
		SurrenderLossPercent: cfg.Universe.SurrenderLossPercent,
	}
	combatDBAdapter := &combatDBAdapter{db: database}
	combatSystem := combat.NewCombatSystem(combatCfg, combatDBAdapter, log.Logger)
	combatAdapter := &combatSystemAdapter{CombatSystem: combatSystem}
	log.Info().Msg("Combat subsystem initialized")

	// Initialize mission subsystem
	log.Info().Msg("Initializing mission subsystem")
	missionCfg := &missions.Config{}
	missionDBAdapter := &missionDBAdapter{db: database}
	missionSystem := missions.NewMissionManager(missionCfg, missionDBAdapter, log.Logger)
	missionAdapter := &missionSystemAdapter{MissionManager: missionSystem}
	
	// Load mission definitions
	if err := missionSystem.LoadMissions(cfg.Server.MissionsConfigDir); err != nil {
		log.Error().Err(err).Msg("Failed to load missions")
		os.Exit(1)
	}
	log.Info().Msg("Mission subsystem initialized")

	engineCfg := engine.Config{
		TickIntervalMs:        cfg.Server.TickIntervalMs,
		SnapshotIntervalTicks: cfg.Server.SnapshotIntervalTicks,
		ServerName:            "Black Sector",
		InitialTickNumber:     tickNumber,
	}

	tickEngine := engine.NewTickEngine(engineCfg, database, sessionMgr, navSystem, econSystem, combatAdapter, missionAdapter, log.Logger)
	tickEngine.Start()

	log.Info().
		Int64("initial_tick", tickNumber).
		Int("tick_interval_ms", cfg.Server.TickIntervalMs).
		Int("snapshot_interval_ticks", cfg.Server.SnapshotIntervalTicks).
		Msg("Tick engine initialized")

	return tickEngine
}

// initRegistrar creates and initializes the registration system
func initRegistrar(cfg *config.Config, database *db.Database, tickEngine *engine.TickEngine) *registration.Registrar {
	log.Info().Msg("Initializing registration system")

	// Get the origin system and port from the tick engine's world
	// For Phase 1, we use system 1 (Nexus Prime) and port 1 as the starting location
	originSystemID := 1
	originPortID := 1

	// Create registration database wrapper
	regDB := registration.NewDatabase(database.Conn(), log.Logger)

	// Create world wrapper with origin location
	regWorld := registration.NewWorld(originSystemID, originPortID)

	// Create config wrapper with starting credits and ship class
	startingCredits := 10000 // Default starting credits
	if cfg.Player.StartingCredits > 0 {
		startingCredits = cfg.Player.StartingCredits
	}
	registrationRateLimit := 3 // Default rate limit (3 per hour)
	if cfg.Security.RegistrationRateLimitPerHour > 0 {
		registrationRateLimit = cfg.Security.RegistrationRateLimitPerHour
	}
	regConfig := registration.NewConfig(startingCredits, "courier", registrationRateLimit)

	// Create registrar
	registrar := registration.NewRegistrar(regDB, regWorld, regConfig, log.Logger)

	log.Info().
		Int("starting_credits", startingCredits).
		Int("origin_system_id", originSystemID).
		Int("origin_port_id", originPortID).
		Msg("Registration system initialized")

	return registrar
}

// initAdminCLI creates and starts the admin CLI
func initAdminCLI(tickEngine *engine.TickEngine, sessionMgr *session.SessionManager) *admin.Handler {
	log.Info().Msg("Initializing admin CLI")

	cli := admin.NewCLI(os.Stdin, os.Stdout, log.Logger)
	cli.Start()

	handler := admin.NewHandler(cli, tickEngine, sessionMgr, log.Logger)
	handler.Start()

	log.Info().Msg("Admin CLI initialized")
	return handler
}


// performGracefulShutdown executes the graceful shutdown sequence
// Requirements: 12.1, 12.2, 12.3, 12.4, 12.5, 12.6, 12.7, 12.8, 12.9, 12.10
func performGracefulShutdown(tickEngine *engine.TickEngine, sshServer *sshserver.Server, sessionMgr *session.SessionManager, database *db.Database) {
	log.Info().Msg("Initiating graceful shutdown")

	// Step 1: Stop the tick engine (sets running flag to false)
	// Requirement 12.1: Set tick engine running flag to false
	tickEngine.Stop()

	// Step 2: Wait for current tick to complete
	// Requirement 12.2: Wait for current tick to complete
	// The tick loop will exit naturally after the current tick finishes
	log.Info().Msg("Waiting for current tick to complete")
	
	// Give the tick loop time to finish (up to 5 seconds)
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			log.Warn().Msg("Timeout waiting for tick loop to stop")
			goto continueShutdown
		case <-ticker.C:
			if !tickEngine.IsRunning() {
				log.Info().Msg("Tick loop stopped")
				goto continueShutdown
			}
		}
	}

continueShutdown:
	// Step 3: Create final snapshot synchronously
	// Requirements 12.3, 12.4: Create and write final snapshot synchronously
	log.Info().Msg("Creating final snapshot")
	if err := tickEngine.CreateFinalSnapshot(); err != nil {
		log.Error().Err(err).Msg("Failed to create final snapshot")
	} else {
		log.Info().Msg("Final snapshot created successfully")
	}

	// Step 4: Send server_shutdown messages to all active sessions
	// Requirement 12.5: Send server_shutdown messages to all active sessions
	log.Info().Msg("Sending shutdown messages to active sessions")
	if err := sessionMgr.SendShutdownMessages(); err != nil {
		log.Error().Err(err).Msg("Failed to send shutdown messages")
	}

	// Step 5: Close all SSH connections
	// Requirement 12.6: Close all SSH connections
	log.Info().Msg("Closing SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := sshServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown SSH server gracefully")
	} else {
		log.Info().Msg("SSH server closed")
	}

	// Step 6: Update all session states to TERMINATED
	// Requirement 12.7: Update all session states to TERMINATED
	log.Info().Msg("Terminating all sessions")
	if err := sessionMgr.TerminateAllSessions(); err != nil {
		log.Error().Err(err).Msg("Failed to terminate all sessions")
	}

	// Step 7: Close database connection
	// Requirement 12.8: Close database connection
	log.Info().Msg("Closing database connection")
	if err := database.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close database")
	} else {
		log.Info().Msg("Database closed")
	}

	// Requirement 12.9: Log "Server stopped" at INFO level
	// (This is done in main() after this function returns)
}
