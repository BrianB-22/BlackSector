package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette for BlackSector TUI
var (
	// Primary colors
	colorPrimary   = lipgloss.Color("#00D9FF") // Cyan
	colorSecondary = lipgloss.Color("#FF6B9D") // Pink
	colorAccent    = lipgloss.Color("#FFD700") // Gold

	// Status colors
	colorSuccess = lipgloss.Color("#00FF00") // Green
	colorWarning = lipgloss.Color("#FFA500") // Orange
	colorDanger  = lipgloss.Color("#FF0000") // Red
	colorInfo    = lipgloss.Color("#87CEEB") // Sky blue

	// UI colors
	colorBackground = lipgloss.Color("#1A1A1A") // Dark gray
	colorForeground = lipgloss.Color("#E0E0E0") // Light gray
	colorBorder     = lipgloss.Color("#404040") // Medium gray
	colorMuted      = lipgloss.Color("#808080") // Gray
)

// Status bar style - displays system, credits, ship stats
var statusBarStyle = lipgloss.NewStyle().
	Foreground(colorForeground).
	Background(colorPrimary).
	Bold(true).
	Padding(0, 1)

// Command prompt style
var promptStyle = lipgloss.NewStyle().
	Foreground(colorPrimary).
	Bold(true)

// Error message style
var errorStyle = lipgloss.NewStyle().
	Foreground(colorDanger).
	Bold(true)

// Info message style
var infoStyle = lipgloss.NewStyle().
	Foreground(colorInfo)

// Success message style
var successStyle = lipgloss.NewStyle().
	Foreground(colorSuccess).
	Bold(true)

// Warning message style
var warningStyle = lipgloss.NewStyle().
	Foreground(colorWarning).
	Bold(true)

// Title style for view headers
var titleStyle = lipgloss.NewStyle().
	Foreground(colorAccent).
	Bold(true).
	Underline(true)

// Subtitle style for section headers
var subtitleStyle = lipgloss.NewStyle().
	Foreground(colorSecondary).
	Bold(true)

// Table header style
var tableHeaderStyle = lipgloss.NewStyle().
	Foreground(colorPrimary).
	Bold(true).
	Underline(true)

// Table row style (normal)
var tableRowStyle = lipgloss.NewStyle().
	Foreground(colorForeground)

// Table row style (alternate for zebra striping)
var tableRowAltStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

// Border style for boxes and panels
var borderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorBorder).
	Padding(1, 2)

// Highlight style for selected items
var highlightStyle = lipgloss.NewStyle().
	Foreground(colorAccent).
	Bold(true)

// Muted text style for less important information
var mutedStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

// Combat styles
var combatPlayerStyle = lipgloss.NewStyle().
	Foreground(colorPrimary).
	Bold(true)

var combatEnemyStyle = lipgloss.NewStyle().
	Foreground(colorDanger).
	Bold(true)

// Ship status styles based on health percentage
func getHealthStyle(current, max int) lipgloss.Style {
	if max == 0 {
		return mutedStyle
	}
	
	percentage := float64(current) / float64(max)
	
	switch {
	case percentage > 0.7:
		return lipgloss.NewStyle().Foreground(colorSuccess)
	case percentage > 0.3:
		return lipgloss.NewStyle().Foreground(colorWarning)
	default:
		return lipgloss.NewStyle().Foreground(colorDanger)
	}
}
