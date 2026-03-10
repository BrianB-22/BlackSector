package tui

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/BrianB-22/BlackSector/internal/engine"
)

// ParsedCommand represents a parsed and validated command
type ParsedCommand struct {
	Type    string // "jump", "dock", "undock", "buy", "sell", "system", "market", "cargo", "help"
	Payload []byte // JSON payload for engine commands
	IsLocal bool   // true for local commands (help, market, cargo, system)
}

// ParseCommand parses a raw command string into a ParsedCommand
// Returns error if command is invalid or has incorrect parameters
func ParseCommand(input string) (*ParsedCommand, error) {
	// Trim and normalize input
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty command")
	}

	// Split into command and arguments
	parts := strings.Fields(input)
	cmdName := strings.ToLower(parts[0])
	args := parts[1:]

	// Parse based on command type
	switch cmdName {
	case "jump":
		return parseJumpCommand(args)
	case "dock":
		return parseDockCommand(args)
	case "undock":
		return parseUndockCommand(args)
	case "buy":
		return parseBuyCommand(args)
	case "sell":
		return parseSellCommand(args)
	case "attack":
		return parseAttackCommand(args)
	case "flee":
		return parseFleeCommand(args)
	case "surrender":
		return parseSurrenderCommand(args)
	case "mission_accept", "missions_accept":
		return parseMissionAcceptCommand(args)
	case "mission_abandon", "missions_abandon":
		return parseMissionAbandonCommand(args)
	case "mission_list", "missions", "missions_list":
		return parseMissionsCommand(args)
	case "system":
		return parseSystemCommand(args)
	case "market":
		return parseMarketCommand(args)
	case "cargo":
		return parseCargoCommand(args)
	case "help":
		return parseHelpCommand(args)
	default:
		return nil, fmt.Errorf("unknown command: %s", cmdName)
	}
}

// parseJumpCommand parses "jump <system_id>"
func parseJumpCommand(args []string) (*ParsedCommand, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("jump requires exactly one argument: jump <system_id>")
	}

	systemID, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid system_id: must be an integer")
	}

	if systemID <= 0 {
		return nil, fmt.Errorf("invalid system_id: must be positive")
	}

	// Create payload
	payload := engine.JumpPayload{
		TargetSystemID: systemID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal jump payload: %w", err)
	}

	return &ParsedCommand{
		Type:    "jump",
		Payload: payloadBytes,
		IsLocal: false,
	}, nil
}

// parseDockCommand parses "dock <port_id>"
func parseDockCommand(args []string) (*ParsedCommand, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("dock requires exactly one argument: dock <port_id>")
	}

	portID, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid port_id: must be an integer")
	}

	if portID <= 0 {
		return nil, fmt.Errorf("invalid port_id: must be positive")
	}

	// Create payload
	payload := engine.DockPayload{
		PortID: portID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal dock payload: %w", err)
	}

	return &ParsedCommand{
		Type:    "dock",
		Payload: payloadBytes,
		IsLocal: false,
	}, nil
}

// parseUndockCommand parses "undock" (no parameters)
func parseUndockCommand(args []string) (*ParsedCommand, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("undock takes no arguments")
	}

	// Create empty payload
	payload := engine.UndockPayload{}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal undock payload: %w", err)
	}

	return &ParsedCommand{
		Type:    "undock",
		Payload: payloadBytes,
		IsLocal: false,
	}, nil
}

// parseBuyCommand parses "buy <commodity> <quantity>"
func parseBuyCommand(args []string) (*ParsedCommand, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buy requires two arguments: buy <commodity> <quantity>")
	}

	commodityID := strings.ToLower(args[0])
	quantity, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: must be an integer")
	}

	if quantity <= 0 {
		return nil, fmt.Errorf("invalid quantity: must be positive")
	}

	// Validate commodity name (must match economy.CommodityType constants)
	if !isValidCommodity(commodityID) {
		return nil, fmt.Errorf("invalid commodity: %s (valid: food_supplies, fuel_cells, raw_ore, refined_ore, machinery, electronics, luxury_goods)", commodityID)
	}

	// Create payload - note: port_id will be filled in by the session layer
	// since the parser doesn't have access to current docked port
	payload := engine.BuyPayload{
		PortID:      0, // Will be set by session layer
		CommodityID: commodityID,
		Quantity:    quantity,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal buy payload: %w", err)
	}

	return &ParsedCommand{
		Type:    "buy",
		Payload: payloadBytes,
		IsLocal: false,
	}, nil
}

// parseSellCommand parses "sell <commodity> <quantity>"
func parseSellCommand(args []string) (*ParsedCommand, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("sell requires two arguments: sell <commodity> <quantity>")
	}

	commodityID := strings.ToLower(args[0])
	quantity, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: must be an integer")
	}

	if quantity <= 0 {
		return nil, fmt.Errorf("invalid quantity: must be positive")
	}

	// Validate commodity name
	if !isValidCommodity(commodityID) {
		return nil, fmt.Errorf("invalid commodity: %s (valid: food_supplies, fuel_cells, raw_ore, refined_ore, machinery, electronics, luxury_goods)", commodityID)
	}

	// Create payload - port_id will be filled in by session layer
	payload := engine.SellPayload{
		PortID:      0, // Will be set by session layer
		CommodityID: commodityID,
		Quantity:    quantity,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal sell payload: %w", err)
	}

	return &ParsedCommand{
		Type:    "sell",
		Payload: payloadBytes,
		IsLocal: false,
	}, nil
}

// parseSystemCommand parses "system" (no parameters)
func parseSystemCommand(args []string) (*ParsedCommand, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("system takes no arguments")
	}

	return &ParsedCommand{
		Type:    "system",
		Payload: nil,
		IsLocal: true,
	}, nil
}

// parseMarketCommand parses "market" (no parameters)
func parseMarketCommand(args []string) (*ParsedCommand, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("market takes no arguments")
	}

	return &ParsedCommand{
		Type:    "market",
		Payload: nil,
		IsLocal: true,
	}, nil
}

// parseCargoCommand parses "cargo" (no parameters)
func parseCargoCommand(args []string) (*ParsedCommand, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("cargo takes no arguments")
	}

	return &ParsedCommand{
		Type:    "cargo",
		Payload: nil,
		IsLocal: true,
	}, nil
}

// parseHelpCommand parses "help" (no parameters)
func parseHelpCommand(args []string) (*ParsedCommand, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("help takes no arguments")
	}

	return &ParsedCommand{
		Type:    "help",
		Payload: nil,
		IsLocal: true,
	}, nil
}

// isValidCommodity checks if a commodity ID is valid
func isValidCommodity(commodityID string) bool {
	validCommodities := map[string]bool{
		"food_supplies": true,
		"fuel_cells":    true,
		"raw_ore":       true,
		"refined_ore":   true,
		"machinery":     true,
		"electronics":   true,
		"luxury_goods":  true,
	}
	return validCommodities[commodityID]
}

// parseAttackCommand parses "attack" (no parameters)
func parseAttackCommand(args []string) (*ParsedCommand, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("attack takes no arguments")
	}

	// Create empty payload - combat_id will be filled in by session layer
	payload := engine.AttackPayload{}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal attack payload: %w", err)
	}

	return &ParsedCommand{
		Type:    "attack",
		Payload: payloadBytes,
		IsLocal: false,
	}, nil
}

// parseFleeCommand parses "flee" (no parameters)
func parseFleeCommand(args []string) (*ParsedCommand, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("flee takes no arguments")
	}

	// Create empty payload - combat_id will be filled in by session layer
	payload := engine.FleePayload{}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal flee payload: %w", err)
	}

	return &ParsedCommand{
		Type:    "flee",
		Payload: payloadBytes,
		IsLocal: false,
	}, nil
}

// parseSurrenderCommand parses "surrender" (no parameters)
func parseSurrenderCommand(args []string) (*ParsedCommand, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("surrender takes no arguments")
	}

	// Create empty payload - combat_id will be filled in by session layer
	payload := engine.SurrenderPayload{}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal surrender payload: %w", err)
	}

	return &ParsedCommand{
		Type:    "surrender",
		Payload: payloadBytes,
		IsLocal: false,
	}, nil
}

// parseMissionAcceptCommand parses "mission_accept <mission_id>"
func parseMissionAcceptCommand(args []string) (*ParsedCommand, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mission_accept requires exactly one argument: mission_accept <mission_id>")
	}

	missionID := args[0]

	// Create payload
	payload := engine.MissionAcceptPayload{
		MissionID: missionID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal mission_accept payload: %w", err)
	}

	return &ParsedCommand{
		Type:    "mission_accept",
		Payload: payloadBytes,
		IsLocal: false,
	}, nil
}

// parseMissionAbandonCommand parses "mission_abandon" (no parameters)
func parseMissionAbandonCommand(args []string) (*ParsedCommand, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("mission_abandon takes no arguments")
	}

	// Create empty payload
	payload := engine.MissionAbandonPayload{}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal mission_abandon payload: %w", err)
	}

	return &ParsedCommand{
		Type:    "mission_abandon",
		Payload: payloadBytes,
		IsLocal: false,
	}, nil
}

// parseMissionsCommand parses "missions" (no parameters) - local command to view mission board
func parseMissionsCommand(args []string) (*ParsedCommand, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("missions takes no arguments")
	}

	return &ParsedCommand{
		Type:    "missions",
		Payload: nil,
		IsLocal: true,
	}, nil
}
