package missions

import (
	"fmt"
)

// ValidateMissionDefinition validates a mission definition for correctness
func ValidateMissionDefinition(mission *MissionDefinition) error {
	if mission == nil {
		return fmt.Errorf("mission definition is nil")
	}

	// Validate required fields
	if mission.MissionID == "" {
		return fmt.Errorf("mission_id is required")
	}
	if mission.Name == "" {
		return fmt.Errorf("mission %s: name is required", mission.MissionID)
	}
	if mission.Description == "" {
		return fmt.Errorf("mission %s: description is required", mission.MissionID)
	}

	// Validate security zones
	if len(mission.SecurityZones) == 0 {
		return fmt.Errorf("mission %s: at least one security zone is required", mission.MissionID)
	}

	for _, zone := range mission.SecurityZones {
		if !isValidSecurityZone(zone) {
			return fmt.Errorf("mission %s: invalid security zone '%s'", mission.MissionID, zone)
		}
	}

	// Validate objectives
	if len(mission.Objectives) == 0 {
		return fmt.Errorf("mission %s: at least one objective is required", mission.MissionID)
	}

	for i, obj := range mission.Objectives {
		if err := ValidateObjectiveDefinition(obj, mission.MissionID, i); err != nil {
			return err
		}
	}

	// Validate rewards
	if mission.Rewards == nil {
		return fmt.Errorf("mission %s: rewards definition is required", mission.MissionID)
	}
	if mission.Rewards.Credits < 0 {
		return fmt.Errorf("mission %s: reward credits cannot be negative", mission.MissionID)
	}

	// Validate expiry ticks if present
	if mission.ExpiryTicks != nil && *mission.ExpiryTicks <= 0 {
		return fmt.Errorf("mission %s: expiry_ticks must be positive", mission.MissionID)
	}

	// Validate repeat cooldown if repeatable
	if mission.Repeatable && mission.RepeatCooldownTicks < 0 {
		return fmt.Errorf("mission %s: repeat_cooldown_ticks cannot be negative", mission.MissionID)
	}

	return nil
}

// ValidateObjectiveDefinition validates a mission objective definition
func ValidateObjectiveDefinition(obj *ObjectiveDefinition, missionID string, index int) error {
	if obj == nil {
		return fmt.Errorf("mission %s: objective %d is nil", missionID, index)
	}

	if obj.ObjectiveID == "" {
		return fmt.Errorf("mission %s: objective %d: objective_id is required", missionID, index)
	}
	if obj.Type == "" {
		return fmt.Errorf("mission %s: objective %d: type is required", missionID, index)
	}
	if obj.Description == "" {
		return fmt.Errorf("mission %s: objective %d: description is required", missionID, index)
	}

	// Validate objective type and parameters
	switch obj.Type {
	case "deliver_commodity":
		if err := validateDeliverCommodityObjective(obj, missionID, index); err != nil {
			return err
		}
	case "navigate_to":
		if err := validateNavigateToObjective(obj, missionID, index); err != nil {
			return err
		}
	case "kill":
		// Phase 2 - not implemented in Phase 1
		return fmt.Errorf("mission %s: objective %d: 'kill' objectives not supported in Phase 1", missionID, index)
	default:
		return fmt.Errorf("mission %s: objective %d: unknown objective type '%s'", missionID, index, obj.Type)
	}

	return nil
}

// validateDeliverCommodityObjective validates deliver_commodity objective parameters
func validateDeliverCommodityObjective(obj *ObjectiveDefinition, missionID string, index int) error {
	if obj.Parameters == nil {
		return fmt.Errorf("mission %s: objective %d: parameters required for deliver_commodity", missionID, index)
	}

	// Validate required parameters
	commodityID, ok := obj.Parameters["commodity_id"].(string)
	if !ok || commodityID == "" {
		return fmt.Errorf("mission %s: objective %d: commodity_id parameter required", missionID, index)
	}

	quantity, ok := obj.Parameters["quantity"].(float64) // JSON numbers are float64
	if !ok || quantity <= 0 {
		return fmt.Errorf("mission %s: objective %d: quantity parameter must be positive", missionID, index)
	}

	destinationPortID, ok := obj.Parameters["destination_port_id"].(float64)
	if !ok || destinationPortID <= 0 {
		return fmt.Errorf("mission %s: objective %d: destination_port_id parameter required", missionID, index)
	}

	return nil
}

// validateNavigateToObjective validates navigate_to objective parameters
func validateNavigateToObjective(obj *ObjectiveDefinition, missionID string, index int) error {
	if obj.Parameters == nil {
		return fmt.Errorf("mission %s: objective %d: parameters required for navigate_to", missionID, index)
	}

	// Validate required parameters
	systemID, ok := obj.Parameters["system_id"].(float64)
	if !ok || systemID <= 0 {
		return fmt.Errorf("mission %s: objective %d: system_id parameter required", missionID, index)
	}

	return nil
}

// isValidSecurityZone checks if a security zone string is valid
func isValidSecurityZone(zone string) bool {
	validZones := map[string]bool{
		"federated_space": true,
		"high_security":   true,
		"low_security":    true,
		"medium_security": true, // Phase 2
		"black_sector":    true, // Phase 2
	}
	return validZones[zone]
}

// ValidateMissionFile validates an entire mission file
func ValidateMissionFile(file *MissionFile) error {
	if file == nil {
		return fmt.Errorf("mission file is nil")
	}

	if len(file.Missions) == 0 {
		return fmt.Errorf("mission file contains no missions")
	}

	// Track mission IDs to ensure uniqueness
	missionIDs := make(map[string]bool)

	for i, mission := range file.Missions {
		if err := ValidateMissionDefinition(mission); err != nil {
			return fmt.Errorf("mission %d: %w", i, err)
		}

		// Check for duplicate mission IDs
		if missionIDs[mission.MissionID] {
			return fmt.Errorf("duplicate mission_id: %s", mission.MissionID)
		}
		missionIDs[mission.MissionID] = true
	}

	return nil
}
