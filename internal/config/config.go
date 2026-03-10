package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the complete server configuration from server.json
type Config struct {
	Server   ServerConfig   `json:"server"`
	Logging  LoggingConfig  `json:"logging"`
	Universe UniverseConfig `json:"universe"`
	Combat   CombatConfig   `json:"combat"`
	Economy  EconomyConfig  `json:"economy"`
	Banking  BankingConfig  `json:"banking"`
	Player   PlayerConfig   `json:"player"`
	IRN      IRNConfig      `json:"irn"`
	Security SecurityConfig `json:"security"`
}

// ServerConfig contains core server settings
type ServerConfig struct {
	SSHPort               int    `json:"ssh_port"`
	TickIntervalMs        int    `json:"tick_interval_ms"`
	SnapshotIntervalTicks int    `json:"snapshot_interval_ticks"`
	MaxConcurrentPlayers  int    `json:"max_concurrent_players"`
	DBPath                string `json:"db_path"`
	WorldConfigPath       string `json:"world_config_path"`
	MissionsConfigDir     string `json:"missions_config_dir"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level                    string `json:"level"`
	LogFile                  string `json:"log_file"`
	DebugEnabled             bool   `json:"debug_log_enabled"`
	DebugLogPath             string `json:"debug_log_path"`
	DebugLogIncludeSQL       bool   `json:"debug_log_include_sql"`
	DebugLogIncludeTickDetail bool  `json:"debug_log_include_tick_detail"`
}

// UniverseConfig contains universe generation settings
type UniverseConfig struct {
	UniverseSeed              int64   `json:"universe_seed"`
	PirateSpawnIntervalTicks  int     `json:"pirate_spawn_interval_ticks"`
	PirateActivityBase        float64 `json:"pirate_activity_base"`
	SurrenderLossPercent      int     `json:"surrender_loss_percent"`
}

// CombatConfig contains combat system settings
type CombatConfig struct {
	CourierAccuracy         float64 `json:"courier_accuracy"`
	ShieldRegenPerTick      int     `json:"shield_regen_per_tick"`
	HullRepairCostPerPoint  int     `json:"hull_repair_cost_per_point"`
	FleeBaseChance          float64 `json:"flee_base_chance"`
	FleeEnergyBonus         float64 `json:"flee_energy_bonus"`
	MissileUnitPrice        int     `json:"missile_unit_price"`
	InsurancePayout         int     `json:"insurance_payout"`
}

// EconomyConfig contains economy settings
type EconomyConfig struct {
	LowSecPriceMultiplier float64 `json:"low_sec_price_multiplier"`
	BuyMarkup             float64 `json:"buy_markup"`
	SellMarkdown          float64 `json:"sell_markdown"`
}

// BankingConfig contains banking system settings
type BankingConfig struct {
	InterestApplyIntervalTicks int `json:"interest_apply_interval_ticks"`
}

// PlayerConfig contains player initialization settings
type PlayerConfig struct {
	StartingCredits   int    `json:"starting_credits"`
	StartingShipClass string `json:"starting_ship_class"`
	StartingSystemID  string `json:"starting_system_id"`
}

// IRNConfig contains Interstellar Relay Network settings
type IRNConfig struct {
	BaseDelayTicks           int `json:"base_delay_ticks"`
	MaxDelayTicks            int `json:"max_delay_ticks"`
	BroadcastRateLimitPerHour int `json:"broadcast_rate_limit_per_hour"`
}

// SecurityConfig contains security and rate limiting settings
type SecurityConfig struct {
	RegistrationRateLimitPerHour    int `json:"registration_rate_limit_per_hour"`
	AuthFailureRateLimitPerMinute   int `json:"auth_failure_rate_limit_per_minute"`
	CommandRateLimitPerTick         int `json:"command_rate_limit_per_tick"`
}

// LoadConfig loads and validates the server configuration from the specified path
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate checks that all configuration values are within acceptable ranges
func (c *Config) Validate() error {
	// Server validation
	if c.Server.SSHPort < 1024 || c.Server.SSHPort > 65535 {
		return fmt.Errorf("server.ssh_port must be between 1024 and 65535, got %d", c.Server.SSHPort)
	}
	if c.Server.TickIntervalMs < 100 || c.Server.TickIntervalMs > 10000 {
		return fmt.Errorf("server.tick_interval_ms must be between 100 and 10000, got %d", c.Server.TickIntervalMs)
	}
	if c.Server.SnapshotIntervalTicks <= 0 {
		return fmt.Errorf("server.snapshot_interval_ticks must be greater than 0, got %d", c.Server.SnapshotIntervalTicks)
	}
	if c.Server.MaxConcurrentPlayers < 1 || c.Server.MaxConcurrentPlayers > 1000 {
		return fmt.Errorf("server.max_concurrent_players must be between 1 and 1000, got %d", c.Server.MaxConcurrentPlayers)
	}
	if c.Server.DBPath == "" {
		return fmt.Errorf("server.db_path must be non-empty")
	}
	if c.Server.WorldConfigPath == "" {
		return fmt.Errorf("server.world_config_path must be non-empty")
	}
	if c.Server.MissionsConfigDir == "" {
		return fmt.Errorf("server.missions_config_dir must be non-empty")
	}

	// Logging validation
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("logging.level must be one of: debug, info, warn, error; got %s", c.Logging.Level)
	}
	if c.Logging.LogFile == "" {
		return fmt.Errorf("logging.log_file must be non-empty")
	}
	if c.Logging.DebugEnabled && c.Logging.DebugLogPath == "" {
		return fmt.Errorf("logging.debug_log_path must be non-empty when debug_log_enabled is true")
	}

	// Universe validation
	if c.Universe.PirateSpawnIntervalTicks <= 0 {
		return fmt.Errorf("universe.pirate_spawn_interval_ticks must be greater than 0, got %d", c.Universe.PirateSpawnIntervalTicks)
	}
	if c.Universe.PirateActivityBase < 0.0 || c.Universe.PirateActivityBase > 1.0 {
		return fmt.Errorf("universe.pirate_activity_base must be between 0.0 and 1.0, got %f", c.Universe.PirateActivityBase)
	}
	if c.Universe.SurrenderLossPercent < 0 || c.Universe.SurrenderLossPercent > 100 {
		return fmt.Errorf("universe.surrender_loss_percent must be between 0 and 100, got %d", c.Universe.SurrenderLossPercent)
	}

	// Combat validation
	if c.Combat.CourierAccuracy < 0.0 || c.Combat.CourierAccuracy > 1.0 {
		return fmt.Errorf("combat.courier_accuracy must be between 0.0 and 1.0, got %f", c.Combat.CourierAccuracy)
	}
	if c.Combat.ShieldRegenPerTick < 0 {
		return fmt.Errorf("combat.shield_regen_per_tick must be non-negative, got %d", c.Combat.ShieldRegenPerTick)
	}
	if c.Combat.HullRepairCostPerPoint < 0 {
		return fmt.Errorf("combat.hull_repair_cost_per_point must be non-negative, got %d", c.Combat.HullRepairCostPerPoint)
	}
	if c.Combat.FleeBaseChance < 0.0 || c.Combat.FleeBaseChance > 1.0 {
		return fmt.Errorf("combat.flee_base_chance must be between 0.0 and 1.0, got %f", c.Combat.FleeBaseChance)
	}
	if c.Combat.FleeEnergyBonus < 0.0 {
		return fmt.Errorf("combat.flee_energy_bonus must be non-negative, got %f", c.Combat.FleeEnergyBonus)
	}
	if c.Combat.MissileUnitPrice < 0 {
		return fmt.Errorf("combat.missile_unit_price must be non-negative, got %d", c.Combat.MissileUnitPrice)
	}
	if c.Combat.InsurancePayout < 0 {
		return fmt.Errorf("combat.insurance_payout must be non-negative, got %d", c.Combat.InsurancePayout)
	}

	// Economy validation
	if c.Economy.LowSecPriceMultiplier <= 0.0 {
		return fmt.Errorf("economy.low_sec_price_multiplier must be positive, got %f", c.Economy.LowSecPriceMultiplier)
	}
	if c.Economy.BuyMarkup <= 0.0 {
		return fmt.Errorf("economy.buy_markup must be positive, got %f", c.Economy.BuyMarkup)
	}
	if c.Economy.SellMarkdown <= 0.0 {
		return fmt.Errorf("economy.sell_markdown must be positive, got %f", c.Economy.SellMarkdown)
	}

	// Banking validation
	if c.Banking.InterestApplyIntervalTicks <= 0 {
		return fmt.Errorf("banking.interest_apply_interval_ticks must be greater than 0, got %d", c.Banking.InterestApplyIntervalTicks)
	}

	// Player validation
	if c.Player.StartingCredits < 0 {
		return fmt.Errorf("player.starting_credits must be non-negative, got %d", c.Player.StartingCredits)
	}
	if c.Player.StartingShipClass == "" {
		return fmt.Errorf("player.starting_ship_class must be non-empty")
	}
	if c.Player.StartingSystemID == "" {
		return fmt.Errorf("player.starting_system_id must be non-empty")
	}

	// IRN validation
	if c.IRN.BaseDelayTicks < 0 {
		return fmt.Errorf("irn.base_delay_ticks must be non-negative, got %d", c.IRN.BaseDelayTicks)
	}
	if c.IRN.MaxDelayTicks < c.IRN.BaseDelayTicks {
		return fmt.Errorf("irn.max_delay_ticks must be >= base_delay_ticks, got max=%d base=%d", c.IRN.MaxDelayTicks, c.IRN.BaseDelayTicks)
	}
	if c.IRN.BroadcastRateLimitPerHour <= 0 {
		return fmt.Errorf("irn.broadcast_rate_limit_per_hour must be greater than 0, got %d", c.IRN.BroadcastRateLimitPerHour)
	}

	// Security validation
	if c.Security.RegistrationRateLimitPerHour <= 0 {
		return fmt.Errorf("security.registration_rate_limit_per_hour must be greater than 0, got %d", c.Security.RegistrationRateLimitPerHour)
	}
	if c.Security.AuthFailureRateLimitPerMinute <= 0 {
		return fmt.Errorf("security.auth_failure_rate_limit_per_minute must be greater than 0, got %d", c.Security.AuthFailureRateLimitPerMinute)
	}
	if c.Security.CommandRateLimitPerTick <= 0 {
		return fmt.Errorf("security.command_rate_limit_per_tick must be greater than 0, got %d", c.Security.CommandRateLimitPerTick)
	}

	return nil
}
