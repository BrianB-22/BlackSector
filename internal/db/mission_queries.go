package db

import (
	"database/sql"
	"fmt"
)

// MissionInstance represents an active mission for a player
type MissionInstance struct {
	InstanceID            string
	MissionID             string
	PlayerID              string
	Status                string
	AcceptedTick          int64
	StartedTick           *int64
	CompletedTick         *int64
	FailedReason          *string
	ExpiresAtTick         *int64
}

// ObjectiveProgress tracks progress on a mission objective
type ObjectiveProgress struct {
	InstanceID     string
	ObjectiveIndex int
	Status         string
	CurrentValue   int
	RequiredValue  int
}

// CreateMissionInstance creates a new mission instance
func (db *Database) CreateMissionInstance(instance *MissionInstance) error {
	query := `
		INSERT INTO mission_instances (
			instance_id, mission_id, player_id, status,
			accepted_tick, started_tick, completed_tick, failed_reason, expires_at_tick
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query,
		instance.InstanceID,
		instance.MissionID,
		instance.PlayerID,
		instance.Status,
		instance.AcceptedTick,
		instance.StartedTick,
		instance.CompletedTick,
		instance.FailedReason,
		instance.ExpiresAtTick,
	)

	if err != nil {
		return fmt.Errorf("create mission instance: %w", err)
	}

	return nil
}

// GetMissionInstance retrieves a mission instance by ID
func (db *Database) GetMissionInstance(instanceID string) (*MissionInstance, error) {
	query := `
		SELECT instance_id, mission_id, player_id, status,
		       accepted_tick, started_tick, completed_tick, failed_reason, expires_at_tick
		FROM mission_instances
		WHERE instance_id = ?
	`

	var instance MissionInstance
	err := db.conn.QueryRow(query, instanceID).Scan(
		&instance.InstanceID,
		&instance.MissionID,
		&instance.PlayerID,
		&instance.Status,
		&instance.AcceptedTick,
		&instance.StartedTick,
		&instance.CompletedTick,
		&instance.FailedReason,
		&instance.ExpiresAtTick,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get mission instance: %w", err)
	}

	return &instance, nil
}

// GetActiveMissionByPlayer retrieves the active mission for a player
func (db *Database) GetActiveMissionByPlayer(playerID string) (*MissionInstance, error) {
	query := `
		SELECT instance_id, mission_id, player_id, status,
		       accepted_tick, started_tick, completed_tick, failed_reason, expires_at_tick
		FROM mission_instances
		WHERE player_id = ? AND status = 'IN_PROGRESS'
		LIMIT 1
	`

	var instance MissionInstance
	err := db.conn.QueryRow(query, playerID).Scan(
		&instance.InstanceID,
		&instance.MissionID,
		&instance.PlayerID,
		&instance.Status,
		&instance.AcceptedTick,
		&instance.StartedTick,
		&instance.CompletedTick,
		&instance.FailedReason,
		&instance.ExpiresAtTick,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active mission by player: %w", err)
	}

	return &instance, nil
}

// GetAllInProgressMissions retrieves all missions with IN_PROGRESS status
func (db *Database) GetAllInProgressMissions() ([]*MissionInstance, error) {
	query := `
		SELECT instance_id, mission_id, player_id, status,
		       accepted_tick, started_tick, completed_tick, failed_reason, expires_at_tick
		FROM mission_instances
		WHERE status = 'IN_PROGRESS'
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get all in progress missions: %w", err)
	}
	defer rows.Close()

	var missions []*MissionInstance
	for rows.Next() {
		var instance MissionInstance
		err := rows.Scan(
			&instance.InstanceID,
			&instance.MissionID,
			&instance.PlayerID,
			&instance.Status,
			&instance.AcceptedTick,
			&instance.StartedTick,
			&instance.CompletedTick,
			&instance.FailedReason,
			&instance.ExpiresAtTick,
		)
		if err != nil {
			return nil, fmt.Errorf("scan mission instance: %w", err)
		}
		missions = append(missions, &instance)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mission instances: %w", err)
	}

	return missions, nil
}

// GetCompletedMissionsByPlayer retrieves completed missions for a player
func (db *Database) GetCompletedMissionsByPlayer(playerID string) ([]*MissionInstance, error) {
	query := `
		SELECT instance_id, mission_id, player_id, status,
		       accepted_tick, started_tick, completed_tick, failed_reason, expires_at_tick
		FROM mission_instances
		WHERE player_id = ? AND status = 'COMPLETED'
		ORDER BY completed_tick DESC
	`

	rows, err := db.conn.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("get completed missions by player: %w", err)
	}
	defer rows.Close()

	var missions []*MissionInstance
	for rows.Next() {
		var instance MissionInstance
		err := rows.Scan(
			&instance.InstanceID,
			&instance.MissionID,
			&instance.PlayerID,
			&instance.Status,
			&instance.AcceptedTick,
			&instance.StartedTick,
			&instance.CompletedTick,
			&instance.FailedReason,
			&instance.ExpiresAtTick,
		)
		if err != nil {
			return nil, fmt.Errorf("scan mission instance: %w", err)
		}
		missions = append(missions, &instance)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mission instances: %w", err)
	}

	return missions, nil
}

// UpdateMissionStatus updates the status of a mission instance
func (db *Database) UpdateMissionStatus(instanceID string, status string, tick int64) error {
	query := `
		UPDATE mission_instances
		SET status = ?,
		    completed_tick = CASE WHEN ? IN ('COMPLETED', 'FAILED', 'EXPIRED', 'ABANDONED') THEN ? ELSE completed_tick END
		WHERE instance_id = ?
	`

	result, err := db.conn.Exec(query, status, status, tick, instanceID)
	if err != nil {
		return fmt.Errorf("update mission status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update mission status: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("update mission status: mission instance not found")
	}

	return nil
}

// UpdateMissionObjectiveIndex updates the current objective index
func (db *Database) UpdateMissionObjectiveIndex(instanceID string, objectiveIndex int) error {
	// Note: The schema doesn't have a current_objective_index column
	// This is tracked via objective_progress status instead
	// This method is kept for interface compatibility but doesn't need to do anything
	return nil
}

// DeleteMissionInstance deletes a mission instance and its objective progress
func (db *Database) DeleteMissionInstance(instanceID string) error {
	// Delete objective progress first (foreign key constraint)
	_, err := db.conn.Exec("DELETE FROM objective_progress WHERE instance_id = ?", instanceID)
	if err != nil {
		return fmt.Errorf("delete objective progress: %w", err)
	}

	// Delete mission instance
	_, err = db.conn.Exec("DELETE FROM mission_instances WHERE instance_id = ?", instanceID)
	if err != nil {
		return fmt.Errorf("delete mission instance: %w", err)
	}

	return nil
}

// CreateObjectiveProgress creates objective progress tracking
func (db *Database) CreateObjectiveProgress(progress *ObjectiveProgress) error {
	query := `
		INSERT INTO objective_progress (
			instance_id, objective_index, status, current_value, required_value
		) VALUES (?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query,
		progress.InstanceID,
		progress.ObjectiveIndex,
		progress.Status,
		progress.CurrentValue,
		progress.RequiredValue,
	)

	if err != nil {
		return fmt.Errorf("create objective progress: %w", err)
	}

	return nil
}

// GetObjectiveProgress retrieves progress for a specific objective
func (db *Database) GetObjectiveProgress(instanceID string, objectiveIndex int) (*ObjectiveProgress, error) {
	query := `
		SELECT instance_id, objective_index, status, current_value, required_value
		FROM objective_progress
		WHERE instance_id = ? AND objective_index = ?
	`

	var progress ObjectiveProgress
	err := db.conn.QueryRow(query, instanceID, objectiveIndex).Scan(
		&progress.InstanceID,
		&progress.ObjectiveIndex,
		&progress.Status,
		&progress.CurrentValue,
		&progress.RequiredValue,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get objective progress: %w", err)
	}

	return &progress, nil
}

// GetAllObjectiveProgress retrieves all objective progress for a mission instance
func (db *Database) GetAllObjectiveProgress(instanceID string) ([]*ObjectiveProgress, error) {
	query := `
		SELECT instance_id, objective_index, status, current_value, required_value
		FROM objective_progress
		WHERE instance_id = ?
		ORDER BY objective_index ASC
	`

	rows, err := db.conn.Query(query, instanceID)
	if err != nil {
		return nil, fmt.Errorf("get all objective progress: %w", err)
	}
	defer rows.Close()

	var progressList []*ObjectiveProgress
	for rows.Next() {
		var progress ObjectiveProgress
		err := rows.Scan(
			&progress.InstanceID,
			&progress.ObjectiveIndex,
			&progress.Status,
			&progress.CurrentValue,
			&progress.RequiredValue,
		)
		if err != nil {
			return nil, fmt.Errorf("scan objective progress: %w", err)
		}
		progressList = append(progressList, &progress)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate objective progress: %w", err)
	}

	return progressList, nil
}

// UpdateObjectiveProgress updates objective progress
func (db *Database) UpdateObjectiveProgress(instanceID string, objectiveIndex int, status string, currentValue int) error {
	query := `
		UPDATE objective_progress
		SET status = ?, current_value = ?
		WHERE instance_id = ? AND objective_index = ?
	`

	result, err := db.conn.Exec(query, status, currentValue, instanceID, objectiveIndex)
	if err != nil {
		return fmt.Errorf("update objective progress: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update objective progress: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("update objective progress: objective not found")
	}

	return nil
}

// DeleteObjectiveProgress deletes all objective progress for a mission instance
func (db *Database) DeleteObjectiveProgress(instanceID string) error {
	_, err := db.conn.Exec("DELETE FROM objective_progress WHERE instance_id = ?", instanceID)
	if err != nil {
		return fmt.Errorf("delete objective progress: %w", err)
	}
	return nil
}

// GetPortByID retrieves a port by ID
func (db *Database) GetPortByID(portID int) (*Port, error) {
	query := `
		SELECT port_id, system_id, name, port_type, security_level,
		       docking_fee, has_bank, has_shipyard, has_upgrade_market, has_repair, has_fuel
		FROM ports
		WHERE port_id = ?
	`

	var port Port
	err := db.conn.QueryRow(query, portID).Scan(
		&port.PortID,
		&port.SystemID,
		&port.Name,
		&port.PortType,
		&port.SecurityLevel,
		&port.DockingFee,
		&port.HasBank,
		&port.HasShipyard,
		&port.HasUpgradeMarket,
		&port.HasRepair,
		&port.HasFuel,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get port by id: %w", err)
	}

	return &port, nil
}

// GetSystemSecurityLevel retrieves the security level of a system
func (db *Database) GetSystemSecurityLevel(systemID int) (float64, error) {
	query := `
		SELECT security_level
		FROM systems
		WHERE system_id = ?
	`

	var securityLevel float64
	err := db.conn.QueryRow(query, systemID).Scan(&securityLevel)

	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("get system security level: system not found")
	}
	if err != nil {
		return 0, fmt.Errorf("get system security level: %w", err)
	}

	return securityLevel, nil
}

// GetCargoByShipID retrieves all cargo for a ship
func (db *Database) GetCargoByShipID(shipID string) ([]*CargoSlot, error) {
	query := `
		SELECT ship_id, slot_index, commodity_id, quantity
		FROM ship_cargo
		WHERE ship_id = ?
		ORDER BY slot_index ASC
	`

	rows, err := db.conn.Query(query, shipID)
	if err != nil {
		return nil, fmt.Errorf("get cargo by ship id: %w", err)
	}
	defer rows.Close()

	var cargo []*CargoSlot
	for rows.Next() {
		var slot CargoSlot
		err := rows.Scan(
			&slot.ShipID,
			&slot.SlotIndex,
			&slot.CommodityID,
			&slot.Quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("scan cargo slot: %w", err)
		}
		cargo = append(cargo, &slot)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cargo: %w", err)
	}

	return cargo, nil
}

// GetAllMissionInstances retrieves all mission instances (for snapshot creation)
func (db *Database) GetAllMissionInstances() ([]MissionInstance, error) {
	query := `
		SELECT instance_id, mission_id, player_id, status,
		       accepted_tick, started_tick, completed_tick, failed_reason, expires_at_tick
		FROM mission_instances
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get all mission instances: %w", err)
	}
	defer rows.Close()

	var missions []MissionInstance
	for rows.Next() {
		var instance MissionInstance
		err := rows.Scan(
			&instance.InstanceID,
			&instance.MissionID,
			&instance.PlayerID,
			&instance.Status,
			&instance.AcceptedTick,
			&instance.StartedTick,
			&instance.CompletedTick,
			&instance.FailedReason,
			&instance.ExpiresAtTick,
		)
		if err != nil {
			return nil, fmt.Errorf("scan mission instance: %w", err)
		}
		missions = append(missions, instance)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mission instances: %w", err)
	}

	return missions, nil
}

// GetAllObjectiveProgressForSnapshot retrieves all objective progress records (for snapshot creation)
func (db *Database) GetAllObjectiveProgressForSnapshot() ([]ObjectiveProgress, error) {
	query := `
		SELECT instance_id, objective_index, status, current_value, required_value
		FROM objective_progress
		ORDER BY instance_id, objective_index
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get all objective progress: %w", err)
	}
	defer rows.Close()

	var progressList []ObjectiveProgress
	for rows.Next() {
		var progress ObjectiveProgress
		err := rows.Scan(
			&progress.InstanceID,
			&progress.ObjectiveIndex,
			&progress.Status,
			&progress.CurrentValue,
			&progress.RequiredValue,
		)
		if err != nil {
			return nil, fmt.Errorf("scan objective progress: %w", err)
		}
		progressList = append(progressList, progress)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate objective progress: %w", err)
	}

	return progressList, nil
}
