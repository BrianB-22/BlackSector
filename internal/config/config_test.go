package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid config",
			content: `{
				"server": {
					"ssh_port": 2222,
					"tick_interval_ms": 2000,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 50,
					"db_path": "test.db",
					"world_config_path": "config/world/test.json",
					"missions_config_dir": "config/missions/"
				},
				"logging": {
					"level": "info",
					"log_file": "test.log",
					"debug_log_enabled": false,
					"debug_log_path": "debug.log",
					"debug_log_include_sql": false,
					"debug_log_include_tick_detail": true
				},
				"universe": {
					"universe_seed": 1701,
					"pirate_spawn_interval_ticks": 10,
					"pirate_activity_base": 0.30,
					"surrender_loss_percent": 40
				},
				"combat": {
					"courier_accuracy": 0.65,
					"shield_regen_per_tick": 5,
					"hull_repair_cost_per_point": 10,
					"flee_base_chance": 0.35,
					"flee_energy_bonus": 0.25,
					"missile_unit_price": 200,
					"insurance_payout": 5000
				},
				"economy": {
					"low_sec_price_multiplier": 1.18,
					"buy_markup": 1.1,
					"sell_markdown": 0.9
				},
				"banking": {
					"interest_apply_interval_ticks": 500
				},
				"player": {
					"starting_credits": 1000,
					"starting_ship_class": "courier",
					"starting_system_id": "nexus_prime"
				},
				"irn": {
					"base_delay_ticks": 5,
					"max_delay_ticks": 20,
					"broadcast_rate_limit_per_hour": 2
				},
				"security": {
					"registration_rate_limit_per_hour": 10,
					"auth_failure_rate_limit_per_minute": 5,
					"command_rate_limit_per_tick": 10
				}
			}`,
			wantErr: false,
		},
		{
			name: "invalid tick_interval_ms - zero",
			content: `{
				"server": {
					"ssh_port": 2222,
					"tick_interval_ms": 0,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 50,
					"db_path": "test.db",
					"world_config_path": "config/world/test.json",
					"missions_config_dir": "config/missions/"
				},
				"logging": {"level": "info", "log_file": "test.log"},
				"universe": {"universe_seed": 1701, "pirate_spawn_interval_ticks": 10, "pirate_activity_base": 0.3, "surrender_loss_percent": 40},
				"combat": {"courier_accuracy": 0.65, "shield_regen_per_tick": 5, "hull_repair_cost_per_point": 10, "flee_base_chance": 0.35, "flee_energy_bonus": 0.25, "missile_unit_price": 200, "insurance_payout": 5000},
				"economy": {"low_sec_price_multiplier": 1.18, "buy_markup": 1.1, "sell_markdown": 0.9},
				"banking": {"interest_apply_interval_ticks": 500},
				"player": {"starting_credits": 1000, "starting_ship_class": "courier", "starting_system_id": "nexus_prime"},
				"irn": {"base_delay_ticks": 5, "max_delay_ticks": 20, "broadcast_rate_limit_per_hour": 2},
				"security": {"registration_rate_limit_per_hour": 10, "auth_failure_rate_limit_per_minute": 5, "command_rate_limit_per_tick": 10}
			}`,
			wantErr: true,
			errMsg:  "tick_interval_ms must be between 100 and 10000",
		},
		{
			name: "invalid tick_interval_ms - too low",
			content: `{
				"server": {
					"ssh_port": 2222,
					"tick_interval_ms": 50,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 50,
					"db_path": "test.db",
					"world_config_path": "config/world/test.json",
					"missions_config_dir": "config/missions/"
				},
				"logging": {"level": "info", "log_file": "test.log"},
				"universe": {"universe_seed": 1701, "pirate_spawn_interval_ticks": 10, "pirate_activity_base": 0.3, "surrender_loss_percent": 40},
				"combat": {"courier_accuracy": 0.65, "shield_regen_per_tick": 5, "hull_repair_cost_per_point": 10, "flee_base_chance": 0.35, "flee_energy_bonus": 0.25, "missile_unit_price": 200, "insurance_payout": 5000},
				"economy": {"low_sec_price_multiplier": 1.18, "buy_markup": 1.1, "sell_markdown": 0.9},
				"banking": {"interest_apply_interval_ticks": 500},
				"player": {"starting_credits": 1000, "starting_ship_class": "courier", "starting_system_id": "nexus_prime"},
				"irn": {"base_delay_ticks": 5, "max_delay_ticks": 20, "broadcast_rate_limit_per_hour": 2},
				"security": {"registration_rate_limit_per_hour": 10, "auth_failure_rate_limit_per_minute": 5, "command_rate_limit_per_tick": 10}
			}`,
			wantErr: true,
			errMsg:  "tick_interval_ms must be between 100 and 10000",
		},
		{
			name: "invalid ssh_port - too low",
			content: `{
				"server": {
					"ssh_port": 80,
					"tick_interval_ms": 2000,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 50,
					"db_path": "test.db",
					"world_config_path": "config/world/test.json",
					"missions_config_dir": "config/missions/"
				},
				"logging": {"level": "info", "log_file": "test.log"},
				"universe": {"universe_seed": 1701, "pirate_spawn_interval_ticks": 10, "pirate_activity_base": 0.3, "surrender_loss_percent": 40},
				"combat": {"courier_accuracy": 0.65, "shield_regen_per_tick": 5, "hull_repair_cost_per_point": 10, "flee_base_chance": 0.35, "flee_energy_bonus": 0.25, "missile_unit_price": 200, "insurance_payout": 5000},
				"economy": {"low_sec_price_multiplier": 1.18, "buy_markup": 1.1, "sell_markdown": 0.9},
				"banking": {"interest_apply_interval_ticks": 500},
				"player": {"starting_credits": 1000, "starting_ship_class": "courier", "starting_system_id": "nexus_prime"},
				"irn": {"base_delay_ticks": 5, "max_delay_ticks": 20, "broadcast_rate_limit_per_hour": 2},
				"security": {"registration_rate_limit_per_hour": 10, "auth_failure_rate_limit_per_minute": 5, "command_rate_limit_per_tick": 10}
			}`,
			wantErr: true,
			errMsg:  "ssh_port must be between 1024 and 65535",
		},
		{
			name: "invalid max_concurrent_players - zero",
			content: `{
				"server": {
					"ssh_port": 2222,
					"tick_interval_ms": 2000,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 0,
					"db_path": "test.db",
					"world_config_path": "config/world/test.json",
					"missions_config_dir": "config/missions/"
				},
				"logging": {"level": "info", "log_file": "test.log"},
				"universe": {"universe_seed": 1701, "pirate_spawn_interval_ticks": 10, "pirate_activity_base": 0.3, "surrender_loss_percent": 40},
				"combat": {"courier_accuracy": 0.65, "shield_regen_per_tick": 5, "hull_repair_cost_per_point": 10, "flee_base_chance": 0.35, "flee_energy_bonus": 0.25, "missile_unit_price": 200, "insurance_payout": 5000},
				"economy": {"low_sec_price_multiplier": 1.18, "buy_markup": 1.1, "sell_markdown": 0.9},
				"banking": {"interest_apply_interval_ticks": 500},
				"player": {"starting_credits": 1000, "starting_ship_class": "courier", "starting_system_id": "nexus_prime"},
				"irn": {"base_delay_ticks": 5, "max_delay_ticks": 20, "broadcast_rate_limit_per_hour": 2},
				"security": {"registration_rate_limit_per_hour": 10, "auth_failure_rate_limit_per_minute": 5, "command_rate_limit_per_tick": 10}
			}`,
			wantErr: true,
			errMsg:  "max_concurrent_players must be between 1 and 1000",
		},
		{
			name: "invalid log_level",
			content: `{
				"server": {
					"ssh_port": 2222,
					"tick_interval_ms": 2000,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 50,
					"db_path": "test.db",
					"world_config_path": "config/world/test.json",
					"missions_config_dir": "config/missions/"
				},
				"logging": {"level": "invalid", "log_file": "test.log"},
				"universe": {"universe_seed": 1701, "pirate_spawn_interval_ticks": 10, "pirate_activity_base": 0.3, "surrender_loss_percent": 40},
				"combat": {"courier_accuracy": 0.65, "shield_regen_per_tick": 5, "hull_repair_cost_per_point": 10, "flee_base_chance": 0.35, "flee_energy_bonus": 0.25, "missile_unit_price": 200, "insurance_payout": 5000},
				"economy": {"low_sec_price_multiplier": 1.18, "buy_markup": 1.1, "sell_markdown": 0.9},
				"banking": {"interest_apply_interval_ticks": 500},
				"player": {"starting_credits": 1000, "starting_ship_class": "courier", "starting_system_id": "nexus_prime"},
				"irn": {"base_delay_ticks": 5, "max_delay_ticks": 20, "broadcast_rate_limit_per_hour": 2},
				"security": {"registration_rate_limit_per_hour": 10, "auth_failure_rate_limit_per_minute": 5, "command_rate_limit_per_tick": 10}
			}`,
			wantErr: true,
			errMsg:  "logging.level must be one of: debug, info, warn, error",
		},
		{
			name: "empty db_path",
			content: `{
				"server": {
					"ssh_port": 2222,
					"tick_interval_ms": 2000,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 50,
					"db_path": "",
					"world_config_path": "config/world/test.json",
					"missions_config_dir": "config/missions/"
				},
				"logging": {"level": "info", "log_file": "test.log"},
				"universe": {"universe_seed": 1701, "pirate_spawn_interval_ticks": 10, "pirate_activity_base": 0.3, "surrender_loss_percent": 40},
				"combat": {"courier_accuracy": 0.65, "shield_regen_per_tick": 5, "hull_repair_cost_per_point": 10, "flee_base_chance": 0.35, "flee_energy_bonus": 0.25, "missile_unit_price": 200, "insurance_payout": 5000},
				"economy": {"low_sec_price_multiplier": 1.18, "buy_markup": 1.1, "sell_markdown": 0.9},
				"banking": {"interest_apply_interval_ticks": 500},
				"player": {"starting_credits": 1000, "starting_ship_class": "courier", "starting_system_id": "nexus_prime"},
				"irn": {"base_delay_ticks": 5, "max_delay_ticks": 20, "broadcast_rate_limit_per_hour": 2},
				"security": {"registration_rate_limit_per_hour": 10, "auth_failure_rate_limit_per_minute": 5, "command_rate_limit_per_tick": 10}
			}`,
			wantErr: true,
			errMsg:  "db_path must be non-empty",
		},
		{
			name: "empty world_config_path",
			content: `{
				"server": {
					"ssh_port": 2222,
					"tick_interval_ms": 2000,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 50,
					"db_path": "test.db",
					"world_config_path": "",
					"missions_config_dir": "config/missions/"
				},
				"logging": {"level": "info", "log_file": "test.log"},
				"universe": {"universe_seed": 1701, "pirate_spawn_interval_ticks": 10, "pirate_activity_base": 0.3, "surrender_loss_percent": 40},
				"combat": {"courier_accuracy": 0.65, "shield_regen_per_tick": 5, "hull_repair_cost_per_point": 10, "flee_base_chance": 0.35, "flee_energy_bonus": 0.25, "missile_unit_price": 200, "insurance_payout": 5000},
				"economy": {"low_sec_price_multiplier": 1.18, "buy_markup": 1.1, "sell_markdown": 0.9},
				"banking": {"interest_apply_interval_ticks": 500},
				"player": {"starting_credits": 1000, "starting_ship_class": "courier", "starting_system_id": "nexus_prime"},
				"irn": {"base_delay_ticks": 5, "max_delay_ticks": 20, "broadcast_rate_limit_per_hour": 2},
				"security": {"registration_rate_limit_per_hour": 10, "auth_failure_rate_limit_per_minute": 5, "command_rate_limit_per_tick": 10}
			}`,
			wantErr: true,
			errMsg:  "world_config_path must be non-empty",
		},
		{
			name: "empty missions_config_dir",
			content: `{
				"server": {
					"ssh_port": 2222,
					"tick_interval_ms": 2000,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 50,
					"db_path": "test.db",
					"world_config_path": "config/world/test.json",
					"missions_config_dir": ""
				},
				"logging": {"level": "info", "log_file": "test.log"},
				"universe": {"universe_seed": 1701, "pirate_spawn_interval_ticks": 10, "pirate_activity_base": 0.3, "surrender_loss_percent": 40},
				"combat": {"courier_accuracy": 0.65, "shield_regen_per_tick": 5, "hull_repair_cost_per_point": 10, "flee_base_chance": 0.35, "flee_energy_bonus": 0.25, "missile_unit_price": 200, "insurance_payout": 5000},
				"economy": {"low_sec_price_multiplier": 1.18, "buy_markup": 1.1, "sell_markdown": 0.9},
				"banking": {"interest_apply_interval_ticks": 500},
				"player": {"starting_credits": 1000, "starting_ship_class": "courier", "starting_system_id": "nexus_prime"},
				"irn": {"base_delay_ticks": 5, "max_delay_ticks": 20, "broadcast_rate_limit_per_hour": 2},
				"security": {"registration_rate_limit_per_hour": 10, "auth_failure_rate_limit_per_minute": 5, "command_rate_limit_per_tick": 10}
			}`,
			wantErr: true,
			errMsg:  "missions_config_dir must be non-empty",
		},
		{
			name: "invalid pirate_activity_base - too high",
			content: `{
				"server": {
					"ssh_port": 2222,
					"tick_interval_ms": 2000,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 50,
					"db_path": "test.db",
					"world_config_path": "config/world/test.json",
					"missions_config_dir": "config/missions/"
				},
				"logging": {"level": "info", "log_file": "test.log"},
				"universe": {"universe_seed": 1701, "pirate_spawn_interval_ticks": 10, "pirate_activity_base": 1.5, "surrender_loss_percent": 40},
				"combat": {"courier_accuracy": 0.65, "shield_regen_per_tick": 5, "hull_repair_cost_per_point": 10, "flee_base_chance": 0.35, "flee_energy_bonus": 0.25, "missile_unit_price": 200, "insurance_payout": 5000},
				"economy": {"low_sec_price_multiplier": 1.18, "buy_markup": 1.1, "sell_markdown": 0.9},
				"banking": {"interest_apply_interval_ticks": 500},
				"player": {"starting_credits": 1000, "starting_ship_class": "courier", "starting_system_id": "nexus_prime"},
				"irn": {"base_delay_ticks": 5, "max_delay_ticks": 20, "broadcast_rate_limit_per_hour": 2},
				"security": {"registration_rate_limit_per_hour": 10, "auth_failure_rate_limit_per_minute": 5, "command_rate_limit_per_tick": 10}
			}`,
			wantErr: true,
			errMsg:  "pirate_activity_base must be between 0.0 and 1.0",
		},
		{
			name: "invalid surrender_loss_percent - too high",
			content: `{
				"server": {
					"ssh_port": 2222,
					"tick_interval_ms": 2000,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 50,
					"db_path": "test.db",
					"world_config_path": "config/world/test.json",
					"missions_config_dir": "config/missions/"
				},
				"logging": {"level": "info", "log_file": "test.log"},
				"universe": {"universe_seed": 1701, "pirate_spawn_interval_ticks": 10, "pirate_activity_base": 0.3, "surrender_loss_percent": 150},
				"combat": {"courier_accuracy": 0.65, "shield_regen_per_tick": 5, "hull_repair_cost_per_point": 10, "flee_base_chance": 0.35, "flee_energy_bonus": 0.25, "missile_unit_price": 200, "insurance_payout": 5000},
				"economy": {"low_sec_price_multiplier": 1.18, "buy_markup": 1.1, "sell_markdown": 0.9},
				"banking": {"interest_apply_interval_ticks": 500},
				"player": {"starting_credits": 1000, "starting_ship_class": "courier", "starting_system_id": "nexus_prime"},
				"irn": {"base_delay_ticks": 5, "max_delay_ticks": 20, "broadcast_rate_limit_per_hour": 2},
				"security": {"registration_rate_limit_per_hour": 10, "auth_failure_rate_limit_per_minute": 5, "command_rate_limit_per_tick": 10}
			}`,
			wantErr: true,
			errMsg:  "surrender_loss_percent must be between 0 and 100",
		},
		{
			name: "invalid courier_accuracy - too high",
			content: `{
				"server": {
					"ssh_port": 2222,
					"tick_interval_ms": 2000,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 50,
					"db_path": "test.db",
					"world_config_path": "config/world/test.json",
					"missions_config_dir": "config/missions/"
				},
				"logging": {"level": "info", "log_file": "test.log"},
				"universe": {"universe_seed": 1701, "pirate_spawn_interval_ticks": 10, "pirate_activity_base": 0.3, "surrender_loss_percent": 40},
				"combat": {"courier_accuracy": 1.5, "shield_regen_per_tick": 5, "hull_repair_cost_per_point": 10, "flee_base_chance": 0.35, "flee_energy_bonus": 0.25, "missile_unit_price": 200, "insurance_payout": 5000},
				"economy": {"low_sec_price_multiplier": 1.18, "buy_markup": 1.1, "sell_markdown": 0.9},
				"banking": {"interest_apply_interval_ticks": 500},
				"player": {"starting_credits": 1000, "starting_ship_class": "courier", "starting_system_id": "nexus_prime"},
				"irn": {"base_delay_ticks": 5, "max_delay_ticks": 20, "broadcast_rate_limit_per_hour": 2},
				"security": {"registration_rate_limit_per_hour": 10, "auth_failure_rate_limit_per_minute": 5, "command_rate_limit_per_tick": 10}
			}`,
			wantErr: true,
			errMsg:  "courier_accuracy must be between 0.0 and 1.0",
		},
		{
			name: "negative economy multiplier",
			content: `{
				"server": {
					"ssh_port": 2222,
					"tick_interval_ms": 2000,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 50,
					"db_path": "test.db",
					"world_config_path": "config/world/test.json",
					"missions_config_dir": "config/missions/"
				},
				"logging": {"level": "info", "log_file": "test.log"},
				"universe": {"universe_seed": 1701, "pirate_spawn_interval_ticks": 10, "pirate_activity_base": 0.3, "surrender_loss_percent": 40},
				"combat": {"courier_accuracy": 0.65, "shield_regen_per_tick": 5, "hull_repair_cost_per_point": 10, "flee_base_chance": 0.35, "flee_energy_bonus": 0.25, "missile_unit_price": 200, "insurance_payout": 5000},
				"economy": {"low_sec_price_multiplier": -1.18, "buy_markup": 1.1, "sell_markdown": 0.9},
				"banking": {"interest_apply_interval_ticks": 500},
				"player": {"starting_credits": 1000, "starting_ship_class": "courier", "starting_system_id": "nexus_prime"},
				"irn": {"base_delay_ticks": 5, "max_delay_ticks": 20, "broadcast_rate_limit_per_hour": 2},
				"security": {"registration_rate_limit_per_hour": 10, "auth_failure_rate_limit_per_minute": 5, "command_rate_limit_per_tick": 10}
			}`,
			wantErr: true,
			errMsg:  "low_sec_price_multiplier must be positive",
		},
		{
			name: "empty starting_ship_class",
			content: `{
				"server": {
					"ssh_port": 2222,
					"tick_interval_ms": 2000,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 50,
					"db_path": "test.db",
					"world_config_path": "config/world/test.json",
					"missions_config_dir": "config/missions/"
				},
				"logging": {"level": "info", "log_file": "test.log"},
				"universe": {"universe_seed": 1701, "pirate_spawn_interval_ticks": 10, "pirate_activity_base": 0.3, "surrender_loss_percent": 40},
				"combat": {"courier_accuracy": 0.65, "shield_regen_per_tick": 5, "hull_repair_cost_per_point": 10, "flee_base_chance": 0.35, "flee_energy_bonus": 0.25, "missile_unit_price": 200, "insurance_payout": 5000},
				"economy": {"low_sec_price_multiplier": 1.18, "buy_markup": 1.1, "sell_markdown": 0.9},
				"banking": {"interest_apply_interval_ticks": 500},
				"player": {"starting_credits": 1000, "starting_ship_class": "", "starting_system_id": "nexus_prime"},
				"irn": {"base_delay_ticks": 5, "max_delay_ticks": 20, "broadcast_rate_limit_per_hour": 2},
				"security": {"registration_rate_limit_per_hour": 10, "auth_failure_rate_limit_per_minute": 5, "command_rate_limit_per_tick": 10}
			}`,
			wantErr: true,
			errMsg:  "starting_ship_class must be non-empty",
		},
		{
			name: "invalid IRN max_delay less than base_delay",
			content: `{
				"server": {
					"ssh_port": 2222,
					"tick_interval_ms": 2000,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 50,
					"db_path": "test.db",
					"world_config_path": "config/world/test.json",
					"missions_config_dir": "config/missions/"
				},
				"logging": {"level": "info", "log_file": "test.log"},
				"universe": {"universe_seed": 1701, "pirate_spawn_interval_ticks": 10, "pirate_activity_base": 0.3, "surrender_loss_percent": 40},
				"combat": {"courier_accuracy": 0.65, "shield_regen_per_tick": 5, "hull_repair_cost_per_point": 10, "flee_base_chance": 0.35, "flee_energy_bonus": 0.25, "missile_unit_price": 200, "insurance_payout": 5000},
				"economy": {"low_sec_price_multiplier": 1.18, "buy_markup": 1.1, "sell_markdown": 0.9},
				"banking": {"interest_apply_interval_ticks": 500},
				"player": {"starting_credits": 1000, "starting_ship_class": "courier", "starting_system_id": "nexus_prime"},
				"irn": {"base_delay_ticks": 20, "max_delay_ticks": 5, "broadcast_rate_limit_per_hour": 2},
				"security": {"registration_rate_limit_per_hour": 10, "auth_failure_rate_limit_per_minute": 5, "command_rate_limit_per_tick": 10}
			}`,
			wantErr: true,
			errMsg:  "max_delay_ticks must be >= base_delay_ticks",
		},
		{
			name: "invalid security rate limit - zero",
			content: `{
				"server": {
					"ssh_port": 2222,
					"tick_interval_ms": 2000,
					"snapshot_interval_ticks": 100,
					"max_concurrent_players": 50,
					"db_path": "test.db",
					"world_config_path": "config/world/test.json",
					"missions_config_dir": "config/missions/"
				},
				"logging": {"level": "info", "log_file": "test.log"},
				"universe": {"universe_seed": 1701, "pirate_spawn_interval_ticks": 10, "pirate_activity_base": 0.3, "surrender_loss_percent": 40},
				"combat": {"courier_accuracy": 0.65, "shield_regen_per_tick": 5, "hull_repair_cost_per_point": 10, "flee_base_chance": 0.35, "flee_energy_bonus": 0.25, "missile_unit_price": 200, "insurance_payout": 5000},
				"economy": {"low_sec_price_multiplier": 1.18, "buy_markup": 1.1, "sell_markdown": 0.9},
				"banking": {"interest_apply_interval_ticks": 500},
				"player": {"starting_credits": 1000, "starting_ship_class": "courier", "starting_system_id": "nexus_prime"},
				"irn": {"base_delay_ticks": 5, "max_delay_ticks": 20, "broadcast_rate_limit_per_hour": 2},
				"security": {"registration_rate_limit_per_hour": 0, "auth_failure_rate_limit_per_minute": 5, "command_rate_limit_per_tick": 10}
			}`,
			wantErr: true,
			errMsg:  "registration_rate_limit_per_hour must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "test_config.json")
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			// Load config
			cfg, err := LoadConfig(configPath)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				assert.Equal(t, 2222, cfg.Server.SSHPort)
				assert.Equal(t, 2000, cfg.Server.TickIntervalMs)
			}
		})
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/config.json")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")
	err := os.WriteFile(configPath, []byte("not valid json {{{"), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to parse config JSON")
}
