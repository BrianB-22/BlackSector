package missions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateMissionDefinition tests mission definition validation
func TestValidateMissionDefinition(t *testing.T) {
	tests := []struct {
		name    string
		mission *MissionDefinition
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid mission with all required fields",
			mission: &MissionDefinition{
				MissionID:           "test_mission",
				Name:                "Test Mission",
				Description:         "A test mission",
				SecurityZones:       []string{"high_security"},
				Objectives:          []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:             &RewardDefinition{Credits: 1000},
				Enabled:             true,
				Repeatable:          false,
				RepeatCooldownTicks: 0,
			},
			wantErr: false,
		},
		{
			name:    "nil mission",
			mission: nil,
			wantErr: true,
			errMsg:  "mission definition is nil",
		},
		{
			name: "missing mission_id",
			mission: &MissionDefinition{
				Name:          "Test Mission",
				Description:   "A test mission",
				SecurityZones: []string{"high_security"},
				Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:       &RewardDefinition{Credits: 1000},
			},
			wantErr: true,
			errMsg:  "mission_id is required",
		},
		{
			name: "missing name",
			mission: &MissionDefinition{
				MissionID:     "test_mission",
				Description:   "A test mission",
				SecurityZones: []string{"high_security"},
				Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:       &RewardDefinition{Credits: 1000},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing description",
			mission: &MissionDefinition{
				MissionID:     "test_mission",
				Name:          "Test Mission",
				SecurityZones: []string{"high_security"},
				Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:       &RewardDefinition{Credits: 1000},
			},
			wantErr: true,
			errMsg:  "description is required",
		},
		{
			name: "empty security zones",
			mission: &MissionDefinition{
				MissionID:     "test_mission",
				Name:          "Test Mission",
				Description:   "A test mission",
				SecurityZones: []string{},
				Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:       &RewardDefinition{Credits: 1000},
			},
			wantErr: true,
			errMsg:  "at least one security zone is required",
		},
		{
			name: "invalid security zone",
			mission: &MissionDefinition{
				MissionID:     "test_mission",
				Name:          "Test Mission",
				Description:   "A test mission",
				SecurityZones: []string{"invalid_zone"},
				Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:       &RewardDefinition{Credits: 1000},
			},
			wantErr: true,
			errMsg:  "invalid security zone 'invalid_zone'",
		},
		{
			name: "multiple valid security zones",
			mission: &MissionDefinition{
				MissionID:     "test_mission",
				Name:          "Test Mission",
				Description:   "A test mission",
				SecurityZones: []string{"high_security", "low_security"},
				Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:       &RewardDefinition{Credits: 1000},
			},
			wantErr: false,
		},
		{
			name: "empty objectives",
			mission: &MissionDefinition{
				MissionID:     "test_mission",
				Name:          "Test Mission",
				Description:   "A test mission",
				SecurityZones: []string{"high_security"},
				Objectives:    []*ObjectiveDefinition{},
				Rewards:       &RewardDefinition{Credits: 1000},
			},
			wantErr: true,
			errMsg:  "at least one objective is required",
		},
		{
			name: "nil rewards",
			mission: &MissionDefinition{
				MissionID:     "test_mission",
				Name:          "Test Mission",
				Description:   "A test mission",
				SecurityZones: []string{"high_security"},
				Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:       nil,
			},
			wantErr: true,
			errMsg:  "rewards definition is required",
		},
		{
			name: "negative reward credits",
			mission: &MissionDefinition{
				MissionID:     "test_mission",
				Name:          "Test Mission",
				Description:   "A test mission",
				SecurityZones: []string{"high_security"},
				Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:       &RewardDefinition{Credits: -100},
			},
			wantErr: true,
			errMsg:  "reward credits cannot be negative",
		},
		{
			name: "zero reward credits is valid",
			mission: &MissionDefinition{
				MissionID:     "test_mission",
				Name:          "Test Mission",
				Description:   "A test mission",
				SecurityZones: []string{"high_security"},
				Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:       &RewardDefinition{Credits: 0},
			},
			wantErr: false,
		},
		{
			name: "negative expiry ticks",
			mission: &MissionDefinition{
				MissionID:     "test_mission",
				Name:          "Test Mission",
				Description:   "A test mission",
				SecurityZones: []string{"high_security"},
				Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:       &RewardDefinition{Credits: 1000},
				ExpiryTicks:   intPtr(-100),
			},
			wantErr: true,
			errMsg:  "expiry_ticks must be positive",
		},
		{
			name: "zero expiry ticks",
			mission: &MissionDefinition{
				MissionID:     "test_mission",
				Name:          "Test Mission",
				Description:   "A test mission",
				SecurityZones: []string{"high_security"},
				Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:       &RewardDefinition{Credits: 1000},
				ExpiryTicks:   intPtr(0),
			},
			wantErr: true,
			errMsg:  "expiry_ticks must be positive",
		},
		{
			name: "valid expiry ticks",
			mission: &MissionDefinition{
				MissionID:     "test_mission",
				Name:          "Test Mission",
				Description:   "A test mission",
				SecurityZones: []string{"high_security"},
				Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:       &RewardDefinition{Credits: 1000},
				ExpiryTicks:   intPtr(500),
			},
			wantErr: false,
		},
		{
			name: "repeatable with negative cooldown",
			mission: &MissionDefinition{
				MissionID:           "test_mission",
				Name:                "Test Mission",
				Description:         "A test mission",
				SecurityZones:       []string{"high_security"},
				Objectives:          []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:             &RewardDefinition{Credits: 1000},
				Repeatable:          true,
				RepeatCooldownTicks: -100,
			},
			wantErr: true,
			errMsg:  "repeat_cooldown_ticks cannot be negative",
		},
		{
			name: "repeatable with zero cooldown",
			mission: &MissionDefinition{
				MissionID:           "test_mission",
				Name:                "Test Mission",
				Description:         "A test mission",
				SecurityZones:       []string{"high_security"},
				Objectives:          []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:             &RewardDefinition{Credits: 1000},
				Repeatable:          true,
				RepeatCooldownTicks: 0,
			},
			wantErr: false,
		},
		{
			name: "repeatable with valid cooldown",
			mission: &MissionDefinition{
				MissionID:           "test_mission",
				Name:                "Test Mission",
				Description:         "A test mission",
				SecurityZones:       []string{"high_security"},
				Objectives:          []*ObjectiveDefinition{validDeliverObjective()},
				Rewards:             &RewardDefinition{Credits: 1000},
				Repeatable:          true,
				RepeatCooldownTicks: 600,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMissionDefinition(tt.mission)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateObjectiveDefinition tests objective validation
func TestValidateObjectiveDefinition(t *testing.T) {
	tests := []struct {
		name      string
		objective *ObjectiveDefinition
		missionID string
		index     int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid deliver_commodity objective",
			objective: validDeliverObjective(),
			missionID: "test_mission",
			index:     0,
			wantErr:   false,
		},
		{
			name:      "nil objective",
			objective: nil,
			missionID: "test_mission",
			index:     0,
			wantErr:   true,
			errMsg:    "objective 0 is nil",
		},
		{
			name: "missing objective_id",
			objective: &ObjectiveDefinition{
				Type:        "deliver_commodity",
				Description: "Deliver goods",
				Parameters: map[string]interface{}{
					"commodity_id":        "food_supplies",
					"quantity":            float64(10),
					"destination_port_id": float64(1),
				},
			},
			missionID: "test_mission",
			index:     0,
			wantErr:   true,
			errMsg:    "objective_id is required",
		},
		{
			name: "missing type",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Description: "Deliver goods",
				Parameters: map[string]interface{}{
					"commodity_id":        "food_supplies",
					"quantity":            float64(10),
					"destination_port_id": float64(1),
				},
			},
			missionID: "test_mission",
			index:     0,
			wantErr:   true,
			errMsg:    "type is required",
		},
		{
			name: "missing description",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Parameters: map[string]interface{}{
					"commodity_id":        "food_supplies",
					"quantity":            float64(10),
					"destination_port_id": float64(1),
				},
			},
			missionID: "test_mission",
			index:     0,
			wantErr:   true,
			errMsg:    "description is required",
		},
		{
			name: "unknown objective type",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "unknown_type",
				Description: "Do something",
				Parameters:  map[string]interface{}{},
			},
			missionID: "test_mission",
			index:     0,
			wantErr:   true,
			errMsg:    "unknown objective type 'unknown_type'",
		},
		{
			name: "kill objective not supported in Phase 1",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "kill",
				Description: "Kill enemies",
				Parameters:  map[string]interface{}{},
			},
			missionID: "test_mission",
			index:     0,
			wantErr:   true,
			errMsg:    "'kill' objectives not supported in Phase 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateObjectiveDefinition(tt.objective, tt.missionID, tt.index)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateDeliverCommodityObjective tests deliver_commodity objective validation
func TestValidateDeliverCommodityObjective(t *testing.T) {
	tests := []struct {
		name      string
		objective *ObjectiveDefinition
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid deliver_commodity objective",
			objective: validDeliverObjective(),
			wantErr:   false,
		},
		{
			name: "missing parameters",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver goods",
				Parameters:  nil,
			},
			wantErr: true,
			errMsg:  "parameters required for deliver_commodity",
		},
		{
			name: "missing commodity_id",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver goods",
				Parameters: map[string]interface{}{
					"quantity":            float64(10),
					"destination_port_id": float64(1),
				},
			},
			wantErr: true,
			errMsg:  "commodity_id parameter required",
		},
		{
			name: "empty commodity_id",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver goods",
				Parameters: map[string]interface{}{
					"commodity_id":        "",
					"quantity":            float64(10),
					"destination_port_id": float64(1),
				},
			},
			wantErr: true,
			errMsg:  "commodity_id parameter required",
		},
		{
			name: "missing quantity",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver goods",
				Parameters: map[string]interface{}{
					"commodity_id":        "food_supplies",
					"destination_port_id": float64(1),
				},
			},
			wantErr: true,
			errMsg:  "quantity parameter must be positive",
		},
		{
			name: "zero quantity",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver goods",
				Parameters: map[string]interface{}{
					"commodity_id":        "food_supplies",
					"quantity":            float64(0),
					"destination_port_id": float64(1),
				},
			},
			wantErr: true,
			errMsg:  "quantity parameter must be positive",
		},
		{
			name: "negative quantity",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver goods",
				Parameters: map[string]interface{}{
					"commodity_id":        "food_supplies",
					"quantity":            float64(-5),
					"destination_port_id": float64(1),
				},
			},
			wantErr: true,
			errMsg:  "quantity parameter must be positive",
		},
		{
			name: "missing destination_port_id",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver goods",
				Parameters: map[string]interface{}{
					"commodity_id": "food_supplies",
					"quantity":     float64(10),
				},
			},
			wantErr: true,
			errMsg:  "destination_port_id parameter required",
		},
		{
			name: "zero destination_port_id",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver goods",
				Parameters: map[string]interface{}{
					"commodity_id":        "food_supplies",
					"quantity":            float64(10),
					"destination_port_id": float64(0),
				},
			},
			wantErr: true,
			errMsg:  "destination_port_id parameter required",
		},
		{
			name: "negative destination_port_id",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver goods",
				Parameters: map[string]interface{}{
					"commodity_id":        "food_supplies",
					"quantity":            float64(10),
					"destination_port_id": float64(-1),
				},
			},
			wantErr: true,
			errMsg:  "destination_port_id parameter required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDeliverCommodityObjective(tt.objective, "test_mission", 0)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateNavigateToObjective tests navigate_to objective validation
func TestValidateNavigateToObjective(t *testing.T) {
	tests := []struct {
		name      string
		objective *ObjectiveDefinition
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid navigate_to objective",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "navigate_to",
				Description: "Navigate to system",
				Parameters: map[string]interface{}{
					"system_id": float64(5),
				},
			},
			wantErr: false,
		},
		{
			name: "missing parameters",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "navigate_to",
				Description: "Navigate to system",
				Parameters:  nil,
			},
			wantErr: true,
			errMsg:  "parameters required for navigate_to",
		},
		{
			name: "missing system_id",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "navigate_to",
				Description: "Navigate to system",
				Parameters:  map[string]interface{}{},
			},
			wantErr: true,
			errMsg:  "system_id parameter required",
		},
		{
			name: "zero system_id",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "navigate_to",
				Description: "Navigate to system",
				Parameters: map[string]interface{}{
					"system_id": float64(0),
				},
			},
			wantErr: true,
			errMsg:  "system_id parameter required",
		},
		{
			name: "negative system_id",
			objective: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "navigate_to",
				Description: "Navigate to system",
				Parameters: map[string]interface{}{
					"system_id": float64(-1),
				},
			},
			wantErr: true,
			errMsg:  "system_id parameter required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNavigateToObjective(tt.objective, "test_mission", 0)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestIsValidSecurityZone tests security zone validation
func TestIsValidSecurityZone(t *testing.T) {
	tests := []struct {
		name string
		zone string
		want bool
	}{
		{"federated_space", "federated_space", true},
		{"high_security", "high_security", true},
		{"low_security", "low_security", true},
		{"medium_security", "medium_security", true},
		{"black_sector", "black_sector", true},
		{"invalid zone", "invalid_zone", false},
		{"empty string", "", false},
		{"high with space", "high security", false},
		{"uppercase", "HIGH_SECURITY", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidSecurityZone(tt.zone)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestValidateMissionFile tests mission file validation
func TestValidateMissionFile(t *testing.T) {
	tests := []struct {
		name    string
		file    *MissionFile
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid mission file with one mission",
			file: &MissionFile{
				Missions: []*MissionDefinition{
					{
						MissionID:     "mission1",
						Name:          "Mission 1",
						Description:   "First mission",
						SecurityZones: []string{"high_security"},
						Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
						Rewards:       &RewardDefinition{Credits: 1000},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid mission file with multiple missions",
			file: &MissionFile{
				Missions: []*MissionDefinition{
					{
						MissionID:     "mission1",
						Name:          "Mission 1",
						Description:   "First mission",
						SecurityZones: []string{"high_security"},
						Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
						Rewards:       &RewardDefinition{Credits: 1000},
					},
					{
						MissionID:     "mission2",
						Name:          "Mission 2",
						Description:   "Second mission",
						SecurityZones: []string{"low_security"},
						Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
						Rewards:       &RewardDefinition{Credits: 2000},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "nil mission file",
			file:    nil,
			wantErr: true,
			errMsg:  "mission file is nil",
		},
		{
			name: "empty missions array",
			file: &MissionFile{
				Missions: []*MissionDefinition{},
			},
			wantErr: true,
			errMsg:  "mission file contains no missions",
		},
		{
			name: "duplicate mission IDs",
			file: &MissionFile{
				Missions: []*MissionDefinition{
					{
						MissionID:     "mission1",
						Name:          "Mission 1",
						Description:   "First mission",
						SecurityZones: []string{"high_security"},
						Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
						Rewards:       &RewardDefinition{Credits: 1000},
					},
					{
						MissionID:     "mission1",
						Name:          "Mission 1 Duplicate",
						Description:   "Duplicate mission",
						SecurityZones: []string{"low_security"},
						Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
						Rewards:       &RewardDefinition{Credits: 2000},
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate mission_id: mission1",
		},
		{
			name: "invalid mission in file",
			file: &MissionFile{
				Missions: []*MissionDefinition{
					{
						MissionID:     "mission1",
						Name:          "Mission 1",
						Description:   "First mission",
						SecurityZones: []string{"high_security"},
						Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
						Rewards:       &RewardDefinition{Credits: 1000},
					},
					{
						MissionID:     "", // Invalid: missing mission_id
						Name:          "Mission 2",
						Description:   "Second mission",
						SecurityZones: []string{"low_security"},
						Objectives:    []*ObjectiveDefinition{validDeliverObjective()},
						Rewards:       &RewardDefinition{Credits: 2000},
					},
				},
			},
			wantErr: true,
			errMsg:  "mission_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMissionFile(tt.file)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions

// validDeliverObjective returns a valid deliver_commodity objective for testing
func validDeliverObjective() *ObjectiveDefinition {
	return &ObjectiveDefinition{
		ObjectiveID: "deliver_obj",
		Type:        "deliver_commodity",
		Description: "Deliver commodity",
		Parameters: map[string]interface{}{
			"commodity_id":        "food_supplies",
			"quantity":            float64(10),
			"destination_port_id": float64(1),
		},
	}
}

// intPtr returns a pointer to an int
func intPtr(i int) *int {
	return &i
}
