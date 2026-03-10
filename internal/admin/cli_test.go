package admin

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestCLI_ParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Command
	}{
		{
			name:     "status command",
			input:    "status",
			expected: &Command{Type: CommandStatus},
		},
		{
			name:     "sessions command",
			input:    "sessions",
			expected: &Command{Type: CommandSessions},
		},
		{
			name:     "shutdown command",
			input:    "shutdown",
			expected: &Command{Type: CommandShutdown},
		},
		{
			name:     "unknown command",
			input:    "invalid",
			expected: nil,
		},
		{
			name:     "empty command",
			input:    "",
			expected: nil,
		},
	}

	logger := zerolog.Nop()
	cli := NewCLI(nil, nil, logger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cli.parseCommand(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Type, result.Type)
			}
		})
	}
}

func TestCLI_ReadLoop(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedCmds   []string
		expectedOutput string
	}{
		{
			name:         "valid status command",
			input:        "status\n",
			expectedCmds: []string{CommandStatus},
		},
		{
			name:         "valid sessions command",
			input:        "sessions\n",
			expectedCmds: []string{CommandSessions},
		},
		{
			name:         "valid shutdown command",
			input:        "shutdown\n",
			expectedCmds: []string{CommandShutdown},
		},
		{
			name:           "unknown command",
			input:          "invalid\n",
			expectedCmds:   nil,
			expectedOutput: "Unknown command\n",
		},
		{
			name:         "multiple commands",
			input:        "status\nsessions\nshutdown\n",
			expectedCmds: []string{CommandStatus, CommandSessions, CommandShutdown},
		},
		{
			name:         "empty lines ignored",
			input:        "\nstatus\n\nsessions\n\n",
			expectedCmds: []string{CommandStatus, CommandSessions},
		},
		{
			name:           "mixed valid and invalid",
			input:          "status\ninvalid\nsessions\n",
			expectedCmds:   []string{CommandStatus, CommandSessions},
			expectedOutput: "Unknown command\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}
			logger := zerolog.Nop()

			cli := NewCLI(input, output, logger)
			cli.Start()

			// Give the goroutine time to process
			time.Sleep(50 * time.Millisecond)

			// Collect commands from channel
			var receivedCmds []string
			for {
				select {
				case cmd := <-cli.CommandChannel():
					receivedCmds = append(receivedCmds, cmd.Type)
				case <-time.After(10 * time.Millisecond):
					goto done
				}
			}
		done:

			assert.Equal(t, tt.expectedCmds, receivedCmds)
			if tt.expectedOutput != "" {
				assert.Contains(t, output.String(), tt.expectedOutput)
			}
		})
	}
}

func TestCLI_WriteOutput(t *testing.T) {
	output := &bytes.Buffer{}
	logger := zerolog.Nop()
	cli := NewCLI(nil, output, logger)

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "message with newline",
			message:  "test message\n",
			expected: "test message\n",
		},
		{
			name:     "message without newline",
			message:  "test message",
			expected: "test message\n",
		},
		{
			name:     "empty message",
			message:  "",
			expected: "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output.Reset()
			cli.WriteOutput(tt.message)
			assert.Equal(t, tt.expected, output.String())
		})
	}
}

func TestCLI_WriteOutputf(t *testing.T) {
	output := &bytes.Buffer{}
	logger := zerolog.Nop()
	cli := NewCLI(nil, output, logger)

	cli.WriteOutputf("Tick: %d, Sessions: %d", 100, 5)
	assert.Equal(t, "Tick: 100, Sessions: 5\n", output.String())
}
