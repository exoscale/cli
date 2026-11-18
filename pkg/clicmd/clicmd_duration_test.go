package clicmd

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type testDurationCLICmd struct {
	_ bool `cli-cmd:"test-duration"`

	Arg time.Duration `cli-arg:"?" cli-usage:"DURATION"`

	Expires time.Duration `cli-short:"e" cli-usage:"expiration"`
}

func (c *testDurationCLICmd) CmdAliases() []string { return nil }
func (c *testDurationCLICmd) CmdShort() string     { return "" }
func (c *testDurationCLICmd) CmdLong() string      { return "" }
func (c *testDurationCLICmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}
func (c *testDurationCLICmd) CmdRun(_ *cobra.Command, _ []string) error { return nil }

// errString returns the message of err, or "" if err is nil.
func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func Test_cliCommandFlagSet_duration(t *testing.T) {
	fs, err := cliCommandFlagSet(&testDurationCLICmd{Expires: 15 * time.Minute})
	assert.NoError(t, err)

	flag := fs.Lookup("expires")
	if assert.NotNil(t, flag) {
		// Must not be registered as int64, time.Duration's reflect.Kind.
		assert.Equal(t, "duration", flag.Value.Type())
		assert.Equal(t, "e", flag.Shorthand)
		assert.Equal(t, "15m0s", flag.DefValue)
	}
}

func Test_registerCobraCommand_duration(t *testing.T) {
	tests := []struct {
		name            string
		defaultExpires  time.Duration
		args            []string
		expectedExpires time.Duration
		expectedArg     time.Duration
		wantErr         string
	}{
		{
			name:            "default value",
			defaultExpires:  900 * time.Second,
			args:            []string{"test-duration"},
			expectedExpires: 900 * time.Second,
		},
		{
			name:            "long flag",
			args:            []string{"test-duration", "--expires", "15m"},
			expectedExpires: 15 * time.Minute,
		},
		{
			name:            "short flag, compound duration",
			args:            []string{"test-duration", "-e", "1h30m"},
			expectedExpires: 90 * time.Minute,
		},
		{
			name:        "positional arg",
			args:        []string{"test-duration", "2h"},
			expectedArg: 2 * time.Hour,
		},
		{
			name:    "invalid flag value",
			args:    []string{"test-duration", "--expires", "soon"},
			wantErr: `invalid argument "soon" for "-e, --expires" flag: time: invalid duration "soon"`,
		},
		{
			name:    "plain integer flag value is rejected like cobra's duration flags",
			args:    []string{"test-duration", "--expires", "900"},
			wantErr: `invalid argument "900" for "-e, --expires" flag: time: missing unit in duration "900"`,
		},
		{
			name:    "invalid positional arg",
			args:    []string{"test-duration", "soon"},
			wantErr: `invalid value for DURATION: time: invalid duration "soon"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := &cobra.Command{SilenceErrors: true, SilenceUsage: true}
			testCmd := &testDurationCLICmd{Expires: tt.defaultExpires}
			assert.NoError(t, RegisterCLICommand(rootCmd, testCmd))

			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()

			assert.Equal(t, tt.wantErr, errString(err))
			assert.Equal(t, tt.expectedExpires, testCmd.Expires)
			assert.Equal(t, tt.expectedArg, testCmd.Arg)
		})
	}
}
