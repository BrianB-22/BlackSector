package registration

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

const (
	// Bcrypt cost factors per requirements
	PasswordCost = 12
	TokenCost    = 10

	// Token size in bytes (32 bytes = 256 bits)
	TokenBytes = 32

	// Validation constraints
	MinDisplayNameLength = 3
	MaxDisplayNameLength = 20
	MinPasswordLength    = 8
)

// Registrar handles player registration and authentication
type Registrar struct {
	db          *Database
	world       *World
	config      *Config
	rateLimiter *RateLimiter
	logger      zerolog.Logger
}

// Database wraps database operations needed for registration
type Database struct {
	conn   *sql.DB
	logger zerolog.Logger
}

// World wraps world data needed for registration
type World struct {
	originSystemID int
	originPortID   int
}

// Config wraps configuration values
type Config struct {
	startingCredits              int
	startingShipClass            string
	registrationRateLimitPerHour int
}

// Player represents a registered player account
type Player struct {
	PlayerID     string
	PlayerName   string
	SSHUsername  string
	TokenHash    string
	PasswordHash string
	Credits      int64
	CreatedAt    int64
	LastLoginAt  *int64
	IsBanned     bool
}

// Ship represents a player's vessel
type Ship struct {
	ShipID           string
	PlayerID         string
	ShipClass        string
	HullPoints       int
	MaxHullPoints    int
	ShieldPoints     int
	MaxShieldPoints  int
	EnergyPoints     int
	MaxEnergyPoints  int
	CargoCapacity    int
	MissilesCurrent  int
	CurrentSystemID  int
	PositionX        float64
	PositionY        float64
	Status           string
	DockedAtPortID   *int
	LastUpdatedTick  int64
}

// RegistrationRequest contains data for new player registration
type RegistrationRequest struct {
	SSHUsername string
	DisplayName string
	Password    string
	RemoteAddr  string
}

// RegistrationResult contains the result of successful registration
type RegistrationResult struct {
	PlayerID     string
	PlayerToken  string // Plaintext token shown once
	StartingShip *Ship
}

// NewRegistrar creates a new registration handler
func NewRegistrar(db *Database, world *World, config *Config, logger zerolog.Logger) *Registrar {
	rateLimiter := NewRateLimiter(config.registrationRateLimitPerHour, logger)
	return &Registrar{
		db:          db,
		world:       world,
		config:      config,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// NewDatabase creates a database wrapper for registration
func NewDatabase(conn *sql.DB, logger zerolog.Logger) *Database {
	return &Database{
		conn:   conn,
		logger: logger,
	}
}

// NewWorld creates a world wrapper for registration
func NewWorld(originSystemID int, originPortID int) *World {
	return &World{
		originSystemID: originSystemID,
		originPortID:   originPortID,
	}
}

// NewConfig creates a config wrapper for registration
func NewConfig(startingCredits int, startingShipClass string, registrationRateLimitPerHour int) *Config {
	return &Config{
		startingCredits:              startingCredits,
		startingShipClass:            startingShipClass,
		registrationRateLimitPerHour: registrationRateLimitPerHour,
	}
}

// CheckPlayerExists verifies if a player exists by SSH username
func (r *Registrar) CheckPlayerExists(sshUsername string) (*Player, error) {
	player, err := r.db.GetPlayerBySSHUsername(sshUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to check player existence: %w", err)
	}
	return player, nil
}

// ValidateRegistrationRequest validates registration input
func (r *Registrar) ValidateRegistrationRequest(req *RegistrationRequest) error {
	// Validate display name
	if len(req.DisplayName) < MinDisplayNameLength {
		return fmt.Errorf("display name must be at least %d characters", MinDisplayNameLength)
	}
	if len(req.DisplayName) > MaxDisplayNameLength {
		return fmt.Errorf("display name must be at most %d characters", MaxDisplayNameLength)
	}

	// Validate password
	if len(req.Password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}

	// Validate SSH username
	if req.SSHUsername == "" {
		return fmt.Errorf("SSH username is required")
	}

	return nil
}

// GeneratePlayerToken generates a cryptographically random token
func GeneratePlayerToken() (string, error) {
	tokenBytes := make([]byte, TokenBytes)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode as base64url (URL-safe base64)
	token := base64.URLEncoding.EncodeToString(tokenBytes)
	return token, nil
}

// HashPassword hashes a password using bcrypt with cost factor 12
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), PasswordCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// HashToken hashes a token using bcrypt with cost factor 10
func HashToken(token string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), TokenCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash token: %w", err)
	}
	return string(hash), nil
}

// ValidatePassword checks if a password matches its hash
func ValidatePassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ValidateToken checks if a token matches its hash
func ValidateToken(token, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(token))
}

// RegisterNewPlayer creates a new player account and starting ship
func (r *Registrar) RegisterNewPlayer(req *RegistrationRequest) (*RegistrationResult, error) {
	// Check rate limit first
	if err := r.rateLimiter.CheckAndRecord(req.RemoteAddr); err != nil {
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}

	// Validate request
	if err := r.ValidateRegistrationRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if player already exists
	existing, err := r.CheckPlayerExists(req.SSHUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing player: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("player already exists with SSH username: %s", req.SSHUsername)
	}

	// Generate player ID
	playerID := uuid.New().String()

	// Generate and hash token
	token, err := GeneratePlayerToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	tokenHash, err := HashToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to hash token: %w", err)
	}

	// Hash password
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create player record
	player := &Player{
		PlayerID:     playerID,
		PlayerName:   req.DisplayName,
		SSHUsername:  req.SSHUsername,
		TokenHash:    tokenHash,
		PasswordHash: passwordHash,
		Credits:      int64(r.config.startingCredits),
		CreatedAt:    time.Now().Unix(),
		IsBanned:     false,
	}

	// Provision starting ship
	ship := r.ProvisionStarterShip(playerID)

	// Begin transaction
	tx, err := r.db.conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Insert player
	if err := r.db.TxInsertPlayer(tx, player); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to insert player: %w", err)
	}

	// Insert ship
	if err := r.db.TxInsertShip(tx, ship); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to insert ship: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info().
		Str("player_id", playerID).
		Str("player_name", req.DisplayName).
		Str("ssh_username", req.SSHUsername).
		Str("remote_addr", req.RemoteAddr).
		Msg("New player registered")

	return &RegistrationResult{
		PlayerID:     playerID,
		PlayerToken:  token, // Return plaintext token for one-time display
		StartingShip: ship,
	}, nil
}

// ProvisionStarterShip creates a courier-class ship at Federated Space origin
func (r *Registrar) ProvisionStarterShip(playerID string) *Ship {
	// Get ship class stats
	shipClass := r.config.startingShipClass
	stats := GetShipClassStats(shipClass)

	// Create ship docked at origin port
	portID := r.world.originPortID
	ship := &Ship{
		ShipID:          uuid.New().String(),
		PlayerID:        playerID,
		ShipClass:       shipClass,
		HullPoints:      stats.MaxHull,
		MaxHullPoints:   stats.MaxHull,
		ShieldPoints:    stats.MaxShield,
		MaxShieldPoints: stats.MaxShield,
		EnergyPoints:    stats.MaxEnergy,
		MaxEnergyPoints: stats.MaxEnergy,
		CargoCapacity:   stats.CargoCapacity,
		MissilesCurrent: 0,
		CurrentSystemID: r.world.originSystemID,
		PositionX:       0.0,
		PositionY:       0.0,
		Status:          "DOCKED",
		DockedAtPortID:  &portID,
		LastUpdatedTick: 0,
	}

	r.logger.Debug().
		Str("ship_id", ship.ShipID).
		Str("player_id", playerID).
		Str("ship_class", shipClass).
		Int("system_id", r.world.originSystemID).
		Int("port_id", portID).
		Msg("Starter ship provisioned")

	return ship
}

// ShipClassStats defines stats for a ship class
type ShipClassStats struct {
	MaxHull       int
	MaxShield     int
	MaxEnergy     int
	CargoCapacity int
	WeaponDamage  int
}

// GetShipClassStats returns stats for a ship class
func GetShipClassStats(shipClass string) ShipClassStats {
	// Phase 1: Only courier class
	// REQ-SHIP-002: Courier ships SHALL have: Hull=100, Shield=50, Energy=100, Cargo=20, WeaponDamage=15
	switch shipClass {
	case "courier":
		return ShipClassStats{
			MaxHull:       100,
			MaxShield:     50,
			MaxEnergy:     100,
			CargoCapacity: 20,
			WeaponDamage:  15,
		}
	default:
		// Default to courier
		return ShipClassStats{
			MaxHull:       100,
			MaxShield:     50,
			MaxEnergy:     100,
			CargoCapacity: 20,
			WeaponDamage:  15,
		}
	}
}

// GetPlayerBySSHUsername retrieves a player by SSH username
func (db *Database) GetPlayerBySSHUsername(sshUsername string) (*Player, error) {
	query := `
		SELECT player_id, player_name, ssh_username, token_hash, password_hash, credits, created_at, last_login_at, is_banned
		FROM players
		WHERE ssh_username = ?
	`

	var player Player
	var sshUser sql.NullString
	var passwordHash sql.NullString
	var lastLoginAt sql.NullInt64
	var isBanned int

	err := db.conn.QueryRow(query, sshUsername).Scan(
		&player.PlayerID,
		&player.PlayerName,
		&sshUser,
		&player.TokenHash,
		&passwordHash,
		&player.Credits,
		&player.CreatedAt,
		&lastLoginAt,
		&isBanned,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Player not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query player by SSH username: %w", err)
	}

	if sshUser.Valid {
		player.SSHUsername = sshUser.String
	}
	if passwordHash.Valid {
		player.PasswordHash = passwordHash.String
	}
	if lastLoginAt.Valid {
		player.LastLoginAt = &lastLoginAt.Int64
	}
	player.IsBanned = isBanned != 0

	return &player, nil
}

// TxInsertPlayer inserts a player within a transaction
func (db *Database) TxInsertPlayer(tx *sql.Tx, player *Player) error {
	query := `
		INSERT INTO players (
			player_id, player_name, ssh_username, token_hash, password_hash, credits, created_at, last_login_at, is_banned
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	isBanned := 0
	if player.IsBanned {
		isBanned = 1
	}

	_, err := tx.Exec(
		query,
		player.PlayerID,
		player.PlayerName,
		player.SSHUsername,
		player.TokenHash,
		player.PasswordHash,
		player.Credits,
		player.CreatedAt,
		player.LastLoginAt,
		isBanned,
	)

	if err != nil {
		return fmt.Errorf("failed to insert player: %w", err)
	}

	return nil
}

// TxInsertShip inserts a ship within a transaction
func (db *Database) TxInsertShip(tx *sql.Tx, ship *Ship) error {
	query := `
		INSERT INTO ships (
			ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var dockedAtPortID interface{}
	if ship.DockedAtPortID != nil {
		dockedAtPortID = *ship.DockedAtPortID
	} else {
		dockedAtPortID = nil
	}

	_, err := tx.Exec(
		query,
		ship.ShipID,
		ship.PlayerID,
		ship.ShipClass,
		ship.HullPoints,
		ship.MaxHullPoints,
		ship.ShieldPoints,
		ship.MaxShieldPoints,
		ship.EnergyPoints,
		ship.MaxEnergyPoints,
		ship.CargoCapacity,
		ship.MissilesCurrent,
		ship.CurrentSystemID,
		ship.PositionX,
		ship.PositionY,
		ship.Status,
		dockedAtPortID,
		ship.LastUpdatedTick,
	)

	if err != nil {
		return fmt.Errorf("failed to insert ship: %w", err)
	}

	return nil
}

// Stop stops the registrar and cleans up resources
func (r *Registrar) Stop() {
	if r.rateLimiter != nil {
		r.rateLimiter.Stop()
	}
}
