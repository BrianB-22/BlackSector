package admin

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/rs/zerolog"
)

// Command represents an admin command
type Command struct {
	Type string
}

// CommandType constants
const (
	CommandStatus   = "status"
	CommandSessions = "sessions"
	CommandShutdown = "shutdown"
)

// CLI handles admin commands from stdin
type CLI struct {
	input      io.Reader
	output     io.Writer
	commandCh  chan Command
	shutdownCh chan struct{}
	logger     zerolog.Logger
}

// NewCLI creates a new admin CLI
func NewCLI(input io.Reader, output io.Writer, logger zerolog.Logger) *CLI {
	return &CLI{
		input:      input,
		output:     output,
		commandCh:  make(chan Command, 10),
		shutdownCh: make(chan struct{}),
		logger:     logger,
	}
}

// Start begins reading commands from stdin in a separate goroutine
// Requirements: 13.1, 13.6
func (c *CLI) Start() {
	go c.readLoop()
	c.logger.Info().Msg("Admin CLI started")
}

// CommandChannel returns the channel for receiving commands
func (c *CLI) CommandChannel() <-chan Command {
	return c.commandCh
}

// ShutdownChannel returns the channel that signals shutdown
func (c *CLI) ShutdownChannel() <-chan struct{} {
	return c.shutdownCh
}

// readLoop reads commands from stdin and sends them to the command channel
// Runs in a separate goroutine to avoid blocking the tick loop
func (c *CLI) readLoop() {
	scanner := bufio.NewScanner(c.input)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if line == "" {
			continue
		}
		
		// Parse command
		cmd := c.parseCommand(line)
		
		if cmd == nil {
			// Unknown command - display error and continue
			// Requirement 13.5
			c.output.Write([]byte("Unknown command\n"))
			c.logger.Debug().
				Str("input", line).
				Msg("Unknown admin command")
			continue
		}
		
		// Send command to channel
		select {
		case c.commandCh <- *cmd:
			c.logger.Debug().
				Str("command", cmd.Type).
				Msg("Admin command received")
		default:
			c.logger.Warn().
				Str("command", cmd.Type).
				Msg("Admin command channel full, command dropped")
		}
	}
	
	if err := scanner.Err(); err != nil {
		c.logger.Error().
			Err(err).
			Msg("Error reading from stdin")
	}
}

// parseCommand parses a command line and returns a Command or nil if unknown
// Requirements: 13.2, 13.3, 13.4
func (c *CLI) parseCommand(line string) *Command {
	switch line {
	case CommandStatus:
		return &Command{Type: CommandStatus}
	case CommandSessions:
		return &Command{Type: CommandSessions}
	case CommandShutdown:
		return &Command{Type: CommandShutdown}
	default:
		return nil
	}
}

// WriteOutput writes a message to the CLI output
func (c *CLI) WriteOutput(message string) {
	c.output.Write([]byte(message))
	if !strings.HasSuffix(message, "\n") {
		c.output.Write([]byte("\n"))
	}
}

// WriteOutputf writes a formatted message to the CLI output
func (c *CLI) WriteOutputf(format string, args ...interface{}) {
	c.WriteOutput(fmt.Sprintf(format, args...))
}
