package registration

import (
	"database/sql"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *Database {
	conn, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Create schema
	schema := `
		CREATE TABLE players (
			player_id TEXT PRIMARY KEY,
			player_name TEXT NOT NULL UNIQUE,
			ssh_username TEXT,
			token_hash TEXT NOT NULL,
			password_hash TEXT,
			credits INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			last_login_at INTEGER,
			is_banned INTEGER NOT NULL DEFAULT 0
		);
		CREATE UNIQUE INDEX idx_players_ssh_username ON players (ssh_username);

		CREATE TABLE ships (
			ship_id TEXT PRIMARY KEY,
			player_id TEXT NOT NULL,
			ship_class TEXT NOT NULL,
			hull_points INTEGER NOT NULL,
			max_hull_points INTEGER NOT NULL,
			shield_points INTEGER NOT NULL,
			max_shield_points INTEGER NOT NULL,
			energy_points INTEGER NOT NULL,
			max_energy_points INTEGER NOT NULL,
			cargo_capacity INTEGER NOT NULL,
			missiles_current INTEGER NOT NULL DEFAULT 0,
			current_system_id INTEGER,
			position_x REAL NOT NULL DEFAULT 0.0,
			position_y REAL NOT NULL DEFAULT 0.0,
			status TEXT NOT NULL,
			docked_at_port_id INTEGER,
			last_updated_tick INTEGER NOT NULL
		);
	`

	_, err = conn.Exec(schema)
	require.NoError(t, err)

	logger := zerolog.Nop()
	return NewDatabase(conn, logger)
}

func TestGeneratePlayerToken(t *testing.T) {
	// Generate multiple tokens to ensure uniqueness
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := GeneratePlayerToken()
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Check token is unique
		assert.False(t, tokens[token], "token should be unique")
		tokens[token] = true

		// Check token length (32 bytes base64url encoded = 43 chars)
		assert.Greater(t, len(token), 40, "token should be at least 40 chars")
	}
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"simple password", "password123"},
		{"complex password", "P@ssw0rd!#$%"},
		{"long password", "this is a very long password with many characters"},
		{"unicode password", "пароль密码🔒"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			require.NoError(t, err)
			require.NotEmpty(t, hash)

			// Verify hash starts with bcrypt prefix
			assert.Contains(t, hash, "$2a$12$", "hash should use bcrypt cost 12")

			// Verify password validates correctly
			err = ValidatePassword(tt.password, hash)
			assert.NoError(t, err, "password should validate against its hash")

			// Verify wrong password fails
			err = ValidatePassword("wrongpassword", hash)
			assert.Error(t, err, "wrong password should not validate")
		})
	}
}

func TestHashToken(t *testing.T) {
	token, err := GeneratePlayerToken()
	require.NoError(t, err)

	hash, err := HashToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	// Verify hash starts with bcrypt prefix for cost 10
	assert.Contains(t, hash, "$2a$10$", "hash should use bcrypt cost 10")

	// Verify token validates correctly
	err = ValidateToken(token, hash)
	assert.NoError(t, err, "token should validate against its hash")

	// Verify wrong token fails
	wrongToken, _ := GeneratePlayerToken()
	err = ValidateToken(wrongToken, hash)
	assert.Error(t, err, "wrong token should not validate")
}

func TestValidateRegistrationRequest(t *testing.T) {
	logger := zerolog.Nop()
	db := setupTestDB(t)
	world := NewWorld(1, 100)
	config := NewConfig(10000, "courier", 3)
	registrar := NewRegistrar(db, world, config, logger)
	defer registrar.Stop()

	tests := []struct {
		name    string
		req     *RegistrationRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &RegistrationRequest{
				SSHUsername: "testuser",
				DisplayName: "TestPlayer",
				Password:    "password123",
				RemoteAddr:  "127.0.0.1",
			},
			wantErr: false,
		},
		{
			name: "display name too short",
			req: &RegistrationRequest{
				SSHUsername: "testuser",
				DisplayName: "ab",
				Password:    "password123",
			},
			wantErr: true,
			errMsg:  "at least 3 characters",
		},
		{
			name: "display name too long",
			req: &RegistrationRequest{
				SSHUsername: "testuser",
				DisplayName: "thisnameiswaytoolo ng",
				Password:    "password123",
			},
			wantErr: true,
			errMsg:  "at most 20 characters",
		},
		{
			name: "password too short",
			req: &RegistrationRequest{
				SSHUsername: "testuser",
				DisplayName: "TestPlayer",
				Password:    "pass",
			},
			wantErr: true,
			errMsg:  "at least 8 characters",
		},
		{
			name: "missing SSH username",
			req: &RegistrationRequest{
				SSHUsername: "",
				DisplayName: "TestPlayer",
				Password:    "password123",
			},
			wantErr: true,
			errMsg:  "SSH username is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registrar.ValidateRegistrationRequest(tt.req)
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

func TestCheckPlayerExists(t *testing.T) {
	logger := zerolog.Nop()
	db := setupTestDB(t)
	world := NewWorld(1, 100)
	config := NewConfig(10000, "courier", 3)
	registrar := NewRegistrar(db, world, config, logger)
	defer registrar.Stop()

	// Insert a test player
	tx, err := db.conn.Begin()
	require.NoError(t, err)

	testPlayer := &Player{
		PlayerID:     "test-player-id",
		PlayerName:   "TestPlayer",
		SSHUsername:  "testuser",
		TokenHash:    "test-token-hash",
		PasswordHash: "test-password-hash",
		Credits:      10000,
		CreatedAt:    1234567890,
		IsBanned:     false,
	}

	err = db.TxInsertPlayer(tx, testPlayer)
	require.NoError(t, err)
	err = tx.Commit()
	require.NoError(t, err)

	// Test existing player
	player, err := registrar.CheckPlayerExists("testuser")
	require.NoError(t, err)
	require.NotNil(t, player)
	assert.Equal(t, "test-player-id", player.PlayerID)
	assert.Equal(t, "TestPlayer", player.PlayerName)
	assert.Equal(t, "testuser", player.SSHUsername)

	// Test non-existing player
	player, err = registrar.CheckPlayerExists("nonexistent")
	require.NoError(t, err)
	assert.Nil(t, player)
}

func TestRegisterNewPlayer(t *testing.T) {
	logger := zerolog.Nop()
	db := setupTestDB(t)
	world := NewWorld(1, 100)
	config := NewConfig(10000, "courier", 3)
	registrar := NewRegistrar(db, world, config, logger)
	defer registrar.Stop()

	req := &RegistrationRequest{
		SSHUsername: "newuser",
		DisplayName: "NewPlayer",
		Password:    "securepass123",
		RemoteAddr:  "127.0.0.1",
	}

	result, err := registrar.RegisterNewPlayer(req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify player ID is UUID
	assert.NotEmpty(t, result.PlayerID)
	assert.Len(t, result.PlayerID, 36, "player ID should be UUID format")

	// Verify token is returned
	assert.NotEmpty(t, result.PlayerToken)
	assert.Greater(t, len(result.PlayerToken), 40, "token should be at least 40 chars")

	// Verify ship is created
	require.NotNil(t, result.StartingShip)
	assert.Equal(t, result.PlayerID, result.StartingShip.PlayerID)
	assert.Equal(t, "courier", result.StartingShip.ShipClass)
	assert.Equal(t, 100, result.StartingShip.HullPoints)
	assert.Equal(t, 100, result.StartingShip.MaxHullPoints)
	assert.Equal(t, 50, result.StartingShip.ShieldPoints)
	assert.Equal(t, 50, result.StartingShip.MaxShieldPoints)
	assert.Equal(t, 100, result.StartingShip.EnergyPoints)
	assert.Equal(t, 100, result.StartingShip.MaxEnergyPoints)
	assert.Equal(t, 20, result.StartingShip.CargoCapacity)
	assert.Equal(t, "DOCKED", result.StartingShip.Status)
	assert.Equal(t, 1, result.StartingShip.CurrentSystemID)
	assert.NotNil(t, result.StartingShip.DockedAtPortID)
	assert.Equal(t, 100, *result.StartingShip.DockedAtPortID)

	// Verify player is in database
	player, err := registrar.CheckPlayerExists("newuser")
	require.NoError(t, err)
	require.NotNil(t, player)
	assert.Equal(t, result.PlayerID, player.PlayerID)
	assert.Equal(t, "NewPlayer", player.PlayerName)
	assert.Equal(t, "newuser", player.SSHUsername)
	assert.Equal(t, int64(10000), player.Credits)

	// Verify password hash is stored
	assert.NotEmpty(t, player.PasswordHash)
	err = ValidatePassword("securepass123", player.PasswordHash)
	assert.NoError(t, err, "stored password hash should validate")

	// Verify token hash is stored
	assert.NotEmpty(t, player.TokenHash)
	err = ValidateToken(result.PlayerToken, player.TokenHash)
	assert.NoError(t, err, "stored token hash should validate")
}

func TestRegisterNewPlayerDuplicateSSHUsername(t *testing.T) {
	logger := zerolog.Nop()
	db := setupTestDB(t)
	world := NewWorld(1, 100)
	config := NewConfig(10000, "courier", 3)
	registrar := NewRegistrar(db, world, config, logger)
	defer registrar.Stop()

	// Register first player
	req1 := &RegistrationRequest{
		SSHUsername: "duplicateuser",
		DisplayName: "FirstPlayer",
		Password:    "password123",
		RemoteAddr:  "127.0.0.1",
	}

	_, err := registrar.RegisterNewPlayer(req1)
	require.NoError(t, err)

	// Try to register second player with same SSH username
	req2 := &RegistrationRequest{
		SSHUsername: "duplicateuser",
		DisplayName: "SecondPlayer",
		Password:    "password456",
		RemoteAddr:  "127.0.0.1",
	}

	_, err = registrar.RegisterNewPlayer(req2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "player already exists")
}

func TestProvisionStarterShip(t *testing.T) {
	logger := zerolog.Nop()
	db := setupTestDB(t)
	world := NewWorld(1, 100)
	config := NewConfig(10000, "courier", 3)
	registrar := NewRegistrar(db, world, config, logger)
	defer registrar.Stop()

	playerID := "test-player-123"
	ship := registrar.ProvisionStarterShip(playerID)

	require.NotNil(t, ship)
	assert.NotEmpty(t, ship.ShipID)
	assert.Equal(t, playerID, ship.PlayerID)
	assert.Equal(t, "courier", ship.ShipClass)
	assert.Equal(t, 100, ship.HullPoints)
	assert.Equal(t, 100, ship.MaxHullPoints)
	assert.Equal(t, 50, ship.ShieldPoints)
	assert.Equal(t, 50, ship.MaxShieldPoints)
	assert.Equal(t, 100, ship.EnergyPoints)
	assert.Equal(t, 100, ship.MaxEnergyPoints)
	assert.Equal(t, 20, ship.CargoCapacity)
	assert.Equal(t, 0, ship.MissilesCurrent)
	assert.Equal(t, 1, ship.CurrentSystemID)
	assert.Equal(t, 0.0, ship.PositionX)
	assert.Equal(t, 0.0, ship.PositionY)
	assert.Equal(t, "DOCKED", ship.Status)
	require.NotNil(t, ship.DockedAtPortID)
	assert.Equal(t, 100, *ship.DockedAtPortID)
	assert.Equal(t, int64(0), ship.LastUpdatedTick)
}

func TestGetShipClassStats(t *testing.T) {
	tests := []struct {
		name      string
		shipClass string
		want      ShipClassStats
	}{
		{
			name:      "courier class",
			shipClass: "courier",
			want: ShipClassStats{
				MaxHull:       100,
				MaxShield:     50,
				MaxEnergy:     100,
				CargoCapacity: 20,
				WeaponDamage:  15,
			},
		},
		{
			name:      "unknown class defaults to courier",
			shipClass: "unknown",
			want: ShipClassStats{
				MaxHull:       100,
				MaxShield:     50,
				MaxEnergy:     100,
				CargoCapacity: 20,
				WeaponDamage:  15,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetShipClassStats(tt.shipClass)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDatabaseTransactionRollback(t *testing.T) {
	db := setupTestDB(t)

	// Begin transaction
	tx, err := db.conn.Begin()
	require.NoError(t, err)

	// Insert player
	player := &Player{
		PlayerID:     "rollback-test",
		PlayerName:   "RollbackTest",
		SSHUsername:  "rollbackuser",
		TokenHash:    "test-hash",
		PasswordHash: "test-pass",
		Credits:      5000,
		CreatedAt:    1234567890,
		IsBanned:     false,
	}

	err = db.TxInsertPlayer(tx, player)
	require.NoError(t, err)

	// Rollback transaction
	err = tx.Rollback()
	require.NoError(t, err)

	// Verify player was not inserted
	foundPlayer, err := db.GetPlayerBySSHUsername("rollbackuser")
	require.NoError(t, err)
	assert.Nil(t, foundPlayer, "player should not exist after rollback")
}

func TestPasswordAndTokenSecurity(t *testing.T) {
	// Test that same password produces different hashes (salt)
	password := "testpassword123"
	hash1, err := HashPassword(password)
	require.NoError(t, err)
	hash2, err := HashPassword(password)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2, "same password should produce different hashes due to salt")

	// Both hashes should validate the password
	assert.NoError(t, ValidatePassword(password, hash1))
	assert.NoError(t, ValidatePassword(password, hash2))

	// Test that tokens are cryptographically random
	token1, err := GeneratePlayerToken()
	require.NoError(t, err)
	token2, err := GeneratePlayerToken()
	require.NoError(t, err)

	assert.NotEqual(t, token1, token2, "tokens should be unique")
}

func TestRegistrationRateLimiting(t *testing.T) {
	logger := zerolog.Nop()
	db := setupTestDB(t)
	world := NewWorld(1, 100)
	config := NewConfig(10000, "courier", 3)
	registrar := NewRegistrar(db, world, config, logger)
	defer registrar.Stop()

	ipAddr := "192.168.1.100"

	// First 3 registrations should succeed
	for i := 0; i < 3; i++ {
		req := &RegistrationRequest{
			SSHUsername: "user" + string(rune('a'+i)),
			DisplayName: "Player" + string(rune('A'+i)),
			Password:    "password123",
			RemoteAddr:  ipAddr,
		}

		result, err := registrar.RegisterNewPlayer(req)
		require.NoError(t, err, "registration %d should succeed", i+1)
		require.NotNil(t, result)
	}

	// 4th registration from same IP should fail
	req := &RegistrationRequest{
		SSHUsername: "userd",
		DisplayName: "PlayerD",
		Password:    "password123",
		RemoteAddr:  ipAddr,
	}

	result, err := registrar.RegisterNewPlayer(req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "rate limit")
}

func TestRegistrationRateLimitingDifferentIPs(t *testing.T) {
	logger := zerolog.Nop()
	db := setupTestDB(t)
	world := NewWorld(1, 100)
	config := NewConfig(10000, "courier", 3)
	registrar := NewRegistrar(db, world, config, logger)
	defer registrar.Stop()

	// Each IP should have independent rate limits
	for i := 0; i < 3; i++ {
		ipAddr := "192.168.1." + string(rune('1'+i))

		for j := 0; j < 3; j++ {
			req := &RegistrationRequest{
				SSHUsername: "user" + string(rune('a'+i)) + string(rune('0'+j)),
				DisplayName: "Player" + string(rune('A'+i)) + string(rune('0'+j)),
				Password:    "password123",
				RemoteAddr:  ipAddr,
			}

			result, err := registrar.RegisterNewPlayer(req)
			require.NoError(t, err, "registration from IP %s should succeed", ipAddr)
			require.NotNil(t, result)
		}
	}

	// Each IP should now be at limit
	for i := 0; i < 3; i++ {
		ipAddr := "192.168.1." + string(rune('1'+i))

		req := &RegistrationRequest{
			SSHUsername: "userextra" + string(rune('a'+i)),
			DisplayName: "PlayerExtra" + string(rune('A'+i)),
			Password:    "password123",
			RemoteAddr:  ipAddr,
		}

		result, err := registrar.RegisterNewPlayer(req)
		assert.Error(t, err, "registration from IP %s should fail", ipAddr)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "rate limit")
	}
}

func TestRegistrationRateLimitDoesNotAffectValidation(t *testing.T) {
	logger := zerolog.Nop()
	db := setupTestDB(t)
	world := NewWorld(1, 100)
	config := NewConfig(10000, "courier", 3)
	registrar := NewRegistrar(db, world, config, logger)
	defer registrar.Stop()

	ipAddr := "192.168.1.100"

	// Use up rate limit with valid registrations
	for i := 0; i < 3; i++ {
		req := &RegistrationRequest{
			SSHUsername: "user" + string(rune('a'+i)),
			DisplayName: "Player" + string(rune('A'+i)),
			Password:    "password123",
			RemoteAddr:  ipAddr,
		}

		_, err := registrar.RegisterNewPlayer(req)
		require.NoError(t, err)
	}

	// Invalid registration should still fail with rate limit error first
	req := &RegistrationRequest{
		SSHUsername: "userd",
		DisplayName: "ab", // Too short
		Password:    "password123",
		RemoteAddr:  ipAddr,
	}

	_, err := registrar.RegisterNewPlayer(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit")
}

func TestRegistrationRateLimitFailedAttemptStillCounts(t *testing.T) {
	logger := zerolog.Nop()
	db := setupTestDB(t)
	world := NewWorld(1, 100)
	config := NewConfig(10000, "courier", 3)
	registrar := NewRegistrar(db, world, config, logger)
	defer registrar.Stop()

	ipAddr := "192.168.1.100"

	// First registration succeeds
	req1 := &RegistrationRequest{
		SSHUsername: "user1",
		DisplayName: "Player1",
		Password:    "password123",
		RemoteAddr:  ipAddr,
	}
	_, err := registrar.RegisterNewPlayer(req1)
	require.NoError(t, err)

	// Second registration fails (duplicate username) but counts toward rate limit
	req2 := &RegistrationRequest{
		SSHUsername: "user1", // Duplicate
		DisplayName: "Player2",
		Password:    "password123",
		RemoteAddr:  ipAddr,
	}
	_, err = registrar.RegisterNewPlayer(req2)
	assert.Error(t, err)

	// Third registration succeeds
	req3 := &RegistrationRequest{
		SSHUsername: "user3",
		DisplayName: "Player3",
		Password:    "password123",
		RemoteAddr:  ipAddr,
	}
	_, err = registrar.RegisterNewPlayer(req3)
	require.NoError(t, err)

	// Fourth registration should hit rate limit
	req4 := &RegistrationRequest{
		SSHUsername: "user4",
		DisplayName: "Player4",
		Password:    "password123",
		RemoteAddr:  ipAddr,
	}
	_, err = registrar.RegisterNewPlayer(req4)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit")
}
