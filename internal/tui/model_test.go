package tui

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGameView(t *testing.T) {
	// Create channels for testing
	updateChan := make(chan interface{}, 10)
	commandOut := make(chan interface{}, 10)
	logger := zerolog.Nop()

	// Create GameView
	view := NewGameView("session-123", "player-456", updateChan, commandOut, logger)

	// Verify initialization
	require.NotNil(t, view)
	assert.Equal(t, "session-123", view.sessionID)
	assert.Equal(t, "player-456", view.playerID)
	assert.Equal(t, ViewModeCommand, view.viewMode)
	assert.Equal(t, "", view.commandBuffer)
	assert.NotNil(t, view.currentState)
	assert.Equal(t, 80, view.width)
	assert.Equal(t, 24, view.height)
}

func TestGameView_Init(t *testing.T) {
	updateChan := make(chan interface{}, 10)
	commandOut := make(chan interface{}, 10)
	logger := zerolog.Nop()

	view := NewGameView("session-123", "player-456", updateChan, commandOut, logger)

	// Init should return a command
	cmd := view.Init()
	assert.NotNil(t, cmd)
}

func TestViewModeConstants(t *testing.T) {
	// Verify view mode constants are defined correctly
	assert.Equal(t, ViewMode("COMMAND"), ViewModeCommand)
	assert.Equal(t, ViewMode("MARKET"), ViewModeMarket)
	assert.Equal(t, ViewMode("COMBAT"), ViewModeCombat)
	assert.Equal(t, ViewMode("MISSION"), ViewModeMission)
	assert.Equal(t, ViewMode("CARGO"), ViewModeCargo)
	assert.Equal(t, ViewMode("HELP"), ViewModeHelp)
}
