package missions

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDatabase implements the Database interface for testing
type MockDatabase struct {
	ships             map[string]*Ship
	cargo             map[string][]*CargoSlot
	players           map[string]*Player
	ports             map[int]*Port
	missionInstances  map[string]*MissionInstance
	objectiveProgress map[string]map[int]*ObjectiveProgress
	systemSecurity    map[int]float64
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		ships:             make(map[string]*Ship),
		cargo:             make(map[string][]*CargoSlot),
		players:           make(map[string]*Player),
		ports:             make(map[int]*Port),
		missionInstances:  make(map[string]*MissionInstance),
		objectiveProgress: make(map[string]map[int]*ObjectiveProgress),
		systemSecurity:    make(map[int]float64),
	}
}

func (m *MockDatabase) CreateMissionInstance(instance *MissionInstance) error {
	m.missionInstances[instance.InstanceID] = instance
	return nil
}

func (m *MockDatabase) GetMissionInstance(instanceID string) (*MissionInstance, error) {
	instance, exists := m.missionInstances[instanceID]
	if !exists {
		return nil, nil
	}
	return instance, nil
}

func (m *MockDatabase) GetActiveMissionByPlayer(playerID string) (*MissionInstance, error) {
	for _, instance := range m.missionInstances {
		if instance.PlayerID == playerID && instance.Status == MissionInProgress {
			return instance, nil
		}
	}
	return nil, nil
}

func (m *MockDatabase) GetAllInProgressMissions() ([]*MissionInstance, error) {
	var missions []*MissionInstance
	for _, instance := range m.missionInstances {
		if instance.Status == MissionInProgress {
			missions = append(missions, instance)
		}
	}
	return missions, nil
}

func (m *MockDatabase) GetCompletedMissionsByPlayer(playerID string) ([]*MissionInstance, error) {
	var missions []*MissionInstance
	for _, instance := range m.missionInstances {
		if instance.PlayerID == playerID && instance.Status == MissionCompleted {
			missions = append(missions, instance)
		}
	}
	return missions, nil
}

func (m *MockDatabase) UpdateMissionStatus(instanceID string, status string, tick int64) error {
	instance, exists := m.missionInstances[instanceID]
	if !exists {
		return nil
	}
	instance.Status = MissionStatus(status)
	if status == string(MissionCompleted) || status == string(MissionFailed) || 
	   status == string(MissionExpired) || status == string(MissionAbandoned) {
		instance.CompletedTick = &tick
	}
	return nil
}

func (m *MockDatabase) UpdateMissionObjectiveIndex(instanceID string, objectiveIndex int) error {
	return nil
}

func (m *MockDatabase) DeleteMissionInstance(instanceID string) error {
	delete(m.missionInstances, instanceID)
	delete(m.objectiveProgress, instanceID)
	return nil
}

func (m *MockDatabase) CreateObjectiveProgress(progress *ObjectiveProgress) error {
	if m.objectiveProgress[progress.InstanceID] == nil {
		m.objectiveProgress[progress.InstanceID] = make(map[int]*ObjectiveProgress)
	}
	m.objectiveProgress[progress.InstanceID][progress.ObjectiveIndex] = progress
	return nil
}

func (m *MockDatabase) GetObjectiveProgress(instanceID string, objectiveIndex int) (*ObjectiveProgress, error) {
	if m.objectiveProgress[instanceID] == nil {
		return nil, nil
	}
	progress, exists := m.objectiveProgress[instanceID][objectiveIndex]
	if !exists {
		return nil, nil
	}
	return progress, nil
}

func (m *MockDatabase) GetAllObjectiveProgress(instanceID string) ([]*ObjectiveProgress, error) {
	var progressList []*ObjectiveProgress
	if m.objectiveProgress[instanceID] == nil {
		return progressList, nil
	}
	for i := 0; i < len(m.objectiveProgress[instanceID]); i++ {
		if progress, exists := m.objectiveProgress[instanceID][i]; exists {
			progressList = append(progressList, progress)
		}
	}
	return progressList, nil
}

func (m *MockDatabase) UpdateObjectiveProgress(instanceID string, objectiveIndex int, status string, currentValue int) error {
	if m.objectiveProgress[instanceID] == nil {
		return nil
	}
	progress, exists := m.objectiveProgress[instanceID][objectiveIndex]
	if !exists {
		return nil
	}
	progress.Status = status
	progress.CurrentValue = currentValue
	return nil
}

func (m *MockDatabase) DeleteObjectiveProgress(instanceID string) error {
	delete(m.objectiveProgress, instanceID)
	return nil
}

func (m *MockDatabase) GetPlayerByID(playerID string) (*Player, error) {
	player, exists := m.players[playerID]
	if !exists {
		return nil, nil
	}
	return player, nil
}

func (m *MockDatabase) GetShipByPlayerID(playerID string) (*Ship, error) {
	for _, ship := range m.ships {
		if ship.PlayerID == playerID {
			return ship, nil
		}
	}
	return nil, nil
}

func (m *MockDatabase) UpdatePlayerCredits(playerID string, credits int) error {
	player, exists := m.players[playerID]
	if !exists {
		return nil
	}
	player.Credits = int64(credits)
	return nil
}

func (m *MockDatabase) GetCargoByShipID(shipID string) ([]*CargoSlot, error) {
	cargo, exists := m.cargo[shipID]
	if !exists {
		return []*CargoSlot{}, nil
	}
	return cargo, nil
}

func (m *MockDatabase) GetPortByID(portID int) (*Port, error) {
	port, exists := m.ports[portID]
	if !exists {
		return nil, nil
	}
	return port, nil
}

func (m *MockDatabase) GetSystemSecurityLevel(systemID int) (float64, error) {
	level, exists := m.systemSecurity[systemID]
	if !exists {
		return 0, nil
	}
	return level, nil
}

// TestEvaluateDeliverCommodity tests the deliver_commodity objective evaluation
func TestEvaluateDeliverCommodity(t *testing.T) {
	logger := zerolog.Nop()
	
	tests := []struct {
		name           string
		setupDB        func(*MockDatabase)
		objectiveDef   *ObjectiveDefinition
		progress       *ObjectiveProgress
		wantCompleted  bool
		wantErr        bool
		errContains    string
	}{
		{
			name: "objective complete - exact quantity at destination",
			setupDB: func(db *MockDatabase) {
				portID := 101
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 1,
					Status:          "DOCKED",
					DockedAtPortID:  &portID,
				}
				db.cargo["ship1"] = []*CargoSlot{
					{ShipID: "ship1", CommodityID: "WATER", Quantity: 100},
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver water",
				Parameters: map[string]interface{}{
					"commodity_id":        "WATER",
					"quantity":            float64(100),
					"destination_port_id": float64(101),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  100,
			},
			wantCompleted: true,
			wantErr:       false,
		},
		{
			name: "objective complete - more than required quantity",
			setupDB: func(db *MockDatabase) {
				portID := 101
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 1,
					Status:          "DOCKED",
					DockedAtPortID:  &portID,
				}
				db.cargo["ship1"] = []*CargoSlot{
					{ShipID: "ship1", CommodityID: "WATER", Quantity: 150},
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver water",
				Parameters: map[string]interface{}{
					"commodity_id":        "WATER",
					"quantity":            float64(100),
					"destination_port_id": float64(101),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  100,
			},
			wantCompleted: true,
			wantErr:       false,
		},
		{
			name: "objective incomplete - insufficient quantity",
			setupDB: func(db *MockDatabase) {
				portID := 101
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 1,
					Status:          "DOCKED",
					DockedAtPortID:  &portID,
				}
				db.cargo["ship1"] = []*CargoSlot{
					{ShipID: "ship1", CommodityID: "WATER", Quantity: 50},
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver water",
				Parameters: map[string]interface{}{
					"commodity_id":        "WATER",
					"quantity":            float64(100),
					"destination_port_id": float64(101),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  100,
			},
			wantCompleted: false,
			wantErr:       false,
		},
		{
			name: "objective incomplete - wrong port",
			setupDB: func(db *MockDatabase) {
				portID := 102
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 1,
					Status:          "DOCKED",
					DockedAtPortID:  &portID,
				}
				db.cargo["ship1"] = []*CargoSlot{
					{ShipID: "ship1", CommodityID: "WATER", Quantity: 100},
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver water",
				Parameters: map[string]interface{}{
					"commodity_id":        "WATER",
					"quantity":            float64(100),
					"destination_port_id": float64(101),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  100,
			},
			wantCompleted: false,
			wantErr:       false,
		},
		{
			name: "objective incomplete - not docked",
			setupDB: func(db *MockDatabase) {
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 1,
					Status:          "IN_SPACE",
					DockedAtPortID:  nil,
				}
				db.cargo["ship1"] = []*CargoSlot{
					{ShipID: "ship1", CommodityID: "WATER", Quantity: 100},
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver water",
				Parameters: map[string]interface{}{
					"commodity_id":        "WATER",
					"quantity":            float64(100),
					"destination_port_id": float64(101),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  100,
			},
			wantCompleted: false,
			wantErr:       false,
		},
		{
			name: "objective incomplete - wrong commodity",
			setupDB: func(db *MockDatabase) {
				portID := 101
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 1,
					Status:          "DOCKED",
					DockedAtPortID:  &portID,
				}
				db.cargo["ship1"] = []*CargoSlot{
					{ShipID: "ship1", CommodityID: "FUEL", Quantity: 100},
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver water",
				Parameters: map[string]interface{}{
					"commodity_id":        "WATER",
					"quantity":            float64(100),
					"destination_port_id": float64(101),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  100,
			},
			wantCompleted: false,
			wantErr:       false,
		},
		{
			name: "objective incomplete - no cargo",
			setupDB: func(db *MockDatabase) {
				portID := 101
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 1,
					Status:          "DOCKED",
					DockedAtPortID:  &portID,
				}
				db.cargo["ship1"] = []*CargoSlot{}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver water",
				Parameters: map[string]interface{}{
					"commodity_id":        "WATER",
					"quantity":            float64(100),
					"destination_port_id": float64(101),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  100,
			},
			wantCompleted: false,
			wantErr:       false,
		},
		{
			name: "objective complete - multiple cargo slots with same commodity",
			setupDB: func(db *MockDatabase) {
				portID := 101
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 1,
					Status:          "DOCKED",
					DockedAtPortID:  &portID,
				}
				db.cargo["ship1"] = []*CargoSlot{
					{ShipID: "ship1", CommodityID: "WATER", Quantity: 60},
					{ShipID: "ship1", CommodityID: "WATER", Quantity: 40},
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver water",
				Parameters: map[string]interface{}{
					"commodity_id":        "WATER",
					"quantity":            float64(100),
					"destination_port_id": float64(101),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  100,
			},
			wantCompleted: true,
			wantErr:       false,
		},
		{
			name: "error - missing destination_port_id parameter",
			setupDB: func(db *MockDatabase) {
				portID := 101
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 1,
					Status:          "DOCKED",
					DockedAtPortID:  &portID,
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver water",
				Parameters: map[string]interface{}{
					"commodity_id": "WATER",
					"quantity":     float64(100),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  100,
			},
			wantCompleted: false,
			wantErr:       true,
			errContains:   "destination_port_id",
		},
		{
			name: "error - missing commodity_id parameter",
			setupDB: func(db *MockDatabase) {
				portID := 101
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 1,
					Status:          "DOCKED",
					DockedAtPortID:  &portID,
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver water",
				Parameters: map[string]interface{}{
					"quantity":            float64(100),
					"destination_port_id": float64(101),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  100,
			},
			wantCompleted: false,
			wantErr:       true,
			errContains:   "commodity_id",
		},
		{
			name: "error - missing quantity parameter",
			setupDB: func(db *MockDatabase) {
				portID := 101
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 1,
					Status:          "DOCKED",
					DockedAtPortID:  &portID,
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver water",
				Parameters: map[string]interface{}{
					"commodity_id":        "WATER",
					"destination_port_id": float64(101),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  100,
			},
			wantCompleted: false,
			wantErr:       true,
			errContains:   "quantity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewMockDatabase()
			tt.setupDB(db)
			
			manager := NewMissionManager(DefaultConfig(), db, logger)
			
			instance := &MissionInstance{
				InstanceID: "mission1",
				MissionID:  "test_mission",
				PlayerID:   "player1",
				Status:     MissionInProgress,
			}
			
			completed, err := manager.evaluateDeliverCommodity(instance, tt.objectiveDef, tt.progress)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCompleted, completed)
			}
		})
	}
}

// TestEvaluateNavigateTo tests the navigate_to objective evaluation
func TestEvaluateNavigateTo(t *testing.T) {
	logger := zerolog.Nop()
	
	tests := []struct {
		name          string
		setupDB       func(*MockDatabase)
		objectiveDef  *ObjectiveDefinition
		progress      *ObjectiveProgress
		wantCompleted bool
		wantErr       bool
		errContains   string
	}{
		{
			name: "objective complete - at destination system",
			setupDB: func(db *MockDatabase) {
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 5,
					Status:          "IN_SPACE",
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "navigate_to",
				Description: "Navigate to system 5",
				Parameters: map[string]interface{}{
					"destination_system_id": float64(5),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  1,
			},
			wantCompleted: true,
			wantErr:       false,
		},
		{
			name: "objective incomplete - wrong system",
			setupDB: func(db *MockDatabase) {
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 3,
					Status:          "IN_SPACE",
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "navigate_to",
				Description: "Navigate to system 5",
				Parameters: map[string]interface{}{
					"destination_system_id": float64(5),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  1,
			},
			wantCompleted: false,
			wantErr:       false,
		},
		{
			name: "objective complete - docked at port in destination system",
			setupDB: func(db *MockDatabase) {
				portID := 101
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 5,
					Status:          "DOCKED",
					DockedAtPortID:  &portID,
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "navigate_to",
				Description: "Navigate to system 5",
				Parameters: map[string]interface{}{
					"destination_system_id": float64(5),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  1,
			},
			wantCompleted: true,
			wantErr:       false,
		},
		{
			name: "error - missing destination_system_id parameter",
			setupDB: func(db *MockDatabase) {
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 5,
					Status:          "IN_SPACE",
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "navigate_to",
				Description: "Navigate to system 5",
				Parameters:  map[string]interface{}{},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  1,
			},
			wantCompleted: false,
			wantErr:       true,
			errContains:   "destination_system_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewMockDatabase()
			tt.setupDB(db)
			
			manager := NewMissionManager(DefaultConfig(), db, logger)
			
			instance := &MissionInstance{
				InstanceID: "mission1",
				MissionID:  "test_mission",
				PlayerID:   "player1",
				Status:     MissionInProgress,
			}
			
			completed, err := manager.evaluateNavigateTo(instance, tt.objectiveDef, tt.progress)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCompleted, completed)
			}
		})
	}
}

// TestEvaluateObjective tests the main evaluateObjective dispatcher
func TestEvaluateObjective(t *testing.T) {
	logger := zerolog.Nop()
	
	tests := []struct {
		name          string
		setupDB       func(*MockDatabase)
		objectiveDef  *ObjectiveDefinition
		progress      *ObjectiveProgress
		wantCompleted bool
		wantErr       bool
	}{
		{
			name: "deliver_commodity objective",
			setupDB: func(db *MockDatabase) {
				portID := 101
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 1,
					Status:          "DOCKED",
					DockedAtPortID:  &portID,
				}
				db.cargo["ship1"] = []*CargoSlot{
					{ShipID: "ship1", CommodityID: "WATER", Quantity: 100},
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "deliver_commodity",
				Description: "Deliver water",
				Parameters: map[string]interface{}{
					"commodity_id":        "WATER",
					"quantity":            float64(100),
					"destination_port_id": float64(101),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  100,
			},
			wantCompleted: true,
			wantErr:       false,
		},
		{
			name: "navigate_to objective",
			setupDB: func(db *MockDatabase) {
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 5,
					Status:          "IN_SPACE",
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "navigate_to",
				Description: "Navigate to system 5",
				Parameters: map[string]interface{}{
					"destination_system_id": float64(5),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  1,
			},
			wantCompleted: true,
			wantErr:       false,
		},
		{
			name: "unsupported objective type",
			setupDB: func(db *MockDatabase) {
				db.ships["ship1"] = &Ship{
					ShipID:          "ship1",
					PlayerID:        "player1",
					CurrentSystemID: 1,
					Status:          "IN_SPACE",
				}
			},
			objectiveDef: &ObjectiveDefinition{
				ObjectiveID: "obj1",
				Type:        "kill_enemy",
				Description: "Kill 5 pirates",
				Parameters: map[string]interface{}{
					"enemy_type": "pirate",
					"count":      float64(5),
				},
			},
			progress: &ObjectiveProgress{
				InstanceID:     "mission1",
				ObjectiveIndex: 0,
				Status:         string(ObjectiveActive),
				CurrentValue:   0,
				RequiredValue:  5,
			},
			wantCompleted: false,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewMockDatabase()
			tt.setupDB(db)
			
			manager := NewMissionManager(DefaultConfig(), db, logger)
			
			instance := &MissionInstance{
				InstanceID: "mission1",
				MissionID:  "test_mission",
				PlayerID:   "player1",
				Status:     MissionInProgress,
			}
			
			completed, err := manager.evaluateObjective(instance, tt.objectiveDef, tt.progress, 1000)
			
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCompleted, completed)
			}
		})
	}
}

// TestExtractRequiredValue tests the extractRequiredValue helper
func TestExtractRequiredValue(t *testing.T) {
	logger := zerolog.Nop()
	manager := NewMissionManager(DefaultConfig(), NewMockDatabase(), logger)
	
	tests := []struct {
		name         string
		objectiveDef *ObjectiveDefinition
		want         int
	}{
		{
			name: "deliver_commodity with quantity",
			objectiveDef: &ObjectiveDefinition{
				Type: "deliver_commodity",
				Parameters: map[string]interface{}{
					"quantity": float64(100),
				},
			},
			want: 100,
		},
		{
			name: "navigate_to defaults to 1",
			objectiveDef: &ObjectiveDefinition{
				Type:       "navigate_to",
				Parameters: map[string]interface{}{},
			},
			want: 1,
		},
		{
			name: "unknown type defaults to 1",
			objectiveDef: &ObjectiveDefinition{
				Type:       "unknown_type",
				Parameters: map[string]interface{}{},
			},
			want: 1,
		},
		{
			name: "deliver_commodity without quantity defaults to 1",
			objectiveDef: &ObjectiveDefinition{
				Type:       "deliver_commodity",
				Parameters: map[string]interface{}{},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := manager.extractRequiredValue(tt.objectiveDef)
			assert.Equal(t, tt.want, got)
		})
	}
}
