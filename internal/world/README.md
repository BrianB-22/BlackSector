# World Package

The `world` package handles loading and managing the static universe configuration for BlackSector.

## Overview

This package provides:
- Loading of universe data from JSON configuration files
- Validation of world topology (connectivity, references)
- Thread-safe access to world data via RWMutex
- Helper methods for querying systems, ports, regions, and jump connections

## Key Components

### WorldGenerator
Implements the `Generator` interface for loading and validating world configurations.

### Universe
The main data structure containing all regions, systems, ports, and jump connections.
Provides thread-safe getter methods for concurrent access from multiple goroutines.

### Data Structures
- **Region**: Collection of star systems with shared characteristics
- **System**: Individual star system with security level and position
- **Port**: Trading station or starbase within a system
- **JumpConnection**: Navigable route between two systems

## Usage

```go
import (
    "github.com/BrianB-22/BlackSector/internal/world"
    "github.com/rs/zerolog"
)

// Create a world generator
logger := zerolog.New(os.Stdout)
generator := world.NewWorldGenerator(logger)

// Load world configuration
universe, err := generator.LoadWorld("config/world/alpha_sector.json")
if err != nil {
    log.Fatal().Err(err).Msg("failed to load world")
}

// Validate topology
if err := generator.ValidateTopology(universe); err != nil {
    log.Fatal().Err(err).Msg("invalid world topology")
}

// Access world data (thread-safe)
system := universe.GetSystem(1)
ports := universe.GetPortsBySystem(1)
connections := universe.GetJumpConnections(1)
```

## Thread Safety

All getter methods on `Universe` use `RWMutex` for concurrent read access.
The world data is loaded once at server startup and remains read-only during operation.

## Validation

The `ValidateTopology` method ensures:
- All systems reference valid regions
- All ports reference valid systems
- All jump connections reference valid systems
- All systems are reachable via jump connections
- At least one Federated Space system exists (SecurityLevel = 2.0)
- Security levels are within valid ranges (0.0-2.0)

## Configuration Format

See `config/world/alpha_sector.json` for the expected JSON structure.
