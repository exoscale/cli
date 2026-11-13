package clicmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// A list followed by one required argument, e.g. "upload FILE... DESTINATION".
type testListThenArgCLICmd struct {
	_ bool `cli-cmd:"list-then-arg"`

	Files       []string `cli-arg:"#" cli-usage:"FILE"`
	Destination string   `cli-arg:"#" cli-usage:"DESTINATION"`
}

func (c *testListThenArgCLICmd) CmdAliases() []string { return nil }
func (c *testListThenArgCLICmd) CmdShort() string     { return "" }
func (c *testListThenArgCLICmd) CmdLong() string      { return "" }
func (c *testListThenArgCLICmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}
func (c *testListThenArgCLICmd) CmdRun(_ *cobra.Command, _ []string) error { return nil }

// A list followed by two required arguments.
type testListThenTwoArgsCLICmd struct {
	_ bool `cli-cmd:"list-then-two-args"`

	Items  []string `cli-arg:"#" cli-usage:"ITEM"`
	First  string   `cli-arg:"#" cli-usage:"FIRST"`
	Second string   `cli-arg:"#" cli-usage:"SECOND"`
}

func (c *testListThenTwoArgsCLICmd) CmdAliases() []string { return nil }
func (c *testListThenTwoArgsCLICmd) CmdShort() string     { return "" }
func (c *testListThenTwoArgsCLICmd) CmdLong() string      { return "" }
func (c *testListThenTwoArgsCLICmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}
func (c *testListThenTwoArgsCLICmd) CmdRun(_ *cobra.Command, _ []string) error { return nil }

// A list between two required arguments.
type testArgListArgCLICmd struct {
	_ bool `cli-cmd:"arg-list-arg"`

	Target      string   `cli-arg:"#" cli-usage:"TARGET"`
	Keys        []string `cli-arg:"#" cli-usage:"KEY"`
	Destination string   `cli-arg:"#" cli-usage:"DESTINATION"`
}

func (c *testArgListArgCLICmd) CmdAliases() []string { return nil }
func (c *testArgListArgCLICmd) CmdShort() string     { return "" }
func (c *testArgListArgCLICmd) CmdLong() string      { return "" }
func (c *testArgListArgCLICmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}
func (c *testArgListArgCLICmd) CmdRun(_ *cobra.Command, _ []string) error { return nil }

// Invalid: an optional argument after a list.
type testListThenOptionalArgCLICmd struct {
	_ bool `cli-cmd:"list-then-optional-arg"`

	Files       []string `cli-arg:"#" cli-usage:"FILE"`
	Destination string   `cli-arg:"?" cli-usage:"DESTINATION"`
}

func (c *testListThenOptionalArgCLICmd) CmdAliases() []string { return nil }
func (c *testListThenOptionalArgCLICmd) CmdShort() string     { return "" }
func (c *testListThenOptionalArgCLICmd) CmdLong() string      { return "" }
func (c *testListThenOptionalArgCLICmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}
func (c *testListThenOptionalArgCLICmd) CmdRun(_ *cobra.Command, _ []string) error { return nil }

// Invalid: a list, a required argument, then a second list.
type testListArgListCLICmd struct {
	_ bool `cli-cmd:"list-arg-list"`

	Items []string `cli-arg:"#" cli-usage:"ITEM"`
	Arg   string   `cli-arg:"#" cli-usage:"ARG"`
	More  []string `cli-arg:"#" cli-usage:"MORE"`
}

func (c *testListArgListCLICmd) CmdAliases() []string { return nil }
func (c *testListArgListCLICmd) CmdShort() string     { return "" }
func (c *testListArgListCLICmd) CmdLong() string      { return "" }
func (c *testListArgListCLICmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}
func (c *testListArgListCLICmd) CmdRun(_ *cobra.Command, _ []string) error { return nil }

func Test_cliCommandUse_listFollowedByArgs(t *testing.T) {
	tests := []struct {
		name     string
		cmd      cliCommand
		expected string
		wantErr  string
	}{
		{
			name:     "list then one argument",
			cmd:      &testListThenArgCLICmd{},
			expected: "list-then-arg FILE... DESTINATION",
		},
		{
			name:     "list then two arguments",
			cmd:      &testListThenTwoArgsCLICmd{},
			expected: "list-then-two-args ITEM... FIRST SECOND",
		},
		{
			name:     "list between two arguments",
			cmd:      &testArgListArgCLICmd{},
			expected: "arg-list-arg TARGET KEY... DESTINATION",
		},
		{
			name:    "optional argument after a list is rejected",
			cmd:     &testListThenOptionalArgCLICmd{},
			wantErr: "field clicmd.testListThenOptionalArgCLICmd.Destination: only required single arguments can follow a list argument",
		},
		{
			name:    "second list after a list is rejected",
			cmd:     &testListArgListCLICmd{},
			wantErr: "field clicmd.testListArgListCLICmd.More: only required single arguments can follow a list argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := cliCommandUse(tt.cmd)

			// RegisterCLICommand fails the same way, so that the CLI doesn't
			// start with an ambiguous command.
			regErr := RegisterCLICommand(&cobra.Command{}, tt.cmd)

			if tt.wantErr == "" {
				assert.NoError(t, err)
				assert.NoError(t, regErr)
			} else {
				wantErr := cliCommandImplemError{tt.wantErr}.Error()
				assert.EqualError(t, err, wantErr)
				assert.EqualError(t, regErr, "error initializing CLI command: "+wantErr)
			}
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_registerCobraCommand_listFollowedByArgs(t *testing.T) {
	tests := []struct {
		name     string
		cmd      cliCommand
		args     []string
		expected cliCommand
		wantErr  string
	}{
		{
			name:     "list then argument: one item",
			cmd:      &testListThenArgCLICmd{},
			args:     []string{"list-then-arg", "a", "dst"},
			expected: &testListThenArgCLICmd{Files: []string{"a"}, Destination: "dst"},
		},
		{
			name:     "list then argument: several items",
			cmd:      &testListThenArgCLICmd{},
			args:     []string{"list-then-arg", "a", "b", "c", "dst"},
			expected: &testListThenArgCLICmd{Files: []string{"a", "b", "c"}, Destination: "dst"},
		},
		{
			name:     "list then argument: empty list, the only argument goes to the last field",
			cmd:      &testListThenArgCLICmd{},
			args:     []string{"list-then-arg", "dst"},
			expected: &testListThenArgCLICmd{Destination: "dst"},
		},
		{
			name:     "list then argument: no arguments",
			cmd:      &testListThenArgCLICmd{},
			args:     []string{"list-then-arg"},
			expected: &testListThenArgCLICmd{},
			wantErr:  "missing arguments, run with --help for usage",
		},
		{
			name:     "list then two arguments: several items",
			cmd:      &testListThenTwoArgsCLICmd{},
			args:     []string{"list-then-two-args", "x", "y", "a", "b"},
			expected: &testListThenTwoArgsCLICmd{Items: []string{"x", "y"}, First: "a", Second: "b"},
		},
		{
			name:     "list then two arguments: empty list",
			cmd:      &testListThenTwoArgsCLICmd{},
			args:     []string{"list-then-two-args", "a", "b"},
			expected: &testListThenTwoArgsCLICmd{First: "a", Second: "b"},
		},
		{
			name:     "list then two arguments: one argument missing",
			cmd:      &testListThenTwoArgsCLICmd{},
			args:     []string{"list-then-two-args", "a"},
			expected: &testListThenTwoArgsCLICmd{First: "a"},
			wantErr:  "missing arguments, run with --help for usage",
		},
		{
			name:     "list between two arguments: several items",
			cmd:      &testArgListArgCLICmd{},
			args:     []string{"arg-list-arg", "t", "k1", "k2", "dst"},
			expected: &testArgListArgCLICmd{Target: "t", Keys: []string{"k1", "k2"}, Destination: "dst"},
		},
		{
			name:     "list between two arguments: empty list",
			cmd:      &testArgListArgCLICmd{},
			args:     []string{"arg-list-arg", "t", "dst"},
			expected: &testArgListArgCLICmd{Target: "t", Destination: "dst"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := &cobra.Command{SilenceErrors: true, SilenceUsage: true}
			assert.NoError(t, RegisterCLICommand(rootCmd, tt.cmd))

			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()

			assert.Equal(t, tt.wantErr, errString(err))
			assert.Equal(t, tt.expected, tt.cmd)
		})
	}
}
