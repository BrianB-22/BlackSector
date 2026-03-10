package db

import (
	"database/sql"
	"fmt"
	
	"golang.org/x/crypto/bcrypt"
)

// GetPlayerByToken retrieves a player by validating their plaintext token against stored hashes
// This is inefficient for large player bases - in production, consider using a token index
func (db *Database) GetPlayerByToken(plaintextToken string) (*Player, error) {
	// We need to get all players and check the token hash for each one
	// This is because bcrypt hashes are one-way and we can't query by them directly
	query := `
		SELECT player_id, player_name, token_hash, credits, created_at, last_login_at, is_banned
		FROM players
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query players: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var player Player
		var lastLoginAt sql.NullInt64
		var isBanned int

		err := rows.Scan(
			&player.PlayerID,
			&player.PlayerName,
			&player.TokenHash,
			&player.Credits,
			&player.CreatedAt,
			&lastLoginAt,
			&isBanned,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player row: %w", err)
		}

		// Convert nullable fields
		if lastLoginAt.Valid {
			player.LastLoginAt = &lastLoginAt.Int64
		}
		player.IsBanned = isBanned != 0

		// Check if this player's token matches
		// Use bcrypt to compare plaintext token with stored hash
		err = bcrypt.CompareHashAndPassword([]byte(player.TokenHash), []byte(plaintextToken))
		if err == nil {
			// Token matches!
			return &player, nil
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating players: %w", err)
	}

	// No matching player found
	return nil, nil
}

// GetPlayerBySSHUsername retrieves a player by their SSH username
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
		return nil, nil // Player not found, return nil without error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query player by SSH username: %w", err)
	}

	// Convert nullable fields
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
// GetPlayerByID retrieves a player by their player ID
func (db *Database) GetPlayerByID(playerID string) (*Player, error) {
	query := `
		SELECT player_id, player_name, token_hash, credits, created_at, last_login_at, is_banned
		FROM players
		WHERE player_id = ?
	`

	var player Player
	var lastLoginAt sql.NullInt64
	var isBanned int

	err := db.conn.QueryRow(query, playerID).Scan(
		&player.PlayerID,
		&player.PlayerName,
		&player.TokenHash,
		&player.Credits,
		&player.CreatedAt,
		&lastLoginAt,
		&isBanned,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Player not found, return nil without error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query player by ID: %w", err)
	}

	// Convert nullable fields
	if lastLoginAt.Valid {
		player.LastLoginAt = &lastLoginAt.Int64
	}
	player.IsBanned = isBanned != 0

	return &player, nil
}
// InsertPlayer creates a new player record in the database
func (db *Database) InsertPlayer(player *Player) error {
	query := `
		INSERT INTO players (
			player_id, player_name, ssh_username, token_hash, password_hash, credits, created_at, last_login_at, is_banned
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	isBanned := 0
	if player.IsBanned {
		isBanned = 1
	}

	_, err := db.conn.Exec(
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

	db.logger.Debug().
		Str("player_id", player.PlayerID).
		Str("player_name", player.PlayerName).
		Str("ssh_username", player.SSHUsername).
		Msg("Player inserted into database")

	return nil
}



// InsertSession creates a new session record in the database
func (db *Database) InsertSession(session *Session) error {
	query := `
		INSERT INTO sessions (
			session_id, player_id, interface_mode, state,
			connected_at, disconnected_at, linger_expiry_at, last_activity_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(
		query,
		session.SessionID,
		session.PlayerID,
		session.InterfaceMode,
		string(session.State),
		session.ConnectedAt,
		session.DisconnectedAt,
		session.LingerExpiryAt,
		session.LastActivityAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert session: %w", err)
	}

	db.logger.Debug().
		Str("session_id", session.SessionID).
		Str("player_id", session.PlayerID).
		Str("state", string(session.State)).
		Msg("Session inserted into database")

	return nil
}

// UpdateSessionState updates the state of an existing session
func (db *Database) UpdateSessionState(sessionID string, state SessionState) error {
	query := `
		UPDATE sessions
		SET state = ?
		WHERE session_id = ?
	`

	result, err := db.conn.Exec(query, string(state), sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session state: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	db.logger.Debug().
		Str("session_id", sessionID).
		Str("new_state", string(state)).
		Msg("Session state updated")

	return nil
}

// GetActiveSessionByPlayerID retrieves an active session for a player
func (db *Database) GetActiveSessionByPlayerID(playerID string) (*Session, error) {
	query := `
		SELECT session_id, player_id, interface_mode, state,
		       connected_at, disconnected_at, linger_expiry_at, last_activity_at
		FROM sessions
		WHERE player_id = ? AND state = ?
		ORDER BY connected_at DESC
		LIMIT 1
	`

	var session Session
	var disconnectedAt sql.NullInt64
	var lingerExpiryAt sql.NullInt64
	var state string

	err := db.conn.QueryRow(query, playerID, string(SessionConnected)).Scan(
		&session.SessionID,
		&session.PlayerID,
		&session.InterfaceMode,
		&state,
		&session.ConnectedAt,
		&disconnectedAt,
		&lingerExpiryAt,
		&session.LastActivityAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No active session found, return nil without error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query active session by player ID: %w", err)
	}

	// Convert nullable fields
	session.State = SessionState(state)
	if disconnectedAt.Valid {
		session.DisconnectedAt = &disconnectedAt.Int64
	}
	if lingerExpiryAt.Valid {
		session.LingerExpiryAt = &lingerExpiryAt.Int64
	}

	return &session, nil
}

// GetAllPlayers retrieves all players from the database
func (db *Database) GetAllPlayers() ([]Player, error) {
	query := `
		SELECT player_id, player_name, token_hash, credits, created_at, last_login_at, is_banned
		FROM players
		ORDER BY player_id
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all players: %w", err)
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var player Player
		var lastLoginAt sql.NullInt64
		var isBanned int

		err := rows.Scan(
			&player.PlayerID,
			&player.PlayerName,
			&player.TokenHash,
			&player.Credits,
			&player.CreatedAt,
			&lastLoginAt,
			&isBanned,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player row: %w", err)
		}

		// Convert nullable fields
		if lastLoginAt.Valid {
			player.LastLoginAt = &lastLoginAt.Int64
		}
		player.IsBanned = isBanned != 0

		players = append(players, player)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating player rows: %w", err)
	}

	return players, nil
}

// GetAllSessions retrieves all sessions from the database
func (db *Database) GetAllSessions() ([]Session, error) {
	query := `
		SELECT session_id, player_id, interface_mode, state,
		       connected_at, disconnected_at, linger_expiry_at, last_activity_at
		FROM sessions
		ORDER BY connected_at DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var session Session
		var disconnectedAt sql.NullInt64
		var lingerExpiryAt sql.NullInt64
		var state string

		err := rows.Scan(
			&session.SessionID,
			&session.PlayerID,
			&session.InterfaceMode,
			&state,
			&session.ConnectedAt,
			&disconnectedAt,
			&lingerExpiryAt,
			&session.LastActivityAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}

		// Convert nullable fields
		session.State = SessionState(state)
		if disconnectedAt.Valid {
			session.DisconnectedAt = &disconnectedAt.Int64
		}
		if lingerExpiryAt.Valid {
			session.LingerExpiryAt = &lingerExpiryAt.Int64
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating session rows: %w", err)
	}

	return sessions, nil
}


// GetShipByID retrieves a ship by its ship ID
func (db *Database) GetShipByID(shipID string) (*Ship, error) {
	query := `
		SELECT ship_id, player_id, ship_class, hull_points, max_hull_points,
		       shield_points, max_shield_points, energy_points, max_energy_points,
		       cargo_capacity, missiles_current, current_system_id, position_x, position_y,
		       status, docked_at_port_id, last_updated_tick
		FROM ships
		WHERE ship_id = ?
	`

	var ship Ship
	var dockedAtPortID sql.NullInt64

	err := db.conn.QueryRow(query, shipID).Scan(
		&ship.ShipID,
		&ship.PlayerID,
		&ship.ShipClass,
		&ship.HullPoints,
		&ship.MaxHullPoints,
		&ship.ShieldPoints,
		&ship.MaxShieldPoints,
		&ship.EnergyPoints,
		&ship.MaxEnergyPoints,
		&ship.CargoCapacity,
		&ship.MissilesCurrent,
		&ship.CurrentSystemID,
		&ship.PositionX,
		&ship.PositionY,
		&ship.Status,
		&dockedAtPortID,
		&ship.LastUpdatedTick,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Ship not found, return nil without error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query ship by ID: %w", err)
	}

	// Convert nullable field
	if dockedAtPortID.Valid {
		portID := int(dockedAtPortID.Int64)
		ship.DockedAtPortID = &portID
	}

	return &ship, nil
}

// GetShipByPlayerID retrieves a ship by its player ID
// Assumes one ship per player (Phase 1 constraint)
func (db *Database) GetShipByPlayerID(playerID string) (*Ship, error) {
	query := `
		SELECT ship_id, player_id, ship_class, hull_points, max_hull_points,
		       shield_points, max_shield_points, energy_points, max_energy_points,
		       cargo_capacity, missiles_current, current_system_id, position_x, position_y,
		       status, docked_at_port_id, last_updated_tick
		FROM ships
		WHERE player_id = ?
		LIMIT 1
	`

	var ship Ship
	var dockedAtPortID sql.NullInt64

	err := db.conn.QueryRow(query, playerID).Scan(
		&ship.ShipID,
		&ship.PlayerID,
		&ship.ShipClass,
		&ship.HullPoints,
		&ship.MaxHullPoints,
		&ship.ShieldPoints,
		&ship.MaxShieldPoints,
		&ship.EnergyPoints,
		&ship.MaxEnergyPoints,
		&ship.CargoCapacity,
		&ship.MissilesCurrent,
		&ship.CurrentSystemID,
		&ship.PositionX,
		&ship.PositionY,
		&ship.Status,
		&dockedAtPortID,
		&ship.LastUpdatedTick,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Ship not found, return nil without error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query ship by player ID: %w", err)
	}

	// Convert nullable field
	if dockedAtPortID.Valid {
		portID := int(dockedAtPortID.Int64)
		ship.DockedAtPortID = &portID
	}

	return &ship, nil
}

// UpdateShipPosition updates a ship's current system and last updated tick
func (db *Database) UpdateShipPosition(shipID string, systemID int, tick int64) error {
	query := `
		UPDATE ships
		SET current_system_id = ?, last_updated_tick = ?
		WHERE ship_id = ?
	`

	result, err := db.conn.Exec(query, systemID, tick, shipID)
	if err != nil {
		return fmt.Errorf("failed to update ship position: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ship not found: %s", shipID)
	}

	db.logger.Debug().
		Str("ship_id", shipID).
		Int("system_id", systemID).
		Int64("tick", tick).
		Msg("Ship position updated")

	return nil
}

// UpdateShipDockStatus updates a ship's status and docked port
func (db *Database) UpdateShipDockStatus(shipID string, status string, dockedAtPortID *int, tick int64) error {
	query := `
		UPDATE ships
		SET status = ?, docked_at_port_id = ?, last_updated_tick = ?
		WHERE ship_id = ?
	`

	var portID interface{}
	if dockedAtPortID != nil {
		portID = *dockedAtPortID
	} else {
		portID = nil
	}

	result, err := db.conn.Exec(query, status, portID, tick, shipID)
	if err != nil {
		return fmt.Errorf("failed to update ship dock status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ship not found: %s", shipID)
	}

	db.logger.Debug().
		Str("ship_id", shipID).
		Str("status", status).
		Interface("docked_at_port_id", dockedAtPortID).
		Int64("tick", tick).
		Msg("Ship dock status updated")

	return nil
}

// UpdatePlayerCredits updates a player's credit balance
func (db *Database) UpdatePlayerCredits(playerID string, credits int) error {
	query := `
		UPDATE players
		SET credits = ?
		WHERE player_id = ?
	`

	result, err := db.conn.Exec(query, credits, playerID)
	if err != nil {
		return fmt.Errorf("failed to update player credits: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("player not found: %s", playerID)
	}

	db.logger.Debug().
		Str("player_id", playerID).
		Int("credits", credits).
		Msg("Player credits updated")

	return nil
}

// GetShipCargo retrieves all cargo slots for a ship
func (db *Database) GetShipCargo(shipID string) ([]CargoSlot, error) {
	query := `
		SELECT ship_id, slot_index, commodity_id, quantity
		FROM ship_cargo
		WHERE ship_id = ?
		ORDER BY slot_index
	`

	rows, err := db.conn.Query(query, shipID)
	if err != nil {
		return nil, fmt.Errorf("failed to query ship cargo: %w", err)
	}
	defer rows.Close()

	var cargo []CargoSlot
	for rows.Next() {
		var slot CargoSlot
		err := rows.Scan(
			&slot.ShipID,
			&slot.SlotIndex,
			&slot.CommodityID,
			&slot.Quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cargo slot: %w", err)
		}
		cargo = append(cargo, slot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cargo rows: %w", err)
	}

	return cargo, nil
}

// GetCargoTotalQuantity returns the total quantity of all cargo in a ship
func (db *Database) GetCargoTotalQuantity(shipID string) (int, error) {
	query := `
		SELECT COALESCE(SUM(quantity), 0)
		FROM ship_cargo
		WHERE ship_id = ?
	`

	var total int
	err := db.conn.QueryRow(query, shipID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to query cargo total: %w", err)
	}

	return total, nil
}

// GetCargoSlot retrieves a specific cargo slot for a commodity
func (db *Database) GetCargoSlot(shipID string, commodityID string) (*CargoSlot, error) {
	query := `
		SELECT ship_id, slot_index, commodity_id, quantity
		FROM ship_cargo
		WHERE ship_id = ? AND commodity_id = ?
	`

	var slot CargoSlot
	err := db.conn.QueryRow(query, shipID, commodityID).Scan(
		&slot.ShipID,
		&slot.SlotIndex,
		&slot.CommodityID,
		&slot.Quantity,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Slot not found, return nil without error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query cargo slot: %w", err)
	}

	return &slot, nil
}

// AddCargo adds or updates a cargo slot for a ship
func (db *Database) AddCargo(shipID string, slotIndex int, commodityID string, quantity int) error {
	query := `
		INSERT INTO ship_cargo (ship_id, slot_index, commodity_id, quantity)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(ship_id, slot_index) DO UPDATE SET
			commodity_id = excluded.commodity_id,
			quantity = excluded.quantity
	`

	_, err := db.conn.Exec(query, shipID, slotIndex, commodityID, quantity)
	if err != nil {
		return fmt.Errorf("failed to add cargo: %w", err)
	}

	db.logger.Debug().
		Str("ship_id", shipID).
		Int("slot_index", slotIndex).
		Str("commodity_id", commodityID).
		Int("quantity", quantity).
		Msg("Cargo added")

	return nil
}

// UpdateCargoQuantity updates the quantity in a cargo slot
func (db *Database) UpdateCargoQuantity(shipID string, slotIndex int, quantity int) error {
	query := `
		UPDATE ship_cargo
		SET quantity = ?
		WHERE ship_id = ? AND slot_index = ?
	`

	result, err := db.conn.Exec(query, quantity, shipID, slotIndex)
	if err != nil {
		return fmt.Errorf("failed to update cargo quantity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cargo slot not found: ship=%s slot=%d", shipID, slotIndex)
	}

	db.logger.Debug().
		Str("ship_id", shipID).
		Int("slot_index", slotIndex).
		Int("quantity", quantity).
		Msg("Cargo quantity updated")

	return nil
}

// RemoveCargo removes a cargo slot from a ship
func (db *Database) RemoveCargo(shipID string, slotIndex int) error {
	query := `
		DELETE FROM ship_cargo
		WHERE ship_id = ? AND slot_index = ?
	`

	result, err := db.conn.Exec(query, shipID, slotIndex)
	if err != nil {
		return fmt.Errorf("failed to remove cargo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cargo slot not found: ship=%s slot=%d", shipID, slotIndex)
	}

	db.logger.Debug().
		Str("ship_id", shipID).
		Int("slot_index", slotIndex).
		Msg("Cargo removed")

	return nil
}

// GetPortInventory retrieves a specific commodity inventory at a port
func (db *Database) GetPortInventory(portID int, commodityID string) (*PortInventory, error) {
	query := `
		SELECT port_id, commodity_id, quantity, buy_price, sell_price, updated_tick
		FROM port_inventory
		WHERE port_id = ? AND commodity_id = ?
	`

	var inv PortInventory
	err := db.conn.QueryRow(query, portID, commodityID).Scan(
		&inv.PortID,
		&inv.CommodityID,
		&inv.Quantity,
		&inv.BuyPrice,
		&inv.SellPrice,
		&inv.UpdatedTick,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Inventory not found, return nil without error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query port inventory: %w", err)
	}

	return &inv, nil
}

// GetAllPortInventory retrieves all commodity inventory at a port
func (db *Database) GetAllPortInventory(portID int) ([]PortInventory, error) {
	query := `
		SELECT port_id, commodity_id, quantity, buy_price, sell_price, updated_tick
		FROM port_inventory
		WHERE port_id = ?
		ORDER BY commodity_id
	`

	rows, err := db.conn.Query(query, portID)
	if err != nil {
		return nil, fmt.Errorf("failed to query port inventory: %w", err)
	}
	defer rows.Close()

	var inventory []PortInventory
	for rows.Next() {
		var inv PortInventory
		err := rows.Scan(
			&inv.PortID,
			&inv.CommodityID,
			&inv.Quantity,
			&inv.BuyPrice,
			&inv.SellPrice,
			&inv.UpdatedTick,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory row: %w", err)
		}
		inventory = append(inventory, inv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating inventory rows: %w", err)
	}

	return inventory, nil
}

// UpdatePortInventory updates the quantity and prices for a port commodity
func (db *Database) UpdatePortInventory(portID int, commodityID string, quantity int, tick int64) error {
	query := `
		UPDATE port_inventory
		SET quantity = ?, updated_tick = ?
		WHERE port_id = ? AND commodity_id = ?
	`

	result, err := db.conn.Exec(query, quantity, tick, portID, commodityID)
	if err != nil {
		return fmt.Errorf("failed to update port inventory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("port inventory not found: port=%d commodity=%s", portID, commodityID)
	}

	db.logger.Debug().
		Int("port_id", portID).
		Str("commodity_id", commodityID).
		Int("quantity", quantity).
		Int64("tick", tick).
		Msg("Port inventory updated")

	return nil
}

// Transaction-aware query methods

// TxUpdatePlayerCredits updates a player's credit balance within a transaction
func (db *Database) TxUpdatePlayerCredits(tx *sql.Tx, playerID string, credits int) error {
	query := `
		UPDATE players
		SET credits = ?
		WHERE player_id = ?
	`

	result, err := tx.Exec(query, credits, playerID)
	if err != nil {
		return fmt.Errorf("failed to update player credits: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("player not found: %s", playerID)
	}

	return nil
}

// TxGetPlayerByID retrieves a player by ID within a transaction
func (db *Database) TxGetPlayerByID(tx *sql.Tx, playerID string) (*Player, error) {
	query := `
		SELECT player_id, player_name, token_hash, credits, created_at, last_login_at, is_banned
		FROM players
		WHERE player_id = ?
	`

	var player Player
	var lastLoginAt sql.NullInt64
	var isBanned int

	err := tx.QueryRow(query, playerID).Scan(
		&player.PlayerID,
		&player.PlayerName,
		&player.TokenHash,
		&player.Credits,
		&player.CreatedAt,
		&lastLoginAt,
		&isBanned,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query player by ID: %w", err)
	}

	if lastLoginAt.Valid {
		player.LastLoginAt = &lastLoginAt.Int64
	}
	player.IsBanned = isBanned != 0

	return &player, nil
}

// TxGetShipByID retrieves a ship by ID within a transaction
func (db *Database) TxGetShipByID(tx *sql.Tx, shipID string) (*Ship, error) {
	query := `
		SELECT ship_id, player_id, ship_class, hull_points, max_hull_points,
		       shield_points, max_shield_points, energy_points, max_energy_points,
		       cargo_capacity, missiles_current, current_system_id, position_x, position_y,
		       status, docked_at_port_id, last_updated_tick
		FROM ships
		WHERE ship_id = ?
	`

	var ship Ship
	var dockedAtPortID sql.NullInt64

	err := tx.QueryRow(query, shipID).Scan(
		&ship.ShipID,
		&ship.PlayerID,
		&ship.ShipClass,
		&ship.HullPoints,
		&ship.MaxHullPoints,
		&ship.ShieldPoints,
		&ship.MaxShieldPoints,
		&ship.EnergyPoints,
		&ship.MaxEnergyPoints,
		&ship.CargoCapacity,
		&ship.MissilesCurrent,
		&ship.CurrentSystemID,
		&ship.PositionX,
		&ship.PositionY,
		&ship.Status,
		&dockedAtPortID,
		&ship.LastUpdatedTick,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query ship by ID: %w", err)
	}

	if dockedAtPortID.Valid {
		portID := int(dockedAtPortID.Int64)
		ship.DockedAtPortID = &portID
	}

	return &ship, nil
}

// TxGetCargoTotalQuantity returns the total quantity within a transaction
func (db *Database) TxGetCargoTotalQuantity(tx *sql.Tx, shipID string) (int, error) {
	query := `
		SELECT COALESCE(SUM(quantity), 0)
		FROM ship_cargo
		WHERE ship_id = ?
	`

	var total int
	err := tx.QueryRow(query, shipID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to query cargo total: %w", err)
	}

	return total, nil
}

// TxGetCargoSlot retrieves a cargo slot within a transaction
func (db *Database) TxGetCargoSlot(tx *sql.Tx, shipID string, commodityID string) (*CargoSlot, error) {
	query := `
		SELECT ship_id, slot_index, commodity_id, quantity
		FROM ship_cargo
		WHERE ship_id = ? AND commodity_id = ?
	`

	var slot CargoSlot
	err := tx.QueryRow(query, shipID, commodityID).Scan(
		&slot.ShipID,
		&slot.SlotIndex,
		&slot.CommodityID,
		&slot.Quantity,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query cargo slot: %w", err)
	}

	return &slot, nil
}

// TxAddCargo adds or updates cargo within a transaction
func (db *Database) TxAddCargo(tx *sql.Tx, shipID string, slotIndex int, commodityID string, quantity int) error {
	query := `
		INSERT INTO ship_cargo (ship_id, slot_index, commodity_id, quantity)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(ship_id, slot_index) DO UPDATE SET
			commodity_id = excluded.commodity_id,
			quantity = excluded.quantity
	`

	_, err := tx.Exec(query, shipID, slotIndex, commodityID, quantity)
	if err != nil {
		return fmt.Errorf("failed to add cargo: %w", err)
	}

	return nil
}

// TxUpdateCargoQuantity updates cargo quantity within a transaction
func (db *Database) TxUpdateCargoQuantity(tx *sql.Tx, shipID string, slotIndex int, quantity int) error {
	query := `
		UPDATE ship_cargo
		SET quantity = ?
		WHERE ship_id = ? AND slot_index = ?
	`

	result, err := tx.Exec(query, quantity, shipID, slotIndex)
	if err != nil {
		return fmt.Errorf("failed to update cargo quantity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cargo slot not found: ship=%s slot=%d", shipID, slotIndex)
	}

	return nil
}

// TxRemoveCargo removes cargo within a transaction
func (db *Database) TxRemoveCargo(tx *sql.Tx, shipID string, slotIndex int) error {
	query := `
		DELETE FROM ship_cargo
		WHERE ship_id = ? AND slot_index = ?
	`

	result, err := tx.Exec(query, shipID, slotIndex)
	if err != nil {
		return fmt.Errorf("failed to remove cargo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cargo slot not found: ship=%s slot=%d", shipID, slotIndex)
	}

	return nil
}

// TxGetPortInventory retrieves port inventory within a transaction
func (db *Database) TxGetPortInventory(tx *sql.Tx, portID int, commodityID string) (*PortInventory, error) {
	query := `
		SELECT port_id, commodity_id, quantity, buy_price, sell_price, updated_tick
		FROM port_inventory
		WHERE port_id = ? AND commodity_id = ?
	`

	var inv PortInventory
	err := tx.QueryRow(query, portID, commodityID).Scan(
		&inv.PortID,
		&inv.CommodityID,
		&inv.Quantity,
		&inv.BuyPrice,
		&inv.SellPrice,
		&inv.UpdatedTick,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query port inventory: %w", err)
	}

	return &inv, nil
}

// TxUpdatePortInventory updates port inventory within a transaction
func (db *Database) TxUpdatePortInventory(tx *sql.Tx, portID int, commodityID string, quantity int, tick int64) error {
	query := `
		UPDATE port_inventory
		SET quantity = ?, updated_tick = ?
		WHERE port_id = ? AND commodity_id = ?
	`

	result, err := tx.Exec(query, quantity, tick, portID, commodityID)
	if err != nil {
		return fmt.Errorf("failed to update port inventory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("port inventory not found: port=%d commodity=%s", portID, commodityID)
	}

	return nil
}

// InsertShip creates a new ship record in the database
func (db *Database) InsertShip(ship *Ship) error {
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

	_, err := db.conn.Exec(
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

	db.logger.Debug().
		Str("ship_id", ship.ShipID).
		Str("player_id", ship.PlayerID).
		Str("ship_class", ship.ShipClass).
		Int("system_id", ship.CurrentSystemID).
		Msg("Ship inserted into database")

	return nil
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
