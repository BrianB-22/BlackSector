package db

import (
	"database/sql"
	"fmt"
)

// CombatInstance represents an active combat encounter
type CombatInstance struct {
	CombatID       string
	PlayerShipID   string
	PirateShipID   string
	SystemID       int
	StartTick      int64
	Status         string
	TurnNumber     int
}

// CreateCombatInstance creates a new combat instance
func (db *Database) CreateCombatInstance(combatID, playerShipID, pirateShipID string, systemID int, startTick int64) error {
	query := `
		INSERT INTO combat_instances (
			combat_id, player_ship_id, pirate_ship_id, system_id, start_tick, status, turn_number
		) VALUES (?, ?, ?, ?, ?, 'ACTIVE', 0)
	`

	_, err := db.conn.Exec(query, combatID, playerShipID, pirateShipID, systemID, startTick)
	if err != nil {
		return fmt.Errorf("create combat instance: %w", err)
	}

	db.logger.Debug().
		Str("combat_id", combatID).
		Str("player_ship_id", playerShipID).
		Str("pirate_ship_id", pirateShipID).
		Int("system_id", systemID).
		Int64("start_tick", startTick).
		Msg("Combat instance created")

	return nil
}

// GetActiveCombatByPlayerShip retrieves the active combat for a player's ship
func (db *Database) GetActiveCombatByPlayerShip(playerShipID string) (*CombatInstance, error) {
	query := `
		SELECT combat_id, player_ship_id, pirate_ship_id, system_id, start_tick, status, turn_number
		FROM combat_instances
		WHERE player_ship_id = ? AND status = 'ACTIVE'
		LIMIT 1
	`

	var combat CombatInstance
	err := db.conn.QueryRow(query, playerShipID).Scan(
		&combat.CombatID,
		&combat.PlayerShipID,
		&combat.PirateShipID,
		&combat.SystemID,
		&combat.StartTick,
		&combat.Status,
		&combat.TurnNumber,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active combat by player ship: %w", err)
	}

	return &combat, nil
}

// GetActiveCombats retrieves all active combats for tick processing
func (db *Database) GetActiveCombats() ([]*CombatInstance, error) {
	query := `
		SELECT combat_id, player_ship_id, pirate_ship_id, system_id, start_tick, status, turn_number
		FROM combat_instances
		WHERE status = 'ACTIVE'
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get active combats: %w", err)
	}
	defer rows.Close()

	var combats []*CombatInstance
	for rows.Next() {
		var combat CombatInstance
		err := rows.Scan(
			&combat.CombatID,
			&combat.PlayerShipID,
			&combat.PirateShipID,
			&combat.SystemID,
			&combat.StartTick,
			&combat.Status,
			&combat.TurnNumber,
		)
		if err != nil {
			return nil, fmt.Errorf("scan combat instance: %w", err)
		}
		combats = append(combats, &combat)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate combat instances: %w", err)
	}

	return combats, nil
}

// UpdateCombatStatus updates the status of a combat instance
func (db *Database) UpdateCombatStatus(combatID, status string) error {
	query := `
		UPDATE combat_instances
		SET status = ?
		WHERE combat_id = ?
	`

	result, err := db.conn.Exec(query, status, combatID)
	if err != nil {
		return fmt.Errorf("update combat status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update combat status: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("update combat status: combat instance not found")
	}

	db.logger.Debug().
		Str("combat_id", combatID).
		Str("status", status).
		Msg("Combat status updated")

	return nil
}

// UpdateCombatTurn increments the turn counter for a combat instance
func (db *Database) UpdateCombatTurn(combatID string, turnNumber int) error {
	query := `
		UPDATE combat_instances
		SET turn_number = ?
		WHERE combat_id = ?
	`

	result, err := db.conn.Exec(query, turnNumber, combatID)
	if err != nil {
		return fmt.Errorf("update combat turn: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update combat turn: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("update combat turn: combat instance not found")
	}

	db.logger.Debug().
		Str("combat_id", combatID).
		Int("turn_number", turnNumber).
		Msg("Combat turn updated")

	return nil
}

// DeleteCombat removes a combat instance
func (db *Database) DeleteCombat(combatID string) error {
	query := `
		DELETE FROM combat_instances
		WHERE combat_id = ?
	`

	result, err := db.conn.Exec(query, combatID)
	if err != nil {
		return fmt.Errorf("delete combat: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete combat: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("delete combat: combat instance not found")
	}

	db.logger.Debug().
		Str("combat_id", combatID).
		Msg("Combat instance deleted")

	return nil
}

// GetAllCombatInstances retrieves all combat instances (for snapshot creation)
func (db *Database) GetAllCombatInstances() ([]CombatInstance, error) {
	query := `
		SELECT combat_id, player_ship_id, pirate_ship_id, system_id, start_tick, status, turn_number
		FROM combat_instances
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get all combat instances: %w", err)
	}
	defer rows.Close()

	var combats []CombatInstance
	for rows.Next() {
		var combat CombatInstance
		err := rows.Scan(
			&combat.CombatID,
			&combat.PlayerShipID,
			&combat.PirateShipID,
			&combat.SystemID,
			&combat.StartTick,
			&combat.Status,
			&combat.TurnNumber,
		)
		if err != nil {
			return nil, fmt.Errorf("scan combat instance: %w", err)
		}
		combats = append(combats, combat)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate combat instances: %w", err)
	}

	return combats, nil
}
