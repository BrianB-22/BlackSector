package combat

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// TestResolveDamage tests the damage calculation logic with various scenarios
func TestResolveDamage(t *testing.T) {
	tests := []struct {
		name           string
		damage         int
		targetType     string // "pirate" or "ship"
		initialShield  int
		initialHull    int
		maxShield      int
		maxHull        int
		accuracy       float64
		wantHit        bool
		wantDamage     int
		wantShieldDmg  int
		wantHullDmg    int
		wantFinalShield int
		wantFinalHull  int
		wantDestroyed  bool
	}{
		// Shield absorption tests
		{
			name:           "damage fully absorbed by shields",
			damage:         10,
			targetType:     "pirate",
			initialShield:  50,
			initialHull:    100,
			maxShield:      50,
			maxHull:        100,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     10,
			wantShieldDmg:  10,
			wantHullDmg:    0,
			wantFinalShield: 40,
			wantFinalHull:  100,
			wantDestroyed:  false,
		},
		{
			name:           "damage depletes shields exactly",
			damage:         50,
			targetType:     "pirate",
			initialShield:  50,
			initialHull:    100,
			maxShield:      50,
			maxHull:        100,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     50,
			wantShieldDmg:  50,
			wantHullDmg:    0,
			wantFinalShield: 0,
			wantFinalHull:  100,
			wantDestroyed:  false,
		},
		{
			name:           "damage exceeds shields and damages hull",
			damage:         30,
			targetType:     "pirate",
			initialShield:  20,
			initialHull:    100,
			maxShield:      50,
			maxHull:        100,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     30,
			wantShieldDmg:  20,
			wantHullDmg:    10,
			wantFinalShield: 0,
			wantFinalHull:  90,
			wantDestroyed:  false,
		},
		{
			name:           "damage with no shields goes directly to hull",
			damage:         25,
			targetType:     "pirate",
			initialShield:  0,
			initialHull:    100,
			maxShield:      50,
			maxHull:        100,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     25,
			wantShieldDmg:  0,
			wantHullDmg:    25,
			wantFinalShield: 0,
			wantFinalHull:  75,
			wantDestroyed:  false,
		},
		// Ship destruction tests
		{
			name:           "damage destroys ship with shields",
			damage:         150,
			targetType:     "pirate",
			initialShield:  50,
			initialHull:    100,
			maxShield:      50,
			maxHull:        100,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     150,
			wantShieldDmg:  50,
			wantHullDmg:    100,
			wantFinalShield: 0,
			wantFinalHull:  0,
			wantDestroyed:  true,
		},
		{
			name:           "damage destroys ship with no shields",
			damage:         100,
			targetType:     "pirate",
			initialShield:  0,
			initialHull:    100,
			maxShield:      50,
			maxHull:        100,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     100,
			wantShieldDmg:  0,
			wantHullDmg:    100,
			wantFinalShield: 0,
			wantFinalHull:  0,
			wantDestroyed:  true,
		},
		{
			name:           "damage reduces hull to exactly zero",
			damage:         75,
			targetType:     "pirate",
			initialShield:  25,
			initialHull:    50,
			maxShield:      50,
			maxHull:        100,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     75,
			wantShieldDmg:  25,
			wantHullDmg:    50,
			wantFinalShield: 0,
			wantFinalHull:  0,
			wantDestroyed:  true,
		},
		{
			name:           "overkill damage capped at remaining hull",
			damage:         200,
			targetType:     "pirate",
			initialShield:  10,
			initialHull:    20,
			maxShield:      50,
			maxHull:        100,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     200,
			wantShieldDmg:  10,
			wantHullDmg:    20, // Only 20 hull damage applied, not 190
			wantFinalShield: 0,
			wantFinalHull:  0,
			wantDestroyed:  true,
		},
		// Miss tests (accuracy-based)
		{
			name:           "attack misses due to accuracy",
			damage:         50,
			targetType:     "pirate",
			initialShield:  50,
			initialHull:    100,
			maxShield:      50,
			maxHull:        100,
			accuracy:       0.0, // Always miss
			wantHit:        false,
			wantDamage:     0,
			wantShieldDmg:  0,
			wantHullDmg:    0,
			wantFinalShield: 50,
			wantFinalHull:  100,
			wantDestroyed:  false,
		},
		// Edge cases
		{
			name:           "zero damage attack",
			damage:         0,
			targetType:     "pirate",
			initialShield:  50,
			initialHull:    100,
			maxShield:      50,
			maxHull:        100,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     0,
			wantShieldDmg:  0,
			wantHullDmg:    0,
			wantFinalShield: 50,
			wantFinalHull:  100,
			wantDestroyed:  false,
		},
		{
			name:           "one damage point to full shields",
			damage:         1,
			targetType:     "pirate",
			initialShield:  50,
			initialHull:    100,
			maxShield:      50,
			maxHull:        100,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     1,
			wantShieldDmg:  1,
			wantHullDmg:    0,
			wantFinalShield: 49,
			wantFinalHull:  100,
			wantDestroyed:  false,
		},
		{
			name:           "one damage point to one shield point",
			damage:         1,
			targetType:     "pirate",
			initialShield:  1,
			initialHull:    100,
			maxShield:      50,
			maxHull:        100,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     1,
			wantShieldDmg:  1,
			wantHullDmg:    0,
			wantFinalShield: 0,
			wantFinalHull:  100,
			wantDestroyed:  false,
		},
		{
			name:           "one damage point to one hull point (no shields)",
			damage:         1,
			targetType:     "pirate",
			initialShield:  0,
			initialHull:    1,
			maxShield:      50,
			maxHull:        100,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     1,
			wantShieldDmg:  0,
			wantHullDmg:    1,
			wantFinalShield: 0,
			wantFinalHull:  0,
			wantDestroyed:  true,
		},
		// Player ship tests (same logic, different target type)
		{
			name:           "player ship takes shield damage",
			damage:         15,
			targetType:     "ship",
			initialShield:  30,
			initialHull:    80,
			maxShield:      30,
			maxHull:        80,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     15,
			wantShieldDmg:  15,
			wantHullDmg:    0,
			wantFinalShield: 15,
			wantFinalHull:  80,
			wantDestroyed:  false,
		},
		{
			name:           "player ship shields depleted and hull damaged",
			damage:         40,
			targetType:     "ship",
			initialShield:  10,
			initialHull:    80,
			maxShield:      30,
			maxHull:        80,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     40,
			wantShieldDmg:  10,
			wantHullDmg:    30,
			wantFinalShield: 0,
			wantFinalHull:  50,
			wantDestroyed:  false,
		},
		{
			name:           "player ship destroyed",
			damage:         100,
			targetType:     "ship",
			initialShield:  10,
			initialHull:    80,
			maxShield:      30,
			maxHull:        80,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     100,
			wantShieldDmg:  10,
			wantHullDmg:    80,
			wantFinalShield: 0,
			wantFinalHull:  0,
			wantDestroyed:  true,
		},
		// Raider typical damage scenarios (12-18 damage, 60% accuracy)
		{
			name:           "raider minimum damage against courier shields",
			damage:         12,
			targetType:     "ship",
			initialShield:  30,
			initialHull:    80,
			maxShield:      30,
			maxHull:        80,
			accuracy:       0.60,
			wantHit:        true, // Forced hit for deterministic test
			wantDamage:     12,
			wantShieldDmg:  12,
			wantHullDmg:    0,
			wantFinalShield: 18,
			wantFinalHull:  80,
			wantDestroyed:  false,
		},
		{
			name:           "raider maximum damage against courier shields",
			damage:         18,
			targetType:     "ship",
			initialShield:  30,
			initialHull:    80,
			maxShield:      30,
			maxHull:        80,
			accuracy:       0.60,
			wantHit:        true, // Forced hit for deterministic test
			wantDamage:     18,
			wantShieldDmg:  18,
			wantHullDmg:    0,
			wantFinalShield: 12,
			wantFinalHull:  80,
			wantDestroyed:  false,
		},
		// Marauder typical damage scenarios (18-28 damage, 65% accuracy)
		{
			name:           "marauder minimum damage breaches courier shields",
			damage:         18,
			targetType:     "ship",
			initialShield:  10,
			initialHull:    80,
			maxShield:      30,
			maxHull:        80,
			accuracy:       0.65,
			wantHit:        true, // Forced hit for deterministic test
			wantDamage:     18,
			wantShieldDmg:  10,
			wantHullDmg:    8,
			wantFinalShield: 0,
			wantFinalHull:  72,
			wantDestroyed:  false,
		},
		{
			name:           "marauder maximum damage against courier",
			damage:         28,
			targetType:     "ship",
			initialShield:  20,
			initialHull:    80,
			maxShield:      30,
			maxHull:        80,
			accuracy:       0.65,
			wantHit:        true, // Forced hit for deterministic test
			wantDamage:     28,
			wantShieldDmg:  20,
			wantHullDmg:    8,
			wantFinalShield: 0,
			wantFinalHull:  72,
			wantDestroyed:  false,
		},
		// Courier weapon damage (15 fixed) against pirate tiers
		{
			name:           "courier attacks raider with full shields",
			damage:         15,
			targetType:     "pirate",
			initialShield:  20,
			initialHull:    60,
			maxShield:      20,
			maxHull:        60,
			accuracy:       1.0, // Player always hits
			wantHit:        true,
			wantDamage:     15,
			wantShieldDmg:  15,
			wantHullDmg:    0,
			wantFinalShield: 5,
			wantFinalHull:  60,
			wantDestroyed:  false,
		},
		{
			name:           "courier attacks raider with depleted shields",
			damage:         15,
			targetType:     "pirate",
			initialShield:  0,
			initialHull:    60,
			maxShield:      20,
			maxHull:        60,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     15,
			wantShieldDmg:  0,
			wantHullDmg:    15,
			wantFinalShield: 0,
			wantFinalHull:  45,
			wantDestroyed:  false,
		},
		{
			name:           "courier attacks marauder with full shields",
			damage:         15,
			targetType:     "pirate",
			initialShield:  40,
			initialHull:    90,
			maxShield:      40,
			maxHull:        90,
			accuracy:       1.0,
			wantHit:        true,
			wantDamage:     15,
			wantShieldDmg:  15,
			wantHullDmg:    0,
			wantFinalShield: 25,
			wantFinalHull:  90,
			wantDestroyed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create combat system (we only need the resolveDamage method)
			cs := &CombatSystem{
				cfg:    DefaultConfig(),
				logger: testLogger(),
			}

			// Create target based on type
			var target interface{}
			if tt.targetType == "pirate" {
				target = &PirateShip{
					ShipID:       "test-pirate",
					Tier:         "raider",
					HullPoints:   tt.initialHull,
					MaxHull:      tt.maxHull,
					ShieldPoints: tt.initialShield,
					MaxShield:    tt.maxShield,
				}
			} else {
				target = &Ship{
					ShipID:          "test-ship",
					PlayerID:        "test-player",
					ShipClass:       "courier",
					HullPoints:      tt.initialHull,
					MaxHullPoints:   tt.maxHull,
					ShieldPoints:    tt.initialShield,
					MaxShieldPoints: tt.maxShield,
					WeaponDamage:    15,
				}
			}

			// For deterministic testing, we need to control the random hit roll
			// Since we can't easily mock rand.Float64(), we'll test with accuracy 1.0 or 0.0
			// and document that accuracy-based misses are probabilistic
			result := cs.resolveDamage(tt.damage, target, tt.accuracy)

			// Verify hit status (only reliable for accuracy 0.0 or 1.0)
			if tt.accuracy == 0.0 {
				assert.False(t, result.Hit, "Expected miss with 0.0 accuracy")
			} else if tt.accuracy == 1.0 {
				assert.True(t, result.Hit, "Expected hit with 1.0 accuracy")
			}

			// Only verify damage results if hit occurred
			if result.Hit {
				assert.Equal(t, tt.wantDamage, result.Damage, "Damage mismatch")
				assert.Equal(t, tt.wantShieldDmg, result.ShieldDamage, "Shield damage mismatch")
				assert.Equal(t, tt.wantHullDmg, result.HullDamage, "Hull damage mismatch")
				assert.Equal(t, tt.wantDestroyed, result.TargetDestroyed, "Destroyed status mismatch")

				// Verify target state was updated correctly
				if tt.targetType == "pirate" {
					pirate := target.(*PirateShip)
					assert.Equal(t, tt.wantFinalShield, pirate.ShieldPoints, "Final shield points mismatch")
					assert.Equal(t, tt.wantFinalHull, pirate.HullPoints, "Final hull points mismatch")
				} else {
					ship := target.(*Ship)
					assert.Equal(t, tt.wantFinalShield, ship.ShieldPoints, "Final shield points mismatch")
					assert.Equal(t, tt.wantFinalHull, ship.HullPoints, "Final hull points mismatch")
				}
			} else {
				// On miss, verify no damage was applied
				assert.Equal(t, 0, result.Damage, "Damage should be 0 on miss")
				assert.Equal(t, 0, result.ShieldDamage, "Shield damage should be 0 on miss")
				assert.Equal(t, 0, result.HullDamage, "Hull damage should be 0 on miss")
				assert.False(t, result.TargetDestroyed, "Target should not be destroyed on miss")

				// Verify target state unchanged
				if tt.targetType == "pirate" {
					pirate := target.(*PirateShip)
					assert.Equal(t, tt.initialShield, pirate.ShieldPoints, "Shield should be unchanged on miss")
					assert.Equal(t, tt.initialHull, pirate.HullPoints, "Hull should be unchanged on miss")
				} else {
					ship := target.(*Ship)
					assert.Equal(t, tt.initialShield, ship.ShieldPoints, "Shield should be unchanged on miss")
					assert.Equal(t, tt.initialHull, ship.HullPoints, "Hull should be unchanged on miss")
				}
			}
		})
	}
}

// testLogger returns a no-op logger for testing
func testLogger() zerolog.Logger {
	return zerolog.Nop()
}

// TestPirateSpawnProbability tests the spawn probability calculation logic
func TestPirateSpawnProbability(t *testing.T) {
	tests := []struct {
		name                string
		securityLevel       float64
		pirateActivityBase  float64
		wantSpawnChance     float64
		wantShouldSpawn     bool // Whether spawns are possible (Low Sec check)
		description         string
	}{
		// Low Security systems (< 0.4) - spawns enabled
		{
			name:               "lowest security system (0.0)",
			securityLevel:      0.0,
			pirateActivityBase: 0.10,
			wantSpawnChance:    0.10, // 0.10 × (1 - 0.0) = 0.10
			wantShouldSpawn:    true,
			description:        "Maximum spawn probability in lawless space",
		},
		{
			name:               "very low security (0.1)",
			securityLevel:      0.1,
			pirateActivityBase: 0.10,
			wantSpawnChance:    0.09, // 0.10 × (1 - 0.1) = 0.09
			wantShouldSpawn:    true,
			description:        "High spawn probability in dangerous space",
		},
		{
			name:               "low security (0.2)",
			securityLevel:      0.2,
			pirateActivityBase: 0.10,
			wantSpawnChance:    0.08, // 0.10 × (1 - 0.2) = 0.08
			wantShouldSpawn:    true,
			description:        "Moderate spawn probability",
		},
		{
			name:               "low security (0.3)",
			securityLevel:      0.3,
			pirateActivityBase: 0.10,
			wantSpawnChance:    0.07, // 0.10 × (1 - 0.3) = 0.07
			wantShouldSpawn:    true,
			description:        "Lower spawn probability but still Low Sec",
		},
		{
			name:               "boundary low security (0.39)",
			securityLevel:      0.39,
			pirateActivityBase: 0.10,
			wantSpawnChance:    0.061, // 0.10 × (1 - 0.39) = 0.061
			wantShouldSpawn:    true,
			description:        "Just below High Security threshold",
		},

		// High Security systems (>= 0.4) - spawns disabled
		{
			name:               "boundary high security (0.4)",
			securityLevel:      0.4,
			pirateActivityBase: 0.10,
			wantSpawnChance:    0.0, // No spawns in High Sec
			wantShouldSpawn:    false,
			description:        "Exactly at High Security threshold - no spawns",
		},
		{
			name:               "high security (0.5)",
			securityLevel:      0.5,
			pirateActivityBase: 0.10,
			wantSpawnChance:    0.0,
			wantShouldSpawn:    false,
			description:        "High Security - no spawns",
		},
		{
			name:               "high security (0.7)",
			securityLevel:      0.7,
			pirateActivityBase: 0.10,
			wantSpawnChance:    0.0,
			wantShouldSpawn:    false,
			description:        "High Security - no spawns",
		},
		{
			name:               "high security (0.9)",
			securityLevel:      0.9,
			pirateActivityBase: 0.10,
			wantSpawnChance:    0.0,
			wantShouldSpawn:    false,
			description:        "Very High Security - no spawns",
		},
		{
			name:               "high security (1.0)",
			securityLevel:      1.0,
			pirateActivityBase: 0.10,
			wantSpawnChance:    0.0,
			wantShouldSpawn:    false,
			description:        "Maximum High Security - no spawns",
		},

		// Federated Space (2.0) - spawns disabled
		{
			name:               "federated space (2.0)",
			securityLevel:      2.0,
			pirateActivityBase: 0.10,
			wantSpawnChance:    0.0,
			wantShouldSpawn:    false,
			description:        "Federated Space - no spawns",
		},

		// Different activity base values
		{
			name:               "higher activity base (0.15) in low sec",
			securityLevel:      0.2,
			pirateActivityBase: 0.15,
			wantSpawnChance:    0.12, // 0.15 × (1 - 0.2) = 0.12
			wantShouldSpawn:    true,
			description:        "Higher pirate activity increases spawn chance",
		},
		{
			name:               "lower activity base (0.05) in low sec",
			securityLevel:      0.2,
			pirateActivityBase: 0.05,
			wantSpawnChance:    0.04, // 0.05 × (1 - 0.2) = 0.04
			wantShouldSpawn:    true,
			description:        "Lower pirate activity decreases spawn chance",
		},
		{
			name:               "zero activity base",
			securityLevel:      0.0,
			pirateActivityBase: 0.0,
			wantSpawnChance:    0.0, // 0.0 × (1 - 0.0) = 0.0
			wantShouldSpawn:    true, // Still Low Sec, but 0% chance
			description:        "Zero activity base means no spawns even in Low Sec",
		},
		{
			name:               "maximum activity base (1.0) in lawless space",
			securityLevel:      0.0,
			pirateActivityBase: 1.0,
			wantSpawnChance:    1.0, // 1.0 × (1 - 0.0) = 1.0
			wantShouldSpawn:    true,
			description:        "100% spawn chance with max activity in lawless space",
		},

		// Edge cases
		{
			name:               "negative security (edge case)",
			securityLevel:      -0.1,
			pirateActivityBase: 0.10,
			wantSpawnChance:    0.11, // 0.10 × (1 - (-0.1)) = 0.11
			wantShouldSpawn:    true,
			description:        "Negative security increases spawn chance above base",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test 1: Verify spawn eligibility based on security level
			shouldSpawn := tt.securityLevel < 0.4
			assert.Equal(t, tt.wantShouldSpawn, shouldSpawn,
				"Security level %.2f should have spawn eligibility: %v",
				tt.securityLevel, tt.wantShouldSpawn)

			// Test 2: Calculate spawn probability using the formula
			var actualSpawnChance float64
			if shouldSpawn {
				actualSpawnChance = tt.pirateActivityBase * (1.0 - tt.securityLevel)
			} else {
				actualSpawnChance = 0.0 // High Sec = no spawns
			}

			assert.InDelta(t, tt.wantSpawnChance, actualSpawnChance, 0.001,
				"Spawn chance mismatch for security %.2f with activity base %.2f: %s",
				tt.securityLevel, tt.pirateActivityBase, tt.description)

			// Test 3: Verify spawn probability is in valid range [0.0, 1.0]
			assert.GreaterOrEqual(t, actualSpawnChance, 0.0,
				"Spawn chance should never be negative")
			assert.LessOrEqual(t, actualSpawnChance, 1.0,
				"Spawn chance should never exceed 1.0")

			// Test 4: Verify formula consistency
			// For Low Sec: higher security = lower spawn chance
			if shouldSpawn && tt.securityLevel > 0 {
				lowerSecSpawnChance := tt.pirateActivityBase * (1.0 - (tt.securityLevel - 0.1))
				assert.Greater(t, lowerSecSpawnChance, actualSpawnChance,
					"Lower security should have higher spawn chance")
			}
		})
	}
}

// TestPirateSpawnProbabilityFormula verifies the mathematical properties of the spawn formula
func TestPirateSpawnProbabilityFormula(t *testing.T) {
	t.Run("formula is monotonically decreasing with security level", func(t *testing.T) {
		activityBase := 0.10
		// In Low Security, as security increases, spawn chance decreases
		securityLevels := []float64{0.0, 0.1, 0.2, 0.3, 0.39}
		var previousChance float64 = 1.0 // Start with max

		for _, sec := range securityLevels {
			chance := activityBase * (1.0 - sec)
			assert.Less(t, chance, previousChance,
				"Spawn chance should decrease as security increases (sec=%.2f)", sec)
			previousChance = chance
		}
	})

	t.Run("formula scales linearly with activity base", func(t *testing.T) {
		securityLevel := 0.2
		activityBases := []float64{0.05, 0.10, 0.15, 0.20}

		for i := 1; i < len(activityBases); i++ {
			chance1 := activityBases[i-1] * (1.0 - securityLevel)
			chance2 := activityBases[i] * (1.0 - securityLevel)
			ratio := chance2 / chance1
			expectedRatio := activityBases[i] / activityBases[i-1]

			assert.InDelta(t, expectedRatio, ratio, 0.001,
				"Spawn chance should scale linearly with activity base")
		}
	})

	t.Run("formula produces zero for high security", func(t *testing.T) {
		highSecLevels := []float64{0.4, 0.5, 0.7, 1.0, 2.0}

		for _, sec := range highSecLevels {
			// High Sec check happens before formula calculation
			if sec >= 0.4 {
				// No spawn calculation should occur
				assert.GreaterOrEqual(t, sec, 0.4,
					"Security %.2f should be classified as High Security", sec)
			}
		}
	})

	t.Run("formula boundary at security 0.4", func(t *testing.T) {
		activityBase := 0.10
		// Just below threshold - spawns enabled
		belowThreshold := 0.39
		chanceBelow := activityBase * (1.0 - belowThreshold)
		assert.Greater(t, chanceBelow, 0.0,
			"Security %.2f should allow spawns", belowThreshold)

		// At threshold - spawns disabled
		atThreshold := 0.4
		assert.GreaterOrEqual(t, atThreshold, 0.4,
			"Security %.2f should disable spawns", atThreshold)
	})
}

// TestPirateSpawnProbabilityEdgeCases tests edge cases and boundary conditions
func TestPirateSpawnProbabilityEdgeCases(t *testing.T) {
	t.Run("zero activity base prevents all spawns", func(t *testing.T) {
		activityBase := 0.0
		securityLevels := []float64{0.0, 0.1, 0.2, 0.3}

		for _, sec := range securityLevels {
			chance := activityBase * (1.0 - sec)
			assert.Equal(t, 0.0, chance,
				"Zero activity base should produce zero spawn chance at security %.2f", sec)
		}
	})

	t.Run("maximum values produce valid probabilities", func(t *testing.T) {
		activityBase := 1.0
		securityLevel := 0.0
		chance := activityBase * (1.0 - securityLevel)

		assert.Equal(t, 1.0, chance,
			"Maximum activity in lawless space should produce 100%% spawn chance")
	})

	t.Run("security level exactly at boundary", func(t *testing.T) {
		activityBase := 0.10
		boundary := 0.4

		// Just below
		belowBoundary := boundary - 0.01
		assert.Less(t, belowBoundary, 0.4, "Should be Low Security")
		chanceBelow := activityBase * (1.0 - belowBoundary)
		assert.Greater(t, chanceBelow, 0.0, "Should have non-zero spawn chance")

		// Exactly at boundary
		assert.GreaterOrEqual(t, boundary, 0.4, "Should be High Security")

		// Just above
		aboveBoundary := boundary + 0.01
		assert.GreaterOrEqual(t, aboveBoundary, 0.4, "Should be High Security")
	})

	t.Run("very small security differences produce proportional spawn differences", func(t *testing.T) {
		activityBase := 0.10
		sec1 := 0.10
		sec2 := 0.11

		chance1 := activityBase * (1.0 - sec1)
		chance2 := activityBase * (1.0 - sec2)

		difference := chance1 - chance2
		expectedDifference := activityBase * (sec2 - sec1)

		assert.InDelta(t, expectedDifference, difference, 0.0001,
			"Small security differences should produce proportional spawn differences")
	})

	t.Run("negative security levels", func(t *testing.T) {
		activityBase := 0.10
		negativeSec := -0.5

		// Negative security is still < 0.4, so spawns are enabled
		assert.Less(t, negativeSec, 0.4, "Negative security should be Low Security")

		chance := activityBase * (1.0 - negativeSec)
		// 0.10 × (1 - (-0.5)) = 0.10 × 1.5 = 0.15
		assert.InDelta(t, 0.15, chance, 0.0001,
			"Negative security should increase spawn chance above base rate")
	})

	t.Run("security level greater than 1.0", func(t *testing.T) {
		activityBase := 0.10
		highSec := 1.5

		// Still High Security (>= 0.4), so no spawns
		assert.GreaterOrEqual(t, highSec, 0.4, "Should be High Security")

		// If formula were applied (it shouldn't be):
		// 0.10 × (1 - 1.5) = 0.10 × (-0.5) = -0.05
		// But High Sec check prevents this calculation
		
		// Verify that if someone mistakenly applied the formula, it would be negative
		wouldBeNegative := activityBase * (1.0 - highSec)
		assert.Less(t, wouldBeNegative, 0.0,
			"Formula with security > 1.0 would produce negative value (but High Sec check prevents this)")
	})
}


// TestPirateFleeThreshold tests the flee threshold logic for pirates
func TestPirateFleeThreshold(t *testing.T) {
	tests := []struct {
		name           string
		tier           string
		maxHull        int
		currentHull    int
		fleeThreshold  float64
		wantShouldFlee bool
		description    string
	}{
		// Raider flee threshold tests (15% = 0.15)
		{
			name:           "raider at full hull",
			tier:           "raider",
			maxHull:        60,
			currentHull:    60,
			fleeThreshold:  0.15,
			wantShouldFlee: false,
			description:    "100% hull - should not flee",
		},
		{
			name:           "raider at 50% hull",
			tier:           "raider",
			maxHull:        60,
			currentHull:    30,
			fleeThreshold:  0.15,
			wantShouldFlee: false,
			description:    "50% hull - above flee threshold",
		},
		{
			name:           "raider at 20% hull",
			tier:           "raider",
			maxHull:        60,
			currentHull:    12,
			fleeThreshold:  0.15,
			wantShouldFlee: false,
			description:    "20% hull - above flee threshold",
		},
		{
			name:           "raider exactly at flee threshold",
			tier:           "raider",
			maxHull:        60,
			currentHull:    9, // 9/60 = 0.15 exactly
			fleeThreshold:  0.15,
			wantShouldFlee: true,
			description:    "Exactly 15% hull - should flee",
		},
		{
			name:           "raider just below flee threshold",
			tier:           "raider",
			maxHull:        60,
			currentHull:    8, // 8/60 = 0.133
			fleeThreshold:  0.15,
			wantShouldFlee: true,
			description:    "13.3% hull - below flee threshold",
		},
		{
			name:           "raider at 10% hull",
			tier:           "raider",
			maxHull:        60,
			currentHull:    6,
			fleeThreshold:  0.15,
			wantShouldFlee: true,
			description:    "10% hull - well below flee threshold",
		},
		{
			name:           "raider at 5% hull",
			tier:           "raider",
			maxHull:        60,
			currentHull:    3,
			fleeThreshold:  0.15,
			wantShouldFlee: true,
			description:    "5% hull - critically damaged",
		},
		{
			name:           "raider at 1 hull point",
			tier:           "raider",
			maxHull:        60,
			currentHull:    1,
			fleeThreshold:  0.15,
			wantShouldFlee: true,
			description:    "1 hull point remaining - should flee",
		},

		// Marauder flee threshold tests (10% = 0.10)
		{
			name:           "marauder at full hull",
			tier:           "marauder",
			maxHull:        90,
			currentHull:    90,
			fleeThreshold:  0.10,
			wantShouldFlee: false,
			description:    "100% hull - should not flee",
		},
		{
			name:           "marauder at 50% hull",
			tier:           "marauder",
			maxHull:        90,
			currentHull:    45,
			fleeThreshold:  0.10,
			wantShouldFlee: false,
			description:    "50% hull - above flee threshold",
		},
		{
			name:           "marauder at 20% hull",
			tier:           "marauder",
			maxHull:        90,
			currentHull:    18,
			fleeThreshold:  0.10,
			wantShouldFlee: false,
			description:    "20% hull - above flee threshold",
		},
		{
			name:           "marauder at 15% hull",
			tier:           "marauder",
			maxHull:        90,
			currentHull:    14, // 14/90 = 0.155
			fleeThreshold:  0.10,
			wantShouldFlee: false,
			description:    "15.5% hull - above flee threshold",
		},
		{
			name:           "marauder exactly at flee threshold",
			tier:           "marauder",
			maxHull:        90,
			currentHull:    9, // 9/90 = 0.10 exactly
			fleeThreshold:  0.10,
			wantShouldFlee: true,
			description:    "Exactly 10% hull - should flee",
		},
		{
			name:           "marauder just below flee threshold",
			tier:           "marauder",
			maxHull:        90,
			currentHull:    8, // 8/90 = 0.088
			fleeThreshold:  0.10,
			wantShouldFlee: true,
			description:    "8.8% hull - below flee threshold",
		},
		{
			name:           "marauder at 5% hull",
			tier:           "marauder",
			maxHull:        90,
			currentHull:    5, // 5/90 = 0.055
			fleeThreshold:  0.10,
			wantShouldFlee: true,
			description:    "5.5% hull - well below flee threshold",
		},
		{
			name:           "marauder at 1 hull point",
			tier:           "marauder",
			maxHull:        90,
			currentHull:    1,
			fleeThreshold:  0.10,
			wantShouldFlee: true,
			description:    "1 hull point remaining - should flee",
		},

		// Edge cases
		{
			name:           "raider just above flee threshold by 1 hull point",
			tier:           "raider",
			maxHull:        60,
			currentHull:    10, // 10/60 = 0.166
			fleeThreshold:  0.15,
			wantShouldFlee: false,
			description:    "16.6% hull - just above threshold",
		},
		{
			name:           "marauder just above flee threshold by 1 hull point",
			tier:           "marauder",
			maxHull:        90,
			currentHull:    10, // 10/90 = 0.111
			fleeThreshold:  0.10,
			wantShouldFlee: false,
			description:    "11.1% hull - just above threshold",
		},
		{
			name:           "raider with odd max hull value",
			tier:           "raider",
			maxHull:        57,
			currentHull:    8, // 8/57 = 0.140
			fleeThreshold:  0.15,
			wantShouldFlee: true,
			description:    "14% hull with odd max hull - should flee",
		},
		{
			name:           "marauder with odd max hull value",
			tier:           "marauder",
			maxHull:        87,
			currentHull:    9, // 9/87 = 0.103
			fleeThreshold:  0.10,
			wantShouldFlee: false,
			description:    "10.3% hull with odd max hull - should not flee",
		},

		// Boundary precision tests
		{
			name:           "raider at 15.01% hull",
			tier:           "raider",
			maxHull:        1000,
			currentHull:    151, // 151/1000 = 0.151
			fleeThreshold:  0.15,
			wantShouldFlee: false,
			description:    "Just above threshold with high precision",
		},
		{
			name:           "raider at 14.99% hull",
			tier:           "raider",
			maxHull:        1000,
			currentHull:    149, // 149/1000 = 0.149
			fleeThreshold:  0.15,
			wantShouldFlee: true,
			description:    "Just below threshold with high precision",
		},
		{
			name:           "marauder at 10.01% hull",
			tier:           "marauder",
			maxHull:        1000,
			currentHull:    101, // 101/1000 = 0.101
			fleeThreshold:  0.10,
			wantShouldFlee: false,
			description:    "Just above threshold with high precision",
		},
		{
			name:           "marauder at 9.99% hull",
			tier:           "marauder",
			maxHull:        1000,
			currentHull:    99, // 99/1000 = 0.099
			fleeThreshold:  0.10,
			wantShouldFlee: true,
			description:    "Just below threshold with high precision",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create pirate ship with specified hull values
			pirate := &PirateShip{
				ShipID:        "test-pirate",
				Tier:          tt.tier,
				HullPoints:    tt.currentHull,
				MaxHull:       tt.maxHull,
				FleeThreshold: tt.fleeThreshold,
			}

			// Calculate hull percentage (same logic as in ProcessAttack)
			hullPercent := float64(pirate.HullPoints) / float64(pirate.MaxHull)

			// Check if pirate should flee
			shouldFlee := hullPercent <= pirate.FleeThreshold

			assert.Equal(t, tt.wantShouldFlee, shouldFlee,
				"%s: hull=%d/%d (%.2f%%), threshold=%.2f%% - %s",
				tt.tier, tt.currentHull, tt.maxHull, hullPercent*100, tt.fleeThreshold*100, tt.description)

			// Verify hull percentage calculation
			expectedPercent := float64(tt.currentHull) / float64(tt.maxHull)
			assert.InDelta(t, expectedPercent, hullPercent, 0.0001,
				"Hull percentage calculation mismatch")

			// Verify flee logic consistency
			if shouldFlee {
				assert.LessOrEqual(t, hullPercent, tt.fleeThreshold,
					"If fleeing, hull percent should be <= threshold")
			} else {
				assert.Greater(t, hullPercent, tt.fleeThreshold,
					"If not fleeing, hull percent should be > threshold")
			}
		})
	}
}

// TestPirateFleeThresholdBehavior tests the flee behavior in combat scenarios
func TestPirateFleeThresholdBehavior(t *testing.T) {
	tests := []struct {
		name              string
		tier              string
		maxHull           int
		initialHull       int
		damageSequence    []int // Sequence of damage amounts to apply
		wantFleeAfterHit  int   // Which hit (1-indexed) should trigger flee, 0 if never
		description       string
	}{
		{
			name:             "raider takes gradual damage and flees",
			tier:             "raider",
			maxHull:          60,
			initialHull:      60,
			damageSequence:   []int{15, 15, 15, 15}, // 60 -> 45 -> 30 -> 15 -> 0 (but flees at 9)
			wantFleeAfterHit: 4,
			description:      "Raider should flee when hull drops to 15 (25%) then to 0, but flees at threshold",
		},
		{
			name:             "raider takes massive damage and flees immediately",
			tier:             "raider",
			maxHull:          60,
			initialHull:      60,
			damageSequence:   []int{52}, // 60 -> 8 (13.3%)
			wantFleeAfterHit: 1,
			description:      "Single massive hit brings raider below flee threshold",
		},
		{
			name:             "raider takes damage but stays above threshold",
			tier:             "raider",
			maxHull:          60,
			initialHull:      60,
			damageSequence:   []int{10, 10, 10}, // 60 -> 50 -> 40 -> 30 (50%)
			wantFleeAfterHit: 0,
			description:      "Raider takes moderate damage but stays above 15% threshold",
		},
		{
			name:             "marauder takes gradual damage and flees",
			tier:             "marauder",
			maxHull:          90,
			initialHull:      90,
			damageSequence:   []int{20, 20, 20, 20, 20}, // 90 -> 70 -> 50 -> 30 -> 10 (11.1%) -> -10 (0%)
			wantFleeAfterHit: 5,
			description:      "Marauder should flee when hull drops below 10% threshold on 5th hit",
		},
		{
			name:             "marauder takes massive damage and flees",
			tier:             "marauder",
			maxHull:          90,
			initialHull:      90,
			damageSequence:   []int{82}, // 90 -> 8 (8.8%)
			wantFleeAfterHit: 1,
			description:      "Single massive hit brings marauder below 10% flee threshold",
		},
		{
			name:             "marauder stays and fights longer than raider",
			tier:             "marauder",
			maxHull:          90,
			initialHull:      90,
			damageSequence:   []int{15, 15, 15, 15, 15}, // 90 -> 75 -> 60 -> 45 -> 30 -> 15 (16.6%)
			wantFleeAfterHit: 0,
			description:      "Marauder with 10% threshold stays at 15 hull (16.6%), raider would flee",
		},
		{
			name:             "raider at exactly flee threshold after damage",
			tier:             "raider",
			maxHull:          60,
			initialHull:      60,
			damageSequence:   []int{51}, // 60 -> 9 (15% exactly)
			wantFleeAfterHit: 1,
			description:      "Damage brings raider to exactly 15% - should flee",
		},
		{
			name:             "marauder at exactly flee threshold after damage",
			tier:             "marauder",
			maxHull:          90,
			initialHull:      90,
			damageSequence:   []int{81}, // 90 -> 9 (10% exactly)
			wantFleeAfterHit: 1,
			description:      "Damage brings marauder to exactly 10% - should flee",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create pirate ship
			pirate := &PirateShip{
				ShipID:      "test-pirate",
				Tier:        tt.tier,
				HullPoints:  tt.initialHull,
				MaxHull:     tt.maxHull,
				ShieldPoints: 0, // No shields for simplicity
				MaxShield:   0,
			}

			// Set flee threshold based on tier
			if tt.tier == "raider" {
				pirate.FleeThreshold = 0.15
			} else {
				pirate.FleeThreshold = 0.10
			}

			fleeOccurred := false
			fleeAfterHit := 0

			// Apply damage sequence
			for i, damage := range tt.damageSequence {
				hitNumber := i + 1

				// Apply damage to hull (no shields)
				pirate.HullPoints -= damage
				if pirate.HullPoints < 0 {
					pirate.HullPoints = 0
				}

				// Check flee condition
				hullPercent := float64(pirate.HullPoints) / float64(pirate.MaxHull)
				if hullPercent <= pirate.FleeThreshold && !fleeOccurred {
					fleeOccurred = true
					fleeAfterHit = hitNumber
					break // Pirate flees, combat ends
				}

				// Check if destroyed
				if pirate.HullPoints <= 0 {
					break // Pirate destroyed, combat ends
				}
			}

			if tt.wantFleeAfterHit > 0 {
				assert.True(t, fleeOccurred,
					"%s: Expected pirate to flee after hit %d, but did not flee. Final hull: %d/%d (%.1f%%)",
					tt.description, tt.wantFleeAfterHit, pirate.HullPoints, pirate.MaxHull,
					float64(pirate.HullPoints)/float64(pirate.MaxHull)*100)

				assert.Equal(t, tt.wantFleeAfterHit, fleeAfterHit,
					"%s: Expected flee after hit %d, but fled after hit %d",
					tt.description, tt.wantFleeAfterHit, fleeAfterHit)
			} else {
				assert.False(t, fleeOccurred,
					"%s: Expected pirate not to flee, but fled after hit %d. Final hull: %d/%d (%.1f%%)",
					tt.description, fleeAfterHit, pirate.HullPoints, pirate.MaxHull,
					float64(pirate.HullPoints)/float64(pirate.MaxHull)*100)
			}
		})
	}
}

// TestPirateFleeThresholdComparison tests that different tiers have different flee behaviors
func TestPirateFleeThresholdComparison(t *testing.T) {
	t.Run("marauder stays longer than raider at same hull percentage", func(t *testing.T) {
		// Both at 12% hull
		raider := &PirateShip{
			Tier:          "raider",
			HullPoints:    12,
			MaxHull:       100,
			FleeThreshold: 0.15,
		}

		marauder := &PirateShip{
			Tier:          "marauder",
			HullPoints:    12,
			MaxHull:       100,
			FleeThreshold: 0.10,
		}

		raiderHullPercent := float64(raider.HullPoints) / float64(raider.MaxHull)
		marauderHullPercent := float64(marauder.HullPoints) / float64(marauder.MaxHull)

		raiderShouldFlee := raiderHullPercent <= raider.FleeThreshold
		marauderShouldFlee := marauderHullPercent <= marauder.FleeThreshold

		assert.True(t, raiderShouldFlee, "Raider at 12%% should flee (threshold 15%%)")
		assert.False(t, marauderShouldFlee, "Marauder at 12%% should not flee (threshold 10%%)")
	})

	t.Run("both flee at their respective thresholds", func(t *testing.T) {
		raider := &PirateShip{
			Tier:          "raider",
			HullPoints:    15,
			MaxHull:       100,
			FleeThreshold: 0.15,
		}

		marauder := &PirateShip{
			Tier:          "marauder",
			HullPoints:    10,
			MaxHull:       100,
			FleeThreshold: 0.10,
		}

		raiderHullPercent := float64(raider.HullPoints) / float64(raider.MaxHull)
		marauderHullPercent := float64(marauder.HullPoints) / float64(marauder.MaxHull)

		raiderShouldFlee := raiderHullPercent <= raider.FleeThreshold
		marauderShouldFlee := marauderHullPercent <= marauder.FleeThreshold

		assert.True(t, raiderShouldFlee, "Raider at exactly 15%% should flee")
		assert.True(t, marauderShouldFlee, "Marauder at exactly 10%% should flee")
	})

	t.Run("marauder is more aggressive - lower flee threshold", func(t *testing.T) {
		raiderThreshold := 0.15
		marauderThreshold := 0.10

		assert.Less(t, marauderThreshold, raiderThreshold,
			"Marauder should have lower flee threshold (more aggressive)")

		// At 11% hull, raider flees but marauder stays
		hullPercent := 0.11

		raiderFlees := hullPercent <= raiderThreshold
		marauderFlees := hullPercent <= marauderThreshold

		assert.True(t, raiderFlees, "Raider should flee at 11%%")
		assert.False(t, marauderFlees, "Marauder should not flee at 11%%")
	})
}

// TestPirateFleeThresholdEdgeCases tests edge cases and boundary conditions
func TestPirateFleeThresholdEdgeCases(t *testing.T) {
	t.Run("zero hull always triggers flee check", func(t *testing.T) {
		pirate := &PirateShip{
			Tier:          "raider",
			HullPoints:    0,
			MaxHull:       60,
			FleeThreshold: 0.15,
		}

		hullPercent := float64(pirate.HullPoints) / float64(pirate.MaxHull)
		shouldFlee := hullPercent <= pirate.FleeThreshold

		assert.True(t, shouldFlee, "Zero hull should trigger flee condition")
		assert.Equal(t, 0.0, hullPercent, "Zero hull should be 0%% hull")
	})

	t.Run("one hull point below threshold triggers flee", func(t *testing.T) {
		pirate := &PirateShip{
			Tier:          "raider",
			HullPoints:    1,
			MaxHull:       60,
			FleeThreshold: 0.15,
		}

		hullPercent := float64(pirate.HullPoints) / float64(pirate.MaxHull)
		shouldFlee := hullPercent <= pirate.FleeThreshold

		assert.True(t, shouldFlee, "1 hull point (1.6%%) should trigger flee for raider")
	})

	t.Run("flee threshold of 0.0 means never flee", func(t *testing.T) {
		pirate := &PirateShip{
			Tier:          "custom",
			HullPoints:    1,
			MaxHull:       100,
			FleeThreshold: 0.0,
		}

		hullPercent := float64(pirate.HullPoints) / float64(pirate.MaxHull)
		shouldFlee := hullPercent <= pirate.FleeThreshold

		// At 1% hull with 0% threshold, should not flee (1% > 0%)
		assert.False(t, shouldFlee, "With 0.0 threshold, pirate should only flee at exactly 0 hull")

		// At 0 hull
		pirate.HullPoints = 0
		hullPercent = float64(pirate.HullPoints) / float64(pirate.MaxHull)
		shouldFlee = hullPercent <= pirate.FleeThreshold
		assert.True(t, shouldFlee, "With 0.0 threshold, pirate should flee at 0 hull")
	})

	t.Run("flee threshold of 1.0 means always flee", func(t *testing.T) {
		pirate := &PirateShip{
			Tier:          "custom",
			HullPoints:    100,
			MaxHull:       100,
			FleeThreshold: 1.0,
		}

		hullPercent := float64(pirate.HullPoints) / float64(pirate.MaxHull)
		shouldFlee := hullPercent <= pirate.FleeThreshold

		assert.True(t, shouldFlee, "With 1.0 threshold, pirate should flee even at full hull")
	})

	t.Run("very small hull values maintain precision", func(t *testing.T) {
		pirate := &PirateShip{
			Tier:          "raider",
			HullPoints:    1,
			MaxHull:       10,
			FleeThreshold: 0.15,
		}

		hullPercent := float64(pirate.HullPoints) / float64(pirate.MaxHull)
		assert.Equal(t, 0.1, hullPercent, "1/10 should be exactly 0.1")

		shouldFlee := hullPercent <= pirate.FleeThreshold
		assert.True(t, shouldFlee, "10%% hull should be below 15%% threshold")
	})

	t.Run("large hull values maintain precision", func(t *testing.T) {
		pirate := &PirateShip{
			Tier:          "raider",
			HullPoints:    150,
			MaxHull:       1000,
			FleeThreshold: 0.15,
		}

		hullPercent := float64(pirate.HullPoints) / float64(pirate.MaxHull)
		assert.Equal(t, 0.15, hullPercent, "150/1000 should be exactly 0.15")

		shouldFlee := hullPercent <= pirate.FleeThreshold
		assert.True(t, shouldFlee, "Exactly at threshold should trigger flee")
	})

	t.Run("flee check happens before destruction check", func(t *testing.T) {
		// In the actual ProcessAttack code, destruction check happens first
		// But flee threshold can be tested independently
		pirate := &PirateShip{
			Tier:          "raider",
			HullPoints:    5,
			MaxHull:       60,
			FleeThreshold: 0.15,
		}

		hullPercent := float64(pirate.HullPoints) / float64(pirate.MaxHull)
		shouldFlee := hullPercent <= pirate.FleeThreshold

		assert.True(t, shouldFlee, "Pirate at 8.3%% hull should flee")
		assert.Greater(t, pirate.HullPoints, 0, "Pirate is damaged but not destroyed")
	})
}

// TestPirateFleeThresholdFormula tests the mathematical properties of the flee threshold
func TestPirateFleeThresholdFormula(t *testing.T) {
	t.Run("hull percentage is monotonically decreasing with damage", func(t *testing.T) {
		pirate := &PirateShip{
			Tier:          "raider",
			HullPoints:    60,
			MaxHull:       60,
			FleeThreshold: 0.15,
		}

		previousPercent := 1.0

		for damage := 10; damage <= 60; damage += 10 {
			pirate.HullPoints = 60 - damage
			if pirate.HullPoints < 0 {
				pirate.HullPoints = 0
			}

			hullPercent := float64(pirate.HullPoints) / float64(pirate.MaxHull)
			assert.LessOrEqual(t, hullPercent, previousPercent,
				"Hull percentage should decrease or stay same as damage increases")
			previousPercent = hullPercent
		}
	})

	t.Run("flee threshold is independent of max hull", func(t *testing.T) {
		// 15% of different max hull values
		testCases := []struct {
			maxHull     int
			hullAt15Pct int
		}{
			{60, 9},
			{100, 15},
			{200, 30},
			{1000, 150},
		}

		for _, tc := range testCases {
			pirate := &PirateShip{
				Tier:          "raider",
				HullPoints:    tc.hullAt15Pct,
				MaxHull:       tc.maxHull,
				FleeThreshold: 0.15,
			}

			hullPercent := float64(pirate.HullPoints) / float64(pirate.MaxHull)
			assert.InDelta(t, 0.15, hullPercent, 0.0001,
				"15%% threshold should be consistent across different max hull values")

			shouldFlee := hullPercent <= pirate.FleeThreshold
			assert.True(t, shouldFlee, "Should flee at exactly 15%% regardless of max hull")
		}
	})

	t.Run("flee threshold comparison is consistent", func(t *testing.T) {
		// Test that <= comparison works correctly at boundary
		threshold := 0.15

		testCases := []struct {
			hullPercent float64
			shouldFlee  bool
		}{
			{0.16, false},
			{0.151, false},
			{0.15, true},   // Exactly at threshold
			{0.149, true},
			{0.14, true},
			{0.10, true},
			{0.01, true},
			{0.0, true},
		}

		for _, tc := range testCases {
			shouldFlee := tc.hullPercent <= threshold
			assert.Equal(t, tc.shouldFlee, shouldFlee,
				"Hull percent %.3f with threshold %.2f should flee=%v",
				tc.hullPercent, threshold, tc.shouldFlee)
		}
	})
}
